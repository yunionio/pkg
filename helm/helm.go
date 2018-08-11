package helm

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig"
	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/downloader"
	"k8s.io/helm/pkg/getter"
	"k8s.io/helm/pkg/helm"
	"k8s.io/helm/pkg/helm/environment"
	"k8s.io/helm/pkg/helm/helmpath"
	"k8s.io/helm/pkg/proto/hapi/chart"
	"k8s.io/helm/pkg/proto/hapi/release"
	rls "k8s.io/helm/pkg/proto/hapi/services"
	"k8s.io/helm/pkg/repo"
	"k8s.io/helm/pkg/strvals"

	"yunion.io/x/jsonutils"
	"yunion.io/x/log"
	yerrors "yunion.io/x/pkg/util/errors"
	"yunion.io/x/pkg/util/sets"
	"yunion.io/x/pkg/util/workqueue"
)

const (
	helmPostFix               string = "helm"
	YunionInfraChart          string = "yunion-infra"
	YunionChartRepoAnnotation string = "helm.yunion.io/chart-repo"
	YunionChartNameAnnotation string = "helm.yunion.io/chart-name"
	YunionRepoURLAnnotation   string = "helm.yunion.io/repo-url"
)

var (
	stateStorePath       string
	ErrRepoNotFound      = errors.New("helm repository not found")
	RunableCheckKeywords = sets.NewString(YunionInfraChart)
)

func InitStateStoreDir(dirPath string) error {
	if len(dirPath) == 0 {
		return fmt.Errorf("Helm state store path must specified")
	}
	if _, err := os.Stat(dirPath); err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll(dirPath, 0755)
			if err != nil {
				return fmt.Errorf("Make directory %s: %v", dirPath, err)
			}
		} else {
			return fmt.Errorf("Get direcotry %s stat: %v", dirPath, err)
		}
	}
	stateStorePath = dirPath
	return nil
}

func EnsureStateStoreDir(dirPath string) {
	err := InitStateStoreDir(dirPath)
	if err != nil {
		log.Fatalf("Init stateStorePath %s: %v", dirPath, err)
	}
}

func generateHelmRepoPath(path string) string {
	return fmt.Sprintf("%s/%s/%s", stateStorePath, path, helmPostFix)
}

func downloadChartFromRepo(name, version, path string) (string, error) {
	settings := createEnvSettings(path)
	dl := downloader.ChartDownloader{
		HelmHome: settings.Home,
		Getters:  getter.All(settings),
	}
	if _, err := os.Stat(settings.Home.Archive()); os.IsNotExist(err) {
		log.Infof("Creating '%s' directory.", settings.Home.Archive())
		os.MkdirAll(settings.Home.Archive(), 0744)
	}

	log.Infof("Downloading helm chart %q to %q, version: %q", name, settings.Home.Archive(), version)
	filename, _, err := dl.DownloadTo(name, version, settings.Home.Archive())
	if err == nil {
		lname, err := filepath.Abs(filename)
		if err != nil {
			return filename, errors.Wrapf(err, "Could not create absolute path from %s", filename)
		}
		log.Debugf("Fetched helm chart '%s' to '%s'", name, filename)
		return lname, nil
	}

	return filename, errors.Wrapf(err, "Failed to download %q", name)
}

func ReposList(regionName string) ([]*repo.Entry, error) {
	repoPath := filepath.Join(generateHelmRepoPath(regionName), "repository", "repositories.yaml")
	log.Debugf("Helm repo path: %s", repoPath)

	f, err := repo.LoadRepositoriesFile(repoPath)
	if err != nil {
		return nil, err
	}
	if len(f.Repositories) == 0 {
		return make([]*repo.Entry, 0), nil
	}

	return f.Repositories, nil
}

