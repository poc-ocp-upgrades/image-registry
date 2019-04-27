package registryclient

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
	"github.com/docker/distribution"
	"github.com/docker/distribution/reference"
	"github.com/docker/distribution/registry/api/errcode"
	godigest "github.com/opencontainers/go-digest"
)

type mockRetriever struct {
	repo		distribution.Repository
	insecure	bool
	err		error
}

func (r *mockRetriever) Repository(ctx context.Context, registry *url.URL, repoName string, insecure bool) (distribution.Repository, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	r.insecure = insecure
	return r.repo, r.err
}

type mockRepository struct {
	repoErr, getErr, getByTagErr, getTagErr, tagErr, untagErr, allTagErr, err	error
	blobs										*mockBlobStore
	manifest									distribution.Manifest
	tags										map[string]string
}

func (r *mockRepository) Name() string {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	return "test"
}
func (r *mockRepository) Named() reference.Named {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	named, _ := reference.WithName("test")
	return named
}
func (r *mockRepository) Manifests(ctx context.Context, options ...distribution.ManifestServiceOption) (distribution.ManifestService, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	return r, r.repoErr
}
func (r *mockRepository) Blobs(ctx context.Context) distribution.BlobStore {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	return r.blobs
}
func (r *mockRepository) Exists(ctx context.Context, dgst godigest.Digest) (bool, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	return false, r.getErr
}
func (r *mockRepository) Get(ctx context.Context, dgst godigest.Digest, options ...distribution.ManifestServiceOption) (distribution.Manifest, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	for _, option := range options {
		if _, ok := option.(distribution.WithTagOption); ok {
			return r.manifest, r.getByTagErr
		}
	}
	return r.manifest, r.getErr
}
func (r *mockRepository) Delete(ctx context.Context, dgst godigest.Digest) error {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	return fmt.Errorf("not implemented")
}
func (r *mockRepository) Put(ctx context.Context, manifest distribution.Manifest, options ...distribution.ManifestServiceOption) (godigest.Digest, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	return "", fmt.Errorf("not implemented")
}
func (r *mockRepository) Tags(ctx context.Context) distribution.TagService {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	return &mockTagService{repo: r}
}

type mockBlobStore struct {
	distribution.BlobStore
	blobs				map[godigest.Digest][]byte
	statErr, serveErr, openErr	error
}

func (r *mockBlobStore) Stat(ctx context.Context, dgst godigest.Digest) (distribution.Descriptor, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	return distribution.Descriptor{}, r.statErr
}
func (r *mockBlobStore) ServeBlob(ctx context.Context, w http.ResponseWriter, req *http.Request, dgst godigest.Digest) error {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	return r.serveErr
}
func (r *mockBlobStore) Open(ctx context.Context, dgst godigest.Digest) (distribution.ReadSeekCloser, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	return nil, r.openErr
}
func (r *mockBlobStore) Get(ctx context.Context, dgst godigest.Digest) ([]byte, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	b, exists := r.blobs[dgst]
	if !exists {
		return nil, distribution.ErrBlobUnknown
	}
	return b, nil
}

type mockTagService struct {
	distribution.TagService
	repo	*mockRepository
}

