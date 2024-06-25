package kubernetes

type kubeconfig struct {
	Clusters       []kubeconfigCluster `yaml:"clusters"`
	Contexts       []kubeconfigContext `yaml:"contexts"`
	CurrentContext string              `yaml:"current-context"`
	Users          []kubeconfigUser    `yaml:"users"`
}

type kubeconfigCluster struct {
	Cluster kubeconfigClusterData `yaml:"cluster"`
	Name    string                `yaml:"name"`
}

type kubeconfigClusterData struct {
	CertificateAuthorityData string `yaml:"certificate-authority-data"`
	Server                   string `yaml:"server"`
}

type kubeconfigContext struct {
	Context kubeconfigContextData `yaml:"context"`
	Name    string                `yaml:"name"`
}

type kubeconfigContextData struct {
	Cluster string `yaml:"cluster"`
	User    string `yaml:"user"`
}

type kubeconfigUser struct {
	User kubeconfigUserData `yaml:"user"`
	Name string             `yaml:"name"`
}

type kubeconfigUserData struct {
	ClientCertificateData string `yaml:"client-certificate-data"`
	ClientKeyData         string `yaml:"client-key-data"`
}