func RepoShow(name, regionName string) (*repo.Entry, error) {
	repoPath := filepath.Join(generateHelmRepoPath(regionName), "repository", "repositories.yaml")
	log.Debugf("Helm repo path: %s", repoPath)
	f, err := repo.LoadRepositoriesFile(repoPath)
	if err != nil {
		return nil, err
	}
	return (&RepoFile{f}).Get(name)
}

func getChartOption(file []byte) (*SpotGuideFile, error) {
	so := &SpotGuideFile{}
	tarReader := tar.NewReader(bytes.NewReader(file))
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if strings.Contains(header.Name, "spotguide.json") {
			valuesContent := new(bytes.Buffer)
			if _, err := io.Copy(valuesContent, tarReader); err != nil {
				return nil, err
			}
			err := json.Unmarshal(valuesContent.Bytes(), so)
			if err != nil {
				return nil, err
			}
			return so, nil
		}
	}
	return so, nil
}

func downloadFile(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	tarContent := new(bytes.Buffer)
	io.Copy(tarContent, resp.Body)
	gzf, err := gzip.NewReader(tarContent)
	if err != nil {
		return nil, err
	}
	rawContent, _ := ioutil.ReadAll(gzf)
	return rawContent, nil
}

//getChartFile Download file from chart repository
func getChartFile(file []byte, fileName string) (string, error) {
	tarReader := tar.NewReader(bytes.NewReader(file))
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}
		if strings.Contains(header.Name, fileName) {
			valuesContent := new(bytes.Buffer)
			if _, err := io.Copy(valuesContent, tarReader); err != nil {
				return "", err
			}
			base64Str := base64.StdEncoding.EncodeToString(valuesContent.Bytes())
			return base64Str, nil
		}
	}
	return "", nil
}

func createEnvSettings(helmRepoHome string) environment.EnvSettings {
	var settings environment.EnvSettings
	settings.Home = helmpath.Home(helmRepoHome)
	return settings
}

func RepoAdd(regionName string, hRepo *repo.Entry) error {
	settings := createEnvSettings(generateHelmRepoPath(regionName))
	repoFile := settings.Home.RepositoryFile()
	var f *repo.RepoFile
	if _, err := os.Stat(repoFile); err != nil {
		log.Infof("Creating %s", repoFile)
		f = repo.NewRepoFile()
	} else {
		f, err = repo.LoadRepositoriesFile(repoFile)
		if err != nil {
			return errors.Wrap(err, "Can't create a new ChartRepo")
		}
		log.Debugf("Profile %q loaded", repoFile)
	}

	for _, n := range f.Repositories {
		log.Debugf("repo: %s", n.Name)
		if n.Name == hRepo.Name {
			return errors.New("Already added")
		}
	}

	c := repo.Entry{
		Name:  hRepo.Name,
		URL:   hRepo.URL,
		Cache: settings.Home.CacheIndex(hRepo.Name),
	}
	r, err := repo.NewChartRepository(&c, getter.All(settings))
	if err != nil {
		return errors.Wrap(err, "Can't create a new ChartRepo")
	}
	log.Debugf("New repo added: %s", hRepo.Name)

	errIdx := r.DownloadIndexFile("")
	if errIdx != nil {
		return errors.Wrap(errIdx, "Repo index download failed")
	}
	f.Add(&c)
	if errW := f.WriteFile(repoFile, 0644); errW != nil {
		return errors.Wrap(errW, "Can't write helm repo profile")
	}
	return nil
}

func RepoDelete(regionName, repoName string) error {
	repoPath := generateHelmRepoPath(regionName)
	settings := createEnvSettings(repoPath)
	repoFile := settings.Home.RepositoryFile()

	r, err := repo.LoadRepositoriesFile(repoFile)
	if err != nil {
		return err
	}

	if !r.Remove(repoName) {
		return ErrRepoNotFound
	}
	if err := r.WriteFile(repoFile, 0644); err != nil {
		return err
	}

	if _, err := os.Stat(settings.Home.CacheIndex(repoName)); err == nil {
		err = os.Remove(settings.Home.CacheIndex(repoName))
		if err != nil {
			return err
		}
	}
	return nil
}

