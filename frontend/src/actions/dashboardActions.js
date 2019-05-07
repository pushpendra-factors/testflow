import { get, post, del } from "./request.js";
import { getHostURL } from "../util";

var host = getHostURL();

export function fetchDashboards(projectId) {
  return function(dispatch){
    return new Promise((resolve, reject) => {
      get(dispatch, host + "projects/" + projectId + "/dashboards")
        .then((r) => {
          if (r.ok) dispatch({type: "FETCH_DASHBOARDS_FULFILLED", payload: r.data });
          else dispatch({type: "FETCH_DASHBOARDS_REJECTED", payload: r.data.error });
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

export function createDashboard(projectId, payload) {
  return function(dispatch) { 
    return new Promise((resolve, reject) => {
      post(dispatch, host + "projects/" + projectId + "/dashboards", payload)
        .then((r) => {
          if (r.ok) {
            dispatch({
              type: "CREATE_DASHBOARD_FULFILLED",
              payload: r.data
            });
          } else {
            dispatch({
              type: "CREATE_DASHBOARD_REJECTED",
              error: r.data.error
            });
          }
          resolve(r);
        })
        .catch((r) => {
          dispatch({
            type: "CREATE_DASHBOARD_REJECTED",
            error: r
          })
          reject(r);
        });
    })
  }
}

export function fetchDashboardUnits(projectId, dashboardId) {
    return function(dispatch){
      return new Promise((resolve, reject) => {
        return get(dispatch, host + "projects/" + projectId + "/dashboards/" + dashboardId + "/units")
          .then((r) => {
            if (r.ok) dispatch({type: "FETCH_DASHBOARD_UNITS_FULFILLED", payload: r.data });
            else dispatch({type: "FETCH_DASHBOARD_UNITS_REJECTED", payload: r.data.error });
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
            type: "CREATE_DASHBOARD_UNIT_REJECTED",
            error: r
          })
          reject(r);
        });
    })
  }
}

export function deleteDashboardUnit(projectId, dashboardId, unitId) {
  return function(dispatch){
    return new Promise((resolve, reject) => {
      return del(dispatch, host + "projects/" + projectId + "/dashboards/" + dashboardId + "/units/" + unitId)
        .then((r) => {
          if (r.ok) {
            let data = { project_id: projectId, dashboard_id: dashboardId, id: unitId };
            dispatch({ type: "DELETE_DASHBOARD_UNIT_FULFILLED", payload: data}); 
            resolve(data);
          } else {
            dispatch({ type: "DELETE_DASHBOARD_UNIT_REJECTED", payload: r.data.error, code: r.status });
            resolve(r.data.error); 
          }
        })
        .catch((r) => {
          if (r.status) {
            dispatch({ type: "DELETE_DASHBOARD_UNIT_REJECTED", payload: r.data, code: r.status });
            reject(r);      
          } else {
            console.error("Network error");
          }
        });
      });
  }
} 


