package main

import (
	"encoding/base64"

	"github.com/getgauge-contrib/json-report/gauge_messages"
)

type tokenKind string
type status string
type errorType string
type parameterType string

const (
	pass                   status        = "passed"
	fail                   status        = "failed"
	skip                   status        = "skipped"
	notExecuted            status        = "not executed"
	stepKind               tokenKind     = "step"
	conceptKind            tokenKind     = "concept"
	assertionErrorType     errorType     = "assertion"
	verificationErrorType  errorType     = "verification"
	staticParameter        parameterType = "static"
	dynamicParameter       parameterType = "dynamic"
	specialStringParameter parameterType = "special_string"
	specialTableParameter  parameterType = "special_table"
	tableParameter         parameterType = "table"
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
	SpecResults            []spec       `json:"specResults"`
	BeforeSuiteHookFailure *hookFailure `json:"beforeSuiteHookFailure"`
	AfterSuiteHookFailure  *hookFailure `json:"afterSuiteHookFailure"`
	PassedSpecsCount       int          `json:"passedSpecsCount"`
	FailedSpecsCount       int          `json:"failedSpecsCount"`
	SkippedSpecsCount      int          `json:"skippedSpecsCount"`
	PassedScenariosCount   int          `json:"passedScenariosCount"`
	FailedScenariosCount   int          `json:"failedScenariosCount"`
	SkippedScenariosCount  int          `json:"skippedScenariosCount"`
}

type spec struct {
	SpecHeading           string       `json:"specHeading"`
	FileName              string       `json:"fileName"`
	Tags                  []string     `json:"tags"`
	ExecutionTime         int64        `json:"executionTime"`
	ExecutionStatus       status       `json:"executionStatus"`
	Scenarios             []scenario   `json:"scenarios"`
	IsTableDriven         bool         `json:"isTableDriven"`
	Datatable             *table       `json:"datatable"`
	BeforeSpecHookFailure *hookFailure `json:"beforeSpecHookFailure"`
	AfterSpecHookFailure  *hookFailure `json:"afterSpecHookFailure"`
	PassedScenarioCount   int          `json:"passedScenarioCount"`
	FailedScenarioCount   int          `json:"failedScenarioCount"`
	SkippedScenarioCount  int          `json:"skippedScenarioCount"`
}

type scenario struct {
	Heading                   string       `json:"scenarioHeading"`
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
	Parameters            []Parameter  `json:"parameters"`
	Table                 *table       `json:"table"`
	BeforeStepHookFailure *hookFailure `json:"beforeStepHookFailure"`
	AfterStepHookFailure  *hookFailure `json:"afterStepHookFailure"`
	Result                *result      `json:"result"`
}

func (s *step) kind() tokenKind {
	return stepKind
}

type Parameter struct {
	ParameterType parameterType `json:"parameterType"`
	Name          string        `json:"name"`
	Value         string        `json:"value"`
	Table         *table        `json:"table"`
}

type result struct {
	Status                status    `json:"status"`
	StackTrace            string    `json:"stackTrace"`
	Screenshot            string    `json:"screenshot"`
	ErrorMessage          string    `json:"errorMessage"`
	ExecutionTime         int64     `json:"executionTime"`
	SkippedReason         string    `json:"skippedReason"`
	Messages              []string  `json:"messages"`
	ErrorType             errorType `json:"errorType"`
	FailureScreenshotFile string    `json:"ScreenshotFile"`
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
	Headers []string `json:"headers"`
	Rows    []row    `json:"rows"`
}

type row struct {
	Cells []string `json:"cells"`
}

