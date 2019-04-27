package credentialprovider

import (
	"encoding/base64"
	"fmt"
	"reflect"
	"testing"
	dockertypes "github.com/docker/docker/api/types"
)

func TestUrlsMatch(t *testing.T) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	tests := []struct {
		globUrl		string
		targetUrl	string
		matchExpected	bool
	}{{globUrl: "*.kubernetes.io", targetUrl: "prefix.kubernetes.io", matchExpected: true}, {globUrl: "prefix.*.io", targetUrl: "prefix.kubernetes.io", matchExpected: true}, {globUrl: "prefix.kubernetes.*", targetUrl: "prefix.kubernetes.io", matchExpected: true}, {globUrl: "*-good.kubernetes.io", targetUrl: "prefix-good.kubernetes.io", matchExpected: true}, {globUrl: "*.kubernetes.io/blah", targetUrl: "prefix.kubernetes.io/blah", matchExpected: true}, {globUrl: "prefix.*.io/foo", targetUrl: "prefix.kubernetes.io/foo/bar", matchExpected: true}, {globUrl: "*.kubernetes.io:1111/blah", targetUrl: "prefix.kubernetes.io:1111/blah", matchExpected: true}, {globUrl: "prefix.*.io:1111/foo", targetUrl: "prefix.kubernetes.io:1111/foo/bar", matchExpected: true}, {globUrl: "*.kubernetes.io", targetUrl: "kubernetes.io", matchExpected: false}, {globUrl: "*.*.kubernetes.io", targetUrl: "prefix.kubernetes.io", matchExpected: false}, {globUrl: "*.*.kubernetes.io", targetUrl: "kubernetes.io", matchExpected: false}, {globUrl: "kubernetes.io", targetUrl: "kubernetes.com", matchExpected: false}, {globUrl: "k*.io", targetUrl: "quay.io", matchExpected: false}, {globUrl: "*.kubernetes.io:1234/blah", targetUrl: "prefix.kubernetes.io:1111/blah", matchExpected: false}, {globUrl: "prefix.*.io/foo", targetUrl: "prefix.kubernetes.io:1111/foo/bar", matchExpected: false}}
	for _, test := range tests {
		matched, _ := urlsMatchStr(test.globUrl, test.targetUrl)
		if matched != test.matchExpected {
			t.Errorf("Expected match result of %s and %s to be %t, but was %t", test.globUrl, test.targetUrl, test.matchExpected, matched)
		}
	}
}
func TestDockerKeyringForGlob(t *testing.T) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	tests := []struct {
		globUrl		string
		targetUrl	string
	}{{globUrl: "https://hello.kubernetes.io", targetUrl: "hello.kubernetes.io"}, {globUrl: "https://*.docker.io", targetUrl: "prefix.docker.io"}, {globUrl: "https://prefix.*.io", targetUrl: "prefix.docker.io"}, {globUrl: "https://prefix.docker.*", targetUrl: "prefix.docker.io"}, {globUrl: "https://*.docker.io/path", targetUrl: "prefix.docker.io/path"}, {globUrl: "https://prefix.*.io/path", targetUrl: "prefix.docker.io/path/subpath"}, {globUrl: "https://prefix.docker.*/path", targetUrl: "prefix.docker.io/path"}, {globUrl: "https://*.docker.io:8888", targetUrl: "prefix.docker.io:8888"}, {globUrl: "https://prefix.*.io:8888", targetUrl: "prefix.docker.io:8888"}, {globUrl: "https://prefix.docker.*:8888", targetUrl: "prefix.docker.io:8888"}, {globUrl: "https://*.docker.io/path:1111", targetUrl: "prefix.docker.io/path:1111"}, {globUrl: "https://*.docker.io/v1/", targetUrl: "prefix.docker.io/path:1111"}, {globUrl: "https://*.docker.io/v2/", targetUrl: "prefix.docker.io/path:1111"}, {globUrl: "https://prefix.docker.*/path:1111", targetUrl: "prefix.docker.io/path:1111"}, {globUrl: "prefix.docker.io:1111", targetUrl: "prefix.docker.io:1111/path"}, {globUrl: "*.docker.io:1111", targetUrl: "prefix.docker.io:1111/path"}}
	for i, test := range tests {
		email := "foo@bar.baz"
		username := "foo"
		password := "bar"
		auth := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", username, password)))
		sampleDockerConfig := fmt.Sprintf(`{
   "%s": {
     "email": %q,
     "auth": %q
   }
}`, test.globUrl, email, auth)
		keyring := &BasicDockerKeyring{}
		if cfg, err := readDockerConfigFileFromBytes([]byte(sampleDockerConfig)); err != nil {
			t.Errorf("Error processing json blob %q, %v", sampleDockerConfig, err)
		} else {
			keyring.Add(cfg)
		}
		creds, ok := keyring.Lookup(test.targetUrl + "/foo/bar")
		if !ok {
			t.Errorf("%d: Didn't find expected URL: %s", i, test.targetUrl)
			continue
		}
		val := creds[0]
		if username != val.Username {
			t.Errorf("Unexpected username value, want: %s, got: %s", username, val.Username)
		}
		if password != val.Password {
			t.Errorf("Unexpected password value, want: %s, got: %s", password, val.Password)
		}
		if email != val.Email {
			t.Errorf("Unexpected email value, want: %s, got: %s", email, val.Email)
		}
	}
}
func TestKeyringMiss(t *testing.T) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	tests := []struct {
		globUrl		string
		lookupUrl	string
	}{{globUrl: "https://hello.kubernetes.io", lookupUrl: "world.mesos.org/foo/bar"}, {globUrl: "https://*.docker.com", lookupUrl: "prefix.docker.io"}, {globUrl: "https://suffix.*.io", lookupUrl: "prefix.docker.io"}, {globUrl: "https://prefix.docker.c*", lookupUrl: "prefix.docker.io"}, {globUrl: "https://prefix.*.io/path:1111", lookupUrl: "prefix.docker.io/path/subpath:1111"}, {globUrl: "suffix.*.io", lookupUrl: "prefix.docker.io"}}
	for _, test := range tests {
		email := "foo@bar.baz"
		username := "foo"
		password := "bar"
		auth := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", username, password)))
		sampleDockerConfig := fmt.Sprintf(`{
   "%s": {
     "email": %q,
     "auth": %q
   }
}`, test.globUrl, email, auth)
		keyring := &BasicDockerKeyring{}
		if cfg, err := readDockerConfigFileFromBytes([]byte(sampleDockerConfig)); err != nil {
			t.Errorf("Error processing json blob %q, %v", sampleDockerConfig, err)
		} else {
			keyring.Add(cfg)
		}
		_, ok := keyring.Lookup(test.lookupUrl + "/foo/bar")
		if ok {
			t.Errorf("Expected not to find URL %s, but found", test.lookupUrl)
		}
	}
}
func TestKeyringMissWithDockerHubCredentials(t *testing.T) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	url := defaultRegistryKey
	email := "foo@bar.baz"
	username := "foo"
	password := "bar"
	auth := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", username, password)))
	sampleDockerConfig := fmt.Sprintf(`{
   "https://%s": {
     "email": %q,
     "auth": %q
   }
}`, url, email, auth)
	keyring := &BasicDockerKeyring{}
	if cfg, err := readDockerConfigFileFromBytes([]byte(sampleDockerConfig)); err != nil {
		t.Errorf("Error processing json blob %q, %v", sampleDockerConfig, err)
	} else {
		keyring.Add(cfg)
	}
	val, ok := keyring.Lookup("world.mesos.org/foo/bar")
	if ok {
		t.Errorf("Found unexpected credential: %+v", val)
	}
}
func TestKeyringHitWithUnqualifiedDockerHub(t *testing.T) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	url := defaultRegistryKey
	email := "foo@bar.baz"
	username := "foo"
	password := "bar"
	auth := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", username, password)))
	sampleDockerConfig := fmt.Sprintf(`{
   "https://%s": {
     "email": %q,
     "auth": %q
   }
}`, url, email, auth)
	keyring := &BasicDockerKeyring{}
	if cfg, err := readDockerConfigFileFromBytes([]byte(sampleDockerConfig)); err != nil {
		t.Errorf("Error processing json blob %q, %v", sampleDockerConfig, err)
	} else {
		keyring.Add(cfg)
	}
	creds, ok := keyring.Lookup("google/docker-registry")
	if !ok {
		t.Errorf("Didn't find expected URL: %s", url)
		return
	}
	if len(creds) > 1 {
		t.Errorf("Got more hits than expected: %s", creds)
	}
	val := creds[0]
	if username != val.Username {
		t.Errorf("Unexpected username value, want: %s, got: %s", username, val.Username)
	}
	if password != val.Password {
		t.Errorf("Unexpected password value, want: %s, got: %s", password, val.Password)
	}
	if email != val.Email {
		t.Errorf("Unexpected email value, want: %s, got: %s", email, val.Email)
	}
}
func TestKeyringHitWithUnqualifiedLibraryDockerHub(t *testing.T) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	url := defaultRegistryKey
	email := "foo@bar.baz"
	username := "foo"
	password := "bar"
	auth := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", username, password)))
	sampleDockerConfig := fmt.Sprintf(`{
   "https://%s": {
     "email": %q,
     "auth": %q
   }
}`, url, email, auth)
	keyring := &BasicDockerKeyring{}
	if cfg, err := readDockerConfigFileFromBytes([]byte(sampleDockerConfig)); err != nil {
		t.Errorf("Error processing json blob %q, %v", sampleDockerConfig, err)
	} else {
		keyring.Add(cfg)
	}
	creds, ok := keyring.Lookup("jenkins")
	if !ok {
		t.Errorf("Didn't find expected URL: %s", url)
		return
	}
	if len(creds) > 1 {
		t.Errorf("Got more hits than expected: %s", creds)
	}
	val := creds[0]
	if username != val.Username {
		t.Errorf("Unexpected username value, want: %s, got: %s", username, val.Username)
	}
	if password != val.Password {
		t.Errorf("Unexpected password value, want: %s, got: %s", password, val.Password)
	}
	if email != val.Email {
		t.Errorf("Unexpected email value, want: %s, got: %s", email, val.Email)
	}
}
func TestKeyringHitWithQualifiedDockerHub(t *testing.T) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	url := defaultRegistryKey
	email := "foo@bar.baz"
	username := "foo"
	password := "bar"
	auth := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", username, password)))
	sampleDockerConfig := fmt.Sprintf(`{
   "https://%s": {
     "email": %q,
     "auth": %q
   }
}`, url, email, auth)
	keyring := &BasicDockerKeyring{}
	if cfg, err := readDockerConfigFileFromBytes([]byte(sampleDockerConfig)); err != nil {
		t.Errorf("Error processing json blob %q, %v", sampleDockerConfig, err)
	} else {
		keyring.Add(cfg)
	}
	creds, ok := keyring.Lookup(url + "/google/docker-registry")
	if !ok {
		t.Errorf("Didn't find expected URL: %s", url)
		return
	}
	if len(creds) > 2 {
		t.Errorf("Got more hits than expected: %s", creds)
	}
	val := creds[0]
	if username != val.Username {
		t.Errorf("Unexpected username value, want: %s, got: %s", username, val.Username)
	}
	if password != val.Password {
		t.Errorf("Unexpected password value, want: %s, got: %s", password, val.Password)
	}
	if email != val.Email {
		t.Errorf("Unexpected email value, want: %s, got: %s", email, val.Email)
	}
}
func TestIsDefaultRegistryMatch(t *testing.T) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	samples := []map[bool]string{{true: "foo/bar"}, {true: "docker.io/foo/bar"}, {true: "index.docker.io/foo/bar"}, {true: "foo"}, {false: ""}, {false: "registry.tld/foo/bar"}, {false: "registry:5000/foo/bar"}, {false: "myhostdocker.io/foo/bar"}}
	for _, sample := range samples {
		for expected, imageName := range sample {
			if got := isDefaultRegistryMatch(imageName); got != expected {
				t.Errorf("Expected '%s' to be %t, got %t", imageName, expected, got)
			}
		}
	}
}

