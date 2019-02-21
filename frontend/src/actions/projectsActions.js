import {get, post, del, put} from "./request.js";
import { getHostURL } from "../util";

var host = getHostURL();

export function changeProject(projectId) {
  return function(dispatch) {
    dispatch({type: "CHANGE_PROJECT", payload: projectId});
  }
}

export function createProject(projectName){
  return function(dispatch){
    return new Promise((resolve, reject)=>{
      post(dispatch, host + "projects", {name:projectName})
      .then((response)=>{
        resolve(dispatch({type: "CREATE_PROJECT_FULFILLED", payload: response.data}))
      }).catch((err)=>{
        reject(dispatch({type:"CREATE_PROJECT_REJECTED", payload: err}))
      })
    })
  }
}

export function fetchProjects() {
  return function(dispatch) {
    return new Promise((resolve,reject) => {
      get(dispatch,host + "projects")
      .then((response)=>{        
        resolve(dispatch({type:"FETCH_PROJECTS_FULFILLED", payload: response.data}))
      }).catch((err)=>{        
        reject(dispatch({type:"FETCH_PROJECTS_REJECTED", payload: err}))
      });
    });
  }
}

export function fetchProjectEvents(projectId) {
  return function(dispatch) {
    return new Promise((resolve, reject) => {
      get(dispatch, host + "projects/" + projectId + "/event_names")
        .then((response) => {
          resolve(dispatch({type: "FETCH_PROJECT_EVENTS_FULFILLED",
                  payload: { currentProjectId: projectId, eventNames: response.data,
                    eventPropertiesMap: {} }}));
        })
        .catch((err) => {
          reject(dispatch({type: "FETCH_PROJECT_EVENTS_REJECTED",
                  payload: { currentProjectId: projectId, eventNames: [],
                    eventPropertiesMap: {}, err: err }}));
                    
        });
    });
  }
}

export function fetchProjectSettings(projectId) {
  return function(dispatch) {
    return new Promise((resolve, reject) => {
      get(dispatch, host + "projects/" + projectId + "/settings")
        .then((response) => {
          resolve(dispatch({
            type: "FETCH_PROJECT_SETTINGS_FULFILLED", 
            payload: {
              currentProjectId: projectId,
              settings: response.data
            }
          }));
        })
        .catch((err) => {
          reject(
            dispatch({
            type: "FETCH_PROJECT_SETTINGS_REJECTED", 
            payload: {
              currentProjectId: projectId, 
              settings: {}, 
              err: err
            }
          }));
        });
      });
  }
}

export function udpateProjectSettings(projectId, payload) {
  return function(dispatch) {
    return put(dispatch, host + "projects/" + projectId + "/settings", payload)
     .then((response) => {
        return dispatch({
          type: "UPDATE_PROJECT_SETTINGS_FULFILLED", 
          payload: {
            updatedSettings: response.data
          }
        });
      })
      .catch((err) => {
        return dispatch({
          type: "UPDATE_PROJECT_SETTINGS_REJECTED", 
          payload: {
            updatedSettings: {}, 
            err: err
          }
        });
      });
  }
}

export function fetchProjectEventProperties(projectId, eventName) {
  return function(dispatch) {
    get(dispatch, host + "projects/" + projectId +
              "/event_names/" + eventName + "/properties")
      .then((response) => {
        dispatch({type: "FETCH_PROJECT_EVENT_PROPERTIES_FULFILLED",
                 payload: { eventName: eventName, eventProperties: response.data }})
      })
      .catch((err) => {
        dispatch({type: "FETCH_PROJECT_EVENT_PROPERTIES_REJECTED",
                 payload: { eventName: eventName, eventProperties: {}, err: err }})
      })
  }
}

export function fetchProjectEventPropertyValues(projectId, eventName, propertyName) {
  return function(dispatch) {
    get(dispatch, host + "projects/" + projectId +
              "/event_names/" + eventName + "/properties/" + propertyName +
              "/values")
      .then((response) => {
        dispatch({type: "FETCH_PROJECT_EVENT_PROPERTY_VALUES_FULFILLED",
                 payload: { eventName: eventName, propertyName: propertyName,
                  eventPropertyValues: response.data }})
      })
      .catch((err) => {
        dispatch({type: "FETCH_PROJECT_EVENT_PROPERTY_VALUES_REJECTED",
                 payload: { eventName: eventName, propertyName: propertyName,
                  eventPropertyValues: [], err: err }})
      })
  }
}

