package mario

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
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
	subscriptionID   string
	resouceGroupName string
	factoryName      string
	factoryClient    *armdatafactory.ClientFactory
}

func Compare(name1 string, name2 string) {
	defer timer("Compare")()
	factory := getFactoryClient()
	ctx := context.Background()
	wg := sync.WaitGroup{}
	wg.Add(2)

	pipelineChan := make(chan armdatafactory.PipelinesClientGetResponse, 2)
	go getPipeline(name1, factory, ctx, pipelineChan, &wg)
	go getPipeline(name2, factory, ctx, pipelineChan, &wg)

	wg.Wait()
	close(pipelineChan)

	pipeline1 := <-pipelineChan
	pipeline2 := <-pipelineChan

	pipeline1Map := parsePipeline(
		pipeline1,
		[]string{"id", "etag", "name", "type"},
	)
	pipeline2Map := parsePipeline(
		pipeline2,
		[]string{"id", "etag", "name", "type"},
	)

	// try out different equality checks
	diffRaw := deep.Equal(pipeline1Map, pipeline2Map)
	differencesExist := diffRaw != nil

	var diff []string
	if differencesExist {
		diff = formatDiff(
			diffRaw,
			[]string{"properties", "activities", "slice"},
		)
	} else {
		diff = []string{"No differences found"}
	}

	printDiffOutput(diff, differencesExist, name1, name2)
}

func printDiffOutput(
	diff []string,
	differencesExist bool,
	name1 string,
	name2 string,
) {

	check := color.New(color.FgGreen).SprintFunc()
	cross := color.New(color.FgRed).SprintFunc()
	headerLength := 80

	pipeline1Color := color.New(color.FgYellow).SprintFunc()
	pipeline2Color := color.New(color.FgCyan).SprintFunc()

	header := createHeader("COMPARE", headerLength, color.New(color.FgBlue), "=", true)
	footer := createHeader("", headerLength, color.New(color.FgWhite), "=", true)

	fmt.Print("\n", header, "\n")

	if differencesExist {
		fmt.Println(
			"[",
			cross("\u2718"),
			"]",
			pipeline1Color(name1),
			"|",
			pipeline2Color(name2),
		)
	} else {
		fmt.Println("[", check("\u2714"), "]", pipeline1Color(name1), "|", pipeline2Color(name2))
		fmt.Println(check("No differences found"))
		fmt.Println(footer)
		return
	}

	fmt.Print("\n", cross("Differences found"), "\n")
	for i, d := range diff {
		diffSplit := strings.Split(d, "\n")

		location := diffSplit[:len(diffSplit)-1]
		value := diffSplit[len(diffSplit)-1]
		value = strings.Trim(value, " ")

		valueSplit := strings.Split(value, "!=")

		value1 := valueSplit[1]
		value2 := valueSplit[0]
		valueFormatted := pipeline1Color(
			value1,
		) + " != " + pipeline2Color(
			value2,
		)

		diffMessage := "[" + strconv.Itoa(
			i+1,
		) + "/" + strconv.Itoa(
			len(diff),
		) + "]"

		diffHeader := createHeader(
			diffMessage,
			headerLength,
			color.New(color.FgWhite),
			"-",
			false,
		)
		fmt.Print(diffHeader)

		fmt.Print("\n", strings.Join(location, "\n"))
		fmt.Print(":", valueFormatted, "\n\n")

	}
	fmt.Println(footer)
}

func formatDiff(diff []string, keysToExclude []string) []string {
	formattedDiff := make([]string, len(diff))

	for i, d := range diff {
		diffSplit := strings.Split(d, ": ")
		location := diffSplit[0]
		locationParts := strings.Split(location, ".")

		formattedPartsRaw := make([]string, len(locationParts))
		// initial string cleaning
		for i, part := range locationParts {
			formattedPart := strings.Replace(part, "]", "", -1)
			formattedPart = strings.Replace(formattedPart, "map[", "", -1)
			formattedPart = strings.Replace(formattedPart, "[", "", -1)
			formattedPartsRaw[i] = formattedPart
		}

		// remove keys that we don't want to show
		formattedParts := []string{}
		for _, part := range formattedPartsRaw {
			for _, key := range keysToExclude {
				if strings.Contains(part, key) {
					part = ""
				}
			}
			if part == "" {
				continue
			}
			formattedParts = append(formattedParts, part)
		}

		// add indentation based on the depth of the key
		for i, part := range formattedParts {
			part = strings.Repeat(" ", i*2) + part + "\n"
			formattedParts[i] = part
		}

		formattedPartsString := strings.Join(formattedParts, "")

		value := diffSplit[1]
		formattedValue := strings.Repeat(" ", len(formattedParts)*2) + value
		formattedDiff[i] = formattedPartsString + formattedValue
	}
	return formattedDiff
}

