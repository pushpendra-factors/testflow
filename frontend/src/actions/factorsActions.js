import axios from "axios";
import appConfig from "../config/appConfig"

export function fetchFactors(currentProjectId, query, queryParams) {
  return function(dispatch) {
    dispatch({type: "FETCH_FACTORS"});

    axios.post(appConfig.API_PATH + "projects/" + currentProjectId + "/factor" + queryParams, query)
      .then((response) => {
        dispatch({type: "FETCH_FACTORS_FULFILLED", payload: response.data})
      })
      .catch((err) => {
        dispatch({type: "FETCH_FACTORS_REJECTED", payload: err})
      })
  }
}
