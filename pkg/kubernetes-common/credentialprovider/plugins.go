package credentialprovider

import (
	"sync"
	"github.com/golang/glog"
)

var providersMutex sync.Mutex
var providers = make(map[string]DockerConfigProvider)

func RegisterCredentialProvider(name string, provider DockerConfigProvider) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	providersMutex.Lock()
	defer providersMutex.Unlock()
	_, found := providers[name]
	if found {
		glog.Fatalf("Credential provider %q was registered twice", name)
	}
	glog.V(4).Infof("Registered credential provider %q", name)
	providers[name] = provider
}
func NewDockerKeyring() DockerKeyring {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	keyring := &lazyDockerKeyring{Providers: make([]DockerConfigProvider, 0)}
	for name, provider := range providers {
		if provider.Enabled() {
			glog.V(4).Infof("Registering credential provider: %v", name)
			keyring.Providers = append(keyring.Providers, provider)
		}
	}
	return keyring
}
