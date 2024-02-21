from pathlib import Path
from azure.mgmt.datafactory import DataFactoryManagementClient
from azure.mgmt.datafactory.models import PipelineResource
from azure.identity import DefaultAzureCredential
import json

from mario_adf.pipeline_common import get_env

env = get_env()

adf_client = DataFactoryManagementClient(
    credential=DefaultAzureCredential(), subscription_id=env["AZ_SUBSCRIPTION_ID"]
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
        resource_group_name=env["AZ_RESOURCE_GROUP"],
        factory_name=env["AZ_DATAFACTORY_NAME"],
        pipeline_name=pipeline_resource.name,
        pipeline=pipeline_resource,
    )
