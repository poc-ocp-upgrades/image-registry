package wrapped

import "io"

type readCloser struct {
	readCloser	io.ReadCloser
	wrapper		SimpleWrapper
}

var _ io.ReadCloser = &readCloser{}

func NewReadCloser(r io.ReadCloser, wrapper SimpleWrapper) io.ReadCloser {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	return &readCloser{readCloser: r, wrapper: wrapper}
}
func (r *readCloser) Read(p []byte) (n int, err error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	err = r.wrapper("ReadCloser.Read", func() error {
		n, err = r.readCloser.Read(p)
		return err
	})
	return
}
func (r *readCloser) Close() error {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	return r.wrapper("ReadCloser.Close", func() error {
		return r.readCloser.Close()
	})
}
