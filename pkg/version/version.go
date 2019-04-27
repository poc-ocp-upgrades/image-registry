package version

import (
	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/apimachinery/pkg/version"
)

var (
	commitFromGit	string
	versionFromGit	string
	majorFromGit	string
	minorFromGit	string
	buildDate	string
)

func Get() version.Info {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	return version.Info{Major: majorFromGit, Minor: minorFromGit, GitCommit: commitFromGit, GitVersion: versionFromGit, BuildDate: buildDate}
}
func init() {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	buildInfo := prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "imageregistry_build_info", Help: "A metric with a constant '1' value labeled by major, minor, git commit & git version from which the image registry was built."}, []string{"major", "minor", "gitCommit", "gitVersion"})
	buildInfo.WithLabelValues(majorFromGit, minorFromGit, commitFromGit, versionFromGit).Set(1)
	prometheus.MustRegister(buildInfo)
}
