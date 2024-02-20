from pathlib import Path
from azure.common.credentials import ServicePrincipalCredentials
from azure.mgmt.resource import ResourceManagementClient
from azure.mgmt.datafactory import DataFactoryManagementClient
from azure.mgmt.datafactory.models import *
import time

# TODO: build this out and add more pipelines
# adf_client = 

pipeline_path = Path(__file__).parent / "pipelines"

all_pipelines = pipeline_path.glob("*.json")

for pipeline in all_pipelines:
