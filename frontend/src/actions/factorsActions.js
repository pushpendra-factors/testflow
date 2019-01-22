import axios from "axios";
import { getHostURL } from "../util";

var host = getHostURL();

export function fetchFactors(currentProjectId, query, queryParams) {
  return function(dispatch) {
    dispatch({type: "FETCH_FACTORS"});

    axios.post(host + "projects/" + currentProjectId + "/factor" + queryParams, query)
      .then((response) => {
        dispatch({type: "FETCH_FACTORS_FULFILLED", payload: response.data})
      })
      .catch((err) => {
        dispatch({type: "FETCH_FACTORS_REJECTED", payload: err})
      })
  }
}
