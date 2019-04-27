package image

import (
	"fmt"
	"net/url"
	"strings"
	"github.com/opencontainers/go-digest"
)

const (
	DockerDefaultNamespace	= "library"
	DockerDefaultRegistry	= "docker.io"
	DockerDefaultV1Registry	= "index." + DockerDefaultRegistry
	DockerDefaultV2Registry	= "registry-1." + DockerDefaultRegistry
)

func ParseImageStreamImageName(input string) (name string, id string, err error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	segments := strings.SplitN(input, "@", 3)
	switch len(segments) {
	case 2:
		name = segments[0]
		id = segments[1]
		if len(name) == 0 || len(id) == 0 {
			err = fmt.Errorf("image stream image name %q must have a name and ID", input)
		}
	default:
		err = fmt.Errorf("expected exactly one @ in the isimage name %q", input)
	}
	return
}
func IsRegistryDockerHub(registry string) bool {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	switch registry {
	case DockerDefaultRegistry, DockerDefaultV1Registry, DockerDefaultV2Registry:
		return true
	default:
		return false
	}
}
func ParseDockerImageReference(spec string) (DockerImageReference, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	var ref DockerImageReference
	namedRef, err := parseNamedDockerImageReference(spec)
	if err != nil {
		return ref, err
	}
	ref.Registry = namedRef.Registry
	ref.Namespace = namedRef.Namespace
	ref.Name = namedRef.Name
	ref.Tag = namedRef.Tag
	ref.ID = namedRef.ID
	return ref, nil
}
func (r DockerImageReference) DockerClientDefaults() DockerImageReference {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	if len(r.Registry) == 0 {
		r.Registry = DockerDefaultRegistry
	}
	if len(r.Namespace) == 0 && IsRegistryDockerHub(r.Registry) {
		r.Namespace = DockerDefaultNamespace
	}
	if len(r.Tag) == 0 {
		r.Tag = DefaultImageTag
	}
	return r
}
func (r DockerImageReference) AsRepository() DockerImageReference {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	r.Tag = ""
	r.ID = ""
	return r
}
func (r DockerImageReference) RepositoryName() string {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	r.Tag = ""
	r.ID = ""
	r.Registry = ""
	return r.Exact()
}
func (r DockerImageReference) RegistryURL() *url.URL {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	return &url.URL{Scheme: "https", Host: r.AsV2().Registry}
}
func (r DockerImageReference) AsV2() DockerImageReference {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	switch r.Registry {
	case DockerDefaultV1Registry, DockerDefaultRegistry:
		r.Registry = DockerDefaultV2Registry
	}
	return r
}
func (r DockerImageReference) NameString() string {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	switch {
	case len(r.Name) == 0:
		return ""
	case len(r.Tag) > 0:
		return r.Name + ":" + r.Tag
	case len(r.ID) > 0:
		var ref string
		if _, err := digest.Parse(r.ID); err == nil {
			ref = "@" + r.ID
		} else {
			ref = ":" + r.ID
		}
		return r.Name + ref
	default:
		return r.Name
	}
}
func (r DockerImageReference) Exact() string {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	name := r.NameString()
	if len(name) == 0 {
		return name
	}
	s := r.Registry
	if len(s) > 0 {
		s += "/"
	}
	if len(r.Namespace) != 0 {
		s += r.Namespace + "/"
	}
	return s + name
}
func (r DockerImageReference) String() string {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	if len(r.Namespace) == 0 && IsRegistryDockerHub(r.Registry) {
		r.Namespace = DockerDefaultNamespace
	}
	return r.Exact()
}
func SplitImageStreamTag(nameAndTag string) (name string, tag string, ok bool) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	parts := strings.SplitN(nameAndTag, ":", 2)
	name = parts[0]
	if len(parts) > 1 {
		tag = parts[1]
	}
	if len(tag) == 0 {
		tag = DefaultImageTag
	}
	return name, tag, len(parts) == 2
}
func JoinImageStreamTag(name, tag string) string {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	if len(tag) == 0 {
		tag = DefaultImageTag
	}
	return fmt.Sprintf("%s:%s", name, tag)
}
func JoinImageStreamImage(name, id string) string {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	return fmt.Sprintf("%s@%s", name, id)
}
func DigestOrImageMatch(image, imageID string) bool {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	if d, err := digest.Parse(image); err == nil {
		return strings.HasPrefix(d.Hex(), imageID) || strings.HasPrefix(image, imageID)
	}
	return strings.HasPrefix(image, imageID)
}
