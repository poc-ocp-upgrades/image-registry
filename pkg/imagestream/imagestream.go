package imagestream

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	dcontext "github.com/docker/distribution/context"
	"github.com/opencontainers/go-digest"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	imageapiv1 "github.com/openshift/api/image/v1"
	"github.com/openshift/image-registry/pkg/dockerregistry/server/client"
	rerrors "github.com/openshift/image-registry/pkg/errors"
	imageapi "github.com/openshift/image-registry/pkg/origin-common/image/apis/image"
	quotautil "github.com/openshift/image-registry/pkg/origin-common/quota/util"
	originutil "github.com/openshift/image-registry/pkg/origin-common/util"
)

const (
	ErrImageStreamCode		= "ImageStream:"
	ErrImageStreamUnknownErrorCode	= ErrImageStreamCode + "Unknown"
	ErrImageStreamNotFoundCode	= ErrImageStreamCode + "NotFound"
	ErrImageStreamImageNotFoundCode	= ErrImageStreamCode + "ImageNotFound"
	ErrImageStreamForbiddenCode	= ErrImageStreamCode + "Forbidden"
)

type ProjectObjectListStore interface {
	Add(namespace string, obj runtime.Object) error
	Get(namespace string) (obj runtime.Object, exists bool, err error)
}
type ImagePullthroughSpec struct {
	DockerImageReference	*imageapi.DockerImageReference
	Insecure		bool
}
type ImageStream interface {
	Reference() string
	Exists(ctx context.Context) (bool, *rerrors.Error)
	GetImageOfImageStream(ctx context.Context, dgst digest.Digest) (*imageapiv1.Image, *rerrors.Error)
	CreateImageStreamMapping(ctx context.Context, userClient client.Interface, tag string, image *imageapiv1.Image) *rerrors.Error
	ResolveImageID(ctx context.Context, dgst digest.Digest) (*imageapiv1.TagEvent, *rerrors.Error)
	HasBlob(ctx context.Context, dgst digest.Digest) (bool, *imageapiv1.ImageStreamLayers, *imageapiv1.Image)
	IdentifyCandidateRepositories(ctx context.Context, primary bool) ([]string, map[string]ImagePullthroughSpec, *rerrors.Error)
	GetLimitRangeList(ctx context.Context, cache ProjectObjectListStore) (*corev1.LimitRangeList, *rerrors.Error)
	GetSecrets() ([]corev1.Secret, *rerrors.Error)
	TagIsInsecure(ctx context.Context, tag string, dgst digest.Digest) (bool, *rerrors.Error)
	Tags(ctx context.Context) (map[string]digest.Digest, *rerrors.Error)
}
type imageStream struct {
	namespace		string
	name			string
	registryOSClient	client.Interface
	imageClient		imageGetter
	imageStreamGetter	*cachedImageStreamGetter
}

var _ ImageStream = &imageStream{}

