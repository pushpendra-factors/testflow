import axios from "axios";
import appConfig from "../config/appConfig"

export function fetchProjects() {
  return function(dispatch) {
    dispatch({type: "FETCH_PROJECTS"});

    axios.get(appConfig.API_PATH + "projects")
      .then((response) => {
        dispatch({type: "FETCH_PROJECTS_FULFILLED", payload: response.data})
      })
      .catch((err) => {
        dispatch({type: "FETCH_PROJECTS_REJECTED", payload: err})
      })
  }
}

export function fetchCurrentProjectEvents(currentProject) {
  return function(dispatch) {
    axios.get(appConfig.API_PATH + "projects/" + currentProject.value + "/event_names")
      .then((response) => {
        dispatch({type: "FETCH_CURRENT_PROJECT_EVENTS_FULFILLED",
                 payload: { currentProject: currentProject, currentProjectEventNames: response.data,
                   eventPropertiesMap: {} }})
      })
      .catch((err) => {
        dispatch({type: "FETCH_CURRENT_PROJECT_EVENTS_REJECTED",
                 payload: { currentProject: currentProject, currentProjectEventNames: [],
                   eventPropertiesMap: {}, err: err }})
      })
  }
}

export function fetchProjectEventProperties(currentProjectId, eventName) {
  return function(dispatch) {
    axios.get(appConfig.API_PATH + "projects/" + currentProjectId +
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

export function fetchProjectEventPropertyValues(currentProjectId, eventName, propertyName) {
  return function(dispatch) {
    axios.get(appConfig.API_PATH + "projects/" + currentProjectId +
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
