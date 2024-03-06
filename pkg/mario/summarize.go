package mario

import (
	"context"
	"fmt"
	"log"
	"slices"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/datafactory/armdatafactory/v3"
	"github.com/fatih/color"
	"github.com/rodaine/table"
)

type FactoryPipelineSummary struct {
	factoryName           string
	folder                string
	nPipelines            int
	nActivities           int
	nCopyActivities       int
	nDatabricksActivities int
}

func SummarizePipelines() {
	defer timer("SummarizePipelines")()
	factory := getFactoryClient()
	ctx := context.Background()

	pipelines := getAllPipelines(&factory, ctx)
	pipelineDetailsSummary := summarizePipelineDetails(factory, pipelines)
	printPipelineDetailsSummary(pipelineDetailsSummary)
}

func printPipelineDetailsSummary(pipelineSummary []FactoryPipelineSummary) {

	defer timer("printPipelineRunSummary")()
	headerLength := 80

	header := createHeader(
		"SUMMARIZE",
		headerLength,
		color.New(color.FgBlue),
		"=",
		true,
	)
	footer := createHeader(
		"",
		headerLength,
		color.New(color.FgHiCyan),
		"=",
		true,
	)
	fmt.Print("\n", header, "\n")

	headerFmt := color.New(color.Underline).SprintfFunc()
	columnFmt := color.New(color.FgYellow).SprintfFunc()

	tbl := table.New(
		"Factory",
		"Folder",
		"Pipelines",
		"Activities",
		"copy",
		"databricks",
	)

	tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt)

	for _, summary := range pipelineSummary {
		tbl.AddRow(
			summary.factoryName,
			summary.folder,
			summary.nPipelines,
			summary.nActivities,
			summary.nCopyActivities,
			summary.nDatabricksActivities,
		)
	}

	tbl.Print()

	fmt.Println(footer)
}

func summarizePipelineDetails(
	factory Factory,
	pipelines []*armdatafactory.PipelineResource,
) []FactoryPipelineSummary {
	defer timer("summarizePipelineDetails")()

	pipelineSummary := []FactoryPipelineSummary{}
	uniqueFolders := []string{}
	pipelinesByFolder := make(map[string][]armdatafactory.PipelineResource)

	for _, pipeline := range pipelines {
		pipelineFolder := ""
		pipelineFolderPointer := pipeline.Properties.Folder
		if pipelineFolderPointer == nil {
			pipelineFolder = "root"
		} else {
			pipelineFolder = *pipelineFolderPointer.Name
		}

		if !slices.Contains(uniqueFolders, pipelineFolder) {
			// add folder to map for the first time
			pipelinesByFolder[pipelineFolder] = []armdatafactory.PipelineResource{
				*pipeline,
			}
			uniqueFolders = append(uniqueFolders, pipelineFolder)
		} else {
			// append to existing folder
			pipelinesByFolder[pipelineFolder] = append(pipelinesByFolder[pipelineFolder], *pipeline)
		}
	}

	for folder, folderPipelines := range pipelinesByFolder {
		nPipelines := len(folderPipelines)

		nActivities := 0
		nCopyActivities := 0
		nDatabricksActivities := 0

		for _, pipeline := range folderPipelines {
			if pipeline.Properties.Activities == nil {
				continue
			}
			if len(pipeline.Properties.Activities) == 0 {
				continue
			}

			for _, activity := range pipeline.Properties.Activities {
				nActivities++
				switch *activity.GetActivity().Type {
				case "Copy":
					nCopyActivities++
				case "DatabricksNotebook":
					nDatabricksActivities++
				}
			}
		}

		pipelineSummary = append(pipelineSummary, FactoryPipelineSummary{
			factoryName:           factory.factoryName,
			folder:                folder,
			nPipelines:            nPipelines,
			nActivities:           nActivities,
			nCopyActivities:       nCopyActivities,
			nDatabricksActivities: nDatabricksActivities,
		})
	}
	return pipelineSummary
}

