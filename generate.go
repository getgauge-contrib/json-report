package main

import (
	"encoding/base64"

	"github.com/getgauge/json-report/gauge_messages"
)

type tokenKind string
type status string

const (
	pass                    status    = "pass"
	fail                    status    = "fail"
	skip                    status    = "skip"
	notExecuted             status    = "not executed"
	tagKind                 tokenKind = "tag"
	scenarioKind            tokenKind = "scenario"
	tableDrivenScenarioKind tokenKind = "tableDrivenScenario"
	commentKind             tokenKind = "comment"
	stepKind                tokenKind = "step"
	conceptKind             tokenKind = "concept"
	tableKind               tokenKind = "table"
)

type item interface {
	kind() tokenKind
}

type suiteResult struct {
	Specs                  []*spec      `json:"specs"`
	BeforeSuiteHookFailure *hookFailure `json:"beforeSuiteHookFailure,omitempty"`
	AfterSuiteHookFailure  *hookFailure `json:"afterSuiteHookFailure,omitempty"`
	ExecutionStatus        status       `json:"executionStatus"`
	ExecutionTime          int64        `json:"executionTime"`
	PassedSpecsCount       int          `json:"passedSpecsCount"`
	FailedSpecsCount       int          `json:"failedSpecsCount"`
	SkippedSpecsCount      int          `json:"skippedSpecsCount"`
	UnhandledErrors        []error      `json:"unhandledErrors,omitempty"`
	Environment            string       `json:"environment"`
	Tags                   string       `json:"tags"`
	ProjectName            string       `json:"projectName"`
	Timestamp              string       `json:"timestamp"`
	SuccessRate            int          `json:"successRate"`
}

type spec struct {
	SpecHeading           string       `json:"specHeading"`
	IsTableDriven         bool         `json:"isTableDriven"`
	FileName              string       `json:"fileName"`
	Tags                  []string     `json:"tags"`
	Items                 []item       `json:"items"`
	BeforeSpecHookFailure *hookFailure `json:"beforeSpecHookFailure,omitempty"`
	AfterSpecHookFailure  *hookFailure `json:"afterSpecHookFailure,omitempty"`
	ExecutionStatus       status       `json:"executionStatus"`
	ExecutionTime         int64        `json:"executionTime"`
	ScenarioFailedCount   int          `json:"scenarioFailedCount"`
	ScenarioSkippedCount  int          `json:"scenarioSkippedCount"`
}

type scenario struct {
	ItemType                  tokenKind    `json:"itemType"`
	Heading                   string       `json:"heading"`
	Contexts                  []item       `json:"contexts"`
	Teardowns                 []item       `json:"teardowns"`
	Items                     []item       `json:"items"`
	ExecutionStatus           status       `json:"executionStatus"`
	ExecutionTime             int64        `json:"executionTime"`
	BeforeScenarioHookFailure *hookFailure `json:"beforeScenarioHookFailure,omitempty"`
	AfterScenarioHookFailure  *hookFailure `json:"afterScenarioHookFailure,omitempty"`
	Tags                      []string     `json:"tags"`
	SkipErrors                []string     `json:"skipErrors"`
}

func (s *scenario) kind() tokenKind {
	return scenarioKind
}

type tableDrivenScenario struct {
	ItemType      tokenKind `json:"itemType"`
	Scenario      *scenario `json:"scenario"`
	TableRowIndex int       `json:"tableRowIndex"`
}

func (t *tableDrivenScenario) kind() tokenKind {
	return tableDrivenScenarioKind
}

type result struct {
	Status        status   `json:"status"`
	StackTrace    string   `json:"stackTrace"`
	Screenshot    string   `json:"screenshot"`
	ErrorMessage  string   `json:"errorMessage"`
	ExecutionTime int64    `json:"executionTime"`
	SkippedReason string   `json:"skippedReason"`
	Messages      []string `json:"messages"`
}
type hookFailure struct {
	ErrMsg     string `json:"errorMessage"`
	Screenshot string `json:"screenshot"`
	StackTrace string `json:"stackTrace"`
}

type step struct {
	ItemType              tokenKind    `json:"itemType"`
	StepText              string       `json:"StepText"`
	BeforeStepHookFailure *hookFailure `json:"beforeStepHookFailure,omitempty"`
	AfterStepHookFailure  *hookFailure `json:"afterStepHookFailure,omitempty"`
	Result                *result      `json:"result"`
}

func (s *step) kind() tokenKind {
	return stepKind
}

