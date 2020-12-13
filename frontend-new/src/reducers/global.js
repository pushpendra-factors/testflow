/* eslint-disable */

import {
  SET_PROJECTS,
  SET_ACTIVE_PROJECT,
  CREATE_PROJECT_FULFILLED,
  FETCH_PROJECTS_REJECTED,
  SHOW_ANALYTICS_RESULT,
} from "./types";
import { get, post, put } from "../utils/request";

var host = BUILD_CONFIG.backend_host;
host = host[host.length - 1] === "/" ? host : host + "/";

const defaultState = {
  is_funnel_results_visible: false,
  funnel_events: [],
  projects: [],
  active_project: {},
  fetchingProjects: null,
  projectsError: null,
  show_analytics_result: false,
  currentProjectSettings: {},
};

export default function (state = defaultState, action) {
  switch (action.type) {
    case SET_PROJECTS:
      return { ...state, projects: action.payload };
    case SET_ACTIVE_PROJECT:
      return { ...state, active_project: action.payload };
    case CREATE_PROJECT_FULFILLED: {
      let _state = { ...state };
      _state.projects = [..._state.projects, action.payload];
      // Set currentProjectId to this newly created project
      _state.active_project = action.payload;
      return _state;
    }
    case FETCH_PROJECTS_REJECTED: {
      return {
        ...state,
        fetchingProjects: false,
        projectsError: action.payload,
      };
    }
    case "UPDATE_PROJECT_SETTINGS_FULFILLED": {
      let _state = { ...state };
      if (_state.currentProjectSettings)
        _state.currentProjectSettings = {
          ..._state.currentProjectSettings,
          ...action.payload.updatedSettings,
        };
      return _state;
    }
    case "UPDATE_PROJECT_SETTINGS_REJECTED": {
      return {
        ...state,
        projectEventsError: action.payload.err,
      };
    }
    case "UPDATE_PROJECT_DETAILS_FULFILLED": {
      let _state = { ...state };
      if (_state.active_project)
        _state.active_project = {
          ..._state.active_project,
          ...action.payload.updatedDetails,
        };
      return _state;
    }
    case "UPDATE_PROJECT_DETAILS_REJECTED": {
      return {
        ...state,
        projectEventsError: action.payload.err,
      };
    }
    case "FETCH_PROJECT_SETTINGS_FULFILLED": {
      return {
        ...state,
        currentProjectSettings: action.payload.settings,
      };
    }
    case "FETCH_PROJECT_SETTINGS_REJECTED": {
      return {
        ...state,
        projectSettingsError: action.payload.err,
      };
    }
    case SHOW_ANALYTICS_RESULT: {
      return {
        ...state,
        show_analytics_result: action.payload,
      };
    }
    default:
      return state;
  }
}

// Action creators
export function fetchProjectAction(projects, status = "success") {
  return { type: SET_PROJECTS, payload: projects };
}

export const setActiveProject = (project) => {
  return { type: SET_ACTIVE_PROJECT, payload: project };
};

// Service Call

// export function fetchProjects(projects) {
//   return function (dispatch) {
//     return new Promise((resolve, reject) => {
//       get(dispatch, host + 'projects', {})
//         .then((response) => {
//           // dispatch(setActiveProject(response.data.projects[0]));
//           resolve(dispatch(fetchProjectAction(response.data.projects)));
//         }).catch((err) => {
//           resolve(dispatch(fetchProjectAction([])));
//         });
//     });
//   };
// }

export function createProject(projectName) {
  return function (dispatch) {
    return new Promise((resolve, reject) => {
      post(dispatch, host + "projects", { name: projectName })
        .then((r) => {
          if (r.ok) {
            dispatch({ type: "CREATE_PROJECT_FULFILLED", payload: r.data });
            resolve(r);
          } else {
            dispatch({
              type: "CREATE_PROJECT_REJECTED",
              payload: "Failed to create project.",
            });
            reject(r);
          }
        })
        .catch((err) => {
          dispatch({ type: "CREATE_PROJECT_REJECTED", payload: err });
          reject(err);
        });
    });
  };
}

export function fetchProjectSettings(projectId) {
  return function (dispatch) {
    return new Promise((resolve, reject) => {
      get(dispatch, host + "projects/" + projectId + "/settings")
        .then((r) => {
          if (r.ok) {
            dispatch({
              type: "FETCH_PROJECT_SETTINGS_FULFILLED",
              payload: {
                currentProjectId: projectId,
                settings: r.data,
              },
            });

            resolve(r);
          } else {
            dispatch({
              type: "FETCH_PROJECT_SETTINGS_REJECTED",
              payload: {
                currentProjectId: projectId,
                settings: {},
                err: "Failed to get project settings.",
              },
            });

            reject(r);
          }
        })
        .catch((err) => {
          dispatch({
            type: "FETCH_PROJECT_SETTINGS_REJECTED",
            payload: {
              currentProjectId: projectId,
              settings: {},
              err: err,
            },
          });

          reject(err);
        });
    });
  };
}

export function udpateProjectSettings(projectId, payload) {
  return function (dispatch) {
    return new Promise((resolve, reject) => {
      put(dispatch, host + "projects/" + projectId + "/settings", payload)
        .then((response) => {
          dispatch({
            type: "UPDATE_PROJECT_SETTINGS_FULFILLED",
            payload: {
              updatedSettings: response.data,
            },
          });
          resolve(response);
        })
        .catch((err) => {
          dispatch({
            type: "UPDATE_PROJECT_SETTINGS_REJECTED",
            payload: {
              updatedSettings: {},
              err: err,
            },
          });
          reject(err);
        });
    });
  };
}

export function udpateProjectDetails(projectId, payload) {
  return function (dispatch) {
    return new Promise((resolve, reject) => {
      put(dispatch, host + "projects/" + projectId, payload)
        .then((response) => {
          dispatch({
            type: "UPDATE_PROJECT_DETAILS_FULFILLED",
            payload: {
              updatedDetails: response.data,
            },
          });
          resolve(response);
        })
        .catch((err) => {
          dispatch({
            type: "UPDATE_PROJECT_DETAILS_REJECTED",
            payload: {
              updatedDetails: {},
              err: err,
            },
          });
          reject(err);
        });
    });
  };
}
