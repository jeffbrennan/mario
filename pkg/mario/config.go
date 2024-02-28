package mario

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
)

type AZEnv struct {
	SubscriptionID    string
	ResourceGroupName string
	DataFactoryName   string
}

func ConfigSetup() {

	azSubscriptionID := parseInput("Enter your Azure Subscription ID: ")
	azResourceGroupName := parseInput("Enter your Resource Group Name: ")
	azDataFactoryName := parseInput("Enter your Data Factory Name: ")

	azEnv := AZEnv{
		SubscriptionID:    azSubscriptionID,
		ResourceGroupName: azResourceGroupName,
		DataFactoryName:   azDataFactoryName,
	}

	writeConfig(azEnv)
	fmt.Println("Azure environment details added to the CLI.")
}

func writeConfig(azEnv AZEnv) {
	f, err := os.Create(".mariocfg")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	_, err = f.WriteString(azEnv.SubscriptionID)
	if err != nil {
		log.Fatal(err)
	}
	_, err = f.WriteString(azEnv.ResourceGroupName)
	if err != nil {
		log.Fatal(err)
	}
	_, err = f.WriteString(azEnv.DataFactoryName)
	if err != nil {
		log.Fatal(err)
	}
}

func readConfig() AZEnv {
	f, err := os.Open(".mariocfg")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	reader := bufio.NewReader(f)
	azSubscriptionID, err := reader.ReadString('\n')
	if err != nil {
		log.Fatal(err)
	}
	azResourceGroupName, err := reader.ReadString('\n')
	if err != nil {
		log.Fatal(err)
	}
	azDataFactoryName, err := reader.ReadString('\n')
	if err != nil {
		log.Fatal(err)
	}

	azEnv := AZEnv{
		SubscriptionID:    strings.TrimSpace(azSubscriptionID),
		ResourceGroupName: strings.TrimSpace(azResourceGroupName),
		DataFactoryName:   strings.TrimSpace(azDataFactoryName),
	}

	return azEnv
}

func parseInput(input string) string {
	fmt.Println(input)
	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil {
		log.Fatal(err)
	}
	return line
}

func HelloConfig() {
	fmt.Println("Hello from config")
}