func New(ctx context.Context, namespace, name string, client client.Interface) ImageStream {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	return &imageStream{namespace: namespace, name: name, registryOSClient: client, imageClient: newCachedImageGetter(client), imageStreamGetter: &cachedImageStreamGetter{namespace: namespace, name: name, isNamespacer: client}}
}
func (is *imageStream) Reference() string {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	return fmt.Sprintf("%s/%s", is.namespace, is.name)
}
func (is *imageStream) getImage(ctx context.Context, dgst digest.Digest) (*imageapiv1.Image, *rerrors.Error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	image, err := is.imageClient.Get(ctx, dgst)
	switch {
	case kerrors.IsNotFound(err):
		return nil, rerrors.NewError(ErrImageStreamImageNotFoundCode, fmt.Sprintf("getImage: unable to find image digest %s in %s", dgst.String(), is.name), err)
	case err != nil:
		return nil, rerrors.NewError(ErrImageStreamUnknownErrorCode, fmt.Sprintf("getImage: unable to get image digest %s in %s", dgst.String(), is.name), err)
	}
	return image, nil
}
func (is *imageStream) ResolveImageID(ctx context.Context, dgst digest.Digest) (*imageapiv1.TagEvent, *rerrors.Error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	stream, rErr := is.imageStreamGetter.get()
	if rErr != nil {
		return nil, convertImageStreamGetterError(rErr, fmt.Sprintf("ResolveImageID: failed to get image stream %s", is.Reference()))
	}
	tagEvent, err := originutil.ResolveImageID(stream, dgst.String())
	if err != nil {
		code := ErrImageStreamUnknownErrorCode
		if kerrors.IsNotFound(err) {
			code = ErrImageStreamImageNotFoundCode
		}
		return nil, rerrors.NewError(code, fmt.Sprintf("ResolveImageID: unable to resolve ImageID %s in image stream %s", dgst.String(), is.Reference()), err)
	}
	return tagEvent, nil
}
func (is *imageStream) getStoredImageOfImageStream(ctx context.Context, dgst digest.Digest) (*imageapiv1.Image, *imageapiv1.TagEvent, *rerrors.Error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	tagEvent, err := is.ResolveImageID(ctx, dgst)
	if err != nil {
		return nil, nil, err
	}
	image, err := is.getImage(ctx, dgst)
	if err != nil {
		return nil, nil, err
	}
	return image, tagEvent, nil
}
func (is *imageStream) GetImageOfImageStream(ctx context.Context, dgst digest.Digest) (*imageapiv1.Image, *rerrors.Error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	image, tagEvent, err := is.getStoredImageOfImageStream(ctx, dgst)
	if err != nil {
		return nil, err
	}
	img := *image
	img.DockerImageReference = tagEvent.DockerImageReference
	return &img, nil
}
func (is *imageStream) GetSecrets() ([]corev1.Secret, *rerrors.Error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	secrets, err := is.registryOSClient.ImageStreamSecrets(is.namespace).Secrets(is.name, metav1.GetOptions{})
	if err != nil {
		return nil, rerrors.NewError(ErrImageStreamUnknownErrorCode, fmt.Sprintf("GetSecrets: error getting secrets for repository %s", is.Reference()), err)
	}
	return secrets.Items, nil
}
func (is *imageStream) TagIsInsecure(ctx context.Context, tag string, dgst digest.Digest) (bool, *rerrors.Error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	stream, err := is.imageStreamGetter.get()
	if err != nil {
		return false, convertImageStreamGetterError(err, fmt.Sprintf("TagIsInsecure: failed to get image stream %s", is.Reference()))
	}
	if insecure, _ := stream.Annotations[imageapi.InsecureRepositoryAnnotation]; insecure == "true" {
		return true, nil
	}
	if len(tag) == 0 {
		tag, _ = originutil.LatestImageTagEvent(stream, dgst.String())
	}
	if len(tag) != 0 {
		for _, t := range stream.Spec.Tags {
			if t.Name == tag {
				return t.ImportPolicy.Insecure, nil
			}
		}
	}
	return false, nil
}
func (is *imageStream) Exists(ctx context.Context) (bool, *rerrors.Error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	_, rErr := is.imageStreamGetter.get()
	if rErr != nil {
		if rErr.Code == ErrImageStreamGetterNotFoundCode {
			return false, nil
		}
		return false, convertImageStreamGetterError(rErr, fmt.Sprintf("Exists: failed to get image stream %s", is.Reference()))
	}
	return true, nil
}
func (is *imageStream) localRegistry(ctx context.Context) ([]string, *rerrors.Error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	stream, rErr := is.imageStreamGetter.get()
	if rErr != nil {
		return nil, convertImageStreamGetterError(rErr, fmt.Sprintf("localRegistry: failed to get image stream %s", is.Reference()))
	}
	var localNames []string
	local, err := imageapi.ParseDockerImageReference(stream.Status.DockerImageRepository)
	if err != nil {
		dcontext.GetLogger(ctx).Warnf("localRegistry: unable to parse dockerImageRepository %q", stream.Status.DockerImageRepository)
	}
	if len(local.Registry) != 0 {
		localNames = append(localNames, local.Registry)
	}
	public, err := imageapi.ParseDockerImageReference(stream.Status.PublicDockerImageRepository)
	if err != nil {
		dcontext.GetLogger(ctx).Warnf("localRegistry: unable to parse publicDockerImageRepository %q", stream.Status.DockerImageRepository)
	}
	if len(public.Registry) != 0 {
		localNames = append(localNames, public.Registry)
	}
	return localNames, nil
}
func (is *imageStream) IdentifyCandidateRepositories(ctx context.Context, primary bool) ([]string, map[string]ImagePullthroughSpec, *rerrors.Error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	stream, err := is.imageStreamGetter.get()
	if err != nil {
		return nil, nil, convertImageStreamGetterError(err, fmt.Sprintf("IdentifyCandidateRepositories: failed to get image stream %s", is.Reference()))
	}
	localRegistry, _ := is.localRegistry(ctx)
	repositoryCandidates, search := identifyCandidateRepositories(stream, localRegistry, primary)
	return repositoryCandidates, search, nil
}
func (is *imageStream) Tags(ctx context.Context) (map[string]digest.Digest, *rerrors.Error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	stream, err := is.imageStreamGetter.get()
	if err != nil {
		return nil, convertImageStreamGetterError(err, fmt.Sprintf("Tags: failed to get image stream %s", is.Reference()))
	}
	m := make(map[string]digest.Digest)
	for _, history := range stream.Status.Tags {
		if len(history.Items) == 0 {
			continue
		}
		tag := history.Tag
		dgst, err := digest.Parse(history.Items[0].Image)
		if err != nil {
			dcontext.GetLogger(ctx).Errorf("bad digest %s: %v", history.Items[0].Image, err)
			continue
		}
		m[tag] = dgst
	}
	return m, nil
}
func (is *imageStream) CreateImageStreamMapping(ctx context.Context, userClient client.Interface, tag string, image *imageapiv1.Image) *rerrors.Error {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	ism := imageapiv1.ImageStreamMapping{ObjectMeta: metav1.ObjectMeta{Namespace: is.namespace, Name: is.name}, Image: *image, Tag: tag}
	_, err := is.registryOSClient.ImageStreamMappings(is.namespace).Create(&ism)
	if err == nil {
		return nil
	}
	if quotautil.IsErrorQuotaExceeded(err) {
		return rerrors.NewError(ErrImageStreamForbiddenCode, fmt.Sprintf("CreateImageStreamMapping: quota exceeded during creation of %s ImageStreamMapping", is.Reference()), err)
	}
	statusErr, ok := err.(*kerrors.StatusError)
	if !ok {
		return rerrors.NewError(ErrImageStreamUnknownErrorCode, fmt.Sprintf("CreateImageStreamMapping: error creating %s ImageStreamMapping", is.Reference()), err)
	}
	status := statusErr.ErrStatus
	isValidKind := false
	if status.Details != nil {
		switch strings.ToLower(status.Details.Kind) {
		case "imagestream", "imagestreams", "imagestreammappings":
			isValidKind = true
		}
	}
	if !isValidKind || status.Code != http.StatusNotFound || status.Details.Name != is.name {
		return rerrors.NewError(ErrImageStreamUnknownErrorCode, fmt.Sprintf("CreateImageStreamMapping: error creation of %s ImageStreamMapping", is.Reference()), err)
	}
	stream := &imageapiv1.ImageStream{}
	stream.Name = is.name
	_, err = userClient.ImageStreams(is.namespace).Create(stream)
	switch {
	case kerrors.IsAlreadyExists(err), kerrors.IsConflict(err):
	case kerrors.IsForbidden(err), kerrors.IsUnauthorized(err), quotautil.IsErrorQuotaExceeded(err):
		return rerrors.NewError(ErrImageStreamForbiddenCode, fmt.Sprintf("CreateImageStreamMapping: denied creating ImageStream %s", is.Reference()), err)
	case err != nil:
		return rerrors.NewError(ErrImageStreamUnknownErrorCode, fmt.Sprintf("CreateImageStreamMapping: error auto provisioning ImageStream %s", is.Reference()), err)
	}
	dcontext.GetLogger(ctx).Debugf("cache image stream %s/%s", stream.Namespace, stream.Name)
	is.imageStreamGetter.cacheImageStream(stream)
	_, err = is.registryOSClient.ImageStreamMappings(is.namespace).Create(&ism)
	if err == nil {
		return nil
	}
	if quotautil.IsErrorQuotaExceeded(err) {
		return rerrors.NewError(ErrImageStreamForbiddenCode, fmt.Sprintf("CreateImageStreamMapping: quota exceeded during creation of %s ImageStreamMapping second time", is.Reference()), err)
	}
	return rerrors.NewError(ErrImageStreamUnknownErrorCode, fmt.Sprintf("CreateImageStreamMapping: error creating %s ImageStreamMapping second time", is.Reference()), err)
}
func (is *imageStream) GetLimitRangeList(ctx context.Context, cache ProjectObjectListStore) (*corev1.LimitRangeList, *rerrors.Error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	if cache != nil {
		obj, exists, _ := cache.Get(is.namespace)
		if exists {
			return obj.(*corev1.LimitRangeList), nil
		}
	}
	dcontext.GetLogger(ctx).Debugf("listing limit ranges in namespace %s", is.namespace)
	lrs, err := is.registryOSClient.LimitRanges(is.namespace).List(metav1.ListOptions{})
	if err != nil {
		return nil, rerrors.NewError(ErrImageStreamUnknownErrorCode, fmt.Sprintf("GetLimitRangeList: failed to list limitranges for %s", is.Reference()), err)
	}
	if cache != nil {
		err = cache.Add(is.namespace, lrs)
		if err != nil {
			dcontext.GetLogger(ctx).Errorf("GetLimitRangeList: failed to cache limit range list: %v", err)
		}
	}
	return lrs, nil
}
func convertImageStreamGetterError(err *rerrors.Error, msg string) *rerrors.Error {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	code := ErrImageStreamUnknownErrorCode
	switch err.Code {
	case ErrImageStreamGetterNotFoundCode:
		code = ErrImageStreamNotFoundCode
	case ErrImageStreamGetterForbiddenCode:
		code = ErrImageStreamForbiddenCode
	}
	return rerrors.NewError(code, msg, err)
}
