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

export function updateCurrentProject(currentProject) {
  return function(dispatch) {
    axios.get(appConfig.API_PATH + "projects/" + currentProject.value + "/event_names")
      .then((response) => {
        dispatch({type: "FETCH_CURRENT_PROJECT_EVENTS_FULFILLED",
                 payload: { currentProject: currentProject, currentProjectEventNames: response.data }})
      })
      .catch((err) => {
        dispatch({type: "FETCH_CURRENT_PROJECT_EVENTS_FULFILLED",
                 payload: { currentProject: currentProject, currentProjectEventNames: [], err: err }})
      })
  }
}
