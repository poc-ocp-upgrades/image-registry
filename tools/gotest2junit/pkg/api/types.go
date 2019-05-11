package api

import "encoding/xml"

type TestSuites struct {
	XMLName	xml.Name		`xml:"testsuites"`
	Suites	[]*TestSuite	`xml:"testsuite"`
}
type TestSuite struct {
	XMLName		xml.Name				`xml:"testsuite"`
	Name		string					`xml:"name,attr"`
	NumTests	uint					`xml:"tests,attr"`
	NumSkipped	uint					`xml:"skipped,attr"`
	NumFailed	uint					`xml:"failures,attr"`
	Duration	float64					`xml:"time,attr"`
	Properties	[]*TestSuiteProperty	`xml:"properties,omitempty"`
	TestCases	[]*TestCase				`xml:"testcase"`
	Children	[]*TestSuite			`xml:"testsuite"`
}
type TestSuiteProperty struct {
	XMLName	xml.Name	`xml:"property"`
	Name	string		`xml:"name,attr"`
	Value	string		`xml:"value,attr"`
}
type TestCase struct {
	XMLName			xml.Name		`xml:"testcase"`
	Name			string			`xml:"name,attr"`
	Classname		string			`xml:"classname,attr,omitempty"`
	Duration		float64			`xml:"time,attr"`
	SkipMessage		*SkipMessage	`xml:"skipped"`
	FailureOutput	*FailureOutput	`xml:"failure"`
	SystemOut		string			`xml:"system-out,omitempty"`
	SystemErr		string			`xml:"system-err,omitempty"`
}
type SkipMessage struct {
	XMLName	xml.Name	`xml:"skipped"`
	Message	string		`xml:"message,attr,omitempty"`
}
type FailureOutput struct {
	XMLName	xml.Name	`xml:"failure"`
	Message	string		`xml:"message,attr"`
	Output	string		`xml:",chardata"`
}
type TestResult string

const (
	TestResultPass	TestResult	= "pass"
	TestResultSkip	TestResult	= "skip"
	TestResultFail	TestResult	= "fail"
)