type testProvider struct{ Count int }

func (d *testProvider) Enabled() bool {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	return true
}
func (d *testProvider) LazyProvide() *DockerConfigEntry {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	return nil
}
func (d *testProvider) Provide() DockerConfig {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	d.Count += 1
	return DockerConfig{}
}
func TestLazyKeyring(t *testing.T) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	provider := &testProvider{Count: 0}
	lazy := &lazyDockerKeyring{Providers: []DockerConfigProvider{provider}}
	if provider.Count != 0 {
		t.Errorf("Unexpected number of Provide calls: %v", provider.Count)
	}
	lazy.Lookup("foo")
	if provider.Count != 1 {
		t.Errorf("Unexpected number of Provide calls: %v", provider.Count)
	}
	lazy.Lookup("foo")
	if provider.Count != 2 {
		t.Errorf("Unexpected number of Provide calls: %v", provider.Count)
	}
	lazy.Lookup("foo")
	if provider.Count != 3 {
		t.Errorf("Unexpected number of Provide calls: %v", provider.Count)
	}
}
func TestDockerKeyringLookup(t *testing.T) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	ada := LazyAuthConfiguration{AuthConfig: dockertypes.AuthConfig{Username: "ada", Password: "smash", Email: "ada@example.com"}}
	grace := LazyAuthConfiguration{AuthConfig: dockertypes.AuthConfig{Username: "grace", Password: "squash", Email: "grace@example.com"}}
	dk := &BasicDockerKeyring{}
	dk.Add(DockerConfig{"bar.example.com/pong": DockerConfigEntry{Username: grace.Username, Password: grace.Password, Email: grace.Email}, "bar.example.com": DockerConfigEntry{Username: ada.Username, Password: ada.Password, Email: ada.Email}})
	tests := []struct {
		image	string
		match	[]LazyAuthConfiguration
		ok	bool
	}{{"bar.example.com", []LazyAuthConfiguration{ada}, true}, {"bar.example.com/pong", []LazyAuthConfiguration{grace, ada}, true}, {"bar.example.com/ping", []LazyAuthConfiguration{ada}, true}, {"bar.example.com/pongz", []LazyAuthConfiguration{grace, ada}, true}, {"bar.example.com/pong/pang", []LazyAuthConfiguration{grace, ada}, true}, {"example.com", []LazyAuthConfiguration{}, false}, {"foo.example.com", []LazyAuthConfiguration{}, false}}
	for i, tt := range tests {
		match, ok := dk.Lookup(tt.image)
		if tt.ok != ok {
			t.Errorf("case %d: expected ok=%t, got %t", i, tt.ok, ok)
		}
		if !reflect.DeepEqual(tt.match, match) {
			t.Errorf("case %d: expected match=%#v, got %#v", i, tt.match, match)
		}
	}
}
func TestIssue3797(t *testing.T) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	rex := LazyAuthConfiguration{AuthConfig: dockertypes.AuthConfig{Username: "rex", Password: "tiny arms", Email: "rex@example.com"}}
	dk := &BasicDockerKeyring{}
	dk.Add(DockerConfig{"https://quay.io/v1/": DockerConfigEntry{Username: rex.Username, Password: rex.Password, Email: rex.Email}})
	tests := []struct {
		image	string
		match	[]LazyAuthConfiguration
		ok	bool
	}{{"quay.io", []LazyAuthConfiguration{rex}, true}, {"quay.io/foo", []LazyAuthConfiguration{rex}, true}, {"quay.io/foo/bar", []LazyAuthConfiguration{rex}, true}}
	for i, tt := range tests {
		match, ok := dk.Lookup(tt.image)
		if tt.ok != ok {
			t.Errorf("case %d: expected ok=%t, got %t", i, tt.ok, ok)
		}
		if !reflect.DeepEqual(tt.match, match) {
			t.Errorf("case %d: expected match=%#v, got %#v", i, tt.match, match)
		}
	}
}
