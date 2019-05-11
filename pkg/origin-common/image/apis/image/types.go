package image

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	DefaultImageTag											= "latest"
	ManagedByOpenShiftAnnotation							= "openshift.io/image.managed"
	InsecureRepositoryAnnotation							= "openshift.io/image.insecureRepository"
	DockerImageLayersOrderAnnotation						= "image.openshift.io/dockerLayersOrder"
	DockerImageLayersOrderAscending							= "ascending"
	ImageManifestBlobStoredAnnotation						= "image.openshift.io/manifestBlobStored"
	ImageSignatureTypeAtomicImageV1		string				= "AtomicImageV1"
	DockerImageLayersOrderDescending						= "descending"
	LimitTypeImage						corev1.LimitType	= "openshift.io/Image"
)

type Image struct {
	metav1.TypeMeta
	metav1.ObjectMeta
	DockerImageReference			string
	DockerImageMetadata				DockerImage
	DockerImageMetadataVersion		string
	DockerImageManifest				string
	DockerImageLayers				[]ImageLayer
	Signatures						[]ImageSignature
	DockerImageSignatures			[][]byte
	DockerImageManifestMediaType	string
	DockerImageConfig				string
}
type ImageLayer struct {
	Name		string
	LayerSize	int64
	MediaType	string
}
type ImageSignature struct {
	metav1.TypeMeta
	metav1.ObjectMeta
	Type			string
	Content			[]byte
	Conditions		[]SignatureCondition
	ImageIdentity	string
	SignedClaims	map[string]string
	Created			*metav1.Time
	IssuedBy		*SignatureIssuer
	IssuedTo		*SignatureSubject
}
type SignatureConditionType string
type SignatureCondition struct {
	Type				SignatureConditionType
	Status				corev1.ConditionStatus
	LastProbeTime		metav1.Time
	LastTransitionTime	metav1.Time
	Reason				string
	Message				string
}
type SignatureGenericEntity struct {
	Organization	string
	CommonName		string
}
type SignatureIssuer struct{ SignatureGenericEntity }
type SignatureSubject struct {
	SignatureGenericEntity
	PublicKeyID	string
}
type DockerImageReference struct {
	Registry	string
	Namespace	string
	Name		string
	Tag			string
	ID			string
}
