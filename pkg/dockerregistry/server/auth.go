package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	dcontext "github.com/docker/distribution/context"
	registryauth "github.com/docker/distribution/registry/auth"
	authorizationapi "k8s.io/api/authorization/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	imageapi "github.com/openshift/api/image/v1"
	"github.com/openshift/image-registry/pkg/origin-common/util/httprequest"
	"github.com/openshift/image-registry/pkg/dockerregistry/server/audit"
	"github.com/openshift/image-registry/pkg/dockerregistry/server/client"
	"github.com/openshift/image-registry/pkg/dockerregistry/server/configuration"
)

type deferredErrors map[string]error

func (d deferredErrors) Add(ref string, err error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	d[ref] = err
}
func (d deferredErrors) Get(ref string) (error, bool) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	err, exists := d[ref]
	return err, exists
}
func (d deferredErrors) Empty() bool {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	return len(d) == 0
}

const (
	defaultUserName = "anonymous"
)

func WithUserInfoLogger(ctx context.Context, username, userid string) context.Context {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	ctx = context.WithValue(ctx, audit.AuditUserEntry, username)
	if len(userid) > 0 {
		ctx = context.WithValue(ctx, audit.AuditUserIDEntry, userid)
	}
	return dcontext.WithLogger(ctx, dcontext.GetLogger(ctx, audit.AuditUserEntry, audit.AuditUserIDEntry))
}

type AccessController struct {
	realm		string
	tokenRealm	*url.URL
	registryClient	client.RegistryClient
	auditLog	bool
	metricsConfig	configuration.Metrics
}

var _ registryauth.AccessController = &AccessController{}

type authChallenge struct {
	realm	string
	err	error
}

var _ registryauth.Challenge = &authChallenge{}

type tokenAuthChallenge struct {
	realm	string
	service	string
	err	error
}

var _ registryauth.Challenge = &tokenAuthChallenge{}
var (
	ErrTokenRequired		= errors.New("authorization header required")
	ErrTokenInvalid			= errors.New("failed to decode credentials")
	ErrOpenShiftAccessDenied	= errors.New("access denied")
	ErrNamespaceRequired		= errors.New("repository namespace required")
	ErrUnsupportedAction		= errors.New("unsupported action")
	ErrUnsupportedResource		= errors.New("unsupported resource")
)

