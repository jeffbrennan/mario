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
	factoriesClient          *armdatafactory.FactoriesClient
)

func main() {
	defer timer("main")()
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

	factoriesClient = datafactoryClientFactory.NewFactoriesClient()
	dataFactory, err := getDataFactory(ctx, resourceGroupName, dataFactoryName)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("get data factory:", *dataFactory.ID)

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

func getDataFactory(ctx context.Context, resourceGroupName string, dataFactoryName string) (*armdatafactory.Factory, error) {

	resp, err := factoriesClient.Get(ctx, resourceGroupName, dataFactoryName, nil)
	if err != nil {
		return nil, err
	}

	return &resp.Factory, nil
}

func timer(name string) func() {
	start := time.Now()
	return func() {
		fmt.Printf("%s took %v\n", name, time.Since(start))
	}
}
