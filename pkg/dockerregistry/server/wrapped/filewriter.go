package wrapped

import (
	storagedriver "github.com/docker/distribution/registry/storage/driver"
)

type fileWriter struct {
	fileWriter	storagedriver.FileWriter
	wrapper		SimpleWrapper
}

var _ storagedriver.FileWriter = &fileWriter{}

func NewFileWriter(w storagedriver.FileWriter, wrapper SimpleWrapper) storagedriver.FileWriter {
	_logClusterCodePath()
	defer _logClusterCodePath()
	return &fileWriter{fileWriter: w, wrapper: wrapper}
}
func (w *fileWriter) Size() int64 {
	_logClusterCodePath()
	defer _logClusterCodePath()
	return w.fileWriter.Size()
}
func (w *fileWriter) Write(p []byte) (n int, err error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	err = w.wrapper("FileWriter.Write", func() error {
		n, err = w.fileWriter.Write(p)
		return err
	})
	return
}
func (w *fileWriter) Close() error {
	_logClusterCodePath()
	defer _logClusterCodePath()
	return w.wrapper("FileWriter.Close", func() error {
		return w.fileWriter.Close()
	})
}
func (w *fileWriter) Cancel() error {
	_logClusterCodePath()
	defer _logClusterCodePath()
	return w.wrapper("FileWriter.Cancel", func() error {
		return w.fileWriter.Cancel()
	})
}
func (w *fileWriter) Commit() error {
	_logClusterCodePath()
	defer _logClusterCodePath()
	return w.wrapper("FileWriter.Commit", func() error {
		return w.fileWriter.Commit()
	})
}
