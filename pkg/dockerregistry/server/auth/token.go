package auth

import (
	"context"
	"regexp"
	"strings"
	dcontext "github.com/docker/distribution/context"
	"github.com/docker/distribution/registry/auth"
)

func ResolveScopeSpecifiers(ctx context.Context, scopeSpecs []string) []auth.Access {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	requestedAccessSet := make(map[auth.Access]struct{}, 2*len(scopeSpecs))
	requestedAccessList := make([]auth.Access, 0, len(requestedAccessSet))
	for _, scopeSpecifier := range scopeSpecs {
		parts := strings.SplitN(scopeSpecifier, ":", 4)
		if len(parts) != 3 {
			dcontext.GetLogger(ctx).Infof("ignoring unsupported scope format %s", scopeSpecifier)
			continue
		}
		resourceType, resourceClass := splitResourceClass(parts[0])
		if resourceType == "" {
			continue
		}
		resourceName, actions := parts[1], parts[2]
		for _, action := range strings.Split(actions, ",") {
			requestedAccess := auth.Access{Resource: auth.Resource{Type: resourceType, Class: resourceClass, Name: resourceName}, Action: action}
			if _, ok := requestedAccessSet[requestedAccess]; !ok {
				requestedAccessList = append(requestedAccessList, requestedAccess)
				requestedAccessSet[requestedAccess] = struct{}{}
			}
		}
	}
	return requestedAccessList
}

var typeRegexp = regexp.MustCompile(`^([a-z0-9]+)(?:\(([a-z0-9]+)\))?$`)

func splitResourceClass(t string) (string, string) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	matches := typeRegexp.FindStringSubmatch(t)
	if matches != nil {
		return matches[1], matches[2]
	}
	return "", ""
}