func SummarizeRuns(nDays int, name string) {
	defer timer("SummarizeRuns")()
	factory := getFactoryClient()
	ctx := context.Background()

	pipelineRuns, _ := getPipelineRuns(&factory, ctx, nDays, "")
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

	header := createHeader(
		"SUMMARIZE",
		headerLength,
		color.New(color.FgBlue),
		"=",
		true,
	)
	footer := createHeader(
		"",
		headerLength,
		color.New(color.FgHiCyan),
		"=",
		true,
	)

	fmt.Print("\n", header, "\n")

	if len(pipelineRunSummary) == 0 {
		fmt.Println("No pipeline runs found")
		fmt.Println(footer)
		return
	}

	headerFmt := color.New(color.Underline).SprintfFunc()
	columnFmt := color.New(color.FgYellow).SprintfFunc()

	tbl := table.New(
		"Pipeline",
		"Avg Time (min)",
		"Total Time (min)",
		"\u2714",
		"\u2718",
		"\u2022\u2022\u2022",
	)

	tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt)

	for _, summary := range pipelineRunSummary {
		summary.RuntimeAvgMin = summary.RuntimeTotalMin / float32(
			summary.Success+summary.Failed,
		)
		tbl.AddRow(
			summary.PipelineName,
			summary.RuntimeAvgMin,
			summary.RuntimeTotalMin,
			summary.Success,
			summary.Failed,
			summary.InProgress,
		)
	}

	tbl.Print()

	fmt.Println(footer)
}

func getAllPipelines(
	factory *Factory,
	ctx context.Context,
) []*armdatafactory.PipelineResource {
	// list all pipelines in the factory
	defer timer("getAllPipelines")()
	pipelineClient := factory.factoryClient.NewPipelinesClient()
	pager := pipelineClient.NewListByFactoryPager(
		factory.resouceGroupName,
		factory.factoryName,
		nil,
	)

	pipelines := []armdatafactory.PipelineResource{}

	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			log.Fatal(err)
		}

		for _, value := range page.Value {
			pipelines = append(pipelines, *value)
		}
	}

	fmt.Println("obtained", len(pipelines), "pipelines")

	var pipelinePointers []*armdatafactory.PipelineResource
	for _, pipeline := range pipelines {
		pipelinePointers = append(pipelinePointers, &pipeline)
	}

	return pipelinePointers
}

func getPipelineRuns(
	factory *Factory,
	ctx context.Context,
	nDays int,
	name string,
) (armdatafactory.PipelineRunsClientQueryByFactoryResponse, error) {
	defer timer("getPipelineRuns")()
	if nDays < 1 {
		log.Fatalf("nDays must be greater than 0")
	}

	if nDays > 30 {
		log.Fatalf("nDays must be less than 30")
	}

	runsFrom := time.Now().AddDate(0, 0, -nDays)
	runsTo := time.Now().
		AddDate(0, 0, 1)
		// add 1 day to include today and handle timezones

	runFilterParameters := armdatafactory.RunFilterParameters{
		LastUpdatedAfter:  &runsFrom,
		LastUpdatedBefore: &runsTo,
	}

	if name != "" {
		operand := armdatafactory.RunQueryFilterOperandPipelineName
		operator := armdatafactory.RunQueryFilterOperatorEquals
		nameFilter := make([]*string, 1)
		nameFilter[0] = &name
		filter := armdatafactory.RunQueryFilter{
			Operand:  &operand,
			Operator: &operator,
			Values:   nameFilter,
		}

		runFilterParameters = armdatafactory.RunFilterParameters{
			LastUpdatedAfter:  &runsFrom,
			LastUpdatedBefore: &runsTo,
			Filters:           []*armdatafactory.RunQueryFilter{&filter},
		}
	}

	log.Printf(
		"Getting pipeline runs from %s to %s",
		runsFrom.Format("2006-01-02"),
		runsTo.Format("2006-01-02"),
	)

	pipelineRunsClient := factory.factoryClient.NewPipelineRunsClient()
	pipelineRuns, err := pipelineRunsClient.QueryByFactory(
		ctx,
		factory.resouceGroupName,
		factory.factoryName,
		runFilterParameters,
		nil,
	)

	if err != nil {
		log.Fatal(err)
	}

	return pipelineRuns, nil
}
