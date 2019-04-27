package server

import (
	"fmt"
	"time"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/cache"
	"github.com/openshift/image-registry/pkg/imagestream"
)

type projectObjectListCache struct{ store cache.Store }

var _ imagestream.ProjectObjectListStore = &projectObjectListCache{}

func newProjectObjectListCache(ttl time.Duration) imagestream.ProjectObjectListStore {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	return &projectObjectListCache{store: cache.NewTTLStore(metaProjectObjectListKeyFunc, ttl)}
}
func (c *projectObjectListCache) Add(namespace string, obj runtime.Object) error {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	if namespace == "" {
		return fmt.Errorf("namespace cannot be empty")
	}
	no := &namespacedObject{namespace: namespace, object: obj}
	return c.store.Add(no)
}
func (c *projectObjectListCache) Get(namespace string) (runtime.Object, bool, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	entry, exists, err := c.store.GetByKey(namespace)
	if err != nil {
		return nil, exists, err
	}
	if !exists {
		return nil, false, err
	}
	no, ok := entry.(*namespacedObject)
	if !ok {
		return nil, false, fmt.Errorf("%T is not a namespaced object", entry)
	}
	return no.object, true, nil
}

type namespacedObject struct {
	namespace	string
	object		runtime.Object
}

func metaProjectObjectListKeyFunc(obj interface{}) (string, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	if key, ok := obj.(cache.ExplicitKey); ok {
		return string(key), nil
	}
	no, ok := obj.(*namespacedObject)
	if !ok {
		return "", fmt.Errorf("object %T is not a namespaced object", obj)
	}
	return no.namespace, nil
}