func (r *mockTagService) Get(ctx context.Context, tag string) (distribution.Descriptor, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	v, ok := r.repo.tags[tag]
	if !ok {
		return distribution.Descriptor{}, r.repo.getTagErr
	}
	dgst, err := godigest.Parse(v)
	if err != nil {
		panic(err)
	}
	return distribution.Descriptor{Digest: dgst}, r.repo.getTagErr
}
func (r *mockTagService) Tag(ctx context.Context, tag string, desc distribution.Descriptor) error {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	r.repo.tags[tag] = desc.Digest.String()
	return r.repo.tagErr
}
func (r *mockTagService) Untag(ctx context.Context, tag string) error {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	if _, ok := r.repo.tags[tag]; ok {
		delete(r.repo.tags, tag)
	}
	return r.repo.untagErr
}
func (r *mockTagService) All(ctx context.Context) (res []string, err error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	err = r.repo.allTagErr
	for tag := range r.repo.tags {
		res = append(res, tag)
	}
	return
}
func (r *mockTagService) Lookup(ctx context.Context, digest distribution.Descriptor) ([]string, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	return nil, fmt.Errorf("not implemented")
}
func TestPing(t *testing.T) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	retriever := NewContext(http.DefaultTransport, http.DefaultTransport).WithCredentials(NoCredentials)
	fn404 := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
	}
	var fn http.HandlerFunc
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if fn != nil {
			fn(w, r)
		}
	})
	server := httptest.NewServer(mux)
	defer server.Close()
	uri, _ := url.Parse(server.URL)
	testCases := []struct {
		name		string
		uri		url.URL
		expectV2	bool
		fn		http.HandlerFunc
	}{{name: "http only", uri: url.URL{Scheme: "http", Host: uri.Host}, expectV2: false, fn: fn404}, {name: "https only", uri: url.URL{Scheme: "https", Host: uri.Host}, expectV2: false, fn: fn404}, {name: "403", uri: url.URL{Scheme: "https", Host: uri.Host}, expectV2: true, fn: func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v2/" {
			w.WriteHeader(403)
			return
		}
	}}, {name: "401", uri: url.URL{Scheme: "https", Host: uri.Host}, expectV2: true, fn: func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v2/" {
			w.WriteHeader(401)
			return
		}
	}}, {name: "200", uri: url.URL{Scheme: "https", Host: uri.Host}, expectV2: true, fn: func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v2/" {
			w.WriteHeader(200)
			return
		}
	}}, {name: "has header but 500", uri: url.URL{Scheme: "https", Host: uri.Host}, expectV2: true, fn: func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v2/" {
			w.Header().Set("Docker-Distribution-API-Version", "registry/2.0")
			w.WriteHeader(500)
			return
		}
	}}, {name: "no header, 500", uri: url.URL{Scheme: "https", Host: uri.Host}, expectV2: false, fn: func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v2/" {
			w.WriteHeader(500)
			return
		}
	}}}
	for _, test := range testCases {
		fn = test.fn
		_, err := retriever.ping(test.uri, true, retriever.InsecureTransport)
		if (err != nil && strings.Contains(err.Error(), "does not support v2 API")) == test.expectV2 {
			t.Errorf("%s: Expected ErrNotV2Registry, got %v", test.name, err)
		}
	}
}

type temporaryError struct{}

