package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	kapierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	dcontext "github.com/docker/distribution/context"
	"github.com/docker/distribution/registry/api/errcode"
	"github.com/docker/distribution/registry/api/v2"
	"github.com/docker/distribution/registry/handlers"
	imageapiv1 "github.com/openshift/api/image/v1"
	"github.com/openshift/image-registry/pkg/dockerregistry/server/client"
	rerrors "github.com/openshift/image-registry/pkg/errors"
	imageapi "github.com/openshift/image-registry/pkg/origin-common/image/apis/image"
	gorillahandlers "github.com/gorilla/handlers"
)

const (
	errGroup		= "registry.api.v2"
	defaultSchemaVersion	= 2
)

type signature struct {
	Version	int	`json:"schemaVersion"`
	Name	string	`json:"name"`
	Type	string	`json:"type"`
	Content	[]byte	`json:"content"`
}
type signatureList struct {
	Signatures []signature `json:"signatures"`
}

var (
	ErrorCodeSignatureInvalid	= errcode.Register(errGroup, errcode.ErrorDescriptor{Value: "SIGNATURE_INVALID", Message: "invalid image signature", HTTPStatusCode: http.StatusBadRequest})
	ErrorCodeSignatureAlreadyExists	= errcode.Register(errGroup, errcode.ErrorDescriptor{Value: "SIGNATURE_EXISTS", Message: "image signature already exists", HTTPStatusCode: http.StatusConflict})
)

type signatureHandler struct {
	ctx		*handlers.Context
	reference	imageapi.DockerImageReference
	isImageClient	client.ImageStreamImagesNamespacer
}

func NewSignatureDispatcher(isImageClient client.ImageStreamImagesNamespacer) func(*handlers.Context, *http.Request) http.Handler {
	_logClusterCodePath()
	defer _logClusterCodePath()
	return func(ctx *handlers.Context, r *http.Request) http.Handler {
		reference, _ := imageapi.ParseDockerImageReference(dcontext.GetStringValue(ctx, "vars.name") + "@" + dcontext.GetStringValue(ctx, "vars.digest"))
		signatureHandler := &signatureHandler{ctx: ctx, isImageClient: isImageClient, reference: reference}
		return gorillahandlers.MethodHandler{"GET": http.HandlerFunc(signatureHandler.Get), "PUT": http.HandlerFunc(signatureHandler.Put)}
	}
}
func (s *signatureHandler) Put(w http.ResponseWriter, r *http.Request) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	dcontext.GetLogger(s.ctx).Debugf("(*signatureHandler).Put")
	if len(s.reference.String()) == 0 {
		s.handleError(s.ctx, v2.ErrorCodeNameInvalid.WithDetail("missing image name or image ID"), w)
		return
	}
	client, ok := userClientFrom(s.ctx)
	if !ok {
		s.handleError(s.ctx, errcode.ErrorCodeUnknown.WithDetail("unable to get origin client"), w)
		return
	}
	sig := signature{}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		s.handleError(s.ctx, ErrorCodeSignatureInvalid.WithDetail(err.Error()), w)
		return
	}
	if err := json.Unmarshal(body, &sig); err != nil {
		s.handleError(s.ctx, ErrorCodeSignatureInvalid.WithDetail(err.Error()), w)
		return
	}
	if len(sig.Type) == 0 {
		sig.Type = imageapi.ImageSignatureTypeAtomicImageV1
	}
	if sig.Version != defaultSchemaVersion {
		s.handleError(s.ctx, ErrorCodeSignatureInvalid.WithDetail(errors.New("only schemaVersion=2 is currently supported")), w)
		return
	}
	newSig := &imageapiv1.ImageSignature{Content: sig.Content, Type: sig.Type}
	newSig.Name = sig.Name
	_, err = client.ImageSignatures().Create(newSig)
	switch {
	case err == nil:
	case kapierrors.IsUnauthorized(err):
		s.handleError(s.ctx, errcode.ErrorCodeUnauthorized.WithDetail(err.Error()), w)
		return
	case kapierrors.IsBadRequest(err):
		s.handleError(s.ctx, ErrorCodeSignatureInvalid.WithDetail(err.Error()), w)
		return
	case kapierrors.IsNotFound(err):
		w.WriteHeader(http.StatusNotFound)
		return
	case kapierrors.IsAlreadyExists(err):
		s.handleError(s.ctx, ErrorCodeSignatureAlreadyExists.WithDetail(err.Error()), w)
		return
	default:
		s.handleError(s.ctx, errcode.ErrorCodeUnknown.WithDetail(fmt.Sprintf("unable to create image %s signature: %v", s.reference.String(), err)), w)
		return
	}
	w.WriteHeader(http.StatusCreated)
	dcontext.GetLogger(s.ctx).Debugf("(*signatureHandler).Put signature successfully added to %s", s.reference.String())
}
func (s *signatureHandler) Get(w http.ResponseWriter, req *http.Request) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	dcontext.GetLogger(s.ctx).Debugf("(*signatureHandler).Get")
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if len(s.reference.String()) == 0 {
		s.handleError(s.ctx, v2.ErrorCodeNameInvalid.WithDetail("missing image name or image ID"), w)
		return
	}
	if len(s.reference.ID) == 0 {
		s.handleError(s.ctx, v2.ErrorCodeNameInvalid.WithDetail("the image ID must be specified (sha256:<digest>"), w)
		return
	}
	image, err := s.isImageClient.ImageStreamImages(s.reference.Namespace).Get(imageapi.JoinImageStreamImage(s.reference.Name, s.reference.ID), metav1.GetOptions{})
	switch {
	case err == nil:
	case kapierrors.IsUnauthorized(err):
		s.handleError(s.ctx, errcode.ErrorCodeUnauthorized.WithDetail(fmt.Sprintf("not authorized to get image %q signature: %v", s.reference.String(), err)), w)
		return
	case kapierrors.IsNotFound(err):
		w.WriteHeader(http.StatusNotFound)
		return
	default:
		s.handleError(s.ctx, errcode.ErrorCodeUnknown.WithDetail(fmt.Sprintf("unable to get image %q signature: %v", s.reference.String(), err)), w)
		return
	}
	signatures := signatureList{Signatures: []signature{}}
	for _, s := range image.Image.Signatures {
		signatures.Signatures = append(signatures.Signatures, signature{Version: defaultSchemaVersion, Name: s.Name, Type: s.Type, Content: s.Content})
	}
	if data, err := json.Marshal(signatures); err != nil {
		s.handleError(s.ctx, errcode.ErrorCodeUnknown.WithDetail(fmt.Sprintf("failed to serialize image signature %v", err)), w)
	} else {
		_, _ = w.Write(data)
	}
}
func (s *signatureHandler) handleError(ctx context.Context, err error, w http.ResponseWriter) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	rerrors.Handle(ctx, "signature response completed with error", err)
	ctx, w = dcontext.WithResponseWriter(ctx, w)
	if serveErr := errcode.ServeJSON(w, err); serveErr != nil {
		dcontext.GetResponseLogger(ctx).Errorf("error sending error response: %v", serveErr)
		return
	}
}
