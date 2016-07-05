package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type MySuite struct{}

var _ = Suite(&MySuite{})

var now = time.Now()

type testNameGenerator struct {
}

func (T testNameGenerator) randomName() string {
	return now.Format(timeFormat)
}

func (s *MySuite) TestCreatingReport(c *C) {
	reportDir := filepath.Join(os.TempDir(), randomName())
	defer os.RemoveAll(reportDir)

	finalReportDir, err := createJSONReport(reportDir, make([]byte, 0), nil)
	c.Assert(err, IsNil)

	expectedFinalReportDir := filepath.Join(reportDir, jsonReport)
	c.Assert(finalReportDir, Equals, expectedFinalReportDir)
	verifyJSONReportFileIsCopied(expectedFinalReportDir, c)
}

func (s *MySuite) TestCreatingReportWithNoOverWrite(c *C) {
	reportDir := filepath.Join(os.TempDir(), randomName())
	defer os.RemoveAll(reportDir)

	nameGen := testNameGenerator{}
	finalReportDir, err := createJSONReport(reportDir, make([]byte, 0), nameGen)
	c.Assert(err, IsNil)

	expectedFinalReportDir := filepath.Join(reportDir, jsonReport, nameGen.randomName())
	c.Assert(finalReportDir, Equals, expectedFinalReportDir)
	verifyJSONReportFileIsCopied(expectedFinalReportDir, c)
}

func randomName() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func verifyJSONReportFileIsCopied(dest string, c *C) {
	c.Assert(fileExists(filepath.Join(dest, jsonReportFile)), Equals, true)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	return !os.IsNotExist(err)
}

func (s *MySuite) TestCreatingReportShouldOverwriteReportsBasedOnEnv(c *C) {
	os.Setenv(overwriteReportsEnvProperty, "true")
	nameGen := getNameGen()
	c.Assert(nameGen, Equals, nil)

	os.Setenv(overwriteReportsEnvProperty, "false")
	nameGen = getNameGen()
	c.Assert(nameGen, Equals, timeStampedNameGenerator{})
}