func RepoModify(regionName, repoName string, newRepo *repo.Entry) error {
	repoPath := generateHelmRepoPath(regionName)
	settings := createEnvSettings(repoPath)
	repoFile := settings.Home.RepositoryFile()
	log.Debugf("New repo content: %#v", newRepo)

	f, err := repo.LoadRepositoriesFile(repoFile)
	if err != nil {
		return err
	}

	if !f.Has(repoName) {
		return ErrRepoNotFound
	}

	f.Update(newRepo)

	if errW := f.WriteFile(repoFile, 0644); errW != nil {
		return errors.Wrap(errW, "Can't write helm repo profile")
	}
	return nil
}

func ReposUpdate(regionName string, repos []string) error {
	errsChannel := make(chan error, len(repos))
	uf := func(i int) {
		err := RepoUpdate(regionName, repos[i])
		if err != nil {
			errsChannel <- err
		}
	}
	workqueue.Parallelize(4, len(repos), uf)
	if len(errsChannel) > 0 {
		errs := make([]error, 0)
		length := len(errsChannel)
		for ; length > 0; length-- {
			errs = append(errs, <-errsChannel)
		}
		return yerrors.NewAggregate(errs)
	}
	return nil
}

func RepoUpdate(regionName, repoName string) error {
	repoPath := generateHelmRepoPath(regionName)
	settings := createEnvSettings(repoPath)
	repoFile := settings.Home.RepositoryFile()

	f, err := repo.LoadRepositoriesFile(repoFile)
	if err != nil {
		return errors.Wrap(err, "Load ChartRepo")
	}

	for _, cfg := range f.Repositories {
		if cfg.Name == repoName {
			log.Debugf("Updating %q chart repo url: %s", cfg.Name, cfg.URL)
			c, err := repo.NewChartRepository(cfg, getter.All(settings))
			if err != nil {
				return errors.Wrap(err, "Cannot get ChartRepo")
			}
			errIdx := c.DownloadIndexFile("")
			if errIdx != nil {
				return fmt.Errorf("Unable to get an update from the %q chart repo (%s): \n\t%v\n", cfg.Name, cfg.URL, errIdx)
			}
			log.Debugf("Successfully update %q chart repo", cfg.Name)
			return nil
		}
	}

	return ErrRepoNotFound
}

type ReleaseListQuery struct {
	Filter     string `json:"filter"`
	Namespace  string `json:"namespace"`
	All        bool   `json:"all"`
	Admin      bool   `json:"admin"`
	Deployed   bool   `json:"deployed"`
	Deleted    bool   `json:"deleted"`
	Deleting   bool   `json:"deleting"`
	Failed     bool   `json:"failed"`
	Superseded bool   `json:"superseded"`
	Pending    bool   `json:"pending"`
}

func (q ReleaseListQuery) statusCodes() []release.Status_Code {
	if q.All {
		return []release.Status_Code{
			release.Status_UNKNOWN,
			release.Status_DEPLOYED,
			release.Status_DELETED,
			release.Status_DELETING,
			release.Status_FAILED,
			release.Status_PENDING_INSTALL,
			release.Status_PENDING_UPGRADE,
			release.Status_PENDING_ROLLBACK,
		}
	}

	status := []release.Status_Code{}
	if q.Deployed {
		status = append(status, release.Status_DEPLOYED)
	}

	if q.Deleted {
		status = append(status, release.Status_DELETED)
	}

	if q.Deleting {
		status = append(status, release.Status_DELETING)
	}

	if q.Failed {
		status = append(status, release.Status_FAILED)
	}

	if q.Superseded {
		status = append(status, release.Status_SUPERSEDED)
	}

	if q.Pending {
		status = append(status, release.Status_PENDING_INSTALL, release.Status_PENDING_UPGRADE, release.Status_PENDING_UPGRADE)
	}

	if len(status) == 0 {
		// Default case
		status = append(status, release.Status_DEPLOYED, release.Status_FAILED, release.Status_PENDING_INSTALL)
	}

	return status
}

