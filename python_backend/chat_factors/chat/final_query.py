from .kpi import get_transformed_kpi_query, ValueNotFoundError
from tornado.log import logging as log
from tornado.web import HTTPError
import Levenshtein


def get_url_and_query_payload_from_gpt_response(gpt_response, pid, kpi_config):
    log.info("running get_url_and_query_payload_from_gpt_response")
    query_class = gpt_response["qt"]
    query_payload = transform_query(pid, gpt_response, query_class, kpi_config)
    query_url = get_url_from_response(query_class, pid)
    result = {
        "payload": query_payload,
        "url": query_url
    }
    return result


def transform_query(pid, gpt_response, query_class, kpi_config):
    query = None
    if query_class == "kpi":
        query = get_transformed_kpi_query(pid, gpt_response, kpi_config)
    else:
        log.info("query_class did not match")
    return query


def get_url_from_response(query_class, pid):
    url = None
    if query_class == "kpi":
        placeholder_url = "projects/project_id/v1/kpi/query"
        url = placeholder_url.replace("project_id", str(pid))
    return url


def validate_gpt_response(gpt_response):
    valid_query_types = ["kpi"]
    if gpt_response["qt"] not in valid_query_types:
        log.error("incorrect query_type in gpt response")
        raise UnexpectedGptResponseError("unexpected query_type :%s in gpt_response", gpt_response["qt"])


class UnexpectedGptResponseError(Exception):
    pass


