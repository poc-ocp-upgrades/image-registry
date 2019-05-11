package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	dcontext "github.com/docker/distribution/context"
	"github.com/openshift/image-registry/pkg/dockerregistry/server/auth"
	"github.com/openshift/image-registry/pkg/dockerregistry/server/client"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type tokenHandler struct {
	ctx		context.Context
	client	client.RegistryClient
}

func NewTokenHandler(ctx context.Context, client client.RegistryClient) http.Handler {
	_logClusterCodePath()
	defer _logClusterCodePath()
	return &tokenHandler{ctx: ctx, client: client}
}

const anonymousToken = "anonymous"

func (t *tokenHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	ctx := dcontext.WithRequest(t.ctx, req)
	params := req.URL.Query()
	if len(params.Get("scope")) > 0 {
		accessRecords := auth.ResolveScopeSpecifiers(ctx, params["scope"])
		for _, access := range accessRecords {
			switch access.Resource.Type {
			case "repository", "signature":
				_, _, err := getNamespaceName(access.Resource.Name)
				if err != nil {
					dcontext.GetRequestLogger(ctx).Errorf("auth token request for unsupported resource name: %s", access.Resource.Name)
					t.writeError(w, req, err.Error())
					return
				}
			}
		}
	}
	if len(req.Header.Get("Authorization")) == 0 {
		dcontext.GetRequestLogger(ctx).Debugf("anonymous token request")
		t.writeToken(anonymousToken, w, req)
		return
	}
	_, token, ok := req.BasicAuth()
	if !ok {
		dcontext.GetRequestLogger(ctx).Debugf("no basic auth credentials provided")
		t.writeUnauthorized(w, req)
		return
	}
	osClient, err := t.client.ClientFromToken(token)
	if err != nil {
		dcontext.GetRequestLogger(ctx).Errorf("error building client: %v", err)
		t.writeError(w, req, "invalid request")
		return
	}
	if _, err := osClient.Users().Get("~", metav1.GetOptions{}); err != nil {
		dcontext.GetRequestLogger(ctx).Errorf("invalid token: %v", err)
		if kerrors.IsUnauthorized(err) {
			t.writeUnauthorized(w, req)
		} else {
			msg := "unable to validate token"
			if reason := kerrors.ReasonForError(err); reason != metav1.StatusReasonUnknown {
				msg = fmt.Sprintf("%s: %s", msg, reason)
			}
			t.writeError(w, req, msg)
		}
		return
	}
	t.writeToken(token, w, req)
}
func (t *tokenHandler) writeError(w http.ResponseWriter, req *http.Request, msg string) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(401)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"details": msg})
}
func (t *tokenHandler) writeToken(token string, w http.ResponseWriter, req *http.Request) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"token": token, "access_token": token})
}
func (t *tokenHandler) writeUnauthorized(w http.ResponseWriter, req *http.Request) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(401)
}
