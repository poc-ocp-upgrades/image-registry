package server

import (
	"encoding/json"
	"github.com/docker/distribution"
	"github.com/docker/distribution/manifest/schema2"
)

func unmarshalManifestSchema2(content []byte) (distribution.Manifest, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	var deserializedManifest schema2.DeserializedManifest
	if err := json.Unmarshal(content, &deserializedManifest); err != nil {
		return nil, err
	}
	return &deserializedManifest, nil
}
