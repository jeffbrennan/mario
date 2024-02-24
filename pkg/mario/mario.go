package mario

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/datafactory/armdatafactory/v3"
	"github.com/joho/godotenv"
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

func Summarize(nDays int) {
	defer timer("Summarize")()
	// setup
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file: %s", err)
	}
	subscriptionID := getEnvironmentVariable("AZ_SUBSCRIPTION_ID")
	resourceGroupName := getEnvironmentVariable("AZ_RESOURCE_GROUP")
	dataFactoryName := getEnvironmentVariable("AZ_DATAFACTORY_NAME")

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
	printPipelineRunSummary(pipelineSummary)
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
		fmt.Println()
		fmt.Println(summary.PipelineName, "============")
		fmt.Println("Success: ", summary.Success)
		fmt.Println("Failed: ", summary.Failed)
		fmt.Println("In Progress: ", summary.InProgress)
		fmt.Println("Avg Runtime: ", summary.RuntimeAvgMin)
	}

	summaryEnd := strings.Repeat("=", headerLength)
	fmt.Println(summaryEnd)

}

func printPipelineRuns(
	runs armdatafactory.PipelineRunsClientQueryByFactoryResponse,
) {
	for _, run := range runs.Value {
		fmt.Printf(
			"Pipeline Name: %s, Status: %s, Duration: %s\n",
			*run.PipelineName,
			*run.Status,
			fmt.Sprint(*run.DurationInMs),
		)
	}
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

	pipelineRunsClient := datafactoryClientFactory.NewPipelineRunsClient()
	runsFrom := time.Now().AddDate(0, 0, -nDays)
	runsTo := time.Now()

	runFilterParameters := armdatafactory.RunFilterParameters{
		LastUpdatedAfter:  &runsFrom,
		LastUpdatedBefore: &runsTo,
	}

	// print pipeline runs as date
	log.Printf(
		"Getting pipeline runs from %s to %s",
		runsFrom.Format("2006-01-02"),
		runsTo.Format("2006-01-02"),
	)
	return pipelineRunsClient.QueryByFactory(
		ctx,
		resourceGroupName,
		dataFactoryName,
		runFilterParameters,
		nil,
	)
}

func getEnvironmentVariable(key string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		log.Fatalf("Environment variable %s is not set", key)
	}

	if value == "" {
		log.Fatalf("Environment variable %s is empty", key)
	}

	return value
}

func timer(name string) func() {
	start := time.Now()
	return func() {
		fmt.Printf("%s took %v\n", name, time.Since(start))
	}
}
