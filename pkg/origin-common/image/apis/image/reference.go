package image

import (
	"fmt"
	"strings"
	"github.com/docker/distribution/reference"
)

type namedDockerImageReference struct {
	Registry	string
	Namespace	string
	Name		string
	Tag			string
	ID			string
}

func parseNamedDockerImageReference(spec string) (namedDockerImageReference, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	var ref namedDockerImageReference
	parsedRef, err := reference.Parse(spec)
	if err != nil {
		return ref, err
	}
	namedRef, isNamed := parsedRef.(reference.Named)
	if !isNamed {
		return ref, fmt.Errorf("reference %s has no name", parsedRef.String())
	}
	name := namedRef.Name()
	i := strings.IndexRune(name, '/')
	if i == -1 || (!strings.ContainsAny(name[:i], ":.") && name[:i] != "localhost") {
		ref.Name = name
	} else {
		ref.Registry, ref.Name = name[:i], name[i+1:]
	}
	if named, ok := namedRef.(reference.NamedTagged); ok {
		ref.Tag = named.Tag()
	}
	if named, ok := namedRef.(reference.Canonical); ok {
		ref.ID = named.Digest().String()
	}
	if i := strings.IndexRune(ref.Name, '/'); i != -1 {
		ref.Namespace, ref.Name = ref.Name[:i], ref.Name[i+1:]
	}
	return ref, nil
}
