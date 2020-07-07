import { get, post, put, del } from "./request.js";
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

export function updateDashboard(projectId, dashboardId, payload) {
  return function(dispatch) {
    return new Promise((resolve, reject) => {
      // immediate local state update.
      let data = { project_id: projectId, id: dashboardId, ...payload };
      dispatch({ type: "UPDATE_DASHBOARD_FULFILLED", payload: data });

      // lazy remote update.
      put(dispatch, host + "projects/" + projectId + "/dashboards/" + dashboardId, payload)
        .then((r) => {
          if (!r.ok) {
            dispatch({
              type: "UPDATE_DASHBOARD_REJECTED",
              error: r.data.error
            });
          }
          resolve(r);
        })
        .catch((r) => {
          dispatch({
            type: "UPDATE_DASHBOARD_REJECTED",
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

export function updateDashboardUnit(projectId, dashboardId, unitId, payload) {
  return function(dispatch){
    return new Promise((resolve, reject) => {
      return put(dispatch, host + "projects/" + projectId + "/dashboards/" + dashboardId + "/units/" + unitId, payload)
        .then((r) => {
          if (r.ok) {
            let data = { project_id: projectId, dashboard_id: dashboardId, id: unitId, ...payload };
            dispatch({ type: "UPDATE_DASHBOARD_UNIT_FULFILLED", payload: data}); 
            resolve(data);
          } else {
            dispatch({ type: "UPDATE_DASHBOARD_UNIT_REJECTED", payload: {}, code: r.status });
            resolve({ error: "Failed to update unit." });
          }
        })
        .catch((r) => {
          if (r.status) {
            dispatch({ type: "UPDATE_DASHBOARD_UNIT_REJECTED", payload: r.data, code: r.status });
            reject(r);
          } else {
            console.error("Network error");
          }
        });
      });
  }
} 

export function fetchWebAnalyticsResult(projectId,dashboardId,query){
  let url = host+"projects/"+projectId+"/dashboard/"+dashboardId+"/units/query/web_analytics";
  return post(null, url, {...query})
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


