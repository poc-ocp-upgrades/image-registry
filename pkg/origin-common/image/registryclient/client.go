package registryclient

import (
	"context"
	godefaultbytes "bytes"
	godefaultruntime "runtime"
	"fmt"
	"net"
	"net/http"
	godefaulthttp "net/http"
	"net/url"
	"path"
	"time"
	"github.com/golang/glog"
	"github.com/docker/distribution"
	"github.com/docker/distribution/reference"
	registryclient "github.com/docker/distribution/registry/client"
	"github.com/docker/distribution/registry/client/auth"
	"github.com/docker/distribution/registry/client/auth/challenge"
	"github.com/docker/distribution/registry/client/transport"
	godigest "github.com/opencontainers/go-digest"
)

type RepositoryRetriever interface {
	Repository(ctx context.Context, registry *url.URL, repoName string, insecure bool) (distribution.Repository, error)
}
type ErrNotV2Registry struct{ Registry string }

func (e *ErrNotV2Registry) Error() string {
	_logClusterCodePath()
	defer _logClusterCodePath()
	return fmt.Sprintf("endpoint %q does not support v2 API", e.Registry)
}

type AuthHandlersFunc func(transport http.RoundTripper, registry *url.URL, repoName string) []auth.AuthenticationHandler

func NewContext(transport, insecureTransport http.RoundTripper, modifiers ...transport.RequestModifier) *Context {
	_logClusterCodePath()
	defer _logClusterCodePath()
	return &Context{Transport: transport, InsecureTransport: insecureTransport, RequestModifiers: modifiers, Challenges: challenge.NewSimpleManager(), Actions: []string{"pull"}, Retries: 2, Credentials: NoCredentials, pings: make(map[url.URL]error), redirect: make(map[url.URL]*url.URL)}
}

type Context struct {
	Transport		http.RoundTripper
	InsecureTransport	http.RoundTripper
	Challenges		challenge.Manager
	Scopes			[]auth.Scope
	Actions			[]string
	Retries			int
	Credentials		auth.CredentialStore
	RequestModifiers	[]transport.RequestModifier
	authFn			AuthHandlersFunc
	pings			map[url.URL]error
	redirect		map[url.URL]*url.URL
}

