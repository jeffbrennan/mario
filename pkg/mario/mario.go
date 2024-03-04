package mario

import (
	"os"

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

type Factory struct {
	subscriptionID   string
	resouceGroupName string
	factoryName      string
	factoryClient    *armdatafactory.ClientFactory
}

func Exit() {
	os.Exit(0)
}
