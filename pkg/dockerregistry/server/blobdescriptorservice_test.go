package server

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"
	"time"
	"github.com/docker/distribution"
	"github.com/docker/distribution/configuration"
	"github.com/docker/distribution/registry/api/errcode"
	"github.com/docker/distribution/registry/api/v2"
	registryauth "github.com/docker/distribution/registry/auth"
	"github.com/opencontainers/go-digest"
	registryclient "github.com/openshift/image-registry/pkg/dockerregistry/server/client"
	srvconfig "github.com/openshift/image-registry/pkg/dockerregistry/server/configuration"
	"github.com/openshift/image-registry/pkg/dockerregistry/server/supermiddleware"
	"github.com/openshift/image-registry/pkg/testutil"
)

type appMiddlewareChain []appMiddleware

func (m appMiddlewareChain) Apply(app supermiddleware.App) supermiddleware.App {
	_logClusterCodePath()
	defer _logClusterCodePath()
	for _, am := range m {
		app = am.Apply(app)
	}
	return app
}
func TestBlobDescriptorServiceIsApplied(t *testing.T) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	ctx := context.Background()
	ctx = testutil.WithTestLogger(ctx, t)
	_, imageClient := testutil.NewFakeOpenShiftWithClient(ctx)
	osclient, err := registryclient.NewFakeRegistryClient(imageClient).Client()
	if err != nil {
		t.Fatal(err)
	}
	m := NewTestBlobDescriptorManager()
	ctx = withAppMiddleware(ctx, &appMiddlewareChain{&fakeAccessControllerMiddleware{t: t, userClient: osclient}, &fakeBlobDescriptorServiceMiddleware{t: t, m: m}})
	dockercfg := &configuration.Configuration{Loglevel: "debug", Auth: map[string]configuration.Parameters{"openshift": nil}, Storage: configuration.Storage{"inmemory": configuration.Parameters{}, "cache": configuration.Parameters{"blobdescriptor": "inmemory"}, "delete": configuration.Parameters{"enabled": true}, "maintenance": configuration.Parameters{"uploadpurging": map[interface{}]interface{}{"enabled": false}}}, Middleware: map[string][]configuration.Middleware{"registry": {{Name: "openshift"}}, "repository": {{Name: "openshift"}}, "storage": {{Name: "openshift"}}}}
	cfg := &srvconfig.Configuration{Server: &srvconfig.Server{Addr: "localhost:5000"}}
	if err := srvconfig.InitExtraConfig(dockercfg, cfg); err != nil {
		t.Fatal(err)
	}
	app := NewApp(ctx, registryclient.NewFakeRegistryClient(imageClient), dockercfg, cfg, nil)
	server := httptest.NewServer(app)
	router := v2.Router()
	serverURL, err := url.Parse(server.URL)
	if err != nil {
		t.Fatalf("error parsing server url: %v", err)
	}
	desc, _, err := testutil.UploadRandomTestBlob(ctx, serverURL.String(), nil, "user/app")
	if err != nil {
		t.Fatal(err)
	}
	repoName := "user/app"
	transport, err := testutil.NewTransport(server.URL, repoName, nil)
	if err != nil {
		t.Fatalf("failed to get transport for %s: %v", repoName, err)
	}
	repo, err := testutil.NewRepository(repoName, server.URL, transport)
	if err != nil {
		t.Fatalf("failed to get repository %s: %v", repoName, err)
	}
	manifest, err := testutil.UploadSchema2Image(ctx, repo, "latest")
	if err != nil {
		t.Fatalf("unable to upload random image: %s", err)
	}
	_, manifestPayload, err := manifest.Payload()
	if err != nil {
		t.Fatal(err)
	}
	manifestDigest := digest.FromBytes(manifestPayload)
	type testCase struct {
		name				string
		method				string
		endpoint			string
		vars				[]string
		expectedStatus			int
		expectedMethodInvocations	map[string]int
	}
	doTest := func(tc testCase) {
		m.clearStats()
		route := router.GetRoute(tc.endpoint).Host(serverURL.Host)
		u, err := route.URL(tc.vars...)
		if err != nil {
			t.Errorf("[%s] failed to build route: %v", tc.name, err)
			return
		}
		req, err := http.NewRequest(tc.method, u.String(), nil)
		if err != nil {
			t.Errorf("[%s] failed to make request: %v", tc.name, err)
		}
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			t.Errorf("[%s] failed to do the request: %v", tc.name, err)
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode != tc.expectedStatus {
			t.Errorf("[%s] unexpected status code: got %v, want %v", tc.name, resp.StatusCode, tc.expectedStatus)
		}
		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
			content, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				t.Errorf("[%s] failed to read body: %v", tc.name, err)
			} else if len(content) > 0 {
				errs := errcode.Errors{}
				err := errs.UnmarshalJSON(content)
				if err != nil {
					t.Logf("[%s] failed to parse body as error: %v", tc.name, err)
					t.Logf("[%s] received body: %v", tc.name, string(content))
				} else {
					t.Logf("[%s] received errors: %#+v", tc.name, errs)
				}
			}
		}
		stats, err := m.getStats(tc.expectedMethodInvocations, time.Second*5)
		if err != nil {
			t.Fatalf("[%s] failed to get stats: %v", tc.name, err)
		}
		for method, exp := range tc.expectedMethodInvocations {
			invoked := stats[method]
			if invoked != exp {
				t.Errorf("[%s] unexpected number of invocations of method %q: %v != %v", tc.name, method, invoked, exp)
			}
		}
		for method, invoked := range stats {
			if _, ok := tc.expectedMethodInvocations[method]; !ok {
				t.Errorf("[%s] unexpected method %q invoked %d times", tc.name, method, invoked)
			}
		}
	}
	for _, tc := range []testCase{{name: "get blob", method: http.MethodGet, endpoint: v2.RouteNameBlob, vars: []string{"name", "user/app", "digest", desc.Digest.String()}, expectedStatus: http.StatusOK, expectedMethodInvocations: map[string]int{"Stat": 3}}, {name: "stat blob", method: http.MethodHead, endpoint: v2.RouteNameBlob, vars: []string{"name", "user/app", "digest", desc.Digest.String()}, expectedStatus: http.StatusOK, expectedMethodInvocations: map[string]int{"Stat": 3}}, {name: "delete blob", method: http.MethodDelete, endpoint: v2.RouteNameBlob, vars: []string{"name", "user/app", "digest", desc.Digest.String()}, expectedStatus: http.StatusAccepted, expectedMethodInvocations: map[string]int{"Stat": 1, "Clear": 1}}, {name: "delete manifest", method: http.MethodDelete, endpoint: v2.RouteNameManifest, vars: []string{"name", "user/app", "reference", string(manifestDigest)}, expectedStatus: http.StatusMethodNotAllowed}, {name: "get manifest", method: http.MethodGet, endpoint: v2.RouteNameManifest, vars: []string{"name", "user/app", "reference", "latest"}, expectedStatus: http.StatusOK, expectedMethodInvocations: map[string]int{"Stat": 3}}} {
		doTest(tc)
	}
}

