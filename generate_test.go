package main

import (
	"fmt"

	"path/filepath"

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
