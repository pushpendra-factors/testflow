import { get, post } from "./request.js";
import { getHostURL } from "../util";

var host = getHostURL();

export function fetchDashboards(projectId) {
  return function(dispatch){
    return new Promise((resolve, reject) => {
      get(dispatch, host + "projects/" + projectId + "/dashboards")
        .then((r) => {
          dispatch({type: "FETCH_DASHBOARDS_FULFILLED", payload: r.data });
          resolve(r.data);
        })
        .catch((r) => {
          if (r.status) {
            dispatch({type: "FETCH_DASHBOARDS_REJECTED", payload: r.data, code: r.status });
            reject(r);      
          } else {
            console.error("Network error");
          }
        });
      });
  }
}

export function fetchDashboardUnits(projectId, dashboardId) {
    return function(dispatch){
      return new Promise((resolve, reject) => {
        return get(dispatch, host + "projects/" + projectId + "/dashboards/" + dashboardId + "/units")
          .then((r) => {
            dispatch({type: "FETCH_DASHBOARD_UNITS_FULFILLED", payload: r.data });
            resolve(r.data);
          })
          .catch((r) => {
            if (r.status) {
              dispatch({type: "FETCH_DASHBOARD_UNITS_REJECTED", payload: r.data, code: r.status });
              reject(r);      
            } else {
              console.error("Network error");
            }
          });
        });
    }
} 

export function createDashboardUnit(projectId, dashboardId, payload) {
  return function(dispatch) { 
    return new Promise((resolve, reject) => {
      post(dispatch, host + "projects/" + projectId + "/dashboards/" + dashboardId + "/units", payload)
        .then((r) => {
          if (r.ok) {
            dispatch({
              type: "CREATE_DASHBOARD_UNIT_FULFILLED",
              payload: r.data
            });
          } else {
            dispatch({
              type: "CREATE_DASHBOARD_UNIT_REJECTED",
              error: r.data.error
            });
          }
          resolve(r);
        })
        .catch((r) => {
          dispatch({
            type: "CREATE_FILTER_REJECTED",
            error: r
          })
          reject(r);
        });
    })
  }
}

