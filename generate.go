package main

import (
	"encoding/base64"

	"github.com/getgauge/json-report/gauge_messages"
)

type tokenKind string
type status string
type errorType string

const (
	pass                  status    = "pass"
	fail                  status    = "fail"
	skip                  status    = "skip"
	notExecuted           status    = "not executed"
	stepKind              tokenKind = "step"
	conceptKind           tokenKind = "concept"
	tableKind             tokenKind = "table"
	assertionErrorType    errorType = "assertion"
	verificationErrorType errorType = "verification"
)

type item interface {
	kind() tokenKind
}

type suiteResult struct {
	ProjectName            string       `json:"projectName"`
	Timestamp              string       `json:"timestamp"`
	SuccessRate            int          `json:"successRate"`
	Environment            string       `json:"environment"`
	Tags                   string       `json:"tags"`
	ExecutionTime          int64        `json:"executionTime"`
	ExecutionStatus        status       `json:"executionStatus"`
	SpecResults            []*spec      `json:"specResults"`
	BeforeSuiteHookFailure *hookFailure `json:"beforeSuiteHookFailure"`
	AfterSuiteHookFailure  *hookFailure `json:"afterSuiteHookFailure"`
	PassedSpecsCount       int          `json:"passedSpecsCount"`
	FailedSpecsCount       int          `json:"failedSpecsCount"`
	SkippedSpecsCount      int          `json:"skippedSpecsCount"`
	UnhandledErrors        []error      `json:"unhandledErrors"`
}

type spec struct {
	SpecHeading           string       `json:"specHeading"`
	FileName              string       `json:"fileName"`
	Tags                  []string     `json:"tags"`
	ExecutionTime         int64        `json:"executionTime"`
	ExecutionStatus       status       `json:"executionStatus"`
	Scenarios             []*scenario  `json:"scenarios"`
	IsTableDriven         bool         `json:"isTableDriven"`
	Datatable             *table       `json:"datatable"`
	BeforeSpecHookFailure *hookFailure `json:"beforeSpecHookFailure"`
	AfterSpecHookFailure  *hookFailure `json:"afterSpecHookFailure"`
	PassedScenarioCount   int          `json:"PassedScenarioCount"`
	FailedScenarioCount   int          `json:"FailedScenarioCount"`
	SkippedScenarioCount  int          `json:"SkippedScenarioCount"`
}

type scenario struct {
	Heading                   string       `json:"heading"`
	Tags                      []string     `json:"tags"`
	ExecutionTime             int64        `json:"executionTime"`
	ExecutionStatus           status       `json:"executionStatus"`
	Contexts                  []item       `json:"contexts"`
	Teardowns                 []item       `json:"teardowns"`
	Items                     []item       `json:"items"`
	BeforeScenarioHookFailure *hookFailure `json:"beforeScenarioHookFailure"`
	AfterScenarioHookFailure  *hookFailure `json:"afterScenarioHookFailure"`
	SkipErrors                []string     `json:"skipErrors"`
	TableRowIndex             int          `json:"tableRowIndex"`
}

type step struct {
	ItemType              tokenKind    `json:"itemType"`
	StepText              string       `json:"stepText"`
	BeforeStepHookFailure *hookFailure `json:"beforeStepHookFailure"`
	AfterStepHookFailure  *hookFailure `json:"afterStepHookFailure"`
	Result                *result      `json:"result"`
}

func (s *step) kind() tokenKind {
	return stepKind
}

type result struct {
	Status        status    `json:"status"`
	StackTrace    string    `json:"stackTrace"`
	Screenshot    string    `json:"screenshot"`
	ErrorMessage  string    `json:"errorMessage"`
	ExecutionTime int64     `json:"executionTime"`
	SkippedReason string    `json:"skippedReason"`
	Messages      []string  `json:"messages"`
	ErrorType     errorType `json:"errorType"`
}

type hookFailure struct {
	ErrMsg     string `json:"errorMessage"`
	Screenshot string `json:"screenshot"`
	StackTrace string `json:"stackTrace"`
}

type concept struct {
	ItemType    tokenKind `json:"itemType"`
	ConceptStep *step     `json:"conceptStep"`
	Items       []item    `json:"items"`
	Result      result    `json:"result"`
}

func (s *concept) kind() tokenKind {
	return conceptKind
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
		suiteResult.SpecResults = append(suiteResult.SpecResults, toSpec(protoSpecRes))
	}
	return suiteResult
}

func toSpec(psr *gauge_messages.ProtoSpecResult) *spec {
	spec := &spec{
		SpecHeading:           psr.GetProtoSpec().GetSpecHeading(),
		IsTableDriven:         psr.GetProtoSpec().GetIsTableDriven(),
		FileName:              psr.GetProtoSpec().GetFileName(),
		Tags:                  psr.GetProtoSpec().GetTags(),
		FailedScenarioCount:   int(psr.GetScenarioFailedCount()),
		SkippedScenarioCount:  int(psr.GetScenarioSkippedCount()),
		PassedScenarioCount:   int(psr.GetScenarioCount() - psr.GetScenarioFailedCount() - psr.GetScenarioSkippedCount()),
		ExecutionTime:         psr.GetExecutionTime(),
		ExecutionStatus:       getStatus(psr.GetFailed(), psr.GetSkipped()),
		BeforeSpecHookFailure: toHookFailure(psr.GetProtoSpec().GetPreHookFailure()),
		AfterSpecHookFailure:  toHookFailure(psr.GetProtoSpec().GetPostHookFailure()),
	}
	for _, item := range psr.GetProtoSpec().GetItems() {
		switch item.GetItemType() {
		case gauge_messages.ProtoItem_Scenario:
			spec.Scenarios = append(spec.Scenarios, toScenario(item.GetScenario(), -1))
		case gauge_messages.ProtoItem_TableDrivenScenario:
			spec.Scenarios = append(spec.Scenarios, toScenario(item.GetTableDrivenScenario().GetScenario(), int(item.GetTableDrivenScenario().GetTableRowIndex())))
		case gauge_messages.ProtoItem_Table:
			spec.Datatable = toTable(item.GetTable())
		}
	}
	return spec
}

func toScenario(scn *gauge_messages.ProtoScenario, tableRowIndex int) *scenario {
	return &scenario{
		Heading:                   scn.GetScenarioHeading(),
		ExecutionTime:             scn.GetExecutionTime(),
		Tags:                      scn.GetTags(),
		ExecutionStatus:           getScenarioStatus(scn),
		Contexts:                  toItems(scn.GetContexts()),
		Items:                     toItems(scn.GetScenarioItems()),
		Teardowns:                 toItems(scn.GetTearDownSteps()),
		BeforeScenarioHookFailure: toHookFailure(scn.GetPreHookFailure()),
		AfterScenarioHookFailure:  toHookFailure(scn.GetPostHookFailure()),
		TableRowIndex:             tableRowIndex,
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
		case gauge_messages.ProtoItem_Concept:
			items = append(items, toConcept(i.GetConcept()))
		}
	}
	return items
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
		ErrorType:     getErrorType(res.GetErrorType()),
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
func getErrorType(protoErrType gauge_messages.ProtoExecutionResult_ErrorType) errorType {
	if protoErrType == gauge_messages.ProtoExecutionResult_VERIFICATION {
		return verificationErrorType
	}
	return assertionErrorType
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
