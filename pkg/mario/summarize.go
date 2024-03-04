package mario

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/datafactory/armdatafactory/v3"
	"github.com/fatih/color"
	"github.com/rodaine/table"
)

func Summarize(nDays int, name string) {
	defer timer("Summarize")()
	factory := getFactoryClient()
	ctx := context.Background()

	pipelineRuns, _ := getPipelineRuns(&factory, ctx, nDays)
	pipelineSummary := summarizePipelineRuns(pipelineRuns)

	if name == "" {
		printPipelineRunSummary(pipelineSummary)
		return
	}

	// if name is not empty, filter the pipeline summary
	filteredPipelineSummary := make(map[string]PipelineRunSummary)
	for _, summary := range pipelineSummary {
		if strings.Contains(summary.PipelineName, name) {
			filteredPipelineSummary[summary.PipelineName] = summary
		}
	}
	printPipelineRunSummary(filteredPipelineSummary)
}

func summarizePipelineRuns(
	runs armdatafactory.PipelineRunsClientQueryByFactoryResponse,
) map[string]PipelineRunSummary {
	defer timer("summarizePipelineRuns")()
	pipelineRunSummary := make(map[string]PipelineRunSummary)
	for _, run := range runs.Value {
		summary, exists := pipelineRunSummary[*run.PipelineName]
		if !exists {
			summary = PipelineRunSummary{
				PipelineName: *run.PipelineName,
			}
		}

		switch *run.Status {
		case "Succeeded":
			summary.Success++
		case "Failed":
			summary.Failed++
		case "InProgress":
			summary.InProgress++
		}

		summary.RuntimeTotalMin += float32(*run.DurationInMs) / (1000 * 60)
		pipelineRunSummary[*run.PipelineName] = summary
	}

	return pipelineRunSummary
}

func printPipelineRunSummary(pipelineRunSummary map[string]PipelineRunSummary) {
	defer timer("printPipelineRunSummary")()
	headerLength := 80

	header := createHeader("SUMMARIZE", headerLength, color.New(color.FgBlue), "=", true)
	footer := createHeader("", headerLength, color.New(color.FgHiCyan), "=", true)
	fmt.Print("\n", header, "\n")

	headerFmt := color.New(color.Underline).SprintfFunc()
	columnFmt := color.New(color.FgYellow).SprintfFunc()

	tbl := table.New(
		"Pipeline",
		"Success",
		"Failed",
		"In Progress",
		"Avg Runtime (min)",
	)

	tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt)

	for _, summary := range pipelineRunSummary {
		summary.RuntimeAvgMin = summary.RuntimeTotalMin / float32(
			summary.Success+summary.Failed,
		)
		tbl.AddRow(
			summary.PipelineName,
			summary.Success,
			summary.Failed,
			summary.InProgress,
			summary.RuntimeAvgMin,
		)
	}

	tbl.Print()

	fmt.Println(footer)
}

func getPipelineRuns(
	factory *Factory,
	ctx context.Context,
	nDays int,
) (armdatafactory.PipelineRunsClientQueryByFactoryResponse, error) {
	defer timer("getPipelineRuns")()
	if nDays < 1 {
		log.Fatalf("nDays must be greater than 0")
	}

	if nDays > 30 {
		log.Fatalf("nDays must be less than 30")
	}

	runsFrom := time.Now().AddDate(0, 0, -nDays)
	runsTo := time.Now()

	runFilterParameters := armdatafactory.RunFilterParameters{
		LastUpdatedAfter:  &runsFrom,
		LastUpdatedBefore: &runsTo,
	}

	log.Printf(
		"Getting pipeline runs from %s to %s",
		runsFrom.Format("2006-01-02"),
		runsTo.Format("2006-01-02"),
	)

	pipelineRunsClient := factory.factoryClient.NewPipelineRunsClient()
	query_start := time.Now()
	pipelineRuns, err := pipelineRunsClient.QueryByFactory(
		ctx,
		factory.resouceGroupName,
		factory.factoryName,
		runFilterParameters,
		nil,
	)

	log.Printf("Query took %v", time.Since(query_start))

	if err != nil {
		log.Fatal(err)
	}

	return pipelineRuns, nil
}
