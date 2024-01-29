package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/getgauge-contrib/json-report/logger"
	"github.com/getgauge/common"
	"github.com/getgauge/gauge-proto/go/gauge_messages"
)

const (
	defaultReportsDir           = "reports"
	gaugeReportsDirEnvName      = "gauge_reports_dir" // directory where reports are generated by plugins
	overwriteReportsEnvProperty = "overwrite_reports"
	jsonReportFile              = "result.json"
	jsonReport                  = "json-report"
	setupAction                 = "setup"
	executionAction             = "execution"
	gaugePortEnv                = "plugin_connection_port"
	pluginActionEnv             = "json-report_action"
	timeFormat                  = "2006-01-02 15.04.05"
)

var projectRoot string

type nameGenerator interface {
	randomName() string
}

type timeStampedNameGenerator struct {
}

func (T timeStampedNameGenerator) randomName() string {
	return time.Now().Format(timeFormat)
}

func findProjectRoot() {
	projectRoot = os.Getenv(common.GaugeProjectRootEnv)
	if projectRoot == "" {
		logger.Fatal("Environment variable '%s' is not set. \n", common.GaugeProjectRootEnv)
	}
}

func addDefaultPropertiesToProject() {
	defaultPropertiesFile := getDefaultPropertiesFile()

	reportsDirProperty := &(common.Property{
		Comment:      "The path to the gauge reports directory. Should be either relative to the project directory or an absolute path",
		Name:         gaugeReportsDirEnvName,
		DefaultValue: defaultReportsDir})

	overwriteReportProperty := &(common.Property{
		Comment:      "Set as false if gauge reports should not be overwritten on each execution. A new time-stamped directory will be created on each execution.",
		Name:         overwriteReportsEnvProperty,
		DefaultValue: "true"})

	if !common.FileExists(defaultPropertiesFile) {
		logger.Info("Failed to setup json report plugin in project. Default properties file does not exist at %s. \n", defaultPropertiesFile)
		return
	}
	if err := common.AppendProperties(defaultPropertiesFile, reportsDirProperty, overwriteReportProperty); err != nil {
		logger.Info("Failed to setup json report plugin in project: %s \n", err)
		return
	}
	logger.Info("Successfully added configurations for json-report to env/default/default.properties")
}

func createReport(suiteResult *gauge_messages.SuiteExecutionResult) {
	jsonContents := generateJSONFileContents(suiteResult)
	reportDir, err := createJSONReport(createReportsDirectory(), jsonContents, getNameGen())
	if err != nil {
		logger.Fatal("Report generation failed: %s \n", err)
	} else {
		logger.Info("Successfully generated json-report to => %s\n", reportDir)
	}
}

func generateJSONFileContents(protoSuiteExeResult *gauge_messages.SuiteExecutionResult) []byte {
	var buffer bytes.Buffer
	suiteRes := toSuiteResult(protoSuiteExeResult.GetSuiteResult())
	buffer.WriteString(fmt.Sprintf("%s", marshal(suiteRes)))
	return buffer.Bytes()
}

func marshal(item interface{}) []byte {
	marshalledResult, err := json.MarshalIndent(item, "", "\t")
	if err != nil {
		logger.Fatal("Failed to convert to json :%s\n", err)
	}
	return marshalledResult
}

func createJSONReport(reportsDir string, jsonContents []byte, nameGen nameGenerator) (string, error) {
	var currentReportDir string
	if nameGen != nil {
		currentReportDir = filepath.Join(reportsDir, jsonReport, nameGen.randomName())
	} else {
		currentReportDir = filepath.Join(reportsDir, jsonReport)
	}
	createDirectory(currentReportDir)
	return currentReportDir, writeResultJSONFile(currentReportDir, jsonContents)
}

func writeResultJSONFile(reportDir string, jsonContents []byte) error {
	resultJsPath := filepath.Join(reportDir, jsonReportFile)
	err := ioutil.WriteFile(resultJsPath, jsonContents, common.NewFilePermissions)
	if err != nil {
		return fmt.Errorf("failed to copy file: %s %s", jsonReportFile, err.Error())
	}
	return nil
}

func getNameGen() nameGenerator {
	var nameGen nameGenerator
	if shouldOverwriteReports() {
		nameGen = nil
	} else {
		nameGen = timeStampedNameGenerator{}
	}
	return nameGen
}

func getDefaultPropertiesFile() string {
	return filepath.Join(projectRoot, "env", "default", "default.properties")
}

func shouldOverwriteReports() bool {
	envValue := os.Getenv(overwriteReportsEnvProperty)
	if strings.ToLower(envValue) == "true" {
		return true
	}
	return false
}

func createReportsDirectory() string {
	reportsDir, err := filepath.Abs(os.Getenv(gaugeReportsDirEnvName))
	if reportsDir == "" || err != nil {
		reportsDir = defaultReportsDir
	}
	createDirectory(reportsDir)
	return reportsDir
}

func createDirectory(dir string) {
	if common.DirExists(dir) {
		return
	}
	if err := os.MkdirAll(dir, common.NewDirectoryPermissions); err != nil {
		logger.Fatal("Failed to create directory %s: %s\n", defaultReportsDir, err)
	}
}
