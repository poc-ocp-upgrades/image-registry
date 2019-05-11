package api

import "time"

func (t *TestSuite) AddProperty(name, value string) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	for _, property := range t.Properties {
		if property.Name == name {
			property.Value = value
			return
		}
	}
	t.Properties = append(t.Properties, &TestSuiteProperty{Name: name, Value: value})
}
func (t *TestSuite) AddTestCase(testCase *TestCase) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	t.NumTests += 1
	switch {
	case testCase.SkipMessage != nil:
		t.NumSkipped += 1
	case testCase.FailureOutput != nil:
		t.NumFailed += 1
	default:
		testCase.SystemOut = ""
		testCase.SystemErr = ""
	}
	t.Duration += testCase.Duration
	t.Duration = float64(int(t.Duration*1000)) / 1000
	t.TestCases = append(t.TestCases, testCase)
}
func (t *TestSuite) SetDuration(duration string) error {
	_logClusterCodePath()
	defer _logClusterCodePath()
	parsedDuration, err := time.ParseDuration(duration)
	if err != nil {
		return err
	}
	t.Duration = float64(int(parsedDuration.Seconds()*1000)) / 1000
	return nil
}

type ByName []*TestSuite

func (n ByName) Len() int {
	_logClusterCodePath()
	defer _logClusterCodePath()
	return len(n)
}
func (n ByName) Swap(i, j int) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	n[i], n[j] = n[j], n[i]
}
func (n ByName) Less(i, j int) bool {
	_logClusterCodePath()
	defer _logClusterCodePath()
	return n[i].Name < n[j].Name
}
