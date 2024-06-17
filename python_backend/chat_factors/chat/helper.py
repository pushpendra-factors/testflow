import Levenshtein
from tornado.log import logging as log


class ValueNotFoundError(Exception):
    pass


def get_closest_match(query_string, complete_list, threshold):

    lowest_edit_distance = 100
    best_match = complete_list[0]
    for current_string in complete_list:
        current_distance = Levenshtein.distance(query_string, current_string)
        if current_distance < lowest_edit_distance:
            lowest_edit_distance = current_distance
            best_match = current_string
    if lowest_edit_distance / len(query_string) < threshold:
        return best_match
    else:
        log.error("matching not found in the list")
        return query_string
