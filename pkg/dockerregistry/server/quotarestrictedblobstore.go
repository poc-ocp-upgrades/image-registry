package server

import (
	"context"
	"fmt"
	"github.com/docker/distribution"
	dcontext "github.com/docker/distribution/context"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"github.com/openshift/image-registry/pkg/dockerregistry/server/configuration"
	"github.com/openshift/image-registry/pkg/imagestream"
	imageapi "github.com/openshift/image-registry/pkg/origin-common/image/apis/image"
)

func newQuotaEnforcingConfig(ctx context.Context, quotaCfg *configuration.Quota) *quotaEnforcingConfig {
	_logClusterCodePath()
	defer _logClusterCodePath()
	if !quotaCfg.Enabled {
		dcontext.GetLogger(ctx).Info("quota enforcement disabled")
		return &quotaEnforcingConfig{}
	}
	if quotaCfg.CacheTTL <= 0 {
		dcontext.GetLogger(ctx).Info("not using project caches for quota objects")
		return &quotaEnforcingConfig{enforcementEnabled: true}
	}
	dcontext.GetLogger(ctx).Infof("caching project quota objects with TTL %s", quotaCfg.CacheTTL.String())
	return &quotaEnforcingConfig{enforcementEnabled: true, limitRanges: newProjectObjectListCache(quotaCfg.CacheTTL)}
}

type quotaEnforcingConfig struct {
	enforcementEnabled	bool
	limitRanges		imagestream.ProjectObjectListStore
}
type quotaRestrictedBlobStore struct {
	distribution.BlobStore
	repo	*repository
}

var _ distribution.BlobStore = &quotaRestrictedBlobStore{}

func (bs *quotaRestrictedBlobStore) Create(ctx context.Context, options ...distribution.BlobCreateOption) (distribution.BlobWriter, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	dcontext.GetLogger(ctx).Debug("(*quotaRestrictedBlobStore).Create: starting")
	bw, err := bs.BlobStore.Create(ctx, options...)
	if err != nil {
		return nil, err
	}
	repo := (*bs.repo)
	repo.ctx = ctx
	return &quotaRestrictedBlobWriter{BlobWriter: bw, repo: &repo}, nil
}
func (bs *quotaRestrictedBlobStore) Resume(ctx context.Context, id string) (distribution.BlobWriter, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	dcontext.GetLogger(ctx).Debug("(*quotaRestrictedBlobStore).Resume: starting")
	bw, err := bs.BlobStore.Resume(ctx, id)
	if err != nil {
		return nil, err
	}
	repo := (*bs.repo)
	repo.ctx = ctx
	return &quotaRestrictedBlobWriter{BlobWriter: bw, repo: &repo}, nil
}

type quotaRestrictedBlobWriter struct {
	distribution.BlobWriter
	repo	*repository
}

func (bw *quotaRestrictedBlobWriter) Commit(ctx context.Context, provisional distribution.Descriptor) (canonical distribution.Descriptor, err error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	dcontext.GetLogger(ctx).Debug("(*quotaRestrictedBlobWriter).Commit: starting")
	if err := admitBlobWrite(ctx, bw.repo, provisional.Size); err != nil {
		return distribution.Descriptor{}, err
	}
	return bw.BlobWriter.Commit(ctx, provisional)
}
func admitBlobWrite(ctx context.Context, repo *repository, size int64) error {
	_logClusterCodePath()
	defer _logClusterCodePath()
	if size < 1 {
		return nil
	}
	lrs, err := repo.imageStream.GetLimitRangeList(ctx, repo.app.quotaEnforcing.limitRanges)
	if err != nil {
		return err
	}
	for _, limitrange := range lrs.Items {
		dcontext.GetLogger(ctx).Debugf("processing limit range %s/%s", limitrange.Namespace, limitrange.Name)
		for _, limit := range limitrange.Spec.Limits {
			if err := admitImage(size, limit); err != nil {
				dcontext.GetLogger(ctx).Errorf("refusing to write blob exceeding limit range %s: %s", limitrange.Name, err.Error())
				return distribution.ErrAccessDenied
			}
		}
	}
	return nil
}
func admitImage(size int64, limit corev1.LimitRangeItem) error {
	_logClusterCodePath()
	defer _logClusterCodePath()
	if limit.Type != imageapi.LimitTypeImage {
		return nil
	}
	limitQuantity, ok := limit.Max[corev1.ResourceStorage]
	if !ok {
		return nil
	}
	imageQuantity := resource.NewQuantity(size, resource.BinarySI)
	if limitQuantity.Cmp(*imageQuantity) < 0 {
		return fmt.Errorf("requested usage of %s exceeds the maximum limit per %s (%s > %s)", corev1.ResourceStorage, imageapi.LimitTypeImage, imageQuantity.String(), limitQuantity.String())
	}
	return nil
}