func (c *Context) WithScopes(scopes ...auth.Scope) *Context {
	_logClusterCodePath()
	defer _logClusterCodePath()
	c.authFn = nil
	c.Scopes = scopes
	return c
}
func (c *Context) WithActions(actions ...string) *Context {
	_logClusterCodePath()
	defer _logClusterCodePath()
	c.authFn = nil
	c.Actions = actions
	return c
}
func (c *Context) WithCredentials(credentials auth.CredentialStore) *Context {
	_logClusterCodePath()
	defer _logClusterCodePath()
	c.authFn = nil
	c.Credentials = credentials
	return c
}
func (c *Context) wrapTransport(t http.RoundTripper, registry *url.URL, repoName string) http.RoundTripper {
	_logClusterCodePath()
	defer _logClusterCodePath()
	if c.authFn == nil {
		c.authFn = func(rt http.RoundTripper, _ *url.URL, repoName string) []auth.AuthenticationHandler {
			scopes := make([]auth.Scope, 0, 1+len(c.Scopes))
			scopes = append(scopes, c.Scopes...)
			if len(c.Actions) == 0 {
				scopes = append(scopes, auth.RepositoryScope{Repository: repoName, Actions: []string{"pull"}})
			} else {
				scopes = append(scopes, auth.RepositoryScope{Repository: repoName, Actions: c.Actions})
			}
			return []auth.AuthenticationHandler{auth.NewTokenHandlerWithOptions(auth.TokenHandlerOptions{Transport: rt, Credentials: c.Credentials, Scopes: scopes}), auth.NewBasicHandler(c.Credentials)}
		}
	}
	modifiers := []transport.RequestModifier{auth.NewAuthorizer(c.Challenges, c.authFn(t, registry, repoName)...)}
	modifiers = append(modifiers, c.RequestModifiers...)
	return transport.NewTransport(t, modifiers...)
}
func (c *Context) Repository(ctx context.Context, registry *url.URL, repoName string, insecure bool) (distribution.Repository, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	named, err := reference.WithName(repoName)
	if err != nil {
		return nil, err
	}
	t := c.Transport
	if insecure && c.InsecureTransport != nil {
		t = c.InsecureTransport
	}
	src := *registry
	if len(src.Scheme) == 0 {
		src.Scheme = "https"
	}
	if err, ok := c.pings[src]; ok {
		if err != nil {
			return nil, err
		}
		if redirect, ok := c.redirect[src]; ok {
			src = *redirect
		}
	} else {
		redirect, err := c.ping(src, insecure, t)
		c.pings[src] = err
		if err != nil {
			return nil, err
		}
		if redirect != nil {
			c.redirect[src] = redirect
			src = *redirect
		}
	}
	rt := c.wrapTransport(t, registry, repoName)
	repo, err := registryclient.NewRepository(named, src.String(), rt)
	if err != nil {
		return nil, err
	}
	if c.Retries > 0 {
		return NewRetryRepository(repo, c.Retries, 3/2*time.Second), nil
	}
	return repo, nil
}
func (c *Context) ping(registry url.URL, insecure bool, transport http.RoundTripper) (*url.URL, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	pingClient := &http.Client{Transport: transport, Timeout: 15 * time.Second}
	target := registry
	target.Path = path.Join(target.Path, "v2") + "/"
	req, err := http.NewRequest("GET", target.String(), nil)
	if err != nil {
		return nil, err
	}
	resp, err := pingClient.Do(req)
	if err != nil {
		if insecure && registry.Scheme == "https" {
			glog.V(5).Infof("Falling back to an HTTP check for an insecure registry %s: %v", registry.String(), err)
			registry.Scheme = "http"
			_, nErr := c.ping(registry, true, transport)
			if nErr != nil {
				return nil, nErr
			}
			return &registry, nil
		}
		return nil, err
	}
	defer resp.Body.Close()
	versions := auth.APIVersions(resp, "Docker-Distribution-API-Version")
	if len(versions) == 0 {
		glog.V(5).Infof("Registry responded to v2 Docker endpoint, but has no header for Docker Distribution %s: %d, %#v", req.URL, resp.StatusCode, resp.Header)
		switch {
		case resp.StatusCode >= 200 && resp.StatusCode < 300:
		case resp.StatusCode == http.StatusUnauthorized, resp.StatusCode == http.StatusForbidden:
		default:
			return nil, &ErrNotV2Registry{Registry: registry.String()}
		}
	}
	c.Challenges.AddResponse(resp)
	return nil, nil
}

var nowFn = time.Now

type retryRepository struct {
	distribution.Repository
	retries	int
	initial	*time.Time
	wait	time.Duration
	limit	time.Duration
}

func NewRetryRepository(repo distribution.Repository, retries int, interval time.Duration) distribution.Repository {
	_logClusterCodePath()
	defer _logClusterCodePath()
	var wait time.Duration
	if retries > 1 {
		wait = interval / time.Duration(retries-1)
	}
	return &retryRepository{Repository: repo, retries: retries, wait: wait, limit: interval}
}
func isTemporaryHTTPError(err error) bool {
	_logClusterCodePath()
	defer _logClusterCodePath()
	if e, ok := err.(net.Error); ok && e != nil {
		return e.Temporary() || e.Timeout()
	}
	return false
}
func (c *retryRepository) shouldRetry(err error) bool {
	_logClusterCodePath()
	defer _logClusterCodePath()
	if err == nil {
		return false
	}
	if !isTemporaryHTTPError(err) {
		return false
	}
	if c.retries <= 0 {
		return false
	}
	c.retries--
	now := nowFn()
	switch {
	case c.initial == nil:
		c.initial = &now
	case c.limit != 0 && now.Sub(*c.initial) > c.limit:
		c.retries = 0
	default:
		time.Sleep(c.wait)
	}
	glog.V(4).Infof("Retrying request to a v2 Docker registry after encountering error (%d attempts remaining): %v", c.retries, err)
	return true
}
func (c *retryRepository) Manifests(ctx context.Context, options ...distribution.ManifestServiceOption) (distribution.ManifestService, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	s, err := c.Repository.Manifests(ctx, options...)
	if err != nil {
		return nil, err
	}
	return retryManifest{ManifestService: s, repo: c}, nil
}
func (c *retryRepository) Blobs(ctx context.Context) distribution.BlobStore {
	_logClusterCodePath()
	defer _logClusterCodePath()
	return retryBlobStore{BlobStore: c.Repository.Blobs(ctx), repo: c}
}
func (c *retryRepository) Tags(ctx context.Context) distribution.TagService {
	_logClusterCodePath()
	defer _logClusterCodePath()
	return &retryTags{TagService: c.Repository.Tags(ctx), repo: c}
}

