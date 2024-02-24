package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/datafactory/armdatafactory/v3"
	"github.com/joho/godotenv"
)

var (
	datafactoryClientFactory *armdatafactory.ClientFactory
)

func main() {
	defer timer("main")()
	fmt.Println("============= Mario =============")
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

	datafactoryClientFactory, err = armdatafactory.NewClientFactory(subscriptionID, cred, nil)
	if err != nil {
		log.Fatal(err)
	}

	pipelineRuns, _ := getPipelineRuns(ctx, resourceGroupName, dataFactoryName, 7)

	fmt.Println("Pipeline runs:")
	for _, run := range pipelineRuns.Value {
		fmt.Printf("Run ID: %s, Status: %s, Duration: %s\n", *run.RunID, *run.Status, fmt.Sprint(*run.DurationInMs))
	}

	fmt.Println("=================================")

}

func getPipelineRuns(ctx context.Context, resourceGroupName string, dataFactoryName string, nDays int) (armdatafactory.PipelineRunsClientQueryByFactoryResponse, error) {
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
	log.Printf("Getting pipeline runs from %s to %s", runsFrom.Format("2006-01-02"), runsTo.Format("2006-01-02"))
	return pipelineRunsClient.QueryByFactory(ctx, resourceGroupName, dataFactoryName, runFilterParameters, nil)
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
