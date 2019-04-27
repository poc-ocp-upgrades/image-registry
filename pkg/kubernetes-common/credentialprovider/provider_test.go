package credentialprovider

import (
	"testing"
	"time"
)

func TestCachingProvider(t *testing.T) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	provider := &testProvider{Count: 0}
	cache := &CachingDockerConfigProvider{Provider: provider, Lifetime: 1 * time.Second}
	if provider.Count != 0 {
		t.Errorf("Unexpected number of Provide calls: %v", provider.Count)
	}
	cache.Provide()
	cache.Provide()
	cache.Provide()
	cache.Provide()
	if provider.Count != 1 {
		t.Errorf("Unexpected number of Provide calls: %v", provider.Count)
	}
	time.Sleep(cache.Lifetime)
	cache.Provide()
	cache.Provide()
	cache.Provide()
	cache.Provide()
	if provider.Count != 2 {
		t.Errorf("Unexpected number of Provide calls: %v", provider.Count)
	}
	time.Sleep(cache.Lifetime)
	cache.Provide()
	cache.Provide()
	cache.Provide()
	cache.Provide()
	if provider.Count != 3 {
		t.Errorf("Unexpected number of Provide calls: %v", provider.Count)
	}
}