type testBlobDescriptorManager struct {
	mu	sync.Mutex
	cond	*sync.Cond
	stats	map[string]int
}

func NewTestBlobDescriptorManager() *testBlobDescriptorManager {
	_logClusterCodePath()
	defer _logClusterCodePath()
	m := &testBlobDescriptorManager{stats: make(map[string]int)}
	m.cond = sync.NewCond(&m.mu)
	return m
}
func (m *testBlobDescriptorManager) clearStats() {
	_logClusterCodePath()
	defer _logClusterCodePath()
	m.mu.Lock()
	defer m.mu.Unlock()
	for k := range m.stats {
		delete(m.stats, k)
	}
}
func (m *testBlobDescriptorManager) methodInvoked(methodName string) int {
	_logClusterCodePath()
	defer _logClusterCodePath()
	m.mu.Lock()
	defer m.mu.Unlock()
	newCount := m.stats[methodName] + 1
	m.stats[methodName] = newCount
	m.cond.Signal()
	return newCount
}
func (m *testBlobDescriptorManager) getStats(minimumLimits map[string]int, timeout time.Duration) (map[string]int, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	end := time.Now().Add(timeout)
	stats := make(map[string]int)
	if len(minimumLimits) == 0 {
		m.mu.Lock()
		for k, v := range m.stats {
			stats[k] = v
		}
		m.mu.Unlock()
		return stats, nil
	}
	c := make(chan struct{})
	go func() {
		m.mu.Lock()
		defer m.mu.Unlock()
		for !statsGreaterThanOrEqual(m.stats, minimumLimits) {
			m.cond.Wait()
		}
		c <- struct{}{}
	}()
	var err error
	select {
	case <-time.After(time.Until(end)):
		err = fmt.Errorf("timeout while waiting on expected stats")
	case <-c:
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	for k, v := range m.stats {
		stats[k] = v
	}
	return stats, err
}
func statsGreaterThanOrEqual(stats, minimumLimits map[string]int) bool {
	_logClusterCodePath()
	defer _logClusterCodePath()
	for key, val := range minimumLimits {
		if val > stats[key] {
			return false
		}
	}
	return true
}

type fakeBlobDescriptorServiceMiddleware struct {
	t	*testing.T
	m	*testBlobDescriptorManager
}

func (m *fakeBlobDescriptorServiceMiddleware) Apply(app supermiddleware.App) supermiddleware.App {
	_logClusterCodePath()
	defer _logClusterCodePath()
	return &fakeBlobDescriptorServiceApp{App: app, t: m.t, m: m.m}
}

type fakeBlobDescriptorServiceApp struct {
	supermiddleware.App
	t	*testing.T
	m	*testBlobDescriptorManager
}

func (app *fakeBlobDescriptorServiceApp) Repository(ctx context.Context, repo distribution.Repository, crossmount bool) (distribution.Repository, distribution.BlobDescriptorServiceFactory, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	repo, bdsf, err := app.App.Repository(ctx, repo, crossmount)
	if err != nil {
		return repo, bdsf, err
	}
	return repo, &testBlobDescriptorServiceFactory{upstream: bdsf, t: app.t, m: app.m}, nil
}

type testBlobDescriptorServiceFactory struct {
	upstream	distribution.BlobDescriptorServiceFactory
	t		*testing.T
	m		*testBlobDescriptorManager
}

func (bf *testBlobDescriptorServiceFactory) BlobAccessController(svc distribution.BlobDescriptorService) distribution.BlobDescriptorService {
	_logClusterCodePath()
	defer _logClusterCodePath()
	svc = bf.upstream.BlobAccessController(svc)
	return &testBlobDescriptorService{BlobDescriptorService: svc, t: bf.t, m: bf.m}
}

type testBlobDescriptorService struct {
	distribution.BlobDescriptorService
	t	*testing.T
	m	*testBlobDescriptorManager
}

func (bs *testBlobDescriptorService) Stat(ctx context.Context, dgst digest.Digest) (distribution.Descriptor, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	if bs.m != nil {
		bs.m.methodInvoked("Stat")
	}
	return bs.BlobDescriptorService.Stat(ctx, dgst)
}
func (bs *testBlobDescriptorService) Clear(ctx context.Context, dgst digest.Digest) error {
	_logClusterCodePath()
	defer _logClusterCodePath()
	if bs.m != nil {
		bs.m.methodInvoked("Clear")
	}
	return bs.BlobDescriptorService.Clear(ctx, dgst)
}

type fakeAccessControllerMiddleware struct {
	t		*testing.T
	userClient	registryclient.Interface
}

func (m *fakeAccessControllerMiddleware) Apply(app supermiddleware.App) supermiddleware.App {
	_logClusterCodePath()
	defer _logClusterCodePath()
	return &fakeAccessControllerApp{App: app, t: m.t, userClient: m.userClient}
}

type fakeAccessControllerApp struct {
	supermiddleware.App
	t		*testing.T
	userClient	registryclient.Interface
}

func (app *fakeAccessControllerApp) Auth(map[string]interface{}) (registryauth.AccessController, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	return &fakeAccessController{t: app.t, userClient: app.userClient}, nil
}

type fakeAccessController struct {
	t		*testing.T
	userClient	registryclient.Interface
}

func (f *fakeAccessController) Authorized(ctx context.Context, access ...registryauth.Access) (context.Context, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	for _, access := range access {
		f.t.Logf("fake authorizer: authorizing access to %s:%s:%s", access.Resource.Type, access.Resource.Name, access.Action)
	}
	if f.userClient != nil {
		ctx = withUserClient(ctx, f.userClient)
	}
	ctx = withAuthPerformed(ctx)
	return ctx, nil
}
