package manifesthandler

import (
	"fmt"
	"github.com/opencontainers/go-digest"
)

type ErrManifestBlobBadSize struct {
	Digest		digest.Digest
	ActualSize	int64
	SizeInManifest	int64
}

func (err ErrManifestBlobBadSize) Error() string {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	return fmt.Sprintf("the blob %s has the size (%d) different from the one specified in the manifest (%d)", err.Digest, err.ActualSize, err.SizeInManifest)
}