func toSuiteResult(psr *gauge_messages.ProtoSuiteResult) suiteResult {
	suiteResult := suiteResult{
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
	suiteResult.SpecResults = make([]spec, 0)
	for _, protoSpecRes := range psr.GetSpecResults() {
		spec := toSpec(protoSpecRes)
		suiteResult.SpecResults = append(suiteResult.SpecResults, spec)
		suiteResult.PassedScenariosCount += spec.PassedScenarioCount
		suiteResult.FailedScenariosCount += spec.FailedScenarioCount
		suiteResult.SkippedScenariosCount += spec.SkippedScenarioCount
	}
	return suiteResult
}

func toSpec(psr *gauge_messages.ProtoSpecResult) spec {
	spec := spec{
		SpecHeading:           psr.GetProtoSpec().GetSpecHeading(),
		IsTableDriven:         psr.GetProtoSpec().GetIsTableDriven(),
		FileName:              psr.GetProtoSpec().GetFileName(),
		Tags:                  make([]string, 0),
		FailedScenarioCount:   int(psr.GetScenarioFailedCount()),
		SkippedScenarioCount:  int(psr.GetScenarioSkippedCount()),
		PassedScenarioCount:   int(psr.GetScenarioCount() - psr.GetScenarioFailedCount() - psr.GetScenarioSkippedCount()),
		ExecutionTime:         psr.GetExecutionTime(),
		ExecutionStatus:       getStatus(psr.GetFailed(), psr.GetSkipped()),
		BeforeSpecHookFailure: toSpecHookFailure(psr.GetProtoSpec().GetPreHookFailures()),
		AfterSpecHookFailure:  toSpecHookFailure(psr.GetProtoSpec().GetPostHookFailures()),
	}
	if psr.GetProtoSpec().GetTags() != nil {
		spec.Tags = psr.GetProtoSpec().GetTags()
	}
	spec.Scenarios = make([]scenario, 0)
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

func toScenario(protoSce *gauge_messages.ProtoScenario, tableRowIndex int) scenario {
	sce := scenario{
		Heading:                   protoSce.GetScenarioHeading(),
		ExecutionTime:             protoSce.GetExecutionTime(),
		Tags:                      make([]string, 0),
		ExecutionStatus:           getScenarioStatus(protoSce),
		Contexts:                  make([]item, 0),
		Items:                     make([]item, 0),
		Teardowns:                 make([]item, 0),
		BeforeScenarioHookFailure: toHookFailure(protoSce.GetPreHookFailure()),
		AfterScenarioHookFailure:  toHookFailure(protoSce.GetPostHookFailure()),
		TableRowIndex:             tableRowIndex,
		SkipErrors:                make([]string, 0),
	}
	if protoSce.GetSkipErrors() != nil {
		sce.SkipErrors = protoSce.GetSkipErrors()
	}
	if protoSce.GetTags() != nil {
		sce.Tags = protoSce.GetTags()
	}
	if protoSce.GetContexts() != nil {
		sce.Contexts = toItems(protoSce.GetContexts())
	}
	if protoSce.GetScenarioItems() != nil {
		sce.Items = toItems(protoSce.GetScenarioItems())
	}
	if protoSce.GetTearDownSteps() != nil {
		sce.Teardowns = toItems(protoSce.GetTearDownSteps())
	}
	return sce
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
	rows := make([]row, len(protoTable.GetRows()))
	for i, r := range protoTable.GetRows() {
		rows[i] = row{
			Cells: r.GetCells(),
		}
	}
	headers := make([]string, 0)
	if protoTable.GetHeaders().GetCells() != nil {
		headers = protoTable.GetHeaders().GetCells()
	}
	return &table{Headers: headers, Rows: rows}
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
		Status:                getStepStatus(protoStep.GetStepExecutionResult()),
		Screenshot:            base64.StdEncoding.EncodeToString(res.GetScreenShot()),
		StackTrace:            res.GetStackTrace(),
		ErrorMessage:          res.GetErrorMessage(),
		ExecutionTime:         res.GetExecutionTime(),
		Messages:              make([]string, 0),
		ErrorType:             getErrorType(res.GetErrorType()),
		FailureScreenshotFile: res.GetFailureScreenshotFile(),
	}
	if protoStep.GetStepExecutionResult().GetSkipped() {
		result.SkippedReason = protoStep.GetStepExecutionResult().GetSkippedReason()
	}
	if res.GetMessage() != nil {
		result.Messages = res.GetMessage()
	}
	step := &step{
		ItemType:              stepKind,
		StepText:              protoStep.GetActualText(),
		Result:                result,
		BeforeStepHookFailure: toHookFailure(protoStep.GetStepExecutionResult().GetPreHookFailure()),
		AfterStepHookFailure:  toHookFailure(protoStep.GetStepExecutionResult().GetPostHookFailure()),
	}

	params := make([]Parameter, 0)

	if protoStep.GetFragments() != nil {
		for _, f := range protoStep.GetFragments() {
			if f.GetFragmentType() != gauge_messages.Fragment_Parameter {
				continue
			}
			p := f.GetParameter()
			switch p.GetParameterType() {
			case gauge_messages.Parameter_Table:
				table := toTable(p.GetTable())
				step.Table = table
				params = append(params, Parameter{ParameterType: tableParameter, Name: p.GetName(), Table: table})
				break
			case gauge_messages.Parameter_Special_Table:
				table := toTable(p.GetTable())
				step.Table = table
				params = append(params, Parameter{ParameterType: specialTableParameter, Name: p.GetName(), Table: table})
				break
			case gauge_messages.Parameter_Dynamic:
				params = append(params, Parameter{ParameterType: dynamicParameter, Name: p.GetName(), Value: p.GetValue()})
				break
			case gauge_messages.Parameter_Special_String:
				params = append(params, Parameter{ParameterType: specialStringParameter, Name: p.GetName(), Value: p.GetValue()})
				break
			default:
				params = append(params, Parameter{ParameterType: staticParameter, Name: p.GetName(), Value: p.GetValue()})
				break
			}
		}
	}
	step.Parameters = params
	return step
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

func toSpecHookFailure(failures []*gauge_messages.ProtoHookFailure) *hookFailure {
	if len(failures) > 0 {
		return toHookFailure(failures[0])
	}
	return nil
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
