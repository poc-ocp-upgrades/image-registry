package credentialprovider

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestReadDockerConfigFile(t *testing.T) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	configJsonFileName := "config.json"
	var fileInfo *os.File
	inputDockerconfigJsonFile := "{ \"auths\": { \"http://foo.example.com\":{\"auth\":\"Zm9vOmJhcgo=\",\"email\":\"foo@example.com\"}}}"
	preferredPath, err := ioutil.TempDir("", "test_foo_bar_dockerconfigjson_")
	if err != nil {
		t.Fatalf("Creating tmp dir fail: %v", err)
		return
	}
	defer os.RemoveAll(preferredPath)
	absDockerConfigFileLocation, err := filepath.Abs(filepath.Join(preferredPath, configJsonFileName))
	if err != nil {
		t.Fatalf("While trying to canonicalize %s: %v", preferredPath, err)
	}
	if _, err := os.Stat(absDockerConfigFileLocation); os.IsNotExist(err) {
		fileInfo, err = os.OpenFile(absDockerConfigFileLocation, os.O_CREATE|os.O_RDWR, 0664)
		if err != nil {
			t.Fatalf("While trying to create file %s: %v", absDockerConfigFileLocation, err)
		}
		defer fileInfo.Close()
	}
	fileInfo.WriteString(inputDockerconfigJsonFile)
	orgPreferredPath := GetPreferredDockercfgPath()
	SetPreferredDockercfgPath(preferredPath)
	defer SetPreferredDockercfgPath(orgPreferredPath)
	if _, err := ReadDockerConfigFile(); err != nil {
		t.Errorf("Getting docker config file fail : %v preferredPath : %q", err, preferredPath)
	}
}
func TestDockerConfigJsonJSONDecode(t *testing.T) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	input := []byte(`{"auths": {"http://foo.example.com":{"username": "foo", "password": "bar", "email": "foo@example.com"}, "http://bar.example.com":{"username": "bar", "password": "baz", "email": "bar@example.com"}}}`)
	expect := DockerConfigJson{Auths: DockerConfig(map[string]DockerConfigEntry{"http://foo.example.com": {Username: "foo", Password: "bar", Email: "foo@example.com"}, "http://bar.example.com": {Username: "bar", Password: "baz", Email: "bar@example.com"}})}
	var output DockerConfigJson
	err := json.Unmarshal(input, &output)
	if err != nil {
		t.Errorf("Received unexpected error: %v", err)
	}
	if !reflect.DeepEqual(expect, output) {
		t.Errorf("Received unexpected output. Expected %#v, got %#v", expect, output)
	}
}
func TestDockerConfigJSONDecode(t *testing.T) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	input := []byte(`{"http://foo.example.com":{"username": "foo", "password": "bar", "email": "foo@example.com"}, "http://bar.example.com":{"username": "bar", "password": "baz", "email": "bar@example.com"}}`)
	expect := DockerConfig(map[string]DockerConfigEntry{"http://foo.example.com": {Username: "foo", Password: "bar", Email: "foo@example.com"}, "http://bar.example.com": {Username: "bar", Password: "baz", Email: "bar@example.com"}})
	var output DockerConfig
	err := json.Unmarshal(input, &output)
	if err != nil {
		t.Errorf("Received unexpected error: %v", err)
	}
	if !reflect.DeepEqual(expect, output) {
		t.Errorf("Received unexpected output. Expected %#v, got %#v", expect, output)
	}
}
func TestDockerConfigEntryJSONDecode(t *testing.T) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	tests := []struct {
		input	[]byte
		expect	DockerConfigEntry
		fail	bool
	}{{input: []byte(`{"username": "foo", "password": "bar", "email": "foo@example.com"}`), expect: DockerConfigEntry{Username: "foo", Password: "bar", Email: "foo@example.com"}, fail: false}, {input: []byte(`{"auth": "Zm9vOmJhcg==", "email": "foo@example.com"}`), expect: DockerConfigEntry{Username: "foo", Password: "bar", Email: "foo@example.com"}, fail: false}, {input: []byte(`{"username": "foo", "password": "bar", "auth": "cGluZzpwb25n", "email": "foo@example.com"}`), expect: DockerConfigEntry{Username: "ping", Password: "pong", Email: "foo@example.com"}, fail: false}, {input: []byte(`{"auth": "pants", "email": "foo@example.com"}`), expect: DockerConfigEntry{Username: "", Password: "", Email: "foo@example.com"}, fail: true}, {input: []byte(`{"email": false}`), expect: DockerConfigEntry{Username: "", Password: "", Email: ""}, fail: true}}
	for i, tt := range tests {
		var output DockerConfigEntry
		err := json.Unmarshal(tt.input, &output)
		if (err != nil) != tt.fail {
			t.Errorf("case %d: expected fail=%t, got err=%v", i, tt.fail, err)
		}
		if !reflect.DeepEqual(tt.expect, output) {
			t.Errorf("case %d: expected output %#v, got %#v", i, tt.expect, output)
		}
	}
}
func TestDecodeDockerConfigFieldAuth(t *testing.T) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	tests := []struct {
		input		string
		username	string
		password	string
		fail		bool
	}{{input: "Zm9vOmJhcg==", username: "foo", password: "bar"}, {input: "cGFudHM=", fail: true}, {input: "pants", fail: true}}
	for i, tt := range tests {
		username, password, err := decodeDockerConfigFieldAuth(tt.input)
		if (err != nil) != tt.fail {
			t.Errorf("case %d: expected fail=%t, got err=%v", i, tt.fail, err)
		}
		if tt.username != username {
			t.Errorf("case %d: expected username %q, got %q", i, tt.username, username)
		}
		if tt.password != password {
			t.Errorf("case %d: expected password %q, got %q", i, tt.password, password)
		}
	}
}
func TestDockerConfigEntryJSONCompatibleEncode(t *testing.T) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	tests := []struct {
		input	DockerConfigEntry
		expect	[]byte
	}{{expect: []byte(`{"username":"foo","password":"bar","email":"foo@example.com","auth":"Zm9vOmJhcg=="}`), input: DockerConfigEntry{Username: "foo", Password: "bar", Email: "foo@example.com"}}}
	for i, tt := range tests {
		actual, err := json.Marshal(tt.input)
		if err != nil {
			t.Errorf("case %d: unexpected error: %v", i, err)
		}
		if string(tt.expect) != string(actual) {
			t.Errorf("case %d: expected %v, got %v", i, string(tt.expect), string(actual))
		}
	}
}
