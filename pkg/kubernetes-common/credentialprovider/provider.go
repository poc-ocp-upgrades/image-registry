package credentialprovider

import (
	"os"
	"reflect"
	"sync"
	"time"
	dockertypes "github.com/docker/docker/api/types"
	"github.com/golang/glog"
)

type DockerConfigProvider interface {
	Enabled() bool
	Provide() DockerConfig
	LazyProvide() *DockerConfigEntry
}

func LazyProvide(creds LazyAuthConfiguration) dockertypes.AuthConfig {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	if creds.Provider != nil {
		entry := *creds.Provider.LazyProvide()
		return DockerConfigEntryToLazyAuthConfiguration(entry).AuthConfig
	} else {
		return creds.AuthConfig
	}
}

type defaultDockerConfigProvider struct{}

func init() {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	RegisterCredentialProvider(".dockercfg", &CachingDockerConfigProvider{Provider: &defaultDockerConfigProvider{}, Lifetime: 5 * time.Minute})
}

type CachingDockerConfigProvider struct {
	Provider		DockerConfigProvider
	Lifetime		time.Duration
	cacheDockerConfig	DockerConfig
	expiration		time.Time
	mu			sync.Mutex
}

func (d *defaultDockerConfigProvider) Enabled() bool {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	return true
}
func (d *defaultDockerConfigProvider) Provide() DockerConfig {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	if cfg, err := ReadDockerConfigFile(); err == nil {
		return cfg
	} else if !os.IsNotExist(err) {
		glog.V(4).Infof("Unable to parse Docker config file: %v", err)
	}
	return DockerConfig{}
}
func (d *defaultDockerConfigProvider) LazyProvide() *DockerConfigEntry {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	return nil
}
func (d *CachingDockerConfigProvider) Enabled() bool {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	return d.Provider.Enabled()
}
func (d *CachingDockerConfigProvider) LazyProvide() *DockerConfigEntry {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	return nil
}
func (d *CachingDockerConfigProvider) Provide() DockerConfig {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	d.mu.Lock()
	defer d.mu.Unlock()
	if time.Now().Before(d.expiration) {
		return d.cacheDockerConfig
	}
	glog.V(2).Infof("Refreshing cache for provider: %v", reflect.TypeOf(d.Provider).String())
	d.cacheDockerConfig = d.Provider.Provide()
	d.expiration = time.Now().Add(d.Lifetime)
	return d.cacheDockerConfig
}
