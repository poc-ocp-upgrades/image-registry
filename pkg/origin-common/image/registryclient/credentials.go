package registryclient

import (
	"net/url"
	"strings"
	"sync"
	"github.com/docker/distribution/registry/client/auth"
	"github.com/golang/glog"
	"github.com/openshift/image-registry/pkg/kubernetes-common/credentialprovider"
	corev1 "k8s.io/api/core/v1"
)

var (
	NoCredentials auth.CredentialStore = &noopCredentialStore{}
)

type RefreshTokenStore interface {
	RefreshToken(url *url.URL, service string) string
	SetRefreshToken(url *url.URL, service string, token string)
}

func NewRefreshTokenStore() RefreshTokenStore {
	_logClusterCodePath()
	defer _logClusterCodePath()
	return &refreshTokenStore{}
}

type refreshTokenKey struct {
	url	string
	service	string
}
type refreshTokenStore struct {
	lock	sync.Mutex
	store	map[refreshTokenKey]string
}

func (s *refreshTokenStore) RefreshToken(url *url.URL, service string) string {
	_logClusterCodePath()
	defer _logClusterCodePath()
	s.lock.Lock()
	defer s.lock.Unlock()
	return s.store[refreshTokenKey{url: url.String(), service: service}]
}
func (s *refreshTokenStore) SetRefreshToken(url *url.URL, service string, token string) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.store == nil {
		s.store = make(map[refreshTokenKey]string)
	}
	s.store[refreshTokenKey{url: url.String(), service: service}] = token
}

type noopCredentialStore struct{}

func (s *noopCredentialStore) Basic(url *url.URL) (string, string) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	return "", ""
}
func (s *noopCredentialStore) RefreshToken(url *url.URL, service string) string {
	_logClusterCodePath()
	defer _logClusterCodePath()
	return ""
}
func (s *noopCredentialStore) SetRefreshToken(url *url.URL, service string, token string) {
	_logClusterCodePath()
	defer _logClusterCodePath()
}
func NewBasicCredentials() *BasicCredentials {
	_logClusterCodePath()
	defer _logClusterCodePath()
	return &BasicCredentials{refreshTokenStore: &refreshTokenStore{}}
}

type basicForURL struct {
	url			url.URL
	username, password	string
}
type BasicCredentials struct {
	creds	[]basicForURL
	*refreshTokenStore
}

func (c *BasicCredentials) Add(url *url.URL, username, password string) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	c.creds = append(c.creds, basicForURL{*url, username, password})
}
func (c *BasicCredentials) Basic(url *url.URL) (string, string) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	for _, cred := range c.creds {
		if len(cred.url.Host) != 0 && cred.url.Host != url.Host {
			continue
		}
		if len(cred.url.Path) != 0 && cred.url.Path != url.Path {
			continue
		}
		return cred.username, cred.password
	}
	return "", ""
}

var (
	emptyKeyring = &credentialprovider.BasicDockerKeyring{}
)

func NewCredentialsForSecrets(secrets []corev1.Secret) *SecretCredentialStore {
	_logClusterCodePath()
	defer _logClusterCodePath()
	return &SecretCredentialStore{secrets: secrets, RefreshTokenStore: NewRefreshTokenStore()}
}
func NewLazyCredentialsForSecrets(secretsFn func() ([]corev1.Secret, error)) *SecretCredentialStore {
	_logClusterCodePath()
	defer _logClusterCodePath()
	return &SecretCredentialStore{secretsFn: secretsFn, RefreshTokenStore: NewRefreshTokenStore()}
}

type SecretCredentialStore struct {
	lock		sync.Mutex
	secrets		[]corev1.Secret
	secretsFn	func() ([]corev1.Secret, error)
	err		error
	keyring		credentialprovider.DockerKeyring
	RefreshTokenStore
}

func (s *SecretCredentialStore) Basic(url *url.URL) (string, string) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	return basicCredentialsFromKeyring(s.init(), url)
}
func (s *SecretCredentialStore) Err() error {
	_logClusterCodePath()
	defer _logClusterCodePath()
	s.lock.Lock()
	defer s.lock.Unlock()
	return s.err
}
func (s *SecretCredentialStore) init() credentialprovider.DockerKeyring {
	_logClusterCodePath()
	defer _logClusterCodePath()
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.keyring != nil {
		return s.keyring
	}
	if s.secrets == nil {
		if s.secretsFn != nil {
			s.secrets, s.err = s.secretsFn()
		}
	}
	keyring, err := credentialprovider.MakeDockerKeyring(s.secrets, emptyKeyring)
	if err != nil {
		glog.V(5).Infof("Loading keyring failed for credential store: %v", err)
		s.err = err
		keyring = emptyKeyring
	}
	s.keyring = keyring
	return keyring
}
func basicCredentialsFromKeyring(keyring credentialprovider.DockerKeyring, target *url.URL) (string, string) {
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
			return basicCredentialsFromKeyring(keyring, &url.URL{Host: "index.docker.io", Path: "/v1"})
		}
		if value == "index.docker.io" {
			glog.V(5).Infof("Being asked for %s (%s), trying %s for legacy behavior", target, value, "docker.io")
			return basicCredentialsFromKeyring(keyring, &url.URL{Host: "docker.io"})
		}
		if (strings.HasSuffix(target.Host, ":443") && target.Scheme == "https") || (strings.HasSuffix(target.Host, ":80") && target.Scheme == "http") {
			host := strings.SplitN(target.Host, ":", 2)[0]
			glog.V(5).Infof("Being asked for %s (%s), trying %s without port", target, value, host)
			return basicCredentialsFromKeyring(keyring, &url.URL{Scheme: target.Scheme, Host: host, Path: target.Path})
		}
		glog.V(5).Infof("Unable to find a secret to match %s (%s)", target, value)
		return "", ""
	}
	glog.V(5).Infof("Found secret to match %s (%s): %s", target, value, configs[0].ServerAddress)
	return configs[0].Username, configs[0].Password
}
