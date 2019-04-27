package server

import (
	"context"
	"fmt"
	"net/http"
	"github.com/docker/distribution"
	dcontext "github.com/docker/distribution/context"
	registrystorage "github.com/docker/distribution/registry/storage"
	restclient "k8s.io/client-go/rest"
	"github.com/openshift/image-registry/pkg/dockerregistry/server/audit"
	"github.com/openshift/image-registry/pkg/dockerregistry/server/cache"
	"github.com/openshift/image-registry/pkg/imagestream"
)

var (
	secureTransport		http.RoundTripper
	insecureTransport	http.RoundTripper
)

func init() {
	_logClusterCodePath()
	defer _logClusterCodePath()
	secureTransport = http.DefaultTransport
	var err error
	insecureTransport, err = restclient.TransportFor(&restclient.Config{TLSClientConfig: restclient.TLSClientConfig{Insecure: true}})
	if err != nil {
		panic(fmt.Sprintf("Unable to configure a default transport for importing insecure images: %v", err))
	}
}

type repository struct {
	distribution.Repository
	ctx			context.Context
	app			*App
	crossmount		bool
	imageStream		imagestream.ImageStream
	remoteBlobGetter	BlobGetterService
	cache			cache.RepositoryDigest
}

func (app *App) Repository(ctx context.Context, repo distribution.Repository, crossmount bool) (distribution.Repository, distribution.BlobDescriptorServiceFactory, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	registryOSClient, err := app.registryClient.Client()
	if err != nil {
		return nil, nil, err
	}
	namespace, name, err := getNamespaceName(repo.Named().Name())
	if err != nil {
		return nil, nil, err
	}
	r := &repository{Repository: repo, ctx: ctx, app: app, crossmount: crossmount, imageStream: imagestream.New(ctx, namespace, name, registryOSClient), cache: cache.NewRepositoryDigest(app.cache)}
	r.remoteBlobGetter = NewBlobGetterService(r.imageStream, r.imageStream.GetSecrets, r.cache, r.app.metrics)
	repo = distribution.Repository(r)
	repo = r.app.metrics.Repository(repo, repo.Named().Name())
	bdsf := blobDescriptorServiceFactoryFunc(r.BlobDescriptorService)
	return repo, bdsf, nil
}
func (r *repository) Manifests(ctx context.Context, options ...distribution.ManifestServiceOption) (distribution.ManifestService, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	opts := append(options, registrystorage.SkipLayerVerification())
	ms, err := r.Repository.Manifests(ctx, opts...)
	if err != nil {
		return nil, err
	}
	ms = &manifestService{manifests: ms, blobStore: r.Blobs(ctx), serverAddr: r.app.config.Server.Addr, imageStream: r.imageStream, cache: r.cache, acceptSchema2: r.app.config.Compatibility.AcceptSchema2}
	ms = &pullthroughManifestService{ManifestService: ms, newLocalManifestService: func(ctx context.Context) (distribution.ManifestService, error) {
		return r.Repository.Manifests(ctx, opts...)
	}, imageStream: r.imageStream, cache: r.cache, mirror: r.app.config.Pullthrough.Mirror, registryAddr: r.app.config.Server.Addr, metrics: r.app.metrics}
	ms = newPendingErrorsManifestService(ms, r)
	if audit.LoggerExists(ctx) {
		ms = audit.NewManifestService(ctx, ms)
	}
	return ms, nil
}
func (r *repository) Blobs(ctx context.Context) distribution.BlobStore {
	_logClusterCodePath()
	defer _logClusterCodePath()
	bs := r.Repository.Blobs(ctx)
	if r.app.quotaEnforcing.enforcementEnabled {
		bs = &quotaRestrictedBlobStore{BlobStore: bs, repo: r}
	}
	bs = &pullthroughBlobStore{BlobStore: bs, remoteBlobGetter: r.remoteBlobGetter, writeLimiter: r.app.writeLimiter, mirror: r.app.config.Pullthrough.Mirror, newLocalBlobStore: r.Repository.Blobs}
	bs = newPendingErrorsBlobStore(bs, r)
	if audit.LoggerExists(ctx) {
		bs = audit.NewBlobStore(ctx, bs)
	}
	return bs
}
func (r *repository) Tags(ctx context.Context) distribution.TagService {
	_logClusterCodePath()
	defer _logClusterCodePath()
	ts := r.Repository.Tags(ctx)
	ts = &tagService{TagService: ts, imageStream: r.imageStream}
	ts = newPendingErrorsTagService(ts, r)
	if audit.LoggerExists(ctx) {
		ts = audit.NewTagService(ctx, ts)
	}
	return ts
}
func (r *repository) BlobDescriptorService(svc distribution.BlobDescriptorService) distribution.BlobDescriptorService {
	_logClusterCodePath()
	defer _logClusterCodePath()
	svc = &cache.RepositoryScopedBlobDescriptor{Repo: r.Named().String(), Cache: r.app.cache, Svc: svc}
	svc = &blobDescriptorService{svc, r}
	svc = newPendingErrorsBlobDescriptorService(svc, r)
	return svc
}
func (r *repository) checkPendingErrors(ctx context.Context) error {
	_logClusterCodePath()
	defer _logClusterCodePath()
	if !authPerformed(ctx) {
		return fmt.Errorf("openshift.auth.completed missing from context")
	}
	deferredErrors, haveDeferredErrors := deferredErrorsFrom(ctx)
	if !haveDeferredErrors {
		return nil
	}
	repoErr, haveRepoErr := deferredErrors.Get(r.imageStream.Reference())
	if !haveRepoErr {
		return nil
	}
	dcontext.GetLogger(r.ctx).Debugf("Origin auth: found deferred error for %s: %v", r.imageStream.Reference(), repoErr)
	return repoErr
}
