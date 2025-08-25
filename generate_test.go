package main

import (
	"fmt"
	"testing"

	"path/filepath"

	"github.com/getgauge/gauge-proto/go/gauge_messages"
	"github.com/xeipuuv/gojsonschema"
	. "gopkg.in/check.v1"
)

func (s *MySuite) TestJSONReportWithSchema(c *C) {
	sampleJSON, _ := filepath.Abs(filepath.Join("_testdata", "sample.json"))
	JSONSchema, _ := filepath.Abs("schema.json")
	schemaLoader := gojsonschema.NewReferenceLoader("file:///" + JSONSchema)
	documentLoader := gojsonschema.NewReferenceLoader("file:///" + sampleJSON)
	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		c.Error(err)
	}
	for _, desc := range result.Errors() {
		fmt.Printf("- %s\n", desc)
	}
	c.Assert(result.Valid(), Equals, true)
}

func TestToScenario_SetsRetriesCount(t *testing.T) {
	expectedRetries := int64(3)
	protoSce := &gauge_messages.ProtoScenario{
		// Set the retries count
		RetriesCount: expectedRetries,
	}

	tableRowIndex := 0
	sc := toScenario(protoSce, tableRowIndex)

	if sc.RetriesCount != expectedRetries {
		t.Errorf("Expected RetriesCount to be %d, but got %d", expectedRetries, sc.RetriesCount)
	}
}
