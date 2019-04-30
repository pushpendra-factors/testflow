import { get } from "./request.js";
import { getHostURL } from "../util";

var host = getHostURL();

export function fetchDashboards(projectId) {
  return function(dispatch){
    return get(dispatch, host + "projects/" + projectId + "/dashboards")
      .then((r) => {
        dispatch({type: "FETCH_PROJECT_DASHBOARDS_FULFILLED", payload: r.data });
      })
      .catch((r) => {
        if (r.status) {
          dispatch({type: "FETCH_PROJECT_DASHBOARDS_REJECTED", payload: r.data, code: r.status });        
        } else {
          // network error. Idea: Use a global error component for this.
          console.log("network error");
        }
      });
  }
}