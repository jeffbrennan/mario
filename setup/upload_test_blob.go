package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

var (
	configFileSchema = &hcl.BodySchema{
		Blocks: []hcl.BlockHeaderSchema{
			{
				Type:       "variable",
				LabelNames: []string{"name"},
			},
		},
	}

	variableBlockSchema = &hcl.BodySchema{
		Attributes: []hcl.AttributeSchema{
			{
				Name: "name",
			},
			{
				Name: "default",
			},
			{
				Name: "type",
			},
			{
				Name: "sensitive",
			},
		},
	}
)

type Config struct {
	Variables map[string]Variable
}

type Variable struct {
	Name      string
	Default   string
	Type      string
	Sensitive bool
}

func configFromFile(filePath string) (*Config, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	file, diags := hclsyntax.ParseConfig(content, filePath, hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		return nil, fmt.Errorf("ParseConfig: %v", diags)
	}

	bodyCont, diags := file.Body.Content(configFileSchema)
	if diags.HasErrors() {
		return nil, fmt.Errorf("file content: %v", diags)
	}

	res := &Config{
		Variables: make(map[string]Variable),
	}

	for _, block := range bodyCont.Blocks {
		v := Variable{
			Name: block.Labels[0],
		}

		blockCont, diags := block.Body.Content(variableBlockSchema)
		if diags.HasErrors() {
			return nil, fmt.Errorf("block content: %v", diags)
		}

		if attr, exists := blockCont.Attributes["default"]; exists {
			diags := gohcl.DecodeExpression(attr.Expr, nil, &v.Default)
			if diags.HasErrors() {
				return nil, fmt.Errorf("default attr: %v", diags)
			}
		}

		if attr, exists := blockCont.Attributes["sensitive"]; exists {
			diags := gohcl.DecodeExpression(attr.Expr, nil, &v.Sensitive)
			if diags.HasErrors() {
				return nil, fmt.Errorf("sensitive attr: %v", diags)
			}
		}

		if attr, exists := blockCont.Attributes["type"]; exists {
			v.Type = hcl.ExprAsKeyword(attr.Expr)
			if v.Type == "" {
				return nil, fmt.Errorf("type attr: invalid value")
			}
		}

		res.Variables[v.Name] = v
	}
	return res, nil
}

func handleError(err error) {
	if err != nil {
		log.Fatal(err.Error())
	}
}

func listBlobs(client *azblob.Client, containerName string) {

	fmt.Println("Listing the blobs in the container:")

	pager := client.NewListBlobsFlatPager(containerName, &azblob.ListBlobsFlatOptions{
		Include: azblob.ListBlobsInclude{Snapshots: true, Versions: true},
	})

	for pager.More() {
		resp, err := pager.NextPage(context.TODO())
		handleError(err)

		for _, blob := range resp.Segment.BlobItems {
			fmt.Println(*blob.Name)
		}
	}
}

func main() {
	ctx := context.Background()
	config, _ := configFromFile("tf/variables.tf")

	const (
		sampleFile = "scripts/test_data/iris.csv"
		blobName   = "iris.csv"
		folderName = "bronze"
	)
	url := fmt.Sprintf("https://%s.blob.core.windows.net", config.Variables["storage_account_name"].Default)
	containerName := config.Variables["container_name"].Default
	fileNameFull := fmt.Sprintf("%s/%s", folderName, blobName)

	credential, err := azidentity.NewDefaultAzureCredential(nil)
	handleError(err)

	client, err := azblob.NewClient(url, credential, nil)
	handleError(err)

	file, err := os.OpenFile(sampleFile, os.O_RDONLY, 0)
	handleError(err)

	fmt.Printf("Uploading a blob named %s\n", blobName)
	client.UploadFile(ctx, containerName, fileNameFull, file, &azblob.UploadFileOptions{})

	listBlobs(client, containerName)
}
