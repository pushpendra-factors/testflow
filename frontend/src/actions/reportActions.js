import { get } from "./request.js";
import { getHostURL } from "../util";

var host = getHostURL();

export function fetchProjectReportsList(projectId) {
  return function(dispatch){
    return new Promise((resolve, reject) => {
      get(dispatch, host + "projects/" + projectId + "/reports")
        .then((r) => {
          if (r.ok) dispatch({type: "FETCH_REPORTS_FULFILLED", payload: r.data });
          else dispatch({type: "FETCH_REPORTS_REJECTED", payload: r.data.error });
          resolve(r.data);
        })
        .catch((r) => {
          if (r.status) {
            dispatch({type: "FETCH_REPORTS_REJECTED", payload: r.data, code: r.status });
            reject(r);      
          } else {
            console.error("Network error");
          }
        });
      });
  }
}

export function fetchReport(projectId, reportId){
  return function(dispatch){
    return new Promise((resolve, reject) => {
      get(dispatch, host + "projects/" + projectId + "/reports/" + reportId)
      .then((r) => {
        if (r.ok) dispatch({type: "FETCH_REPORT_FULFILLED", payload: r.data });
        else dispatch({type: "FETCH_REPORT_REJECTED", payload: r.data.error });
        resolve(r.data);
      })
      .catch((r) => {
        if (r.status) {
          dispatch({type: "FETCH_REPORT_REJECTED", payload: r.data, code: r.status });
          reject(r);      
        } else {
          console.error("Network error");
        }
      });
    })
  }
}