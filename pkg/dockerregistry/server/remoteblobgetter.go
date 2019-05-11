package server

import (
	"context"
	"net/http"
	"sync"
	"github.com/docker/distribution"
	dcontext "github.com/docker/distribution/context"
	"github.com/opencontainers/go-digest"
	corev1 "k8s.io/api/core/v1"
	"github.com/openshift/image-registry/pkg/dockerregistry/server/cache"
	"github.com/openshift/image-registry/pkg/dockerregistry/server/metrics"
	rerrors "github.com/openshift/image-registry/pkg/errors"
	"github.com/openshift/image-registry/pkg/imagestream"
	"github.com/openshift/image-registry/pkg/origin-common/image/registryclient"
)

type BlobGetterService interface {
	distribution.BlobStatter
	distribution.BlobProvider
	distribution.BlobServer
}
type secretsGetter func() ([]corev1.Secret, *rerrors.Error)
type digestBlobStoreCache struct {
	mu		sync.RWMutex
	data	map[string]distribution.BlobStore
	metrics	metrics.Cache
}

func newDigestBlobStoreCache(m metrics.Pullthrough) *digestBlobStoreCache {
	_logClusterCodePath()
	defer _logClusterCodePath()
	return &digestBlobStoreCache{data: make(map[string]distribution.BlobStore), metrics: m.DigestBlobStoreCache()}
}
func (c *digestBlobStoreCache) Get(dgst digest.Digest) (bs distribution.BlobStore, ok bool) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	func() {
		c.mu.RLock()
		defer c.mu.RUnlock()
		bs, ok = c.data[dgst.String()]
	}()
	c.metrics.Request(ok)
	return
}
func (c *digestBlobStoreCache) Put(dgst digest.Digest, bs distribution.BlobStore) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[dgst.String()] = bs
}

type remoteBlobGetterService struct {
	imageStream		imagestream.ImageStream
	getSecrets		secretsGetter
	cache			cache.RepositoryDigest
	digestToStore	*digestBlobStoreCache
	metrics			metrics.Pullthrough
}

var _ BlobGetterService = &remoteBlobGetterService{}

