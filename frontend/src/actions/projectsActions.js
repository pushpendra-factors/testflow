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
    return new Promise((resolve, reject) => {
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
      get(dispatch, host + "projects")
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

export function fetchProjectEventProperties(projectId, eventName, modelId="", useStore=true) {
  let url = host + "projects/" + projectId + "/event_names/" + btoa(eventName) + "/properties";
  if (!!modelId) {
    url += "?model_id=" + modelId;
  }

  if (useStore){
    return function(dispatch) {
      // Using base64 encoded event name.
      return get(dispatch, url)
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
  
  return get(null, url);
}

export function fetchProjectEventPropertyValues(projectId, eventName, propertyName, useStore=true) {
  let url = host + "projects/" + projectId + "/event_names/" + btoa(eventName) + "/properties/" + propertyName + "/values";

  if (useStore) {
    return function(dispatch) {
      // Using base64 encoded event name.
      get(dispatch, url)
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

  return get(null, url);
}

export function fetchProjectUserProperties(projectId, modelId="", useStore=true) {
  let url = host + "projects/" + projectId + "/user_properties";
  if (!!modelId) {
    url += "?model_id=" + modelId;
  }

  if (useStore) {
    return function(dispatch) {
      get(dispatch, url)
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
  
  return get(null, url);
}

export function fetchProjectUserPropertyValues(projectId, propertyName, useStore=true) {
  let url = host + "projects/" + projectId + "/user_properties/" + propertyName + "/values";

  if (useStore) {
    return function(dispatch) {
      get(dispatch, url)
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

  return get(null, url);
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
          if (r.ok) {
            dispatch({
              type: "CREATE_FILTER_FULFILLED",
              payload: r.data
            });
          } else {
            dispatch({
              type: "CREATE_FILTER_REJECTED",
              error: r.data
            })
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

export function runQuery(projectId, query) {
  let url = host + "projects/" + projectId + "/query";
  return post(null, url , {query: query});
}

export function fetchProjectAgents(projectId){
  return function(dispatch){
    return get(dispatch, host + "projects/" + projectId + "/agents")
      .then((r) => {
        dispatch({type: "FETCH_PROJECT_AGENTS_FULFILLED", payload: r.data });
      })
      .catch((r) => {
        if (r.status) {
          // use this pattern for error handling. 
          // decided to use redux store.
          dispatch({type: "FETCH_PROJECT_AGENTS_REJECTED", payload: r.data, code: r.status });        
        } else {
          // network error. Idea: Use a global error component for this.
          console.log("network error");
        }
      });
  }
}

export function projectAgentInvite(projectId, emailId){
  return function(dispatch){
    let payload = {"email":emailId};
    return new Promise((resolve, reject) => {
      post(dispatch, host + "projects/" + projectId + "/agents/invite", payload)
      .then((r) => {
        if (r.ok && r.status && r.status == 201){
          dispatch({type: "PROJECT_AGENT_INVITE_FULFILLED", payload: r.data });
          resolve(r.data);
        }else if (r.status && r.status == 409){
          dispatch({type: "PROJECT_AGENT_INVITE_REJECTED", payload: r.data.error }); 
          reject("User Seats limit reached");
        }
        else {
          dispatch({type: "PROJECT_AGENT_INVITE_REJECTED", payload: r.data.error });
          reject(r.data.error);
        }
      })
      .catch((r) => {
        dispatch({type: "PROJECT_AGENT_INVITE_REJECTED", payload: r.data.error });
      });
    });
  }
}

export function projectAgentRemove(projectId, agentUUID){
  return function(dispatch){
    return new Promise((resolve, reject) => {
      put(dispatch, host + "projects/" + projectId +"/agents/remove", {"agent_uuid":agentUUID})
        .then((r) => {
          dispatch({
            type: "PROJECT_AGENT_REMOVE_FULFILLED",
            payload: r.data
          });
          resolve(r.data);
        })
        .catch((r) => {
          dispatch({
            type: "PROJECT_AGENT_REMOVE_REJECTED",
            error: r
          });
          reject({body: r.data, status: r.status});
        });
    })
  }
}