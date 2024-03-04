package mario

import (
	"fmt"
	"log"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/datafactory/armdatafactory/v3"
	"github.com/fatih/color"
)

func createHeader(
	headerTitle string,
	headerLength int,
	headerColor *color.Color,
	spacerChar string,
	centerHeader bool,
) string {
	headerTitleLength := utf8.RuneCountInString(headerTitle)
	header := headerColor.SprintfFunc()

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
