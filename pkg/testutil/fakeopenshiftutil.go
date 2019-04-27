package testutil

import (
	"fmt"
	"testing"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	imageapiv1 "github.com/openshift/api/image/v1"
	imageapi "github.com/openshift/image-registry/pkg/origin-common/image/apis/image"
)

func AddImageStream(t *testing.T, fos *FakeOpenShift, namespace, name string, annotations map[string]string) *imageapiv1.ImageStream {
	_logClusterCodePath()
	defer _logClusterCodePath()
	is := &imageapiv1.ImageStream{}
	is.Name = name
	is.Annotations = annotations
	is, err := fos.CreateImageStream(namespace, is)
	if err != nil {
		t.Fatal(err)
	}
	return is
}
func AddUntaggedImage(t *testing.T, fos *FakeOpenShift, image *imageapiv1.Image) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_, err := fos.CreateImage(image)
	if err != nil {
		t.Fatal(err)
	}
}
func AddImage(t *testing.T, fos *FakeOpenShift, image *imageapiv1.Image, namespace, name, tag string) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_, err := fos.CreateImageStreamMapping(namespace, &imageapiv1.ImageStreamMapping{ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: name}, Image: *image, Tag: tag})
	if err != nil {
		t.Fatal(err)
	}
}
func AddRandomImage(t *testing.T, fos *FakeOpenShift, namespace, name, tag string) *imageapiv1.Image {
	_logClusterCodePath()
	defer _logClusterCodePath()
	image, err := CreateRandomImage(namespace, name)
	if err != nil {
		t.Fatal(err)
	}
	_, err = fos.GetImageStream(namespace, name)
	if err != nil {
		AddImageStream(t, fos, namespace, name, map[string]string{imageapi.InsecureRepositoryAnnotation: "true"})
	}
	AddImage(t, fos, image, namespace, name, tag)
	return image
}
func AddImageStreamTag(t *testing.T, fos *FakeOpenShift, image *imageapiv1.Image, namespace, name string, tag *imageapiv1.TagReference) *imageapiv1.ImageStreamTag {
	_logClusterCodePath()
	defer _logClusterCodePath()
	istag, err := fos.CreateImageStreamTag(namespace, &imageapiv1.ImageStreamTag{ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("%s:%s", name, tag.Name)}, Tag: tag, Image: *image})
	if err != nil {
		t.Fatal(err)
	}
	return istag
}