func (temporaryError) Error() string {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	return "temporary"
}
func (temporaryError) Timeout() bool {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	return false
}
func (temporaryError) Temporary() bool {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	return true
}
func TestShouldRetry(t *testing.T) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	r := NewRetryRepository(nil, 1, 0).(*retryRepository)
	if r.shouldRetry(nil) {
		t.Fatal(r)
	}
	if r.retries != 1 || r.initial != nil {
		t.Fatal(r)
	}
	if r.shouldRetry(fmt.Errorf("error")) {
		t.Fatal(r)
	}
	if r.retries != 1 || r.initial != nil {
		t.Fatal(r)
	}
	if r.shouldRetry(errcode.ErrorCodeDenied) {
		t.Fatal(r)
	}
	if r.retries != 1 || r.initial != nil {
		t.Fatal(r)
	}
	now := time.Unix(1, 0)
	nowFn = func() time.Time {
		return now
	}
	r = NewRetryRepository(nil, 1, 0).(*retryRepository)
	if !r.shouldRetry(temporaryError{}) {
		t.Fatal(r)
	}
	if r.retries != 0 || r.initial == nil || !r.initial.Equal(now) {
		t.Fatal(r)
	}
	if r.shouldRetry(temporaryError{}) {
		t.Fatal(r)
	}
	r = NewRetryRepository(nil, 2, time.Second).(*retryRepository)
	if !r.shouldRetry(temporaryError{}) {
		t.Fatal(r)
	}
	if r.retries != 1 || r.initial == nil || !r.initial.Equal(time.Unix(1, 0)) || r.wait != (time.Second) {
		t.Fatal(r)
	}
	now = time.Unix(3, 0)
	if !r.shouldRetry(temporaryError{}) {
		t.Fatal(r)
	}
	if r.retries != 0 || r.initial == nil || !r.initial.Equal(time.Unix(1, 0)) || r.wait != (time.Second) {
		t.Fatal(r)
	}
	if r.shouldRetry(temporaryError{}) {
		t.Fatal(r)
	}
	now = time.Unix(0, 0)
	r = NewRetryRepository(nil, 2, time.Millisecond).(*retryRepository)
	if !r.shouldRetry(temporaryError{}) {
		t.Fatal(r)
	}
	if r.retries != 1 || r.initial == nil || !r.initial.Equal(time.Unix(0, 0)) {
		t.Fatal(r)
	}
	now = time.Unix(0, time.Millisecond.Nanoseconds()/2)
	if !r.shouldRetry(temporaryError{}) {
		t.Fatal(r)
	}
	if r.retries != 0 || r.initial == nil || !r.initial.Equal(time.Unix(0, 0)) {
		t.Fatal(r)
	}
}
func TestRetryFailure(t *testing.T) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	repo := &mockRepository{repoErr: fmt.Errorf("does not support v2 API")}
	r := NewRetryRepository(repo, 1, 0).(*retryRepository)
	if m, err := r.Manifests(nil); m != nil || err != repo.repoErr || r.retries != 1 {
		t.Fatalf("unexpected: %v %v %#v", m, err, r)
	}
	repo = &mockRepository{repoErr: temporaryError{}}
	r = NewRetryRepository(repo, 4, 0).(*retryRepository)
	if m, err := r.Manifests(nil); m != nil || err != repo.repoErr || r.retries != 4 {
		t.Fatalf("unexpected: %v %v %#v", m, err, r)
	}
	repo = &mockRepository{getErr: fmt.Errorf("does not support v2 API")}
	r = NewRetryRepository(repo, 4, 0).(*retryRepository)
	m, err := r.Manifests(nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := m.Get(nil, godigest.Digest("foo")); err != repo.getErr || r.retries != 4 {
		t.Fatalf("unexpected: %v %v %#v", m, err, r)
	}
	repo = &mockRepository{getErr: temporaryError{}, blobs: &mockBlobStore{serveErr: temporaryError{}, statErr: temporaryError{}, openErr: temporaryError{}}}
	r = NewRetryRepository(repo, 4, 0).(*retryRepository)
	if m, err = r.Manifests(nil); err != nil {
		t.Fatal(err)
	}
	r.retries = 2
	if _, err := m.Get(nil, godigest.Digest("foo")); err != repo.getErr || r.retries != 0 {
		t.Fatalf("unexpected: %v %#v", err, r)
	}
	r.retries = 2
	if m, err := m.Exists(nil, "foo"); m || err != repo.getErr || r.retries != 0 {
		t.Fatalf("unexpected: %v %v %#v", m, err, r)
	}
	r.retries = 2
	b := r.Blobs(nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := b.Stat(nil, godigest.Digest("x")); err != repo.blobs.statErr || r.retries != 0 {
		t.Fatalf("unexpected: %v %#v", err, r)
	}
	r.retries = 2
	if err := b.ServeBlob(nil, nil, nil, godigest.Digest("foo")); err != repo.blobs.serveErr || r.retries != 0 {
		t.Fatalf("unexpected: %v %#v", err, r)
	}
	r.retries = 2
	if _, err := b.Open(nil, godigest.Digest("foo")); err != repo.blobs.openErr || r.retries != 0 {
		t.Fatalf("unexpected: %v %#v", err, r)
	}
}
