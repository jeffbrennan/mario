package mario

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/datafactory/armdatafactory/v3"
)

var (
	datafactoryClientFactory *armdatafactory.ClientFactory
)

type PipelineRunSummary struct {
	PipelineName    string
	Success         int
	Failed          int
	InProgress      int
	RuntimeTotalMin float32
	RuntimeAvgMin   float32
}

func Summarize(nDays int, name string) {
	defer timer("Summarize")()

	azEnv := readConfig()
	subscriptionID := azEnv.SubscriptionID
	resourceGroupName := azEnv.ResourceGroupName
	dataFactoryName := azEnv.DataFactoryName

	ctx := context.Background()

	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		log.Fatal(err)
	}

	datafactoryClientFactory, err = armdatafactory.NewClientFactory(
		subscriptionID,
		cred,
		nil,
	)
	if err != nil {
		log.Fatal(err)
	}

	pipelineRuns, _ := getPipelineRuns(
		ctx,
		resourceGroupName,
		dataFactoryName,
		nDays,
	)

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
	headerLength := 50
	headerTitle := "Pipeline Summary"
	headerSpace := (headerLength - len(headerTitle)) / 2
	fmt.Println(
		strings.Repeat("=", headerSpace),
		headerTitle,
		strings.Repeat("=", headerSpace),
	)

	// TODO: make a table
	for _, summary := range pipelineRunSummary {
		summary.RuntimeAvgMin = summary.RuntimeTotalMin / float32(
			summary.Success+summary.Failed,
		)
		fmt.Println(summary.PipelineName, "============")
		fmt.Println("Success: ", summary.Success)
		fmt.Println("Failed: ", summary.Failed)
		fmt.Println("In Progress: ", summary.InProgress)
		fmt.Println("Avg Runtime: ", summary.RuntimeAvgMin)
	}

	summaryEnd := strings.Repeat("=", headerLength)
	fmt.Println(summaryEnd)

}

func getPipelineRuns(
	ctx context.Context,
	resourceGroupName string,
	dataFactoryName string,
	nDays int,
) (armdatafactory.PipelineRunsClientQueryByFactoryResponse, error) {

	if nDays < 1 {
		log.Fatalf("nDays must be greater than 0")
	}

	if nDays > 30 {
		log.Fatalf("nDays must be less than 30")
	}

	runsFrom := time.Now().AddDate(0, 0, -nDays)
	runsTo := time.Now()

	pipelineRunsClient := datafactoryClientFactory.NewPipelineRunsClient()

	runFilterParameters := armdatafactory.RunFilterParameters{
		LastUpdatedAfter:  &runsFrom,
		LastUpdatedBefore: &runsTo,
	}

	log.Printf(
		"Getting pipeline runs from %s to %s",
		runsFrom.Format("2006-01-02"),
		runsTo.Format("2006-01-02"),
	)
	pipelineRuns, _ := pipelineRunsClient.QueryByFactory(
		ctx,
		resourceGroupName,
		dataFactoryName,
		runFilterParameters,
		nil,
	)
	return pipelineRuns, nil
}
func timer(name string) func() {
	start := time.Now()
	return func() {
		fmt.Printf("%s took %v\n", name, time.Since(start))
	}
}
