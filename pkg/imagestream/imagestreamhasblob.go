package imagestream

import (
	"context"
	"time"
	dcontext "github.com/docker/distribution/context"
	"github.com/opencontainers/go-digest"
	imageapiv1 "github.com/openshift/api/image/v1"
)

func (is *imageStream) HasBlob(ctx context.Context, dgst digest.Digest) (bool, *imageapiv1.ImageStreamLayers, *imageapiv1.Image) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	dcontext.GetLogger(ctx).Debugf("verifying presence of blob %q in image stream %s", dgst.String(), is.Reference())
	started := time.Now()
	logFound := func(found bool, layers *imageapiv1.ImageStreamLayers, image *imageapiv1.Image) (bool, *imageapiv1.ImageStreamLayers, *imageapiv1.Image) {
		elapsed := time.Since(started)
		if found {
			dcontext.GetLogger(ctx).Debugf("verified presence of blob %q in image stream %s after %s", dgst.String(), is.Reference(), elapsed.String())
		} else {
			dcontext.GetLogger(ctx).Debugf("detected absence of blob %q in image stream %s after %s", dgst.String(), is.Reference(), elapsed.String())
		}
		return found, layers, image
	}
	layers, err := is.imageStreamGetter.layers()
	if err != nil {
		dcontext.GetLogger(ctx).Errorf("imageStream.HasBlob: failed to get image stream layers: %v", err)
		return logFound(false, nil, nil)
	}
	if _, ok := layers.Blobs[dgst.String()]; ok {
		return logFound(true, layers, nil)
	}
	if _, ok := layers.Images[dgst.String()]; ok {
		return logFound(true, layers, nil)
	}
	return logFound(false, layers, nil)
}
