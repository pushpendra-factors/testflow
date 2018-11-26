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
