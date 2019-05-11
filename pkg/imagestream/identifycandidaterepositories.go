package imagestream

import (
	"sort"
	imageapiv1 "github.com/openshift/api/image/v1"
	imageapi "github.com/openshift/image-registry/pkg/origin-common/image/apis/image"
)

type byInsecureFlag struct {
	repositories	[]string
	specs			[]*ImagePullthroughSpec
}

func (by *byInsecureFlag) Len() int {
	_logClusterCodePath()
	defer _logClusterCodePath()
	if len(by.specs) < len(by.repositories) {
		return len(by.specs)
	}
	return len(by.repositories)
}
func (by *byInsecureFlag) Swap(i, j int) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	by.repositories[i], by.repositories[j] = by.repositories[j], by.repositories[i]
	by.specs[i], by.specs[j] = by.specs[j], by.specs[i]
}
func (by *byInsecureFlag) Less(i, j int) bool {
	_logClusterCodePath()
	defer _logClusterCodePath()
	if by.specs[i].Insecure == by.specs[j].Insecure {
		switch {
		case by.repositories[i] < by.repositories[j]:
			return true
		case by.repositories[i] > by.repositories[j]:
			return false
		default:
			return by.specs[i].DockerImageReference.Exact() < by.specs[j].DockerImageReference.Exact()
		}
	}
	return !by.specs[i].Insecure
}
func stringListContains(list []string, val string) bool {
	_logClusterCodePath()
	defer _logClusterCodePath()
	for _, x := range list {
		if x == val {
			return true
		}
	}
	return false
}
func identifyCandidateRepositories(is *imageapiv1.ImageStream, localRegistry []string, primary bool) ([]string, map[string]ImagePullthroughSpec) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	insecureByDefault := false
	if insecure, ok := is.Annotations[imageapi.InsecureRepositoryAnnotation]; ok {
		insecureByDefault = insecure == "true"
	}
	insecureRegistries := make(map[string]bool)
	search := make(map[string]*imageapi.DockerImageReference)
	for _, tagEvent := range is.Status.Tags {
		tag := tagEvent.Tag
		var candidates []imageapiv1.TagEvent
		if primary {
			if len(tagEvent.Items) == 0 {
				continue
			}
			candidates = tagEvent.Items[:1]
		} else {
			if len(tagEvent.Items) <= 1 {
				continue
			}
			candidates = tagEvent.Items[1:]
		}
		for _, event := range candidates {
			ref, err := imageapi.ParseDockerImageReference(event.DockerImageReference)
			if err != nil {
				continue
			}
			if stringListContains(localRegistry, ref.Registry) {
				continue
			}
			ref = ref.DockerClientDefaults()
			insecure := insecureByDefault
			for _, t := range is.Spec.Tags {
				if t.Name == tag {
					insecure = insecureByDefault || t.ImportPolicy.Insecure
					break
				}
			}
			if is := insecureRegistries[ref.Registry]; !is && insecure {
				insecureRegistries[ref.Registry] = insecure
			}
			search[ref.AsRepository().Exact()] = &ref
		}
	}
	repositories := make([]string, 0, len(search))
	results := make(map[string]ImagePullthroughSpec)
	specs := []*ImagePullthroughSpec{}
	for repo, ref := range search {
		repositories = append(repositories, repo)
		spec := ImagePullthroughSpec{DockerImageReference: ref, Insecure: insecureRegistries[ref.Registry]}
		results[repo] = spec
		specs = append(specs, &spec)
	}
	sort.Sort(&byInsecureFlag{repositories: repositories, specs: specs})
	return repositories, results
}
