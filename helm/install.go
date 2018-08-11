package helm

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/banzaicloud/banzai-types/components/helm"
	"github.com/pkg/errors"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/helm/cmd/helm/installer"
	"k8s.io/helm/pkg/getter"
	"k8s.io/helm/pkg/helm/environment"
	"k8s.io/helm/pkg/repo"

	"yunion.io/x/log"
)

const (
	stableRepository = "stable"
	yunionRepository = "yunion"
)

var (
	stableRepoURL = ""
	yunionRepoURL = ""
)

func SetStableRepoURL(url string) {
	stableRepoURL = url
}

func SetYunionRepoURL(url string) {
	yunionRepoURL = url
}

func ensureDirectories(env environment.EnvSettings) error {
	home := env.Home
	configDirectories := []string{
		home.String(),
		home.Repository(),
		home.Cache(),
		home.LocalRepository(),
		home.Plugins(),
		home.Starters(),
		home.Archive(),
	}

	log.Infof("Setting up helm directories")

	for _, p := range configDirectories {
		if fi, err := os.Stat(p); err != nil {
			log.Infof("Creating %q", p)
			if err := os.MkdirAll(p, 0755); err != nil {
				return errors.Wrapf(err, "Can't create %q", p)
			}
		} else if !fi.IsDir() {
			return errors.Errorf("%q must be a directory", p)
		}
	}
	return nil
}

func addDefaultRepos(env environment.EnvSettings) error {
	home := env.Home
	repoFile := home.RepositoryFile()

	log.Infof("Setting up default helm repos")

	repos := make([]*repo.Entry, 0)
	defaultRepoMap := map[string]string{
		stableRepository: stableRepoURL,
		yunionRepository: yunionRepoURL,
	}
	if fi, err := os.Stat(repoFile); err != nil {
		log.Infof("Creating %q", repoFile)
		f := repo.NewRepoFile()
		for repoName, url := range defaultRepoMap {
			if len(url) == 0 {
				log.Warningf("%q repo url is empty, skip add it", repoName)
				continue
			}
			sr, err := initRepo(repoName, url, env)
			if err != nil {
				return errors.Wrapf(err, "can't init %q repo", repoName)
			}
			repos = append(repos, sr)
		}
		f.Add(repos...)
		if err := f.WriteFile(repoFile, 0644); err != nil {
			return errors.Wrap(err, "can't create file")
		}
	} else if fi.IsDir() {
		return errors.Errorf("%s must be a file, not a directory", repoFile)
	}
	return nil
}

func initRepo(repoName, repoURL string, env environment.EnvSettings) (*repo.Entry, error) {
	log.Infof("Add %q repo with URL: %q", repoName, repoURL)
	c := repo.Entry{
		Name:  repoName,
		URL:   repoURL,
		Cache: env.Home.CacheIndex(repoName),
	}
	r, err := repo.NewChartRepository(&c, getter.All(env))
	if err != nil {
		return nil, errors.Wrap(err, "Can't create a new ChartRepo")
	}

	// In this case, the cacheFile is always absolute. So passing emtpry
	// string is safe
	if err := r.DownloadIndexFile(""); err != nil {
		return nil, errors.Errorf("Looks like %q is not a valid chart repository or can't be reached: %v", repoURL, err)
	}

	return &c, nil
}

func installHelmClient(path string) error {
	settings := createEnvSettings(generateHelmRepoPath(path))
	if err := ensureDirectories(settings); err != nil {
		return errors.Wrap(err, "Initializing helm directories failed!")
	}

	if err := addDefaultRepos(settings); err != nil {
		return errors.Wrap(err, "Setting up default repos failed")
	}

	log.Infof("Initializing helm client succeeded, happy helming")
	return nil
}