func ReleasesList(q ReleaseListQuery, regionName string) (*rls.ListReleasesResponse, error) {
	hClient, err := GetHelmClient(regionName)
	var sortBy = int32(2)
	var sortOrd = int32(1)
	ops := []helm.ReleaseListOption{
		helm.ReleaseListSort(sortBy),
		helm.ReleaseListOrder(sortOrd),
		//helm.ReleaseListLimit(q.Limit),
		//helm.ReleaseListOffset(fmt.Sprintf("%d", q.Offset)),
		helm.ReleaseListStatuses(q.statusCodes()),
	}
	if len(q.Filter) != 0 {
		log.Debugf("Apply filters: %v", q.Filter)
		ops = append(ops, helm.ReleaseListFilter(q.Filter))
	}
	if len(q.Namespace) != 0 {
		ops = append(ops, helm.ReleaseListNamespace(q.Namespace))
	}
	resp, err := hClient.ListReleases(ops...)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func ReleaseShow(releaseName string, regionName string) (*rls.GetReleaseContentResponse, error) {
	helmClient, err := GetHelmClient(regionName)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't get the helm client")
	}
	return helmClient.ReleaseContent(releaseName)
}

// ReleaseStatus retrieves the status of the passed in release name.
// returns with an error if the release is not found or another error occurs
// in case of error the status is filled with information to classify the error cause
func ReleaseStatus(releaseName string, regionName string) (*rls.GetReleaseStatusResponse, error) {
	helmClient, err := GetHelmClient(regionName)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't get the helm client")
	}
	return helmClient.ReleaseStatus(releaseName)
	//if err != nil {
	//// the release cannot be found
	//return http.StatusNotFound, errors.Wrap(err, "couldn't get the release status")
	//}
	//return int32(releaseStatusResponse.Info.Status.GetCode()), nil
}

func generateName(nameTemplate string) (string, error) {
	t, err := template.New("name-template").Funcs(sprig.TxtFuncMap()).Parse(nameTemplate)
	if err != nil {
		return "", err
	}
	var b bytes.Buffer
	err = t.Execute(&b, nil)
	if err != nil {
		return "", err
	}
	return b.String(), nil
}

