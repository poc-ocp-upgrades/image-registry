package server

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
	"github.com/docker/distribution"
	dcontext "github.com/docker/distribution/context"
	"github.com/opencontainers/go-digest"
	"github.com/openshift/image-registry/pkg/dockerregistry/server/maxconnections"
)

type pullthroughBlobStore struct {
	distribution.BlobStore
	remoteBlobGetter	BlobGetterService
	writeLimiter		maxconnections.Limiter
	mirror				bool
	newLocalBlobStore	func(ctx context.Context) distribution.BlobStore
}

var _ distribution.BlobStore = &pullthroughBlobStore{}

func (pbs *pullthroughBlobStore) Stat(ctx context.Context, dgst digest.Digest) (distribution.Descriptor, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	dcontext.GetLogger(ctx).Debugf("(*pullthroughBlobStore).Stat: starting with dgst=%s", dgst.String())
	desc, err := pbs.BlobStore.Stat(ctx, dgst)
	switch {
	case err == distribution.ErrBlobUnknown:
	case err != nil:
		dcontext.GetLogger(ctx).Errorf("unable to find blob %q: %#v", dgst.String(), err)
		fallthrough
	default:
		return desc, err
	}
	return pbs.remoteBlobGetter.Stat(ctx, dgst)
}
func (pbs *pullthroughBlobStore) ServeBlob(ctx context.Context, w http.ResponseWriter, req *http.Request, dgst digest.Digest) error {
	_logClusterCodePath()
	defer _logClusterCodePath()
	dcontext.GetLogger(ctx).Debugf("(*pullthroughBlobStore).ServeBlob: starting with dgst=%s", dgst.String())
	err := pbs.BlobStore.ServeBlob(ctx, w, req, dgst)
	switch {
	case err == distribution.ErrBlobUnknown:
	case err != nil:
		dcontext.GetLogger(ctx).Errorf("unable to serve blob %q: %#v", dgst.String(), err)
		fallthrough
	default:
		return err
	}
	if pbs.mirror {
		mu.Lock()
		if _, ok := inflight[dgst]; ok {
			mu.Unlock()
			dcontext.GetLogger(ctx).Infof("Serving %q while mirroring in background", dgst)
			_, err := copyContent(ctx, pbs.remoteBlobGetter, dgst, w, req)
			return err
		}
		inflight[dgst] = struct{}{}
		mu.Unlock()
		pbs.storeLocalInBackground(ctx, dgst)
	}
	_, err = copyContent(ctx, pbs.remoteBlobGetter, dgst, w, req)
	return err
}
func (pbs *pullthroughBlobStore) Get(ctx context.Context, dgst digest.Digest) ([]byte, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	dcontext.GetLogger(ctx).Debugf("(*pullthroughBlobStore).Get: starting with dgst=%s", dgst.String())
	data, originalErr := pbs.BlobStore.Get(ctx, dgst)
	if originalErr == nil {
		return data, nil
	}
	return pbs.remoteBlobGetter.Get(ctx, dgst)
}
func setResponseHeaders(w http.ResponseWriter, length int64, mediaType string, digest digest.Digest) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	w.Header().Set("Content-Type", mediaType)
	w.Header().Set("Docker-Content-Digest", digest.String())
	w.Header().Set("Etag", digest.String())
}
func serveRemoteContent(rw http.ResponseWriter, req *http.Request, desc distribution.Descriptor, remoteReader io.ReadSeeker) (bool, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	setResponseHeaders(rw, desc.Size, desc.MediaType, desc.Digest)
	if req == nil {
		return false, nil
	}
	if _, err := remoteReader.Seek(0, io.SeekEnd); err != nil {
		return false, nil
	}
	if _, err := remoteReader.Seek(0, io.SeekStart); err != nil {
		return false, err
	}
	http.ServeContent(rw, req, "", time.Time{}, remoteReader)
	return true, nil
}

var inflight = make(map[digest.Digest]struct{})
var mu sync.Mutex

func copyContent(ctx context.Context, store BlobGetterService, dgst digest.Digest, writer io.Writer, req *http.Request) (distribution.Descriptor, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	desc, err := store.Stat(ctx, dgst)
	if err != nil {
		return distribution.Descriptor{}, err
	}
	remoteReader, err := store.Open(ctx, dgst)
	if err != nil {
		return distribution.Descriptor{}, err
	}
	rw, ok := writer.(http.ResponseWriter)
	if ok {
		contentHandled, err := serveRemoteContent(rw, req, desc, remoteReader)
		if err != nil {
			return distribution.Descriptor{}, err
		}
		if contentHandled {
			return desc, nil
		}
		rw.Header().Set("Content-Length", fmt.Sprintf("%d", desc.Size))
	}
	if _, err = io.CopyN(writer, remoteReader, desc.Size); err != nil {
		return distribution.Descriptor{}, err
	}
	return desc, nil
}
func (pbs *pullthroughBlobStore) storeLocalInBackground(ctx context.Context, dgst digest.Digest) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	newCtx := dcontext.WithLogger(context.Background(), dcontext.GetLogger(ctx))
	localBlobStore := pbs.newLocalBlobStore(newCtx)
	writeLimiter := pbs.writeLimiter
	remoteGetter := pbs.remoteBlobGetter
	go func(dgst digest.Digest) {
		if writeLimiter != nil {
			if !writeLimiter.Start(newCtx) {
				dcontext.GetLogger(newCtx).Infof("Skipped background mirroring of %q because write limits are reached", dgst)
				return
			}
			defer writeLimiter.Done()
		}
		dcontext.GetLogger(newCtx).Infof("Start background mirroring of %q", dgst)
		if err := storeLocal(newCtx, localBlobStore, remoteGetter, dgst); err != nil {
			dcontext.GetLogger(newCtx).Errorf("Background mirroring failed: error committing to storage: %v", err.Error())
			return
		}
		dcontext.GetLogger(newCtx).Infof("Completed mirroring of %q", dgst)
	}(dgst)
}
func storeLocal(ctx context.Context, localBlobStore distribution.BlobStore, remoteGetter BlobGetterService, dgst digest.Digest) (err error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	defer func() {
		mu.Lock()
		delete(inflight, dgst)
		mu.Unlock()
	}()
	var bw distribution.BlobWriter
	bw, err = localBlobStore.Create(ctx)
	if err != nil {
		return err
	}
	defer func() {
		_ = bw.Cancel(ctx)
	}()
	var desc distribution.Descriptor
	desc, err = copyContent(ctx, remoteGetter, dgst, bw, nil)
	if err != nil {
		return err
	}
	_, err = bw.Commit(ctx, desc)
	return
}
