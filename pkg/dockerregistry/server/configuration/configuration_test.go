package configuration

import (
	"bytes"
	"os"
	"reflect"
	"strings"
	"testing"
	"github.com/docker/distribution/configuration"
)

var configYamlV0_1 = `
version: 0.1
http:
  addr: :5000
  relativeurls: true
storage:
  inmemory: {}
openshift:
  version: 1.0
  server:
    addr: :5000
  metrics:
    enabled: true
    secret: TopSecretToken
  auth:
    realm: myrealm
  audit:
    enabled: true
  pullthrough:
    enabled: true
    mirror: true
`

func TestConfigurationParser(t *testing.T) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	configFile := bytes.NewBufferString(configYamlV0_1)
	dockerConfig, extraConfig, err := Parse(configFile)
	if err != nil {
		t.Fatalf("unexpected error parsing configuration file: %s", err)
	}
	if !dockerConfig.HTTP.RelativeURLs {
		t.Fatalf("unexpected value: dockerConfig.HTTP.RelativeURLs != true")
	}
	if extraConfig.Version.Major() != 1 || extraConfig.Version.Minor() != 0 {
		t.Fatalf("unexpected value: extraConfig.Version: %s", extraConfig.Version)
	}
	if !extraConfig.Metrics.Enabled {
		t.Fatalf("unexpected value: extraConfig.Metrics.Enabled != true")
	}
	if extraConfig.Metrics.Secret != "TopSecretToken" {
		t.Fatalf("unexpected value: extraConfig.Metrics.Secret: %s", extraConfig.Metrics.Secret)
	}
	if extraConfig.Auth == nil {
		t.Fatalf("unexpected empty section: extraConfig.Auth")
	} else if extraConfig.Auth.Realm != "myrealm" {
		t.Fatalf("unexpected value: extraConfig.Auth.Realm: %s", extraConfig.Auth.Realm)
	}
	if extraConfig.Audit == nil {
		t.Fatalf("unexpected empty section: extraConfig.Audit")
	} else if !extraConfig.Audit.Enabled {
		t.Fatalf("unexpected value: extraConfig.Audit.Enabled != true")
	}
	if extraConfig.Pullthrough == nil {
		t.Fatalf("unexpected empty section: extraConfig.Pullthrough")
	} else {
		if !extraConfig.Pullthrough.Enabled {
			t.Fatalf("unexpected value: extraConfig.Pullthrough.Enabled != true")
		}
		if !extraConfig.Pullthrough.Mirror {
			t.Fatalf("unexpected value: extraConfig.Pullthrough.Mirror != true")
		}
	}
}
func testConfigurationOverwriteEnv(t *testing.T, config string) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	os.Setenv("REGISTRY_OPENSHIFT_SERVER_ADDR", ":5000")
	defer os.Unsetenv("REGISTRY_OPENSHIFT_SERVER_ADDR")
	os.Setenv("REGISTRY_OPENSHIFT_METRICS_ENABLED", "false")
	defer os.Unsetenv("REGISTRY_OPENSHIFT_METRICS_ENABLED")
	configFile := bytes.NewBufferString(config)
	_, extraConfig, err := Parse(configFile)
	if err != nil {
		t.Fatalf("unexpected error parsing configuration file: %s", err)
	}
	if extraConfig.Metrics.Enabled {
		t.Fatalf("unexpected value: extraConfig.Metrics.Enabled != false")
	}
	if extraConfig.Server == nil {
		t.Fatalf("unexpected empty section extraConfig.Server")
	} else if extraConfig.Server.Addr != ":5000" {
		t.Fatalf("unexpected value: extraConfig.Server.Addr: %s", extraConfig.Server.Addr)
	}
}
func TestConfigurationOverwriteEnv(t *testing.T) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	var configYaml = `
version: 0.1
storage:
  inmemory: {}
openshift:
  version: 1.0
  server:
    addr: :5000
  metrics:
    enabled: true
    secret: TopSecretToken
`
	testConfigurationOverwriteEnv(t, configYaml)
}
func TestConfigurationWithEmptyOpenshiftOverwriteEnv(t *testing.T) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	var configYaml = `
version: 0.1
storage:
  inmemory: {}
`
	testConfigurationOverwriteEnv(t, configYaml)
}
func TestDockerConfigurationError(t *testing.T) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	var badDockerConfigYamlV0_1 = `
version: 0.1
http:
  addr: :5000
  relativeurls: "true"
storage:
  inmemory: {}
`
	configFile := bytes.NewBufferString(badDockerConfigYamlV0_1)
	_, _, err := Parse(configFile)
	if err == nil {
		t.Fatalf("unexpected parser success")
	}
}
func TestExtraConfigurationError(t *testing.T) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	var badExtraConfigYaml = `
version: 0.1
http:
  addr: :5000
storage:
  inmemory: {}
openshift:
  version: 1.0
  metrics:
    enabled: "true"
`
	configFile := bytes.NewBufferString(badExtraConfigYaml)
	_, _, err := Parse(configFile)
	if err == nil {
		t.Fatalf("unexpected parser success")
	}
}
func TestEmptyExtraConfigurationError(t *testing.T) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	var emptyExtraConfigYaml = `
version: 0.1
http:
  addr: :5000
storage:
  inmemory: {}
`
	os.Setenv("REGISTRY_OPENSHIFT_SERVER_ADDR", ":5000")
	defer os.Unsetenv("REGISTRY_OPENSHIFT_SERVER_ADDR")
	configFile := bytes.NewBufferString(emptyExtraConfigYaml)
	_, _, err := Parse(configFile)
	if err != nil {
		t.Fatalf("unexpected parser error: %s", err)
	}
}
func TestExtraConfigurationVersionError(t *testing.T) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	var badExtraConfigYaml = `
version: 0.1
http:
  addr: :5000
storage:
  inmemory: {}
openshift:
  version: 2.0
`
	configFile := bytes.NewBufferString(badExtraConfigYaml)
	_, _, err := Parse(configFile)
	if err == nil {
		t.Fatalf("unexpected parser success")
	}
	if err != ErrUnsupportedVersion {
		t.Fatalf("unexpected parser error: %v", err)
	}
}
func TestDefaultMiddleware(t *testing.T) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	checks := []struct{ title, input, expect string }{{title: "miss all middlewares", input: `
version: 0.1
storage:
  inmemory: {}
`, expect: `
version: 0.1
storage:
  inmemory: {}
middleware:
  registry:
    - name: openshift
  repository:
    - name: openshift
  storage:
    - name: openshift
`}, {title: "miss some middlewares", input: `
version: 0.1
storage:
  inmemory: {}
middleware:
  registry:
    - name: openshift
`, expect: `
version: 0.1
storage:
  inmemory: {}
middleware:
  registry:
    - name: openshift
  repository:
    - name: openshift
  storage:
    - name: openshift
`}, {title: "all middlewares are in place", input: `
version: 0.1
storage:
  inmemory: {}
middleware:
  registry:
    - name: openshift
  repository:
    - name: openshift
  storage:
    - name: openshift
`, expect: `
version: 0.1
storage:
  inmemory: {}
middleware:
  registry:
    - name: openshift
  repository:
    - name: openshift
  storage:
    - name: openshift
`}, {title: "check v1.0.8 config", input: `
version: 0.1
log:
  level: debug
http:
  addr: :5000
storage:
  cache:
    layerinfo: inmemory
  filesystem:
    rootdirectory: /registry
auth:
  openshift:
    realm: openshift
middleware:
  repository:
   - name: openshift
`, expect: `
version: 0.1
log:
  level: debug
http:
  addr: :5000
storage:
  cache:
    layerinfo: inmemory
  filesystem:
    rootdirectory: /registry
auth:
  openshift:
    realm: openshift
middleware:
  registry:
    - name: openshift
  repository:
    - name: openshift
  storage:
    - name: openshift
`}, {title: "check v1.2.1 config", input: `
version: 0.1
log:
  level: debug
http:
  addr: :5000
storage:
  cache:
    layerinfo: inmemory
  filesystem:
    rootdirectory: /registry
  delete:
    enabled: true
auth:
  openshift:
    realm: openshift
middleware:
  repository:
    - name: openshift
      options:
        pullthrough: true
`, expect: `
version: 0.1
log:
  level: debug
http:
  addr: :5000
storage:
  cache:
    layerinfo: inmemory
  filesystem:
    rootdirectory: /registry
  delete:
    enabled: true
auth:
  openshift:
    realm: openshift
middleware:
  registry:
    - name: openshift
  repository:
    - name: openshift
      options:
        pullthrough: true
  storage:
    - name: openshift
`}, {title: "check v1.3.0-alpha.3 config", input: `
version: 0.1
log:
  level: debug
http:
  addr: :5000
storage:
  cache:
    blobdescriptor: inmemory
  filesystem:
    rootdirectory: /registry
  delete:
    enabled: true
auth:
  openshift:
    realm: openshift
middleware:
  registry:
    - name: openshift
  repository:
    - name: openshift
      options:
        acceptschema2: false
        pullthrough: true
        enforcequota: false
        projectcachettl: 1m
        blobrepositorycachettl: 10m
  storage:
    - name: openshift
`, expect: `
version: 0.1
log:
  level: debug
http:
  addr: :5000
storage:
  cache:
    blobdescriptor: inmemory
  filesystem:
    rootdirectory: /registry
  delete:
    enabled: true
auth:
  openshift:
    realm: openshift
middleware:
  registry:
    - name: openshift
  repository:
    - name: openshift
      options:
        acceptschema2: false
        pullthrough: true
        enforcequota: false
        projectcachettl: 1m
        blobrepositorycachettl: 10m
  storage:
    - name: openshift
`}}
	for _, check := range checks {
		currentConfig, err := configuration.Parse(strings.NewReader(check.input))
		if err != nil {
			t.Fatal(err)
		}
		expectConfig, err := configuration.Parse(strings.NewReader(check.expect))
		if err != nil {
			t.Fatal(err)
		}
		setDefaultMiddleware(currentConfig)
		if !reflect.DeepEqual(currentConfig, expectConfig) {
			t.Errorf("%s: expected\n\t%#v\ngot\n\t%#v", check.title, expectConfig, currentConfig)
		}
	}
}
func TestMiddlewareMigration(t *testing.T) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	var inputConfigYaml = `
version: 0.1
log:
  level: debug
http:
  addr: :5000
storage:
  cache:
    blobdescriptor: inmemory
  filesystem:
    rootdirectory: /registry
  delete:
    enabled: true
auth:
  openshift:
    realm: openshift
middleware:
  registry:
    - name: openshift
  repository:
    - name: openshift
      options:
        acceptschema2: true
        pullthrough: true
        enforcequota: false
        projectcachettl: 1m
        blobrepositorycachettl: 10m
  storage:
    - name: openshift
openshift:
  version: 1.0
  server:
    addr: :5000
`
	var expectConfigYaml = `
version: 0.1
log:
  level: debug
http:
  addr: :5000
storage:
  cache:
    blobdescriptor: inmemory
  filesystem:
    rootdirectory: /registry
  delete:
    enabled: true
auth:
  openshift
middleware:
  registry:
    - name: openshift
  repository:
    - name: openshift
  storage:
    - name: openshift
openshift:
  version: 1.0
  server:
    addr: :5000
  auth:
    realm: openshift
  quota:
    enabled: false
    cachettl: 1m
  pullthrough:
    enabled: true
    mirror: true
  cache:
    blobrepositoryttl: 10m
  compatibility:
    acceptschema2: true
`
	_, currentConfig, err := Parse(strings.NewReader(inputConfigYaml))
	if err != nil {
		t.Fatal(err)
	}
	_, expectConfig, err := Parse(strings.NewReader(expectConfigYaml))
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(currentConfig.Server, expectConfig.Server) {
		t.Fatalf("expected server section\n\t%#v\ngot\n\t%#v", expectConfig.Server, currentConfig.Server)
	}
	if !reflect.DeepEqual(currentConfig.Auth, expectConfig.Auth) {
		t.Fatalf("expected auth section\n\t%#v\ngot\n\t%#v", expectConfig.Auth, currentConfig.Auth)
	}
	if !reflect.DeepEqual(currentConfig.Audit, expectConfig.Audit) {
		t.Fatalf("expected audit section\n\t%#v\ngot\n\t%#v", expectConfig.Audit, currentConfig.Audit)
	}
	if !reflect.DeepEqual(currentConfig.Quota, expectConfig.Quota) {
		t.Fatalf("expected quota section\n\t%#v\ngot\n\t%#v", expectConfig.Quota, currentConfig.Quota)
	}
	if !reflect.DeepEqual(currentConfig.Pullthrough, expectConfig.Pullthrough) {
		t.Fatalf("expected pullthrough section\n\t%#v\ngot\n\t%#v", expectConfig.Pullthrough, currentConfig.Pullthrough)
	}
	if !reflect.DeepEqual(currentConfig.Cache, expectConfig.Cache) {
		t.Fatalf("expected cache section\n\t%#v\ngot\n\t%#v", expectConfig.Cache, currentConfig.Cache)
	}
	if !reflect.DeepEqual(currentConfig.Compatibility, expectConfig.Compatibility) {
		t.Fatalf("expected compatibility section\n\t%#v\ngot\n\t%#v", expectConfig.Compatibility, currentConfig.Compatibility)
	}
}
func TestServerAddrEnvOrder(t *testing.T) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	var configYaml = `
version: 0.1
http:
  addr: :5000
storage:
  filesystem:
    rootdirectory: /registry
openshift:
  version: 1.0
`
	type env struct{ name, value string }
	testCases := []struct {
		setenv		[]env
		expected	string
	}{{setenv: []env{{name: "DOCKER_REGISTRY_SERVICE_HOST", value: "DOCKER_REGISTRY_SERVICE_HOST"}, {name: "DOCKER_REGISTRY_SERVICE_PORT", value: "DOCKER_REGISTRY_SERVICE_PORT"}}, expected: "DOCKER_REGISTRY_SERVICE_HOST:DOCKER_REGISTRY_SERVICE_PORT"}, {setenv: []env{{name: "REGISTRY_OPENSHIFT_SERVER_ADDR", value: "REGISTRY_OPENSHIFT_SERVER_ADDR"}}, expected: "REGISTRY_OPENSHIFT_SERVER_ADDR"}, {setenv: []env{{name: "REGISTRY_MIDDLEWARE_REPOSITORY_OPENSHIFT_DOCKERREGISTRYURL", value: "REGISTRY_MIDDLEWARE_REPOSITORY_OPENSHIFT_DOCKERREGISTRYURL"}}, expected: "REGISTRY_MIDDLEWARE_REPOSITORY_OPENSHIFT_DOCKERREGISTRYURL"}, {setenv: []env{{name: "DOCKER_REGISTRY_URL", value: "DOCKER_REGISTRY_URL"}}, expected: "DOCKER_REGISTRY_URL"}, {setenv: []env{{name: "OPENSHIFT_DEFAULT_REGISTRY", value: "OPENSHIFT_DEFAULT_REGISTRY"}}, expected: "OPENSHIFT_DEFAULT_REGISTRY"}}
	for _, test := range testCases {
		for _, env := range test.setenv {
			os.Setenv(env.name, env.value)
			defer os.Unsetenv(env.name)
		}
		_, cfg, err := Parse(strings.NewReader(configYaml))
		if err != nil {
			t.Fatal(err)
		}
		if cfg.Server.Addr != test.expected {
			t.Fatalf("unexpected value: cfg.Server.Addr != %s, got %s", test.expected, cfg.Server.Addr)
		}
	}
}
func TestServerAddrConfigPriority(t *testing.T) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	var configYaml = `
version: 0.1
http:
  addr: :5000
storage:
  filesystem:
    rootdirectory: /registry
openshift:
  version: 1.0
  server:
    addr: from-config
`
	type env struct{ name, value string }
	testCases := []struct {
		setenv		[]env
		expected	string
	}{{setenv: []env{}, expected: "from-config"}, {setenv: []env{{name: "DOCKER_REGISTRY_SERVICE_HOST", value: "DOCKER_REGISTRY_SERVICE_HOST"}, {name: "DOCKER_REGISTRY_SERVICE_PORT", value: "DOCKER_REGISTRY_SERVICE_PORT"}}, expected: "from-config"}, {setenv: []env{{name: "REGISTRY_OPENSHIFT_SERVER_ADDR", value: "REGISTRY_OPENSHIFT_SERVER_ADDR"}}, expected: "REGISTRY_OPENSHIFT_SERVER_ADDR"}, {setenv: []env{{name: "REGISTRY_MIDDLEWARE_REPOSITORY_OPENSHIFT_DOCKERREGISTRYURL", value: "REGISTRY_MIDDLEWARE_REPOSITORY_OPENSHIFT_DOCKERREGISTRYURL"}}, expected: "REGISTRY_MIDDLEWARE_REPOSITORY_OPENSHIFT_DOCKERREGISTRYURL"}, {setenv: []env{{name: "DOCKER_REGISTRY_URL", value: "DOCKER_REGISTRY_URL"}}, expected: "DOCKER_REGISTRY_URL"}, {setenv: []env{{name: "OPENSHIFT_DEFAULT_REGISTRY", value: "OPENSHIFT_DEFAULT_REGISTRY"}}, expected: "OPENSHIFT_DEFAULT_REGISTRY"}}
	for _, test := range testCases {
		for _, env := range test.setenv {
			os.Setenv(env.name, env.value)
			defer os.Unsetenv(env.name)
		}
		_, cfg, err := Parse(strings.NewReader(configYaml))
		if err != nil {
			t.Fatal(err)
		}
		if cfg.Server.Addr != test.expected {
			t.Fatalf("unexpected value: cfg.Server.Addr != %s, got %s", test.expected, cfg.Server.Addr)
		}
	}
}
func TestAudit(t *testing.T) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	var configYaml = `
version: 0.1
http:
  addr: :5000
storage:
  filesystem:
    rootdirectory: /registry
auth:
  openshift:
    audit:
      enabled: true
    realm: openshift
    tokenrealm: https://example.com:5000
openshift:
  version: 1.0
  server:
    addr: "localhost:5000"
`
	_, cfg, err := Parse(strings.NewReader(configYaml))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Audit == nil {
		t.Fatalf("unexpected value: extraConfig.Audit == nil")
	}
	if !cfg.Audit.Enabled {
		t.Fatalf("unexpected value: extraConfig.Audit.Enabled != true")
	}
}
func testDisableInmemoryCacheName(t *testing.T, field string) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	var configYaml = `
version: 0.1
http:
  addr: :5000
storage:
  filesystem:
    rootdirectory: /registry
  cache:
    ` + field + `: inmemory
openshift:
  version: 1.0
  server:
    addr: "localhost:5000"
`
	dockercfg, _, err := Parse(strings.NewReader(configYaml))
	if err != nil {
		t.Fatal(err)
	}
	_, ok := dockercfg.Storage["cache"]
	if ok {
		t.Fatalf("unexpected value: dockerConfig.Storage['cache'] != nil")
	}
}
func testDisableInmemoryCache(t *testing.T) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	testDisableInmemoryCacheName(t, "layerinfo")
	testDisableInmemoryCacheName(t, "blobdescriptor")
}
func testPreserveRedisCacheName(t *testing.T, field string) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	var configYaml = `
version: 0.1
http:
  addr: :5000
storage:
  filesystem:
    rootdirectory: /registry
  cache:
    ` + field + `: redis
openshift:
  version: 1.0
  server:
    addr: "localhost:5000"
`
	dockercfg, _, err := Parse(strings.NewReader(configYaml))
	if err != nil {
		t.Fatal(err)
	}
	cc, ok := dockercfg.Storage["cache"]
	if !ok {
		t.Fatalf("unexpected value: dockerConfig.Storage['cache'] == nil")
	}
	v, ok := cc[field]
	if !ok {
		t.Fatalf("unexpected value: dockerConfig.Storage['cache']['%s'] == nil", field)
	}
	if v != "redis" {
		t.Fatalf("unexpected value: dockerConfig.Storage['cache']['%s'] != redis", field)
	}
}
func TestPreserveRedisCache(t *testing.T) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	testPreserveRedisCacheName(t, "layerinfo")
	testPreserveRedisCacheName(t, "blobdescriptor")
}
func TestDisabledMiddleware(t *testing.T) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	var inputConfigYaml = `
version: 0.1
storage:
  inmemory: {}
middleware:
  registry:
    - name: openshift
  repository:
    - name: openshift
      disabled: true
  storage:
    - name: openshift
openshift:
  version: 1.0
  server:
    addr: "localhost:5000"
`
	var expectConfigYaml = `
version: 0.1
storage:
  inmemory: {}
middleware:
  registry:
    - name: openshift
  repository:
    - name: openshift
  storage:
    - name: openshift
openshift:
  version: 1.0
  server:
    addr: "localhost:5000"
`
	_, currentConfig, err := Parse(strings.NewReader(inputConfigYaml))
	if err != nil {
		t.Fatal(err)
	}
	_, expectConfig, err := Parse(strings.NewReader(expectConfigYaml))
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(currentConfig, expectConfig) {
		t.Fatalf("expected configuration\n\t%#v\ngot\n\t%#v", expectConfig, currentConfig)
	}
}
