package mario

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/datafactory/armdatafactory/v3"
	"github.com/fatih/color"
	"github.com/go-test/deep"
	"github.com/rodaine/table"
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

type Factory struct {
	resouceGroupName string
	factoryName      string
	factoryClient    *armdatafactory.ClientFactory
	ctx              context.Context
}

func Compare(name1 string, name2 string) {
	defer timer("Compare")()
	factory := getFactoryClient()

	wg := sync.WaitGroup{}
	wg.Add(2)

	pipelineChan := make(chan armdatafactory.PipelinesClientGetResponse, 2)

	go getPipeline(name1, factory, pipelineChan, &wg)
	go getPipeline(name2, factory, pipelineChan, &wg)

	wg.Wait()
	close(pipelineChan)

	pipeline1 := <-pipelineChan
	pipeline2 := <-pipelineChan

	pipeline1Json, _ := pipeline1.MarshalJSON()
	pipeline2Json, _ := pipeline2.MarshalJSON()

	pipeline1Map := jsonToMap(string(pipeline1Json))
	pipeline2Map := jsonToMap(string(pipeline2Json))

	diff := deep.Equal(pipeline1Map, pipeline2Map)
	fmt.Println(diff)

}

func jsonToMap(jsonStr string) map[string]interface{} {
	result := make(map[string]interface{})
	json.Unmarshal([]byte(jsonStr), &result)
	return result
}

func getPipeline(
	name string,
	factory Factory,
	pipelineChan chan armdatafactory.PipelinesClientGetResponse,
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	pipelineClient := factory.factoryClient.NewPipelinesClient()
	pipeline, err := pipelineClient.Get(
		factory.ctx,
		factory.resouceGroupName,
		factory.factoryName,
		name,
		nil,
	)

	if err != nil {
		log.Fatal(err)
	}

	pipelineChan <- pipeline
}

func getFactoryClient() Factory {
	azEnv := readConfig()
	subscriptionID := azEnv.SubscriptionID
	resourceGroupName := azEnv.ResourceGroupName
	dataFactoryName := azEnv.DataFactoryName

	ctx := context.Background()

	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		log.Fatal(err)
	}

	datafactoryClientFactory, _ = armdatafactory.NewClientFactory(
		subscriptionID,
		cred,
		nil,
	)

	return Factory{
		resouceGroupName: resourceGroupName,
		factoryName:      dataFactoryName,
		factoryClient:    datafactoryClientFactory,
		ctx:              ctx,
	}

}

func Exit() {
	os.Exit(0)
}

func Summarize(nDays int, name string) {
	defer timer("Summarize")()
	factory := getFactoryClient()

	pipelineRuns, _ := getPipelineRuns(&factory, nDays)
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
	headerFmt := color.New(color.FgGreen, color.Underline).SprintfFunc()
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
}

func getPipelineRuns(
	factory *Factory,
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

	pipelineRuns, err := pipelineRunsClient.QueryByFactory(
		factory.ctx,
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

func timer(name string) func() {
	start := time.Now()
	return func() {
		fmt.Printf("%s took %v\n", name, time.Since(start))
	}
}