func parsePipeline(
	pipeline armdatafactory.PipelinesClientGetResponse,
	keysToDrop []string,
) map[string]interface{} {
	pipelineJson, _ := pipeline.MarshalJSON()
	pipelineMap := jsonToMap(string(pipelineJson))
	pipelineMapClean := cleanMap(pipelineMap, keysToDrop)
	return pipelineMapClean
}

func cleanMap(
	m map[string]interface{},
	keysToDrop []string,
) map[string]interface{} {
	for _, key := range keysToDrop {
		delete(m, key)
	}
	return m
}

func jsonToMap(jsonStr string) map[string]interface{} {
	result := make(map[string]interface{})
	json.Unmarshal([]byte(jsonStr), &result)
	return result
}

func getPipelineHttp(
	factory Factory,
	pipelineName string,
	pipelineChan chan http.Response,
	wg *sync.WaitGroup,
) {
	defer timer("getPipelineHttp")
	defer wg.Done()
	requestString := fmt.Sprintf(
		"https://management.azure.com/subscriptions/%s/resourceGroups/%s/providers/Microsoft.DataFactory/factories/%s/pipelines/%s?api-version=2018-06-01",
		factory.subscriptionID,
		factory.resouceGroupName,
		factory.factoryName,
		pipelineName,
	)
	req, err := http.NewRequest("GET", requestString, nil)
	if err != nil {
		log.Fatal(err)
	}

	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		log.Fatal(err)
	}

	token, err := cred.GetToken(
		context.TODO(),
		policy.TokenRequestOptions{
			Scopes: []string{"https://management.azure.com/.default"},
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("Authorization", "Bearer "+token.Token)
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(resp)

	pipelineChan <- *resp

}

func getPipeline(
	name string,
	factory Factory,
	ctx context.Context,
	pipelineChan chan armdatafactory.PipelinesClientGetResponse,
	wg *sync.WaitGroup,
) {
	defer timer("getPipeline")()
	defer wg.Done()

	pipelineClient := factory.factoryClient.NewPipelinesClient()
	pipeline, err := pipelineClient.Get(
		ctx,
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
	defer timer("getFactoryClient")()
	azEnv := readConfig()
	subscriptionID := azEnv.SubscriptionID
	resourceGroupName := azEnv.ResourceGroupName
	dataFactoryName := azEnv.DataFactoryName

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
		subscriptionID:   subscriptionID,
		resouceGroupName: resourceGroupName,
		factoryName:      dataFactoryName,
		factoryClient:    datafactoryClientFactory,
	}

}

func Exit() {
	os.Exit(0)
}

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

	header := createHeader("SUMMARIZE", headerLength, color.New(color.FgHiCyan), "=", true)
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

func createHeader(
	headerTitle string,
	headerLength int,
	headerColor *color.Color,
	spacerChar string,
	centerHeader bool,
) string {
	headerTitleLength := utf8.RuneCountInString(headerTitle)
	header := headerColor.SprintFunc()

	if headerTitle == "" {
		return strings.Repeat(spacerChar, headerLength)
	}

	if centerHeader {
		spacerLengthStart := ((headerLength - headerTitleLength) / 2) - 1
		spacerLengthEnd := spacerLengthStart
		if spacerLengthStart*2 < headerLength-headerTitleLength {
			spacerLengthStart++
		}

		return strings.Repeat(
			spacerChar,
			spacerLengthStart,
		) + " " + header(
			headerTitle,
		) + " " + strings.Repeat(
			spacerChar,
			spacerLengthEnd,
		)

	}

	spacerLength := headerLength - headerTitleLength
	return header(
		headerTitle,
	) + " " + strings.Repeat(
		spacerChar,
		spacerLength-1,
	)

}

func timer(name string) func() {
	start := time.Now()
	return func() {
		fmt.Printf("%s took %v\n", name, time.Since(start))
	}
}
