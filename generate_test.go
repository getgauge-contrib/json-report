package main

import (
	"encoding/json"
	"fmt"
	"reflect"
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
	scenarioTableRowIndex := -1
	sc := toScenario(protoSce, tableRowIndex, scenarioTableRowIndex)

	if sc.RetriesCount != expectedRetries {
		t.Errorf("Expected RetriesCount to be %d, but got %d", expectedRetries, sc.RetriesCount)
	}
}

func newProtoTable(headers []string, rows ...[]string) *gauge_messages.ProtoTable {
	protoRows := make([]*gauge_messages.ProtoTableRow, len(rows))
	for i, cells := range rows {
		protoRows[i] = &gauge_messages.ProtoTableRow{Cells: cells}
	}
	return &gauge_messages.ProtoTable{
		Headers: &gauge_messages.ProtoTableRow{Cells: headers},
		Rows:    protoRows,
	}
}

func newSpecResult(items ...*gauge_messages.ProtoItem) *gauge_messages.ProtoSpecResult {
	return &gauge_messages.ProtoSpecResult{
		ProtoSpec: &gauge_messages.ProtoSpec{
			SpecHeading: "Spec",
			Items:       items,
		},
	}
}

func newTableDrivenScenarioItem(tds *gauge_messages.ProtoTableDrivenScenario) *gauge_messages.ProtoItem {
	tds.Scenario = &gauge_messages.ProtoScenario{ScenarioHeading: "Scenario"}
	return &gauge_messages.ProtoItem{
		ItemType:            gauge_messages.ProtoItem_TableDrivenScenario,
		TableDrivenScenario: tds,
	}
}

func TestToSpec_ScenarioTableDrivenScenarioContainsScenarioTable(t *testing.T) {
	headers := []string{"word", "count"}
	scenarioDataTable := newProtoTable(headers, []string{"Gauge", "3"}, []string{"Mingle", "2"})
	scenarioTableRow := newProtoTable(headers, []string{"Mingle", "2"})

	s := toSpec(newSpecResult(newTableDrivenScenarioItem(&gauge_messages.ProtoTableDrivenScenario{
		TableRowIndex:         -1,
		ScenarioTableRowIndex: 1,
		IsScenarioTableDriven: true,
		ScenarioDataTable:     scenarioDataTable,
		ScenarioTableRow:      scenarioTableRow,
	})))

	if len(s.Scenarios) != 1 {
		t.Fatalf("Expected 1 scenario, but got %d", len(s.Scenarios))
	}
	sce := s.Scenarios[0]
	if sce.ScenarioTableRowIndex != 1 {
		t.Errorf("Expected ScenarioTableRowIndex to be 1, but got %d", sce.ScenarioTableRowIndex)
	}
	expectedDataTable := &table{
		Headers: headers,
		Rows:    []row{{Cells: []string{"Gauge", "3"}}, {Cells: []string{"Mingle", "2"}}},
	}
	if !reflect.DeepEqual(sce.ScenarioDataTable, expectedDataTable) {
		t.Errorf("Expected ScenarioDataTable to be %+v, but got %+v", expectedDataTable, sce.ScenarioDataTable)
	}
	expectedRow := &table{
		Headers: headers,
		Rows:    []row{{Cells: []string{"Mingle", "2"}}},
	}
	if !reflect.DeepEqual(sce.ScenarioTableRow, expectedRow) {
		t.Errorf("Expected ScenarioTableRow to be %+v, but got %+v", expectedRow, sce.ScenarioTableRow)
	}
}

func TestToSpec_ScenarioTableDrivenScenarioWithinSpecTableDrivenSpecContainsScenarioTable(t *testing.T) {
	headers := []string{"word", "count"}

	s := toSpec(newSpecResult(newTableDrivenScenarioItem(&gauge_messages.ProtoTableDrivenScenario{
		TableRowIndex:         2,
		ScenarioTableRowIndex: 0,
		IsSpecTableDriven:     true,
		IsScenarioTableDriven: true,
		ScenarioDataTable:     newProtoTable(headers, []string{"Gauge", "3"}),
		ScenarioTableRow:      newProtoTable(headers, []string{"Gauge", "3"}),
	})))

	sce := s.Scenarios[0]
	if sce.TableRowIndex != 2 || sce.ScenarioTableRowIndex != 0 {
		t.Errorf("Expected row indexes 2/0, but got %d/%d", sce.TableRowIndex, sce.ScenarioTableRowIndex)
	}
	if sce.ScenarioDataTable == nil || len(sce.ScenarioDataTable.Rows) != 1 {
		t.Errorf("Expected ScenarioDataTable with 1 row, but got %+v", sce.ScenarioDataTable)
	}
	if sce.ScenarioTableRow == nil || !reflect.DeepEqual(sce.ScenarioTableRow.Headers, headers) {
		t.Errorf("Expected ScenarioTableRow with headers %v, but got %+v", headers, sce.ScenarioTableRow)
	}
}

func TestToSpec_SpecTableDrivenOnlyScenarioHasNoScenarioTable(t *testing.T) {
	// Gauge sends an empty scenario table row for scenarios which are only spec table driven
	s := toSpec(newSpecResult(newTableDrivenScenarioItem(&gauge_messages.ProtoTableDrivenScenario{
		TableRowIndex:         0,
		ScenarioTableRowIndex: -1,
		IsSpecTableDriven:     true,
		ScenarioTableRow:      &gauge_messages.ProtoTable{},
	})))

	sce := s.Scenarios[0]
	if sce.ScenarioDataTable != nil {
		t.Errorf("Expected ScenarioDataTable to be nil, but got %+v", sce.ScenarioDataTable)
	}
	if sce.ScenarioTableRow != nil {
		t.Errorf("Expected ScenarioTableRow to be nil, but got %+v", sce.ScenarioTableRow)
	}
}

func TestToSpec_ScenarioWithoutTableHasNoScenarioTable(t *testing.T) {
	s := toSpec(newSpecResult(&gauge_messages.ProtoItem{
		ItemType: gauge_messages.ProtoItem_Scenario,
		Scenario: &gauge_messages.ProtoScenario{ScenarioHeading: "Scenario"},
	}))

	sce := s.Scenarios[0]
	if sce.ScenarioDataTable != nil || sce.ScenarioTableRow != nil {
		t.Errorf("Expected no scenario table, but got %+v / %+v", sce.ScenarioDataTable, sce.ScenarioTableRow)
	}
}

func TestScenarioJSONContainsScenarioTableKeys(t *testing.T) {
	sce := toScenario(&gauge_messages.ProtoScenario{}, -1, -1)
	b, err := json.Marshal(sce)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"scenarioDataTable", "scenarioTableRow"} {
		v, ok := m[key]
		if !ok {
			t.Errorf("Expected key %q in scenario JSON", key)
		} else if v != nil {
			t.Errorf("Expected %q to be null, but got %v", key, v)
		}
	}
}