type concept struct {
	ItemType        tokenKind `json:"itemType"`
	ConceptStep     *step     `json:"conceptStep"`
	Items           []item    `json:"items"`
	ExecutionStatus string    `json:"executionStatus"`
	ExecutionTime   int64     `json:"executionTime"`
	Result          result    `json:"result"`
}

func (s *concept) kind() tokenKind {
	return conceptKind
}

type comment struct {
	ItemType tokenKind `json:"itemType"`
	Text     string    `json:"text"`
}

func (c *comment) kind() tokenKind {
	return commentKind
}

type tag struct {
	ItemType tokenKind `json:"itemType"`
	Tags     []string  `json:"tags"`
}

func (c *tag) kind() tokenKind {
	return tagKind
}

type table struct {
	ItemType tokenKind `json:"itemType"`
	Headers  []string  `json:"headers"`
	Rows     []*row    `json:"rows"`
}

func (t *table) kind() tokenKind {
	return tableKind
}

type row struct {
	Cells []string `json:"cells"`
}

func toSuiteResult(psr *gauge_messages.ProtoSuiteResult) *suiteResult {
	suiteResult := &suiteResult{
		ProjectName:            psr.GetProjectName(),
		Environment:            psr.GetEnvironment(),
		Tags:                   psr.GetTags(),
		ExecutionTime:          psr.GetExecutionTime(),
		PassedSpecsCount:       len(psr.GetSpecResults()) - int(psr.GetSpecsFailedCount()) - int(psr.GetSpecsSkippedCount()),
		FailedSpecsCount:       int(psr.GetSpecsFailedCount()),
		SkippedSpecsCount:      int(psr.GetSpecsSkippedCount()),
		BeforeSuiteHookFailure: toHookFailure(psr.GetPreHookFailure()),
		AfterSuiteHookFailure:  toHookFailure(psr.GetPostHookFailure()),
		SuccessRate:            int(psr.GetSuccessRate()),
		Timestamp:              psr.GetTimestamp(),
		ExecutionStatus:        pass,
	}
	if psr.GetFailed() {
		suiteResult.ExecutionStatus = fail
	}
	for _, protoSpecRes := range psr.GetSpecResults() {
		suiteResult.Specs = append(suiteResult.Specs, toSpec(protoSpecRes))
	}
	return suiteResult
}

func toSpec(psr *gauge_messages.ProtoSpecResult) *spec {
	spec := &spec{
		SpecHeading:           psr.GetProtoSpec().GetSpecHeading(),
		IsTableDriven:         psr.GetProtoSpec().GetIsTableDriven(),
		FileName:              psr.GetProtoSpec().GetFileName(),
		Tags:                  psr.GetProtoSpec().GetTags(),
		ScenarioFailedCount:   int(psr.GetScenarioFailedCount()),
		ScenarioSkippedCount:  int(psr.GetScenarioSkippedCount()),
		ExecutionTime:         psr.GetExecutionTime(),
		ExecutionStatus:       getStatus(psr.GetFailed(), psr.GetSkipped()),
		BeforeSpecHookFailure: toHookFailure(psr.GetProtoSpec().GetPreHookFailure()),
		AfterSpecHookFailure:  toHookFailure(psr.GetProtoSpec().GetPostHookFailure()),
	}
	for _, item := range psr.GetProtoSpec().GetItems() {
		switch item.GetItemType() {
		case gauge_messages.ProtoItem_Comment:
			spec.Items = append(spec.Items, toComment(item.GetComment()))
		case gauge_messages.ProtoItem_Table:
			spec.Items = append(spec.Items, toTable(item.GetTable()))
		case gauge_messages.ProtoItem_Scenario:
			spec.Items = append(spec.Items, toScenario(item.GetScenario()))
		case gauge_messages.ProtoItem_Tags:
			spec.Items = append(spec.Items, toTag(item.GetTags()))
		case gauge_messages.ProtoItem_Step:
			spec.Items = append(spec.Items, toStep(item.GetStep()))
		case gauge_messages.ProtoItem_Concept:
			spec.Items = append(spec.Items, toConcept(item.GetConcept()))
		case gauge_messages.ProtoItem_TableDrivenScenario:
			spec.Items = append(spec.Items, toTableDrivenScenario(item.GetTableDrivenScenario().GetScenario(), int(item.GetTableDrivenScenario().GetTableRowIndex())))
		}
	}
	return spec
}