// PreInstall create ServiceAccount and AccountRoleBinding
func PreInstall(helmInstall *helm.Install, kubeConfig []byte) error {
	log.Infof("Start helm pre-install")

	cli, err := GetK8sConnection(kubeConfig)
	if err != nil {
		log.Errorf("Can't get kubernetes client: %v", err)
		return err
	}

	v1MetaData := metav1.ObjectMeta{
		Name:      helmInstall.ServiceAccount, // "tiller"
		Namespace: helmInstall.Namespace,
	}

	serviceAccount := &apiv1.ServiceAccount{
		ObjectMeta: v1MetaData,
	}
	log.Infof("Create service account: %q, namespace: %q", v1MetaData.Name, v1MetaData.Namespace)
	for i := 0; i <= 5; i++ {
		_, err = cli.CoreV1().ServiceAccounts(helmInstall.Namespace).Create(serviceAccount)
		if err != nil {
			log.Warningf("create service account failed: %v", err)
			if strings.Contains(err.Error(), "etcdserver: request timed out") {
				time.Sleep(time.Duration(40) * time.Second)
				continue
			}
			if !strings.Contains(err.Error(), "already exists") {
				return errors.Wrap(err, "create service account failed")
			}
		}
		break
	}

	clusterRole := &v1.ClusterRole{
		ObjectMeta: v1MetaData,
		Rules: []v1.PolicyRule{{
			APIGroups: []string{
				"*",
			},
			Resources: []string{
				"*",
			},
			Verbs: []string{
				"*",
			},
		},
			{
				NonResourceURLs: []string{
					"*",
				},
				Verbs: []string{
					"*",
				},
			}},
	}
	log.Infof("Create cluster roles: %q, namespace: %q", v1MetaData.Name, v1MetaData.Namespace)
	clusterRoleName := helmInstall.ServiceAccount
	for i := 0; i <= 5; i++ {
		_, err = cli.RbacV1().ClusterRoles().Create(clusterRole)
		if err != nil {
			if strings.Contains(err.Error(), "etcdserver: request timed out") {
				time.Sleep(time.Duration(10) * time.Second)
				continue
			} else if strings.Contains(err.Error(), "is forbidden") {
				_, errGet := cli.RbacV1().ClusterRoles().Get("cluster-admin", metav1.GetOptions{})
				if errGet != nil {
					return fmt.Errorf("clusterrole create error: %v cluster-admin not found: %v", err, errGet)
				}
				clusterRoleName = "cluster-admin"
				break
			}
			log.Warningf("create roles failed: %v", err)
			if !strings.Contains(err.Error(), "already exists") {
				return fmt.Errorf("crate roles failed: %s", err)
			}
		}
		break
	}

	log.Debugf("ClusterRole name: %s", clusterRoleName)
	log.Debugf("ServiceAccount name: %s", helmInstall.ServiceAccount)
	clusterRoleBinding := &v1.ClusterRoleBinding{
		ObjectMeta: v1MetaData,
		RoleRef: v1.RoleRef{
			APIGroup: v1.GroupName,
			Kind:     "ClusterRole",
			Name:     clusterRoleName,
		},
		Subjects: []v1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      helmInstall.ServiceAccount,
				Namespace: helmInstall.Namespace,
			},
		},
	}
	log.Infof("Crate cluster role bindings: %q, namespace: %q, roleRef: %q", v1MetaData.Name, v1MetaData.Namespace, clusterRoleName)
	for i := 0; i <= 5; i++ {
		_, err = cli.RbacV1().ClusterRoleBindings().Create(clusterRoleBinding)
		if err != nil {
			log.Warningf("create role bindings failed: %v", err)
			if strings.Contains(err.Error(), "etcdserver: request timed out") {
				time.Sleep(time.Duration(10) * time.Second)
				continue
			}
			if !strings.Contains(err.Error(), "already exists") {
				return fmt.Errorf("Create role bindings failed: %v", err)
			}
		}
		break
	}
	return nil
}

// Install uses Kubernetes client to install Tiller
func Install(helmInstall *helm.Install, kubeConfig []byte, regionName string) error {
	log.Infof("Install helm client")
	if err := installHelmClient(regionName); err != nil {
		return err
	}

	err := PreInstall(helmInstall, kubeConfig)
	if err != nil {
		return err
	}

	opts := installer.Options{
		Namespace:      helmInstall.Namespace,
		ServiceAccount: helmInstall.ServiceAccount,
		UseCanary:      helmInstall.Canary,
		ImageSpec:      helmInstall.ImageSpec,
		MaxHistory:     helmInstall.MaxHistory,
	}
	kubeClient, err := GetK8sConnection(kubeConfig)
	if err != nil {
		return errors.Wrap(err, "Get k8s client")
	}
	if err := installer.Install(kubeClient, &opts); err != nil {
		if !apierrors.IsAlreadyExists(err) {
			return err
		}
		if helmInstall.Upgrade {
			if err := installer.Upgrade(kubeClient, &opts); err != nil {
				return errors.Wrap(err, "error when upgrading")
			}
			log.Infof("Tiller (the Helm server-side component) has been upgraded to the current version.")
		} else {
			msg := "Tiller is already installed in the cluster."
			return errors.New(msg)
		}
	} else {
		log.Infof("Tiller (the Helm server-side component) has been installed into your Kuberntes Cluster.")
	}
	log.Infof("Helm install finished")
	return nil
}
