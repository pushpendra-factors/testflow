from .kpi import get_transformed_kpi_query
from tornado.log import logging as log


def get_url_and_query_payload_from_gpt_response(gpt_response, pid, kpi_config):
    log.info("running get_url_and_query_payload_from_gpt_response")
    query_class = gpt_response["qt"]
    query_payload = transform_query(gpt_response, query_class, kpi_config)
    query_url = get_url_from_response(query_class, pid)
    result = {
        "payload":query_payload,
        "url": query_url
    }
    return result


def transform_query(gpt_response, query_class, kpi_config):
    query = None
    if query_class == "kpi":
        query = get_transformed_kpi_query(gpt_response, kpi_config)
    else:
        log.info("query_class did not match")
    return query


def get_url_from_response(query_class, pid):
    url = None
    if query_class == "kpi":
        placeholder_url = "projects/project_id/v1/kpi/query"
        url = placeholder_url.replace("project_id", str(pid))
    return url


