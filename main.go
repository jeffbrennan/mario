package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/datafactory/armdatafactory/v3"
	"github.com/joho/godotenv"
)

var (
	datafactoryClientFactory *armdatafactory.ClientFactory
	factoriesClient          *armdatafactory.FactoriesClient
)

func main() {
	// setup
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file: %s", err)
	}

	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()

	subscriptionID := os.Getenv("AZ_SUBSCRIPTION_ID")
	resourceGroupName := os.Getenv("AZ_RESOURCE_GROUP")
	dataFactoryName := os.Getenv("AZ_DATAFACTORY_NAME")

	if dataFactoryName == "" {
		log.Fatal("AZ_DATAFACTORY_NAME is required")
	}

	datafactoryClientFactory, err = armdatafactory.NewClientFactory(subscriptionID, cred, nil)
	if err != nil {
		log.Fatal(err)
	}

	factoriesClient = datafactoryClientFactory.NewFactoriesClient()

	fmt.Println(resourceGroupName)
	fmt.Println(dataFactoryName)

	dataFactory, err := getDataFactory(ctx, resourceGroupName, dataFactoryName)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("get data factory:", *dataFactory.ID)

}

func getDataFactory(ctx context.Context, resourceGroupName string, dataFactoryName string) (*armdatafactory.Factory, error) {

	resp, err := factoriesClient.Get(ctx, resourceGroupName, dataFactoryName, nil)
	if err != nil {
		return nil, err
	}

	return &resp.Factory, nil
}