func NewBlobGetterService(imageStream imagestream.ImageStream, secretsGetter secretsGetter, cache cache.RepositoryDigest, m metrics.Pullthrough) BlobGetterService {
	_logClusterCodePath()
	defer _logClusterCodePath()
	return &remoteBlobGetterService{imageStream: imageStream, getSecrets: secretsGetter, cache: cache, digestToStore: newDigestBlobStoreCache(m), metrics: m}
}
func (rbgs *remoteBlobGetterService) findBlobStore(ctx context.Context, dgst digest.Digest) (distribution.Descriptor, distribution.BlobStore, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	ok, err := rbgs.imageStream.Exists(ctx)
	if err != nil {
		switch err.Code {
		case imagestream.ErrImageStreamNotFoundCode:
			dcontext.GetLogger(ctx).Errorf("findBlobStore: imagestream %s not found: %v", rbgs.imageStream.Reference(), err)
			return distribution.Descriptor{}, nil, distribution.ErrBlobUnknown
		case imagestream.ErrImageStreamForbiddenCode:
			dcontext.GetLogger(ctx).Errorf("findBlobStore: unable get access to imagestream %s: %v", rbgs.imageStream.Reference(), err)
			return distribution.Descriptor{}, nil, distribution.ErrAccessDenied
		}
		return distribution.Descriptor{}, nil, err
	}
	if !ok {
		return distribution.Descriptor{}, nil, distribution.ErrBlobUnknown
	}
	cached := rbgs.cache.Repositories(dgst)
	retriever, impErr := getImportContext(ctx, rbgs.getSecrets, rbgs.metrics)
	if impErr != nil {
		return distribution.Descriptor{}, nil, impErr
	}
	repositoryCandidates, search, err := rbgs.imageStream.IdentifyCandidateRepositories(ctx, true)
	if err != nil {
		return distribution.Descriptor{}, nil, err
	}
	if desc, bs, err := rbgs.findCandidateRepository(ctx, repositoryCandidates, search, cached, dgst, retriever); err == nil {
		return desc, bs, nil
	}
	repositoryCandidates, secondary, err := rbgs.imageStream.IdentifyCandidateRepositories(ctx, false)
	if err != nil {
		return distribution.Descriptor{}, nil, err
	}
	for k := range search {
		delete(secondary, k)
	}
	if desc, bs, err := rbgs.findCandidateRepository(ctx, repositoryCandidates, secondary, cached, dgst, retriever); err == nil {
		return desc, bs, nil
	}
	return distribution.Descriptor{}, nil, distribution.ErrBlobUnknown
}
func (rbgs *remoteBlobGetterService) Stat(ctx context.Context, dgst digest.Digest) (distribution.Descriptor, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	dcontext.GetLogger(ctx).Debugf("(*remoteBlobGetterService).Stat: starting with dgst=%s", dgst)
	bs, ok := rbgs.digestToStore.Get(dgst)
	if ok {
		desc, err := bs.Stat(ctx, dgst)
		if err == nil {
			return desc, nil
		}
		dcontext.GetLogger(ctx).Warnf("Stat: failed to stat blob %s in cached remote repository: %v", dgst, err)
	}
	desc, bs, err := rbgs.findBlobStore(ctx, dgst)
	if err != nil {
		return desc, err
	}
	rbgs.digestToStore.Put(dgst, bs)
	return desc, nil
}
func (rbgs *remoteBlobGetterService) Open(ctx context.Context, dgst digest.Digest) (distribution.ReadSeekCloser, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	dcontext.GetLogger(ctx).Debugf("(*remoteBlobGetterService).Open: starting with dgst=%s", dgst)
	bs, ok := rbgs.digestToStore.Get(dgst)
	if !ok {
		var err error
		_, bs, err = rbgs.findBlobStore(ctx, dgst)
		if err != nil {
			return nil, err
		}
		rbgs.digestToStore.Put(dgst, bs)
	}
	return bs.Open(ctx, dgst)
}
func (rbgs *remoteBlobGetterService) ServeBlob(ctx context.Context, w http.ResponseWriter, req *http.Request, dgst digest.Digest) error {
	_logClusterCodePath()
	defer _logClusterCodePath()
	dcontext.GetLogger(ctx).Debugf("(*remoteBlobGetterService).ServeBlob: starting with dgst=%s", dgst)
	bs, ok := rbgs.digestToStore.Get(dgst)
	if !ok {
		var err error
		_, bs, err = rbgs.findBlobStore(ctx, dgst)
		if err != nil {
			return err
		}
		rbgs.digestToStore.Put(dgst, bs)
	}
	return bs.ServeBlob(ctx, w, req, dgst)
}
func (rbgs *remoteBlobGetterService) proxyStat(ctx context.Context, retriever registryclient.RepositoryRetriever, spec *imagestream.ImagePullthroughSpec, dgst digest.Digest) (distribution.Descriptor, distribution.BlobStore, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	ref := spec.DockerImageReference
	insecureNote := ""
	if spec.Insecure {
		insecureNote = " with a fall-back to insecure transport"
	}
	dcontext.GetLogger(ctx).Infof("Trying to stat %q from %q%s", dgst, ref.AsRepository().Exact(), insecureNote)
	repo, err := retriever.Repository(ctx, ref.RegistryURL(), ref.RepositoryName(), spec.Insecure)
	if err != nil {
		dcontext.GetLogger(ctx).Errorf("Error getting remote repository for image %q: %v", ref.AsRepository().Exact(), err)
		return distribution.Descriptor{}, nil, err
	}
	pullthroughBlobStore := repo.Blobs(ctx)
	desc, err := pullthroughBlobStore.Stat(ctx, dgst)
	if err != nil {
		if err != distribution.ErrBlobUnknown {
			dcontext.GetLogger(ctx).Errorf("Error statting blob %s in remote repository %q: %v", dgst, ref.AsRepository().Exact(), err)
		}
		return distribution.Descriptor{}, nil, err
	}
	return desc, pullthroughBlobStore, nil
}
func (rbgs *remoteBlobGetterService) Get(ctx context.Context, dgst digest.Digest) ([]byte, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	dcontext.GetLogger(ctx).Debugf("(*remoteBlobGetterService).Get: starting with dgst=%s", dgst.String())
	bs, ok := rbgs.digestToStore.Get(dgst)
	if !ok {
		var err error
		_, bs, err = rbgs.findBlobStore(ctx, dgst)
		if err != nil {
			return nil, err
		}
		rbgs.digestToStore.Put(dgst, bs)
	}
	return bs.Get(ctx, dgst)
}
func (rbgs *remoteBlobGetterService) findCandidateRepository(ctx context.Context, repositoryCandidates []string, search map[string]imagestream.ImagePullthroughSpec, cachedRepos []string, dgst digest.Digest, retriever registryclient.RepositoryRetriever) (distribution.Descriptor, distribution.BlobStore, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	if len(search) == 0 {
		return distribution.Descriptor{}, nil, distribution.ErrBlobUnknown
	}
	for _, repo := range cachedRepos {
		spec, ok := search[repo]
		if !ok {
			continue
		}
		desc, bs, err := rbgs.proxyStat(ctx, retriever, &spec, dgst)
		if err != nil {
			delete(search, repo)
			continue
		}
		dcontext.GetLogger(ctx).Infof("Found digest location from cache %q in %q", dgst, repo)
		return desc, bs, nil
	}
	for _, repo := range repositoryCandidates {
		spec, ok := search[repo]
		if !ok {
			continue
		}
		desc, bs, err := rbgs.proxyStat(ctx, retriever, &spec, dgst)
		if err != nil {
			continue
		}
		_ = rbgs.cache.AddDigest(dgst, repo)
		dcontext.GetLogger(ctx).Infof("Found digest location by search %q in %q", dgst, repo)
		return desc, bs, nil
	}
	return distribution.Descriptor{}, nil, distribution.ErrBlobUnknown
}