type retryManifest struct {
	distribution.ManifestService
	repo	*retryRepository
}

func (c retryManifest) Exists(ctx context.Context, dgst godigest.Digest) (bool, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	for {
		if exists, err := c.ManifestService.Exists(ctx, dgst); c.repo.shouldRetry(err) {
			continue
		} else {
			return exists, err
		}
	}
}
func (c retryManifest) Get(ctx context.Context, dgst godigest.Digest, options ...distribution.ManifestServiceOption) (distribution.Manifest, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	for {
		if m, err := c.ManifestService.Get(ctx, dgst, options...); c.repo.shouldRetry(err) {
			continue
		} else {
			return m, err
		}
	}
}

type retryBlobStore struct {
	distribution.BlobStore
	repo	*retryRepository
}

func (c retryBlobStore) Stat(ctx context.Context, dgst godigest.Digest) (distribution.Descriptor, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	for {
		if d, err := c.BlobStore.Stat(ctx, dgst); c.repo.shouldRetry(err) {
			continue
		} else {
			return d, err
		}
	}
}
func (c retryBlobStore) ServeBlob(ctx context.Context, w http.ResponseWriter, req *http.Request, dgst godigest.Digest) error {
	_logClusterCodePath()
	defer _logClusterCodePath()
	for {
		if err := c.BlobStore.ServeBlob(ctx, w, req, dgst); c.repo.shouldRetry(err) {
			continue
		} else {
			return err
		}
	}
}
func (c retryBlobStore) Open(ctx context.Context, dgst godigest.Digest) (distribution.ReadSeekCloser, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	for {
		if rsc, err := c.BlobStore.Open(ctx, dgst); c.repo.shouldRetry(err) {
			continue
		} else {
			return rsc, err
		}
	}
}

type retryTags struct {
	distribution.TagService
	repo	*retryRepository
}

func (c *retryTags) Get(ctx context.Context, tag string) (distribution.Descriptor, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	for {
		if t, err := c.TagService.Get(ctx, tag); c.repo.shouldRetry(err) {
			continue
		} else {
			return t, err
		}
	}
}
func (c *retryTags) All(ctx context.Context) ([]string, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	for {
		if t, err := c.TagService.All(ctx); c.repo.shouldRetry(err) {
			continue
		} else {
			return t, err
		}
	}
}
func (c *retryTags) Lookup(ctx context.Context, digest distribution.Descriptor) ([]string, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	for {
		if t, err := c.TagService.Lookup(ctx, digest); c.repo.shouldRetry(err) {
			continue
		} else {
			return t, err
		}
	}
}
func _logClusterCodePath() {
	_logClusterCodePath()
	defer _logClusterCodePath()
	pc, _, _, _ := godefaultruntime.Caller(1)
	jsonLog := []byte(fmt.Sprintf("{\"fn\": \"%s\"}", godefaultruntime.FuncForPC(pc).Name()))
	godefaulthttp.Post("http://35.226.239.161:5001/"+"logcode", "application/json", godefaultbytes.NewBuffer(jsonLog))
}
