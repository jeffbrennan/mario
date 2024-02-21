from azure.mgmt.datafactory import DataFactoryManagementClient
from azure.identity import DefaultAzureCredential
from pipeline_common import get_env

env = get_env()

adf_client = DataFactoryManagementClient(
    credential=DefaultAzureCredential(), subscription_id=env["AZ_SUBSCRIPTION_ID"]
)

existing_pipeline_gen = adf_client.pipelines.list_by_factory(
    resource_group_name=env["AZ_RESOURCE_GROUP"],
    factory_name=env["AZ_DATAFACTORY_NAME"],
)

existing_pipelines = [i.name for i in existing_pipeline_gen if i.name is not None]
print(f"Starting runs for n={len(existing_pipelines)} pipelines...")
for pipeline in existing_pipelines:
    print(f"Starting pipeline: {pipeline}...")
    adf_client.pipelines.create_run(
        resource_group_name=env["AZ_RESOURCE_GROUP"],
        factory_name=env["AZ_DATAFACTORY_NAME"],
        pipeline_name=pipeline,
    )
