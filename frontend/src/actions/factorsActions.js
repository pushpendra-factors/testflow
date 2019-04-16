import { post } from "./request.js";
import { getHostURL } from "../util";

var host = getHostURL();

export function fetchFactors(currentProjectId, modelId, query, queryParams) {
  return function(dispatch) {
    dispatch({type: "FETCH_FACTORS"});
    var mid = "model_id=" + modelId ;
    
    var separator = "?"
    if(queryParams != ""){
      separator = "&"
    }

    return new Promise((resolve, reject) => {
      post(dispatch, host + "projects/" + currentProjectId + "/factor"+ queryParams + separator + mid, query)
        .then((response) => {
          dispatch({type: "FETCH_FACTORS_FULFILLED", payload: response.data});
          resolve(response);
        })
        .catch((err) => {
          dispatch({type: "FETCH_FACTORS_REJECTED", payload: err});
          reject(err);
        });
    });
  }
}

export function resetFactors() {
  return function(dispatch) {
    dispatch({type: "RESET_FACTORS"});
  }
}