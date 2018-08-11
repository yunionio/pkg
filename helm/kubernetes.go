package helm

import (
	"fmt"
	"io/ioutil"
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/helm/pkg/helm"
	"k8s.io/helm/pkg/helm/portforwarder"
	"k8s.io/helm/pkg/kube"

	"yunion.io/x/log"
	"yunion.io/x/pkg/util/runtime"
	"yunion.io/x/pkg/util/wait"
)

type K8sRegionConfig map[string]string

func NewK8sRegionConfig() K8sRegionConfig {
	return make(map[string]string)
}

func (m K8sRegionConfig) AddConfig(regionName string, conf string) error {
	contents, err := ioutil.ReadFile(conf)
	if err != nil {
		return err
	}
	m[regionName] = string(contents)
	return nil
}

func (m K8sRegionConfig) GetConfigBytes(regionName string) ([]byte, error) {
	conf, err := m.GetConfig(regionName)
	if err != nil {
		return nil, err
	}
	return []byte(conf), err
}

func (m K8sRegionConfig) GetConfig(regionName string) (string, error) {
	conf, ok := m[regionName]
	if !ok {
		return "", fmt.Errorf("Not found config by region %q", regionName)
	}
	if len(conf) == 0 {
		return "", fmt.Errorf("Region %q config is empty", regionName)
	}
	return conf, nil
}

type HelmTunnelClient struct {
	*helm.Client
	tillerTunnel *kube.Tunnel
	k8sClient    kubernetes.Interface
	k8sConfig    *rest.Config
}

func NewHelmTunnelClient(client kubernetes.Interface, config *rest.Config) (*HelmTunnelClient, error) {
	cli := &HelmTunnelClient{
		k8sClient: client,
		k8sConfig: config,
	}
	err := cli.tunnel()
	return cli, err
}

func (c *HelmTunnelClient) tunnel() error {
	log.Debugf("Create kubernetes Tunnel")
	tillerTunnel, err := portforwarder.New("kube-system", c.k8sClient, c.k8sConfig)
	if err != nil {
		return fmt.Errorf("create tunnel failed: %v", err)
	}
	tillerTunnelAddress := fmt.Sprintf("localhost:%d", tillerTunnel.Local)
	log.Debugf("Created kubernetes tunnel on address: %s", tillerTunnelAddress)
	helmClient := helm.NewClient(helm.Host(tillerTunnelAddress))
	c.Client = helmClient
	c.tillerTunnel = tillerTunnel
	return nil
}

func (c *HelmTunnelClient) StartHealthCheck() {
	wait.Forever(c.healthCheck, time.Second*30)
}

func (c *HelmTunnelClient) healthCheck() {
	log.Debugf("Start tiller healthCheck")
	err := c.Client.PingTiller()
	if err == nil {
		log.Debugf("Tiller health ok")
		return
	}
	log.Errorf("Tiller not health: %v, start reconnect...", err)
	c.tillerTunnel.Close()
	err = c.tunnel()
	if err != nil {
		log.Errorf("Reconnect tiller error: %v", err)
	}
}

type HelmClientsManager struct {
	kubeConfigMap K8sRegionConfig
	Clients       map[string]*HelmTunnelClient
}

func NewHelmClientsManager(configMap K8sRegionConfig) *HelmClientsManager {
	// temporarily fix healthCheck panic recover
	runtime.ReallyCrash = false

	m := &HelmClientsManager{
		kubeConfigMap: configMap,
	}
	m.Clients = make(map[string]*HelmTunnelClient)
	for region := range configMap {
		cli, err := m.newHelmClient(region)
		if err != nil {
			log.Errorf("New helm client error: %v", err)
			continue
		}
		m.Clients[region] = cli
		go cli.StartHealthCheck()
	}
	return m
}

func (m *HelmClientsManager) newHelmClient(regionName string) (*HelmTunnelClient, error) {
	conf, err := m.kubeConfigMap.GetConfigBytes(regionName)
	if err != nil {
		return nil, err
	}
	k8sConfig, err := GetK8sClientConfig(conf)
	if err != nil {
		return nil, fmt.Errorf("create kubernetes config failed: %v", err)
	}
	k8sClient, err := GetK8sConnection(conf)
	if err != nil {
		return nil, fmt.Errorf("create kubernetes client failed: %v", err)
	}
	return NewHelmTunnelClient(k8sClient, k8sConfig)
}

func (m *HelmClientsManager) Client(regionName string) (*HelmTunnelClient, error) {
	cli, ok := m.Clients[regionName]
	if !ok {
		return nil, fmt.Errorf("Not found region %q k8s client", regionName)
	}
	return cli, nil
}

func GetK8sConnection(kubeConfig []byte) (*kubernetes.Clientset, error) {
	config, err := GetK8sClientConfig(kubeConfig)
	if err != nil {
		return nil, fmt.Errorf("Create kubernetes config: %v", err)
	}
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("Create kubernetes connection: %v", err)
	}
	return client, nil
}

func GetK8sClientConfig(kubeConfig []byte) (*rest.Config, error) {
	var config *rest.Config
	var err error
	if kubeConfig != nil {
		apiconfig, err := clientcmd.Load(kubeConfig)
		if err != nil {
			return nil, err
		}

		clientConfig := clientcmd.NewDefaultClientConfig(*apiconfig, &clientcmd.ConfigOverrides{})
		config, err = clientConfig.ClientConfig()
		if err != nil {
			return nil, err
		}
		log.Debugf("Use k8s remote cluster config connect to %q", config.Host)
	} else {
		return nil, fmt.Errorf("kubeconfig value is nil")
	}
	if err != nil {
		return nil, fmt.Errorf("crate kubernetes config failed: %v", err)
	}
	return config, nil
}

var defaultHelmMan *HelmClientsManager

func Init(configMap K8sRegionConfig) {
	defaultHelmMan = NewHelmClientsManager(configMap)
}

func GetHelmClient(regionName string) (*HelmTunnelClient, error) {
	if defaultHelmMan == nil {
		return nil, fmt.Errorf("Default helm clients manager not init")
	}
	return defaultHelmMan.Client(regionName)
}