export function fetchProjectUserProperties(projectId) {
  return function(dispatch) {
    get(dispatch, host + "projects/" + projectId +
              "/user_properties")
      .then((response) => {
        dispatch({type: "FETCH_PROJECT_USER_PROPERTIES_FULFILLED",
                 payload: { userProperties: response.data }})
      })
      .catch((err) => {
        dispatch({type: "FETCH_PROJECT_USER_PROPERTIES_REJECTED",
                 payload: { userProperties: {}, err: err }})
      })
  }
}

export function fetchProjectUserPropertyValues(projectId, propertyName) {
  return function(dispatch) {
    get(dispatch, host + "projects/" + projectId +
              "/user_properties/" + propertyName + "/values")
      .then((response) => {
        dispatch({type: "FETCH_PROJECT_USER_PROPERTY_VALUES_FULFILLED",
                 payload: { propertyName: propertyName,
                  userPropertyValues: response.data }})
      })
      .catch((err) => {
        dispatch({type: "FETCH_PROJECT_USER_PROPERTY_VALUES_REJECTED",
                 payload: { propertyName: propertyName,
                  userPropertyValues: [], err: err }})
      })
  }
}

export function fetchFilters(projectId) {
  return function(dispatch) {
    // New promise created to handle use catch on
    // fetch call from component.
    return new Promise((resolve, reject) => {
      get(dispatch, host + "projects/" + projectId +"/filters")
        .then((response) => {
          dispatch({
            type: "FETCH_FILTERS_FULFILLED",
            payload: response.data
          });
          resolve(response.data);
        })
        .catch((err) => {
          dispatch({
            type: "FETCH_FILTERS_REJECTED",
            error: err
          })
          reject(err);
        });
    })
  }
}

export function createFilter(projectId, payload) {
  return function(dispatch) {
    return new Promise((resolve, reject) => {
      post(dispatch, host + "projects/" + projectId +"/filters", payload)
        .then((r) => {
          dispatch({
            type: "CREATE_FILTER_FULFILLED",
            payload: r.data
          });
          resolve(r.data);
        })
        .catch((r) => {
          dispatch({
            type: "CREATE_FILTER_REJECTED",
            error: r
          })
          reject({body: r.data, status: r.status});
        });
    })
  }
}

export function updateFilter(projectId, filterId, payload, storeIndex) {
  return function(dispatch) {
    return new Promise((resolve, reject) => {
      put(dispatch, host + "projects/" + projectId +"/filters/"+filterId, payload)
        .then((r) => {
          dispatch({
            type: "UPDATE_FILTER_FULFILLED",
            payload: {data: r.data, storeIndex: storeIndex}
          });
          resolve(r.data);
        })
        .catch((r) => {
          dispatch({
            type: "UPDATE_FILTER_REJECTED",
            error: r
          })
          reject({body: r.data, status: r.status});
        });
    })
  }
}

export function deleteFilter(projectId, filterId, storeIndex) {
  return function(dispatch) {
    return new Promise((resolve, reject) => {
      del(dispatch, host + "projects/" + projectId +"/filters/"+filterId)
        .then((r) => {
          dispatch({
            type: "DELETE_FILTER_FULFILLED",
            payload: storeIndex
          });
          resolve(r.data);
        })
        .catch((r) => {
          dispatch({
            type: "DELETE_FILTER_REJECTED",
            error: r
          });
          reject({body: r.data, status: r.status});
        });
    })
  }
}

export function fetchProjectModels(projectId){
  return function(dispatch){
    return get(dispatch, host + "projects/" + projectId + "/models")
      .then((r) => {
        dispatch({type: "FETCH_PROJECT_MODELS_FULFILLED", payload: r.data });
      })
      .catch((r) => {
        if (r.status) {
          // use this pattern for error handling. 
          // decided to use redux store.
          dispatch({type: "FETCH_PROJECT_MODELS_REJECTED", payload: r.data, code: r.status });        
        } else {
          // network error. Idea: Use a global error component for this.
          console.log("network error");
        }
      });
  }
}