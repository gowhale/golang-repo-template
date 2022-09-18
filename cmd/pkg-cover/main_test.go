package main

import (
	"fmt"
	"testing"

	"golang-repo-template/pkg/fruit"

	"github.com/stretchr/testify/suite"
)

const (
	runGoTestFunc              = "runGoTest"
	covertOutputToCoverageFunc = "covertOutputToCoverage"
	validateTestOutputFunc     = "validateTestOutput"
)

type mainTest struct {
	suite.Suite

	mockExec *mockExecuter

	executeStruct execute
}

func (m *mainTest) SetupTest() {
	m.mockExec = new(mockExecuter)
	m.executeStruct = execute{}
}

func TestMainTest(t *testing.T) {
	suite.Run(t, new(mainTest))
}

func (m *mainTest) Test_run_Pass() {
	m.mockExec.On(runGoTestFunc).Return(fruit.Apple, nil)
	m.mockExec.On(covertOutputToCoverageFunc, fruit.Apple).Return([]testLine{}, nil)
	m.mockExec.On(validateTestOutputFunc, []testLine{}, fruit.Apple).Return(nil)

	err := run(m.mockExec)
	m.Nil(err)
}

func (m *mainTest) Test_validateOutput_Error() {
	m.mockExec.On(runGoTestFunc).Return(fruit.Apple, nil)
	m.mockExec.On(covertOutputToCoverageFunc, fruit.Apple).Return([]testLine{}, nil)
	m.mockExec.On(validateTestOutputFunc, []testLine{}, fruit.Apple).Return(fmt.Errorf("validate error"))

	err := run(m.mockExec)
	m.EqualError(err, "validate error")
}

func (m *mainTest) Test_run_covertOutputToCoverage_Error() {
	m.mockExec.On(runGoTestFunc).Return(fruit.Apple, nil)
	m.mockExec.On(covertOutputToCoverageFunc, fruit.Apple).Return([]testLine{}, fmt.Errorf("covert output error"))

	err := run(m.mockExec)
	m.EqualError(err, "covert output error")
}

func (m *mainTest) Test_run_runGoTest_Error() {
	m.mockExec.On(runGoTestFunc).Return(fruit.Apple, fmt.Errorf("run go test error"))

	err := run(m.mockExec)
	m.EqualError(err, "run go test error")
}

func (m *mainTest) Test_getCoverage_testLine_Pass() {
	tl, err := getCoverage("ok      cmd/pkg-cover      0.182s  coverage: 38.5% of statements")

	m.Equal(true, tl.coverLine)
	m.Equal(38.5, tl.coverage)
	m.Nil(err)
}

func (m *mainTest) Test_getCoverage_testLine_Error() {
	tl, err := getCoverage("ok      golang-repo-test/cmd/pkg-cover      0.182s  coverage: wrong of statements")

	m.Equal(false, tl.coverLine)
	m.EqualError(err, "strconv.ParseFloat: parsing \"wron\": invalid syntax")
}

func (m *mainTest) Test_getCoverage_noTestLine() {
	tl, err := getCoverage("?       golang-repo-template/pkg/fruit  [no test files]")

	m.Equal(true, tl.coverLine)
	m.Equal(-1.0, tl.coverage)
	m.Nil(err)
}

func (m *mainTest) Test_getCoverage_excludedFile() {
	tl, err := getCoverage("?       golang-repo-template  [no test files]")

	m.Equal(false, tl.coverLine)
	m.Nil(err)
}

func (m *mainTest) Test_covertOutputToCoverage_noTests_Pass() {
	commandOutput := `?       golang-repo-template  [no test files]
	`

	tl, err := m.executeStruct.covertOutputToCoverage(commandOutput)
	m.Equal(tl, []testLine{})
	m.Nil(err)
}

func (m *mainTest) Test_covertOutputToCoverage_testsLine_Pass() {
	commandOutput := `ok      golang-repo-template/cmd/pkg-test      0.179s  coverage: 64.6% of statements
	`
	expectedTl := []testLine{{pkgName: "golang-repo-template/cmd/pkg-test", coverage: 64.6, coverLine: true}}

	tl, err := m.executeStruct.covertOutputToCoverage(commandOutput)
	m.Equal(tl, expectedTl)
	m.Nil(err)
}

func (m *mainTest) Test_covertOutputToCoverage_testsLine_Fail() {
	commandOutput := `ok      golang-repo-template/cmd/pkg-test      0.179s  coverage: wrong% of statements
	`

	tl, err := m.executeStruct.covertOutputToCoverage(commandOutput)
	m.Equal(tl, []testLine(nil))
	m.EqualError(err, "strconv.ParseFloat: parsing \"wrong\": invalid syntax")
}

func (m *mainTest) Test_validateTestOutput_sufficentCov_testLine() {
	expectedTl := []testLine{{pkgName: "golang-repo-template/cmd/pkg-test", coverage: 81, coverLine: true}}

	err := m.executeStruct.validateTestOutput(expectedTl, fruit.Apple)
	m.Nil(err)
}

func (m *mainTest) Test_validateTestOutput_NOT_sufficentCov_testLine() {
	expectedTl := []testLine{{pkgName: "golang-repo-template/cmd/pkg-test", coverage: 79, coverLine: true}}

	err := m.executeStruct.validateTestOutput(expectedTl, fruit.Apple)
	m.EqualError(err,"the following pkgs are not valid: [pkg=golang-repo-template/cmd/pkg-test cov=79.000000 under the 80.000000% minimum line coverage]")
}

func (m *mainTest) Test_validateTestOutput_testLine_missing_tests() {
	expectedTl := []testLine{{coverLine: false}}

	err := m.executeStruct.validateTestOutput(expectedTl, fruit.Apple)
	m.EqualError(err,"the following pkgs are not valid: [pkg= is missing tests]")
}
