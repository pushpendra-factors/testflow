import axios from "axios";
import appConfig from "../config/appConfig"

export function changeProject(projectId) {
  return function(dispatch) {
    dispatch({type: "CHANGE_PROJECT", payload: projectId});
  }
}

export function fetchProjects() {
  return function(dispatch) {
    dispatch({type: "FETCH_PROJECTS"});

    return new Promise((resolve, reject) => {
      axios.get(appConfig.API_PATH + "projects")
        .then((response) => {
          resolve(dispatch({type: "FETCH_PROJECTS_FULFILLED", payload: response.data}));
        })
        .catch((err) => {
          reject(dispatch({type: "FETCH_PROJECTS_REJECTED", payload: err}));
        });
    });
  }
}

export function fetchCurrentProjectEvents(projectId) {
  return function(dispatch) {
    return new Promise((resolve, reject) => {
      axios.get(appConfig.API_PATH + "projects/" + projectId + "/event_names")
        .then((response) => {
          resolve(dispatch({type: "FETCH_CURRENT_PROJECT_EVENTS_FULFILLED",
                  payload: { currentProjectId: projectId, currentProjectEventNames: response.data,
                    eventPropertiesMap: {} }}));
        })
        .catch((err) => {
          reject(dispatch({type: "FETCH_CURRENT_PROJECT_EVENTS_REJECTED",
                  payload: { currentProjectId: projectId, currentProjectEventNames: [],
                    eventPropertiesMap: {}, err: err }}));
        });
    });
  }
}

export function fetchCurrentProjectSettings(projectId) {
  return function(dispatch) {
    return new Promise((resolve, reject) => {
      axios.get(appConfig.API_PATH + "projects/" + projectId + "/settings")
        .then((response) => {
          resolve(dispatch({
            type: "FETCH_CURRENT_PROJECT_SETTINGS_FULFILLED", 
            payload: {
              currentProjectId: projectId,
              settings: response.data
            }
          }));
        })
        .catch((err) => {
          reject(
            dispatch({
            type: "FETCH_CURRENT_PROJECT_SETTINGS_REJECTED", 
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

export function udpateCurrentProjectSettings(projectId, payload) {
  return function(dispatch) {
    return axios.put(appConfig.API_PATH + "projects/" + projectId + "/settings", payload)
     .then((response) => {
        return dispatch({
          type: "UPDATE_CURRENT_PROJECT_SETTINGS_FULFILLED", 
          payload: {
            updatedSettings: response.data
          }
        });
      })
      .catch((err) => {
        return dispatch({
          type: "UPDATE_CURRENT_PROJECT_SETTINGS_REJECTED", 
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
    axios.get(appConfig.API_PATH + "projects/" + projectId +
              "/event_names/" + eventName + "/properties")
      .then((response) => {
        dispatch({type: "FETCH_CURRENT_PROJECT_EVENT_PROPERTIES_FULFILLED",
                 payload: { eventName: eventName, eventProperties: response.data }})
      })
      .catch((err) => {
        dispatch({type: "FETCH_CURRENT_PROJECT_EVENT_PROPERTIES_REJECTED",
                 payload: { eventName: eventName, eventProperties: {}, err: err }})
      })
  }
}

export function fetchProjectEventPropertyValues(projectId, eventName, propertyName) {
  return function(dispatch) {
    axios.get(appConfig.API_PATH + "projects/" + projectId +
              "/event_names/" + eventName + "/properties/" + propertyName +
              "/values")
      .then((response) => {
        dispatch({type: "FETCH_CURRENT_PROJECT_EVENT_PROPERTY_VALUES_FULFILLED",
                 payload: { eventName: eventName, propertyName: propertyName,
                  eventPropertyValues: response.data }})
      })
      .catch((err) => {
        dispatch({type: "FETCH_CURRENT_PROJECT_EVENT_PROPERTY_VALUES_REJECTED",
                 payload: { eventName: eventName, propertyName: propertyName,
                  eventPropertyValues: [], err: err }})
      })
  }
}

export function fetchProjectUserProperties(projectId) {
  return function(dispatch) {
    axios.get(appConfig.API_PATH + "projects/" + projectId +
              "/user_properties")
      .then((response) => {
        dispatch({type: "FETCH_CURRENT_PROJECT_USER_PROPERTIES_FULFILLED",
                 payload: { userProperties: response.data }})
      })
      .catch((err) => {
        dispatch({type: "FETCH_CURRENT_PROJECT_USER_PROPERTIES_REJECTED",
                 payload: { userProperties: {}, err: err }})
      })
  }
}

export function fetchProjectUserPropertyValues(projectId, propertyName) {
  return function(dispatch) {
    axios.get(appConfig.API_PATH + "projects/" + projectId +
              "/user_properties/" + propertyName + "/values")
      .then((response) => {
        dispatch({type: "FETCH_CURRENT_PROJECT_USER_PROPERTY_VALUES_FULFILLED",
                 payload: { propertyName: propertyName,
                  userPropertyValues: response.data }})
      })
      .catch((err) => {
        dispatch({type: "FETCH_CURRENT_PROJECT_USER_PROPERTY_VALUES_REJECTED",
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
      axios.get(appConfig.API_PATH + "projects/" + projectId +"/filters")
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
      axios.post(appConfig.API_PATH + "projects/" + projectId +"/filters", payload)
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
      axios.put(appConfig.API_PATH + "projects/" + projectId +"/filters/"+filterId, payload)
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
      axios.delete(appConfig.API_PATH + "projects/" + projectId +"/filters/"+filterId)
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