func (app *App) Auth(options map[string]interface{}) (registryauth.AccessController, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	tokenRealm, err := configuration.TokenRealm(app.config.Auth.TokenRealm)
	if err != nil {
		return nil, err
	}
	return &AccessController{realm: app.config.Auth.Realm, tokenRealm: tokenRealm, registryClient: app.registryClient, metricsConfig: app.config.Metrics, auditLog: app.config.Audit.Enabled}, nil
}
func (ac *authChallenge) Error() string {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	return ac.err.Error()
}
func (ac *authChallenge) SetHeaders(w http.ResponseWriter) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	str := fmt.Sprintf("Basic realm=%s", ac.realm)
	if ac.err != nil {
		str = fmt.Sprintf("%s,error=%q", str, ac.Error())
	}
	w.Header().Set("WWW-Authenticate", str)
}
func (ac *tokenAuthChallenge) Error() string {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	return ac.err.Error()
}
func (ac *tokenAuthChallenge) SetHeaders(w http.ResponseWriter) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	str := fmt.Sprintf("Bearer realm=%q", ac.realm)
	if ac.service != "" {
		str += fmt.Sprintf(",service=%q", ac.service)
	}
	w.Header().Set("WWW-Authenticate", str)
}
func (ac *AccessController) wrapErr(ctx context.Context, err error) error {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	switch err {
	case ErrTokenRequired:
		if ac.tokenRealm == nil {
			return &authChallenge{realm: ac.realm, err: err}
		}
		if len(ac.tokenRealm.Scheme) > 0 && len(ac.tokenRealm.Host) > 0 {
			return &tokenAuthChallenge{realm: ac.tokenRealm.String(), err: err}
		}
		req, reqErr := dcontext.GetRequest(ctx)
		if reqErr != nil {
			return reqErr
		}
		scheme, host := httprequest.SchemeHost(req)
		tokenRealmCopy := *ac.tokenRealm
		if len(tokenRealmCopy.Scheme) == 0 {
			tokenRealmCopy.Scheme = scheme
		}
		if len(tokenRealmCopy.Host) == 0 {
			tokenRealmCopy.Host = host
		}
		return &tokenAuthChallenge{realm: tokenRealmCopy.String(), err: err}
	case ErrTokenInvalid, ErrOpenShiftAccessDenied:
		return &authChallenge{realm: ac.realm, err: err}
	default:
		return err
	}
}
func (ac *AccessController) Authorized(ctx context.Context, accessRecords ...registryauth.Access) (context.Context, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	req, err := dcontext.GetRequest(ctx)
	if err != nil {
		return nil, ac.wrapErr(ctx, err)
	}
	bearerToken, err := getOpenShiftAPIToken(req)
	if err != nil {
		return nil, ac.wrapErr(ctx, err)
	}
	osClient, err := ac.registryClient.ClientFromToken(bearerToken)
	if err != nil {
		return nil, ac.wrapErr(ctx, err)
	}
	if len(bearerToken) > 0 && !isMetricsBearerToken(ac.metricsConfig, bearerToken) {
		user, userid, err := verifyOpenShiftUser(ctx, osClient)
		if err != nil {
			return nil, ac.wrapErr(ctx, err)
		}
		ctx = WithUserInfoLogger(ctx, user, userid)
	} else {
		ctx = WithUserInfoLogger(ctx, defaultUserName, "")
	}
	if ac.auditLog {
		ctx = audit.WithLogger(ctx, audit.GetLogger(ctx))
	}
	pushChecks := map[string]bool{}
	possibleCrossMountErrors := deferredErrors{}
	verifiedPrune := false
	for _, access := range accessRecords {
		dcontext.GetLogger(ctx).Debugf("Origin auth: checking for access to %s:%s:%s", access.Resource.Type, access.Resource.Name, access.Action)
		switch access.Resource.Type {
		case "repository":
			imageStreamNS, imageStreamName, err := getNamespaceName(access.Resource.Name)
			if err != nil {
				return nil, ac.wrapErr(ctx, err)
			}
			verb := ""
			switch access.Action {
			case "push":
				verb = "update"
				pushChecks[imageStreamNS+"/"+imageStreamName] = true
			case "pull":
				verb = "get"
			case "delete":
				if strings.Contains(req.URL.Path, "/blobs/uploads/") {
					verb = "update"
				} else {
					if !verifiedPrune {
						if err := verifyPruneAccess(ctx, osClient); err != nil {
							return nil, ac.wrapErr(ctx, err)
						}
						verifiedPrune = true
					}
					continue
				}
			default:
				return nil, ac.wrapErr(ctx, ErrUnsupportedAction)
			}
			if err := verifyImageStreamAccess(ctx, imageStreamNS, imageStreamName, verb, osClient); err != nil {
				if access.Action != "pull" {
					return nil, ac.wrapErr(ctx, err)
				}
				possibleCrossMountErrors.Add(imageStreamNS+"/"+imageStreamName, ac.wrapErr(ctx, err))
			}
		case "signature":
			namespace, name, err := getNamespaceName(access.Resource.Name)
			if err != nil {
				return nil, ac.wrapErr(ctx, err)
			}
			switch access.Action {
			case "get":
				if err := verifyImageStreamAccess(ctx, namespace, name, access.Action, osClient); err != nil {
					return nil, ac.wrapErr(ctx, err)
				}
			case "put":
				if err := verifyImageSignatureAccess(ctx, namespace, name, osClient); err != nil {
					return nil, ac.wrapErr(ctx, err)
				}
			default:
				return nil, ac.wrapErr(ctx, ErrUnsupportedAction)
			}
		case "metrics":
			switch access.Action {
			case "get":
				if err := verifyMetricsAccess(ctx, ac.metricsConfig, bearerToken, osClient); err != nil {
					return nil, ac.wrapErr(ctx, err)
				}
			default:
				return nil, ac.wrapErr(ctx, ErrUnsupportedAction)
			}
		case "admin":
			switch access.Action {
			case "prune":
				if verifiedPrune {
					continue
				}
				if err := verifyPruneAccess(ctx, osClient); err != nil {
					return nil, ac.wrapErr(ctx, err)
				}
				verifiedPrune = true
			default:
				return nil, ac.wrapErr(ctx, ErrUnsupportedAction)
			}
		case "registry":
			switch access.Resource.Name {
			case "catalog":
				if access.Action != "*" {
					return nil, ac.wrapErr(ctx, ErrUnsupportedAction)
				}
				if err := verifyCatalogAccess(ctx, osClient); err != nil {
					return nil, ac.wrapErr(ctx, err)
				}
			default:
				return nil, ac.wrapErr(ctx, ErrUnsupportedResource)
			}
		default:
			return nil, ac.wrapErr(ctx, ErrUnsupportedResource)
		}
	}
	for namespaceAndName, err := range possibleCrossMountErrors {
		if len(pushChecks) == 0 {
			return nil, err
		}
		if pushChecks[namespaceAndName] {
			return nil, err
		}
	}
	if !possibleCrossMountErrors.Empty() {
		dcontext.GetLogger(ctx).Debugf("Origin auth: deferring errors: %#v", possibleCrossMountErrors)
		ctx = withDeferredErrors(ctx, possibleCrossMountErrors)
	}
	ctx = withAuthPerformed(ctx)
	return withUserClient(ctx, osClient), nil
}
func getOpenShiftAPIToken(req *http.Request) (string, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	token := ""
	authParts := strings.SplitN(req.Header.Get("Authorization"), " ", 2)
	if len(authParts) != 2 {
		return "", ErrTokenRequired
	}
	switch strings.ToLower(authParts[0]) {
	case "bearer":
		token = authParts[1]
		if token == anonymousToken {
			token = ""
		}
	case "basic":
		_, password, ok := req.BasicAuth()
		if !ok || len(password) == 0 {
			return "", ErrTokenInvalid
		}
		token = password
	default:
		return "", ErrTokenRequired
	}
	return token, nil
}
func verifyOpenShiftUser(ctx context.Context, c client.UsersInterfacer) (string, string, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	userInfo, err := c.Users().Get("~", metav1.GetOptions{})
	if err != nil {
		dcontext.GetLogger(ctx).Errorf("Get user failed with error: %s", err)
		if kerrors.IsUnauthorized(err) || kerrors.IsForbidden(err) {
			return "", "", ErrOpenShiftAccessDenied
		}
		return "", "", err
	}
	return userInfo.GetName(), string(userInfo.GetUID()), nil
}
func verifyWithSAR(ctx context.Context, resource, namespace, name, verb string, c client.SelfSubjectAccessReviewsNamespacer) error {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	sar := authorizationapi.SelfSubjectAccessReview{Spec: authorizationapi.SelfSubjectAccessReviewSpec{ResourceAttributes: &authorizationapi.ResourceAttributes{Namespace: namespace, Verb: verb, Group: imageapi.GroupName, Resource: resource, Name: name}}}
	response, err := c.SelfSubjectAccessReviews().Create(&sar)
	if err != nil {
		dcontext.GetLogger(ctx).Errorf("OpenShift client error: %s", err)
		if kerrors.IsUnauthorized(err) || kerrors.IsForbidden(err) {
			return ErrOpenShiftAccessDenied
		}
		return err
	}
	if !response.Status.Allowed {
		dcontext.GetLogger(ctx).Errorf("OpenShift access denied: %s", response.Status.Reason)
		return ErrOpenShiftAccessDenied
	}
	return nil
}
func verifyWithGlobalSAR(ctx context.Context, resource, subresource, verb string, c client.SelfSubjectAccessReviewsNamespacer) error {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	sar := authorizationapi.SelfSubjectAccessReview{Spec: authorizationapi.SelfSubjectAccessReviewSpec{ResourceAttributes: &authorizationapi.ResourceAttributes{Verb: verb, Group: imageapi.GroupName, Resource: resource, Subresource: subresource}}}
	response, err := c.SelfSubjectAccessReviews().Create(&sar)
	if err != nil {
		dcontext.GetLogger(ctx).Errorf("OpenShift client error: %s", err)
		if kerrors.IsUnauthorized(err) || kerrors.IsForbidden(err) {
			return ErrOpenShiftAccessDenied
		}
		return err
	}
	if !response.Status.Allowed {
		dcontext.GetLogger(ctx).Errorf("OpenShift access denied: %s", response.Status.Reason)
		return ErrOpenShiftAccessDenied
	}
	return nil
}
func verifyImageStreamAccess(ctx context.Context, namespace, imageRepo, verb string, c client.SelfSubjectAccessReviewsNamespacer) error {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	return verifyWithSAR(ctx, "imagestreams/layers", namespace, imageRepo, verb, c)
}
func verifyImageSignatureAccess(ctx context.Context, namespace, imageRepo string, c client.SelfSubjectAccessReviewsNamespacer) error {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	return verifyWithSAR(ctx, "imagesignatures", namespace, imageRepo, "create", c)
}
func verifyPruneAccess(ctx context.Context, c client.SelfSubjectAccessReviewsNamespacer) error {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	return verifyWithGlobalSAR(ctx, "images", "", "delete", c)
}
func verifyCatalogAccess(ctx context.Context, c client.SelfSubjectAccessReviewsNamespacer) error {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	return verifyWithGlobalSAR(ctx, "imagestreams", "", "list", c)
}
func verifyMetricsAccess(ctx context.Context, metrics configuration.Metrics, token string, c client.SelfSubjectAccessReviewsNamespacer) error {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	if !metrics.Enabled {
		return ErrOpenShiftAccessDenied
	}
	if len(metrics.Secret) > 0 {
		if metrics.Secret != token {
			return ErrOpenShiftAccessDenied
		}
		return nil
	}
	if err := verifyWithGlobalSAR(ctx, "registry", "metrics", "get", c); err != nil {
		return err
	}
	return nil
}
func isMetricsBearerToken(metrics configuration.Metrics, token string) bool {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	if metrics.Enabled {
		return metrics.Secret == token
	}
	return false
}
