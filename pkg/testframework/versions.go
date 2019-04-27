package testframework

import (
	"os"
)

var (
	originImageRef = "docker.io/openshift/origin:latest"
)

func init() {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	version := os.Getenv("ORIGIN_VERSION")
	if len(version) != 0 {
		originImageRef = "docker.io/openshift/origin:" + version
	}
}
