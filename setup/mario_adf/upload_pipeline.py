import os
from pathlib import Path
from azure.mgmt.datafactory import DataFactoryManagementClient
from azure.mgmt.datafactory.models import PipelineResource
from azure.identity import DefaultAzureCredential
from dotenv import load_dotenv
import json

load_dotenv(Path(__file__).parent / ".env")
AZ_SUBSCRIPTION_ID = os.environ.get("AZ_SUBSCRIPTION_ID")
AZ_RESOURCE_GROUP = os.environ.get("AZ_RESOURCE_GROUP")
AZ_DATAFACTORY_NAME = os.environ.get("AZ_DATAFACTORY_NAME")

if AZ_SUBSCRIPTION_ID is None:
    raise ValueError("AZ_SUBSCRIPTION_ID is not set")

if AZ_RESOURCE_GROUP is None:
    raise ValueError("AZ_RESOURCE_GROUP is not set")

if AZ_DATAFACTORY_NAME is None:
    raise ValueError("AZ_DATAFACTORY_NAME is not set")


adf_client = DataFactoryManagementClient(
    credential=DefaultAzureCredential(), subscription_id=AZ_SUBSCRIPTION_ID
)

pipeline_path = Path(__file__).parent / "pipelines"

all_pipelines = [i for i in pipeline_path.glob("*.json")]

for pipeline in all_pipelines:
    with open(pipeline, "r") as f:
        pipeline_dict = json.loads(f.read())

    pipeline_resource = PipelineResource.deserialize(pipeline_dict)
    if pipeline_resource is None:
        raise ValueError(f"Pipeline {pipeline} is empty")

    if pipeline_resource.name is None:
        raise ValueError(f"Pipeline {pipeline} has no name")

    print(f"Uploading pipeline: {pipeline_resource.name}...")
    adf_client.pipelines.create_or_update(
        resource_group_name=AZ_RESOURCE_GROUP,
        factory_name=AZ_DATAFACTORY_NAME,
        pipeline_name=pipeline_resource.name,
        pipeline=pipeline_resource,
    )
