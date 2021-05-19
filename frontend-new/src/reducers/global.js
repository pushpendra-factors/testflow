/* eslint-disable */

import {
  SET_PROJECTS,
  SET_ACTIVE_PROJECT,
  CREATE_PROJECT_FULFILLED,
  FETCH_PROJECTS_REJECTED,
} from "./types";
import { get, getHostUrl, post, put } from "../utils/request";

var host = getHostUrl();
host = host[host.length - 1] === "/" ? host : host + "/";

const defaultState = {
  is_funnel_results_visible: false,
  funnel_events: [],
  projects: [],
  active_project: {},
  fetchingProjects: null,
  projectsError: null,
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
    case "ENABLE_FACEBOOK_USER_ID": {
      let fbUserID = action.payload.int_facebook_user_id;

      let _state = {...state};
      _state.currentProjectSettings = {
        ...state.currentProjectSettings,
        int_facebook_user_id: fbUserID,
      }
      return _state;
    }
    case "ENABLE_SALESFORCE_FULFILLED": {
      let enabledAgentUUID = action.payload.int_salesforce_enabled_agent_uuid;
      if (!enabledAgentUUID || enabledAgentUUID == "")
        return state;

      let _state = { ...state };
      _state.currentProjectSettings = {
        ...state.currentProjectSettings,
        int_salesforce_enabled_agent_uuid: enabledAgentUUID,
      }
      return _state;
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


export function addFacebookAccessToken(data) {
  return function(dispatch){
    return new Promise((resolve, reject) => {
      post(dispatch, host +"integrations/facebook/add_access_token", data)
        .then((r) => {
          if (r.ok) {
            dispatch({type:"ENABLE_FACEBOOK_USER_ID", payload: data})
            resolve(r);
          } else {
            reject(r); 
          }
        })
        .catch((err) => {
          reject(err);
        })
    })
  }
}

export function addLinkedinAccessToken(data) {
  return function(dispatch){
    return new Promise((resolve, reject) => {
      post(dispatch, host +"integrations/linkedin/add_access_token", data)
        .then((r) => {
          if (r.ok) {
            dispatch({type:"ENABLE_LINKEDIN_AD_ACCOUNT", payload: data})
            resolve(r);
          } else {
            reject(r); 
          }
        })
        .catch((err) => {
          reject(err);
        })
    })
  }
}

export function enableSalesforceIntegration(projectId) {
  return function(dispatch){
    return new Promise((resolve, reject) => {
      let payload = {"project_id":projectId}
      post(dispatch, host +"integrations/salesforce/enable", payload)
        .then((r) => {
          if (r.ok) {
            dispatch({ type: "ENABLE_SALESFORCE_FULFILLED", payload: r.data })
            resolve(r);
          } else {
            dispatch({ type:"ENABLE_SALESFORCE_REJECTED" });
            reject(r);
          }
        })
        .catch((err) => {
          dispatch({ type:"ENABLE_SALESFORCE_REJECTED", payload: err });
          reject(err);
        })
    })
  }
}


export function fetchSalesforceRedirectURL(projectId, agentUUID){
  return function(dispatch){
    return new Promise((resolve, reject) => {
      let payload = {"project_id":projectId}

      post(dispatch, host +"integrations/salesforce/auth", payload)
        .then((r) => {
          if (r.ok) {
            dispatch({ type: "FETCH_SALESFORCE_REDIRECT_URL_FULFILLED", payload: r.data })
            resolve(r);
          } else {
            dispatch({ type:"FETCH_SALESFORCE_REDIRECT_URL_REJECTED" });
            reject(r);
          }
        })
        .catch((err) => {
          dispatch({ type:"FETCH_SALESFORCE_REDIRECT_URL_REJECTED", payload: err });
          reject(err);
        })
    })
  }
}


export function enableAdwordsIntegration(projectId) {
  return function(dispatch){
    return new Promise((resolve, reject) => {
      let payload = { project_id: projectId.toString() }
      post(dispatch, host + "integrations/adwords/enable", payload)
        .then((r) => {
          if (r.ok) {
            dispatch({ type: "ENABLE_ADWORDS_FULFILLED", payload: r.data })
            resolve(r);
          } else {
            dispatch({ type:"ENABLE_ADWORDS_REJECTED" });
            reject(r); 
          }
        })
        .catch((err) => {
          dispatch({ type:"ENABLE_ADWORDS_REJECTED", payload: err });
          reject(err);
        })
    })
  }
}
export function enableSearchConsoleIntegration(projectId) {
  return function(dispatch){
    return new Promise((resolve, reject) => {
      let payload = { project_id: projectId.toString() }
      post(dispatch, host + "integrations/google_organic/enable", payload)
        .then((r) => {
          if (r.ok) {
            dispatch({ type: "ENABLE_ADWORDS_FULFILLED", payload: r.data })
            resolve(r);
          } else {
            dispatch({ type:"ENABLE_ADWORDS_REJECTED" });
            reject(r); 
          }
        })
        .catch((err) => {
          dispatch({ type:"ENABLE_ADWORDS_REJECTED", payload: err });
          reject(err);
        })
    })
  }
}

export function fetchAdwordsCustomerAccounts(payload) {
  return function(dispatch){
    return new Promise((resolve, reject) => {
      post(dispatch, host +"adwords/v1/get_customer_accounts", payload)
        .then((r) => {
          if (r.ok) {
            dispatch({ type: "FETCH_ADWORDS_CUSTOMER_ACCOUNTS_FULFILLED", payload: r.data })
            resolve(r.data);
          } else {
            dispatch({ type:"FETCH_ADWORDS_CUSTOMER_ACCOUNTS_REJECTED" });
            reject(r);
          }
        })
        .catch((err) => {
          dispatch({ type:"FETCH_ADWORDS_CUSTOMER_ACCOUNTS_REJECTED", payload: err });
          reject(err);
        })
    })
  }
}
export function fetchSearchConsoleCustomerAccounts(payload) {
  return function(dispatch){
    return new Promise((resolve, reject) => {
      post(dispatch, host +"google_organic/v1/get_google_organic_urls", payload)
        .then((r) => {
          if (r.ok) { 
            resolve(r.data);
          } else { 
            reject(r);
          }
        })
        .catch((err) => { 
          reject(err);
        })
    })
  }
}