func toScenario(scn *gauge_messages.ProtoScenario) *scenario {
	return &scenario{
		ItemType:                  scenarioKind,
		Heading:                   scn.GetScenarioHeading(),
		ExecutionTime:             scn.GetExecutionTime(),
		Tags:                      scn.GetTags(),
		ExecutionStatus:           getScenarioStatus(scn),
		Contexts:                  toItems(scn.GetContexts()),
		Items:                     toItems(scn.GetScenarioItems()),
		Teardowns:                 toItems(scn.GetTearDownSteps()),
		BeforeScenarioHookFailure: toHookFailure(scn.GetPreHookFailure()),
		AfterScenarioHookFailure:  toHookFailure(scn.GetPostHookFailure()),
	}
}

func toTableDrivenScenario(scn *gauge_messages.ProtoScenario, tableRowIndex int) *tableDrivenScenario {
	return &tableDrivenScenario{
		ItemType:      tableDrivenScenarioKind,
		Scenario:      toScenario(scn),
		TableRowIndex: tableRowIndex,
	}
}

func getScenarioStatus(scn *gauge_messages.ProtoScenario) status {
	switch scn.GetExecutionStatus() {
	case gauge_messages.ExecutionStatus_FAILED:
		return fail
	case gauge_messages.ExecutionStatus_PASSED:
		return pass
	case gauge_messages.ExecutionStatus_SKIPPED:
		return skip
	default:
		return notExecuted
	}
}

func toTable(protoTable *gauge_messages.ProtoTable) *table {
	rows := make([]*row, len(protoTable.GetRows()))
	for i, r := range protoTable.GetRows() {
		rows[i] = &row{
			Cells: r.GetCells(),
		}
	}
	return &table{ItemType: tableKind, Headers: protoTable.GetHeaders().GetCells(), Rows: rows}
}

func toItems(protoItems []*gauge_messages.ProtoItem) []item {
	items := make([]item, 0)
	for _, i := range protoItems {
		switch i.GetItemType() {
		case gauge_messages.ProtoItem_Step:
			items = append(items, toStep(i.GetStep()))
		case gauge_messages.ProtoItem_Comment:
			items = append(items, toComment(i.GetComment()))
		case gauge_messages.ProtoItem_Concept:
			items = append(items, toConcept(i.GetConcept()))
		}
	}
	return items
}

func toComment(protoComment *gauge_messages.ProtoComment) *comment {
	return &comment{ItemType: commentKind, Text: protoComment.GetText()}
}

func toTag(protoTag *gauge_messages.ProtoTags) *tag {
	return &tag{ItemType: tagKind, Tags: protoTag.GetTags()}
}

func toStep(protoStep *gauge_messages.ProtoStep) *step {
	res := protoStep.GetStepExecutionResult().GetExecutionResult()
	result := &result{
		Status:        getStepStatus(protoStep.GetStepExecutionResult()),
		Screenshot:    base64.StdEncoding.EncodeToString(res.GetScreenShot()),
		StackTrace:    res.GetStackTrace(),
		ErrorMessage:  res.GetErrorMessage(),
		ExecutionTime: res.GetExecutionTime(),
		Messages:      res.GetMessage(),
	}
	if protoStep.GetStepExecutionResult().GetSkipped() {
		result.SkippedReason = protoStep.GetStepExecutionResult().GetSkippedReason()
	}
	return &step{
		ItemType:              stepKind,
		StepText:              protoStep.GetActualText(),
		Result:                result,
		BeforeStepHookFailure: toHookFailure(protoStep.GetStepExecutionResult().GetPreHookFailure()),
		AfterStepHookFailure:  toHookFailure(protoStep.GetStepExecutionResult().GetPostHookFailure()),
	}
}

func toConcept(protoConcept *gauge_messages.ProtoConcept) *concept {
	protoConcept.ConceptStep.StepExecutionResult = protoConcept.GetConceptExecutionResult()
	return &concept{
		ItemType:    conceptKind,
		ConceptStep: toStep(protoConcept.GetConceptStep()),
		Items:       toItems(protoConcept.GetSteps()),
	}
}

func toHookFailure(failure *gauge_messages.ProtoHookFailure) *hookFailure {
	if failure == nil {
		return nil
	}
	return &hookFailure{
		ErrMsg:     failure.GetErrorMessage(),
		Screenshot: base64.StdEncoding.EncodeToString(failure.GetScreenShot()),
		StackTrace: failure.GetStackTrace(),
	}
}

func getStatus(failed, skipped bool) status {
	if failed {
		return fail
	}
	if skipped {
		return skip
	}
	return pass
}

func getStepStatus(res *gauge_messages.ProtoStepExecutionResult) status {
	if res.GetSkipped() {
		return skip
	}
	if res.GetExecutionResult() == nil {
		return notExecuted
	}
	if res.GetExecutionResult().GetFailed() {
		return fail
	}
	return pass
}
