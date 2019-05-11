package api

import "time"

func (t *TestCase) SetDuration(duration string) error {
	_logClusterCodePath()
	defer _logClusterCodePath()
	parsedDuration, err := time.ParseDuration(duration)
	if err != nil {
		return err
	}
	t.Duration = float64(int(parsedDuration.Seconds()*1000)) / 1000
	return nil
}
func (t *TestCase) MarkSkipped(message string) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	t.SkipMessage = &SkipMessage{Message: message}
}
func (t *TestCase) MarkFailed(message, output string) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	t.FailureOutput = &FailureOutput{Message: message, Output: output}
}
