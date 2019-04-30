package credentialprovider

import (
	"encoding/json"
	"net"
	"net/url"
	"path/filepath"
	"sort"
	"strings"
	"github.com/golang/glog"
	dockertypes "github.com/docker/docker/api/types"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"
)

type DockerKeyring interface {
	Lookup(image string) ([]LazyAuthConfiguration, bool)
}
type BasicDockerKeyring struct {
	index	[]string
	creds	map[string][]LazyAuthConfiguration
}
type lazyDockerKeyring struct{ Providers []DockerConfigProvider }
type LazyAuthConfiguration struct {
	dockertypes.AuthConfig
	Provider	DockerConfigProvider
}

func DockerConfigEntryToLazyAuthConfiguration(ident DockerConfigEntry) LazyAuthConfiguration {
	_logClusterCodePath()
	defer _logClusterCodePath()
	return LazyAuthConfiguration{AuthConfig: dockertypes.AuthConfig{Username: ident.Username, Password: ident.Password, Email: ident.Email}}
}
func (dk *BasicDockerKeyring) Add(cfg DockerConfig) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	if dk.index == nil {
		dk.index = make([]string, 0)
		dk.creds = make(map[string][]LazyAuthConfiguration)
	}
	for loc, ident := range cfg {
		var creds LazyAuthConfiguration
		if ident.Provider != nil {
			creds = LazyAuthConfiguration{Provider: ident.Provider}
		} else {
			creds = DockerConfigEntryToLazyAuthConfiguration(ident)
		}
		value := loc
		if !strings.HasPrefix(value, "https://") && !strings.HasPrefix(value, "http://") {
			value = "https://" + value
		}
		parsed, err := url.Parse(value)
		if err != nil {
			glog.Errorf("Entry %q in dockercfg invalid (%v), ignoring", loc, err)
			continue
		}
		effectivePath := parsed.Path
		if strings.HasPrefix(effectivePath, "/v2/") || strings.HasPrefix(effectivePath, "/v1/") {
			effectivePath = effectivePath[3:]
		}
		var key string
		if (len(effectivePath) > 0) && (effectivePath != "/") {
			key = parsed.Host + effectivePath
		} else {
			key = parsed.Host
		}
		dk.creds[key] = append(dk.creds[key], creds)
		dk.index = append(dk.index, key)
	}
	eliminateDupes := sets.NewString(dk.index...)
	dk.index = eliminateDupes.List()
	sort.Sort(sort.Reverse(sort.StringSlice(dk.index)))
}

const (
	defaultRegistryHost	= "index.docker.io"
	defaultRegistryKey	= defaultRegistryHost + "/v1/"
)

