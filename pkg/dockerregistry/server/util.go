package server

import (
	"context"
	"fmt"
	"strings"
	"github.com/docker/distribution"
	dcontext "github.com/docker/distribution/context"
	"github.com/docker/distribution/manifest/schema2"
	"github.com/opencontainers/go-digest"
	dockerapiv10 "github.com/openshift/api/image/docker10"
	imageapiv1 "github.com/openshift/api/image/v1"
	"github.com/openshift/image-registry/pkg/dockerregistry/server/cache"
	"github.com/openshift/image-registry/pkg/dockerregistry/server/metrics"
	"github.com/openshift/image-registry/pkg/origin-common/image/registryclient"
	"github.com/openshift/image-registry/pkg/requesttrace"
)

func getNamespaceName(resourceName string) (string, string, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	repoParts := strings.Split(resourceName, "/")
	if len(repoParts) != 2 {
		return "", "", distribution.ErrRepositoryNameInvalid{Name: resourceName, Reason: fmt.Errorf("it must be of the format <project>/<name>")}
	}
	ns := repoParts[0]
	if len(ns) == 0 {
		return "", "", ErrNamespaceRequired
	}
	name := repoParts[1]
	if len(name) == 0 {
		return "", "", ErrNamespaceRequired
	}
	return ns, name, nil
}
func getImportContext(ctx context.Context, secretsGetter secretsGetter, m metrics.Pullthrough) (registryclient.RepositoryRetriever, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	req, err := dcontext.GetRequest(ctx)
	if err != nil {
		dcontext.GetLogger(ctx).Errorf("unable to get request from context: %v", err)
		return nil, err
	}
	secrets, err := secretsGetter()
	if err != nil {
		dcontext.GetLogger(ctx).Errorf("error getting secrets: %v", err)
	}
	credentials := registryclient.NewCredentialsForSecrets(secrets)
	var retriever registryclient.RepositoryRetriever
	retriever = registryclient.NewContext(secureTransport, insecureTransport, requesttrace.New(ctx, req)).WithCredentials(credentials)
	retriever = m.RepositoryRetriever(retriever)
	return retriever, nil
}
func RememberLayersOfImage(ctx context.Context, cache cache.RepositoryDigest, image *imageapiv1.Image, cacheName string) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	for _, layer := range image.DockerImageLayers {
		_ = cache.AddDigest(digest.Digest(layer.Name), cacheName)
	}
	meta, ok := image.DockerImageMetadata.Object.(*dockerapiv10.DockerImage)
	if !ok {
		dcontext.GetLogger(ctx).Errorf("image %s does not have metadata", image.Name)
		return
	}
	if image.DockerImageManifestMediaType == schema2.MediaTypeManifest && len(meta.ID) > 0 {
		_ = cache.AddDigest(digest.Digest(meta.ID), cacheName)
	}
}
func RememberLayersOfImageStream(ctx context.Context, cache cache.RepositoryDigest, layers *imageapiv1.ImageStreamLayers, cacheName string) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	for dgst := range layers.Blobs {
		_ = cache.AddDigest(digest.Digest(dgst), cacheName)
	}
	for dgst := range layers.Images {
		_ = cache.AddDigest(digest.Digest(dgst), cacheName)
	}
}
