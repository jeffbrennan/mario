import os
from pathlib import Path
from dotenv import load_dotenv

from typing import TypedDict


class AZEnv(TypedDict):
    AZ_SUBSCRIPTION_ID: str
    AZ_RESOURCE_GROUP: str
    AZ_DATAFACTORY_NAME: str


def get_env() -> AZEnv:

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

    return {
        "AZ_SUBSCRIPTION_ID": AZ_SUBSCRIPTION_ID,
        "AZ_RESOURCE_GROUP": AZ_RESOURCE_GROUP,
        "AZ_DATAFACTORY_NAME": AZ_DATAFACTORY_NAME,
    }