func checkDependencies(ch *chart.Chart, reqs *chartutil.Requirements) error {
	missing := []string{}

	deps := ch.GetDependencies()
	for _, r := range reqs.Dependencies {
		found := false
		for _, d := range deps {
			if d.Metadata.Name == r.Name {
				found = true
				break
			}
		}
		if !found {
			missing = append(missing, r.Name)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("found in requirements.yaml, but missing in charts/ directory: %s", strings.Join(missing, ", "))
	}
	return nil
}

type valueFiles []string

func (v valueFiles) String() string {
	return fmt.Sprintf("%s", v)
}

func (v valueFiles) Type() string {
	return "valueFiles"
}

func (v *valueFiles) Set(value string) error {
	for _, fp := range strings.Split(value, ",") {
		*v = append(*v, fp)
	}
	return nil
}

func mergeValues(dest map[string]interface{}, src map[string]interface{}) map[string]interface{} {
	for k, v := range src {
		// If the key doesn't exist already, then just set the key to that value
		if _, exists := dest[k]; !exists {
			dest[k] = v
			continue
		}
		nextMap, ok := v.(map[string]interface{})
		// If it isn't another map, overwrite the value
		if !ok {
			dest[k] = v
			continue
		}
		// If the key doesn't exist already, then just set the key to that value
		if _, exists := dest[k]; !exists {
			dest[k] = nextMap
			continue
		}
		// Edge case: If the key exists in the destination, but isn't a map
		destMap, isMap := dest[k].(map[string]interface{})
		// If the source map has a map for this key, prefer it
		if !isMap {
			dest[k] = v
			continue
		}
		// If we got to this point, it is a map in both, so merge them
		dest[k] = mergeValues(destMap, nextMap)
	}
	return dest
}

func MergeValues(values, stringValues []string) ([]byte, error) {
	return MergeValuesF([]string{}, values, stringValues)
}

func MergeBytesValues(vbytes []byte, values []string) ([]byte, error) {
	base := map[string]interface{}{}
	currentMap := map[string]interface{}{}
	if err := yaml.Unmarshal(vbytes, &currentMap); err != nil {
		return []byte{}, fmt.Errorf("Failed to parse: %s, error: %v", string(vbytes), err)
	}
	base = mergeValues(base, currentMap)

	for _, value := range values {
		if err := strvals.ParseInto(value, base); err != nil {
			return []byte{}, fmt.Errorf("failed parsing set data: %s", err)
		}
	}
	return yaml.Marshal(base)
}

func MergeValuesF(valueFiles valueFiles, values, stringValues []string) ([]byte, error) {
	base := map[string]interface{}{}

	// parse values files
	for _, filePath := range valueFiles {
		currentMap := map[string]interface{}{}

		var bbytes []byte
		var err error
		bbytes, err = ioutil.ReadFile(filePath)
		if err != nil {
			return []byte{}, err
		}

		if err := yaml.Unmarshal(bbytes, &currentMap); err != nil {
			return []byte{}, fmt.Errorf("Failed to parse %s: %s", filePath, err)
		}
		// Merge with the previous map
		base = mergeValues(base, currentMap)
	}

	// parse set values
	for _, value := range values {
		if err := strvals.ParseInto(value, base); err != nil {
			return []byte{}, fmt.Errorf("failed parsing set data: %s", err)
		}
	}

	// parse set string values
	for _, value := range stringValues {
		if err := strvals.ParseIntoString(value, base); err != nil {
			return []byte{}, fmt.Errorf("failed parsing set string: %s", err)
		}
	}

	return yaml.Marshal(base)
}

type CreateUpdateReleaseReq struct {
	ChartName   string   `json:"chart_name"`
	Namespace   string   `json:"namespace"`
	ReleaseName string   `json:"release_name"`
	Version     string   `json:"version"`
	ReUseValues bool     `json:"reuse_values"`
	ResetValues bool     `json:"reset_values"`
	DryRun      bool     `json:"dry_run"`
	Values      string   `json:"values"`
	Sets        []string `json:"sets"`
	Timeout     int64    `json:"timeout"`
}

func (c *CreateUpdateReleaseReq) Vals() ([]byte, error) {
	return MergeBytesValues([]byte(c.Values), c.Sets)
}

func NewCreateUpdateReleaseReq(params jsonutils.JSONObject) (*CreateUpdateReleaseReq, error) {
	var req CreateUpdateReleaseReq
	err := params.Unmarshal(&req)
	if err != nil {
		return nil, err
	}
	if req.Timeout == 0 {
		req.Timeout = 300 // set default 5 mins timeout
	}
	return &req, nil
}

func ReleaseUpgrade(r *CreateUpdateReleaseReq, regionName string) (*rls.UpdateReleaseResponse, error) {
	downloadedChartPath, err := downloadChartFromRepo(r.ChartName, r.Version, generateHelmRepoPath(regionName))
	if err != nil {
		return nil, err
	}
	chartRequested, err := chartutil.Load(downloadedChartPath)
	if err != nil {
		return nil, fmt.Errorf("Error loading chart: %v", err)
	}
	if req, err := chartutil.LoadRequirements(chartRequested); err == nil {
		if err := checkDependencies(chartRequested, req); err != nil {
			return nil, err
		}
	} else if err != chartutil.ErrRequirementsNotFound {
		return nil, fmt.Errorf("cannot load requirements: %v", err)
	}
	//Get cluster based or inCluster kubeconfig
	hClient, err := GetHelmClient(regionName)
	if err != nil {
		return nil, err
	}
	vals, err := r.Vals()
	if err != nil {
		return nil, err
	}
	upgradeRes, err := hClient.UpdateReleaseFromChart(
		r.ReleaseName,
		chartRequested,
		helm.UpdateValueOverrides(vals),
		helm.UpgradeDryRun(r.DryRun),
		helm.UpgradeTimeout(r.Timeout),
		helm.ResetValues(r.ResetValues),
		helm.ReuseValues(r.ReUseValues),
	)
	if err != nil {
		return nil, fmt.Errorf("upgrade failed: %v", err)
	}
	return upgradeRes, nil
}

func setChartAnnotations(chrt *chart.Chart, r *CreateUpdateReleaseReq, regionName string) error {
	chartInfo := strings.Split(r.ChartName, "/")
	if len(chartInfo) != 2 {
		return fmt.Errorf("Invalid chart name: %q", r.ChartName)
	}
	chartRepo := chartInfo[0]
	chartName := chartInfo[1]
	if chrt.Metadata.Annotations == nil {
		chrt.Metadata.Annotations = make(map[string]string)
	}
	repoInfo, err := RepoShow(chartRepo, regionName)
	if err != nil {
		return err
	}
	chrt.Metadata.Annotations[YunionChartRepoAnnotation] = chartRepo
	chrt.Metadata.Annotations[YunionChartNameAnnotation] = chartName
	chrt.Metadata.Annotations[YunionRepoURLAnnotation] = repoInfo.URL
	return nil
}

func ReleaseCreate(r *CreateUpdateReleaseReq, regionName string) (*rls.InstallReleaseResponse, error) {
	log.Infof("Deploying chart=%q, release name=%q.", r.ChartName, r.ReleaseName)
	downloadedChartPath, err := downloadChartFromRepo(r.ChartName, r.Version, generateHelmRepoPath(regionName))
	if err != nil {
		return nil, err
	}
	log.Infof("Loading chart %q", downloadedChartPath)
	chartRequested, err := chartutil.Load(downloadedChartPath)
	if err != nil {
		return nil, fmt.Errorf("Error loading chart: %v", err)
	}
	if req, err := chartutil.LoadRequirements(chartRequested); err == nil {
		if err := checkDependencies(chartRequested, req); err != nil {
			return nil, err
		}
	} else if err != chartutil.ErrRequirementsNotFound {
		return nil, fmt.Errorf("cannot load requirements: %v", err)
	}
	err = setChartAnnotations(chartRequested, r, regionName)
	if err != nil {
		return nil, fmt.Errorf("Set chart annotations: %v", err)
	}
	if len(r.Namespace) == 0 {
		r.Namespace = "default"
	}
	if len(strings.TrimSpace(r.ReleaseName)) == 0 {
		r.ReleaseName, _ = generateName("")
	}
	hClient, err := GetHelmClient(regionName)
	if err != nil {
		return nil, err
	}
	vals, err := r.Vals()
	if err != nil {
		return nil, err
	}
	installRes, err := hClient.InstallReleaseFromChart(
		chartRequested,
		r.Namespace,
		helm.ValueOverrides(vals),
		helm.ReleaseName(r.ReleaseName),
		helm.InstallDryRun(r.DryRun),
		helm.InstallReuseName(true),
		helm.InstallDisableHooks(false),
		helm.InstallTimeout(r.Timeout),
		helm.InstallWait(false))
	if err != nil {
		return nil, fmt.Errorf("Error deploying chart: %v", err)
	}
	return installRes, nil
}

func ReleaseDelete(releaseName, regionName string) error {
	hClient, err := GetHelmClient(regionName)
	if err != nil {
		return err
	}
	// TODO: sophisticate commant options
	opts := []helm.DeleteOption{
		helm.DeletePurge(true),
	}
	_, err = hClient.DeleteRelease(releaseName, opts...)
	if err != nil {
		return err
	}
	return nil
}

// ChartQuery describes a query to get available helm chart's list
type ChartQuery struct {
	Name       string `json:"name"`
	Repo       string `json:"repo"`
	RepoURL    string `json:"repo_url"`
	AllVersion bool   `json:"all_version"`
	Keyword    string `json:"keyword"`
}

// Chart describe a chart
type ChartResult struct {
	Repo  string             `json:"repo"`
	Chart *repo.ChartVersion `json:"chart"`
}

func ChartsList(regionName string, query ChartQuery) ([]*ChartResult, error) {
	repoPath := fmt.Sprintf("%s/repository/repositories.yaml", generateHelmRepoPath(regionName))
	log.Debugf("Helm repo path: %s", repoPath)
	f, err := repo.LoadRepositoriesFile(repoPath)
	if err != nil {
		return nil, err
	}
	if len(f.Repositories) == 0 {
		return nil, nil
	}

	trans := func(repoName string, versions repo.ChartVersions) []*ChartResult {
		ret := make([]*ChartResult, 0)
		for _, c := range versions {
			ret = append(ret, &ChartResult{
				Repo:  repoName,
				Chart: c,
			})
		}
		return ret
	}

	cl := make([]*ChartResult, 0)
	for _, r := range f.Repositories {
		log.Debugf("Repository: %s", r.Name)
		i, errIndx := repo.LoadIndexFile(r.Cache)
		if errIndx != nil {
			return nil, errIndx
		}
		if query.Repo != "" {
			repoMatched := query.Repo == strings.ToLower(r.Name)
			if !repoMatched {
				continue
			}
		}
		if query.RepoURL != "" {
			repoURLMatched := query.RepoURL == r.URL
			if !repoURLMatched {
				continue
			}
		}
		log.Debugf("Repository: %s Matched", r.Name)
		for n := range i.Entries {
			log.Debugf("Chart: %s", n)
			chartMatched, _ := regexp.MatchString(query.Name, strings.ToLower(n))

			kwString := strings.ToLower(strings.Join(i.Entries[n][0].Keywords, " "))
			log.Debugf("kwString: %s", kwString)

			kwMatched, _ := regexp.MatchString(query.Keyword, kwString)
			if (chartMatched || query.Name == "") && (kwMatched || query.Keyword == "") {
				log.Debugf("Chart: %s Matched", n)
				if !query.AllVersion {
					cl = append(cl, &ChartResult{
						Repo:  r.Name,
						Chart: i.Entries[n][0],
					})
				} else {
					cl = append(cl, trans(r.Name, i.Entries[n])...)
				}
			}
		}
	}
	return cl, nil
}

// SpotGuideFile describes a spotguide file with the options
type SpotGuideFile struct {
	Options []SpotguideOptions `json:"options"`
}

// SpotguideOptions describes a spotguide options
type SpotguideOptions struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Default bool   `json:"default"`
	Info    string `json:"info"`
	Key     string `json:"key"`
}

type Runable struct {
	CanDeploy        bool   `json:"can_deploy"`
	Reason           string `json:"reason"`
	Release          string `json:"release"`
	ReleaseNamespace string `json:"release_namespace"`
}

// ChartDetails describes a chart details
type ChartDetails struct {
	Name    string             `json:"name"`
	Repo    string             `json:"repo"`
	Chart   *repo.ChartVersion `json:"chart"`
	Values  string             `json:"values"`
	Readme  string             `json:"readme"`
	Options []SpotguideOptions `json:"options"`
	Runable Runable            `json:"runable"`
}

func ChartShowDetails(regionName, chartRepo, chartName, chartVersion string) (*ChartDetails, error) {
	repoPath := fmt.Sprintf("%s/repository/repositories.yaml", generateHelmRepoPath(regionName))
	log.Debugf("Helm repo path: %s", repoPath)
	chartD := &ChartDetails{}
	f, err := repo.LoadRepositoriesFile(repoPath)
	if err != nil {
		return nil, err
	}
	if len(f.Repositories) == 0 {
		return nil, nil
	}

	dstRepo, err := (&RepoFile{f}).Get(chartRepo)
	if err != nil {
		return nil, err
	}
	if dstRepo == nil {
		return nil, nil
	}
	i, errIndx := repo.LoadIndexFile(dstRepo.Cache)
	if errIndx != nil {
		return nil, errIndx
	}
	dstChart, err := i.Get(chartName, chartVersion)
	if err != nil {
		return nil, err
	}
	if dstChart == nil {
		return nil, nil
	}

	chartSrc := dstChart.URLs[0]
	log.Debugf("Get chartsource: %s", chartSrc)
	reader, err := downloadFile(chartSrc)
	if err != nil {
		return nil, err
	}
	valuesStr, err := getChartFile(reader, "values.yaml")
	if err != nil {
		return nil, err
	}
	options, err := getChartOption(reader)
	if err != nil {
		return nil, err
	}
	log.Debugf("values hash: %s", valuesStr)

	readmeStr, err := getChartFile(reader, "README.md")
	if err != nil {
		return nil, err
	}
	log.Debugf("readme hash: %s", readmeStr)
	runable := Runable{CanDeploy: true}
	chartD = &ChartDetails{
		Name:    chartName,
		Repo:    chartRepo,
		Chart:   dstChart,
		Values:  valuesStr,
		Readme:  readmeStr,
		Options: options.Options,
		Runable: runable,
	}
	if !RunableCheckKeywords.HasAny(dstChart.Metadata.Keywords...) {
		return chartD, nil
	}
	releaseList, err := ReleasesList(ReleaseListQuery{All: true}, regionName)
	if err != nil {
		return nil, err
	}
	runable = ChartRunable(dstChart.Metadata, releaseList.GetReleases())
	chartD.Runable = runable
	return chartD, nil
}