func isDefaultRegistryMatch(image string) bool {
	_logClusterCodePath()
	defer _logClusterCodePath()
	parts := strings.SplitN(image, "/", 2)
	if len(parts[0]) == 0 {
		return false
	}
	if len(parts) == 1 {
		return true
	}
	if parts[0] == "docker.io" || parts[0] == "index.docker.io" {
		return true
	}
	return !strings.ContainsAny(parts[0], ".:")
}
func parseSchemelessUrl(schemelessUrl string) (*url.URL, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	parsed, err := url.Parse("https://" + schemelessUrl)
	if err != nil {
		return nil, err
	}
	parsed.Scheme = ""
	return parsed, nil
}
func splitUrl(url *url.URL) (parts []string, port string) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	host, port, err := net.SplitHostPort(url.Host)
	if err != nil {
		host, port = url.Host, ""
	}
	return strings.Split(host, "."), port
}
func urlsMatchStr(glob string, target string) (bool, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	globUrl, err := parseSchemelessUrl(glob)
	if err != nil {
		return false, err
	}
	targetUrl, err := parseSchemelessUrl(target)
	if err != nil {
		return false, err
	}
	return urlsMatch(globUrl, targetUrl)
}
func urlsMatch(globUrl *url.URL, targetUrl *url.URL) (bool, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	globUrlParts, globPort := splitUrl(globUrl)
	targetUrlParts, targetPort := splitUrl(targetUrl)
	if globPort != targetPort {
		return false, nil
	}
	if len(globUrlParts) != len(targetUrlParts) {
		return false, nil
	}
	if !strings.HasPrefix(targetUrl.Path, globUrl.Path) {
		return false, nil
	}
	for k, globUrlPart := range globUrlParts {
		targetUrlPart := targetUrlParts[k]
		matched, err := filepath.Match(globUrlPart, targetUrlPart)
		if err != nil {
			return false, err
		}
		if !matched {
			return false, nil
		}
	}
	return true, nil
}
func (dk *BasicDockerKeyring) Lookup(image string) ([]LazyAuthConfiguration, bool) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	ret := []LazyAuthConfiguration{}
	for _, k := range dk.index {
		if matched, _ := urlsMatchStr(k, image); !matched {
			continue
		}
		ret = append(ret, dk.creds[k]...)
	}
	if len(ret) > 0 {
		return ret, true
	}
	if isDefaultRegistryMatch(image) {
		if auth, ok := dk.creds[defaultRegistryHost]; ok {
			return auth, true
		}
	}
	return []LazyAuthConfiguration{}, false
}
func (dk *lazyDockerKeyring) Lookup(image string) ([]LazyAuthConfiguration, bool) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	keyring := &BasicDockerKeyring{}
	for _, p := range dk.Providers {
		keyring.Add(p.Provide())
	}
	return keyring.Lookup(image)
}

type FakeKeyring struct {
	auth	[]LazyAuthConfiguration
	ok	bool
}

func (f *FakeKeyring) Lookup(image string) ([]LazyAuthConfiguration, bool) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	return f.auth, f.ok
}

type unionDockerKeyring struct{ keyrings []DockerKeyring }

func (k *unionDockerKeyring) Lookup(image string) ([]LazyAuthConfiguration, bool) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	authConfigs := []LazyAuthConfiguration{}
	for _, subKeyring := range k.keyrings {
		if subKeyring == nil {
			continue
		}
		currAuthResults, _ := subKeyring.Lookup(image)
		authConfigs = append(authConfigs, currAuthResults...)
	}
	return authConfigs, (len(authConfigs) > 0)
}
func MakeDockerKeyring(passedSecrets []v1.Secret, defaultKeyring DockerKeyring) (DockerKeyring, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	passedCredentials := []DockerConfig{}
	for _, passedSecret := range passedSecrets {
		if dockerConfigJsonBytes, dockerConfigJsonExists := passedSecret.Data[v1.DockerConfigJsonKey]; (passedSecret.Type == v1.SecretTypeDockerConfigJson) && dockerConfigJsonExists && (len(dockerConfigJsonBytes) > 0) {
			dockerConfigJson := DockerConfigJson{}
			if err := json.Unmarshal(dockerConfigJsonBytes, &dockerConfigJson); err != nil {
				return nil, err
			}
			passedCredentials = append(passedCredentials, dockerConfigJson.Auths)
		} else if dockercfgBytes, dockercfgExists := passedSecret.Data[v1.DockerConfigKey]; (passedSecret.Type == v1.SecretTypeDockercfg) && dockercfgExists && (len(dockercfgBytes) > 0) {
			dockercfg := DockerConfig{}
			if err := json.Unmarshal(dockercfgBytes, &dockercfg); err != nil {
				return nil, err
			}
			passedCredentials = append(passedCredentials, dockercfg)
		}
	}
	if len(passedCredentials) > 0 {
		basicKeyring := &BasicDockerKeyring{}
		for _, currCredentials := range passedCredentials {
			basicKeyring.Add(currCredentials)
		}
		return &unionDockerKeyring{[]DockerKeyring{basicKeyring, defaultKeyring}}, nil
	}
	return defaultKeyring, nil
}