func chartsWithDependencies(chrt *chart.Chart) []*chart.Chart {
	ret := make([]*chart.Chart, 0)
	if chrt == nil {
		return ret
	}
	ret = append(ret, chrt)
	deps := chrt.GetDependencies()
	if len(deps) == 0 {
		return ret
	}
	for _, dep := range deps {
		depRet := chartsWithDependencies(dep)
		if len(depRet) == 0 {
			continue
		}
		ret = append(ret, depRet...)
	}
	return ret
}

type chartList []*chart.Chart

func (l chartList) Has(meta *chart.Metadata) bool {
	for _, lch := range l {
		if lch.Metadata.Name == meta.Name {
			return true
		}
	}
	return false
}

func releaseCharts(r *release.Release) chartList {
	return chartsWithDependencies(r.GetChart())
}

func ChartRunable(meta *chart.Metadata, releases []*release.Release) (ret Runable) {
	ret = Runable{CanDeploy: true}
	if len(releases) == 0 {
		return
	}
	for _, r := range releases {
		if releaseCharts(r).Has(meta) {
			ret.CanDeploy = false
			ret.Reason = fmt.Sprintf("%s already used by release: %s, namespace: %s", meta.Name, r.Name, r.Namespace)
			ret.Release = r.Name
			ret.ReleaseNamespace = r.Namespace
			return
		}
	}
	return
}

func TillerStatus(regionName string) (string, error) {
	cli, err := GetHelmClient(regionName)
	if err != nil {
		return "", fmt.Errorf("Get helm cliet: %v", err)
	}
	err = cli.PingTiller()
	if err != nil {
		return "", err
	}
	return "Tiller is available", nil
}
