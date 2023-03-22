/* eslint-disable */

import {
  SET_PROJECTS,
  SET_ACTIVE_PROJECT,
  CREATE_PROJECT_FULFILLED,
  FETCH_PROJECTS_REJECTED
} from './types';
import { get, getHostUrl, post, put, del } from '../utils/request';

var host = getHostUrl();
host = host[host.length - 1] === '/' ? host : host + '/';

const defaultState = {
  is_funnel_results_visible: false,
  funnel_events: [],
  projects: [],
  active_project: {},
  fetchingProjects: null,
  projectsError: null,
  currentProjectSettings: {},
  currentProjectSettingsLoading: false,
  contentGroup: [],
  bingAds: {},
  marketo: {}
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
        projectsError: action.payload
      };
    }
    case 'CREATE_PROJECT_TIMEZONE_FULFILLED': {
      let _state = { ...state };
      _state.projects = [..._state.projects, action.payload];
      // Set currentProjectId to this newly created project
      _state.active_project = action.payload;
      //Update timezone
      if (_state.currentProjectSettings)
        _state.currentProjectSettings = {
          ..._state.currentProjectSettings,
          ...action.payload.time_zone
        };
      return _state;
    }
    case 'UPDATE_PROJECT_SETTINGS_FULFILLED': {
      let _state = { ...state };
      if (_state.currentProjectSettings)
        _state.currentProjectSettings = {
          ..._state.currentProjectSettings,
          ...action.payload.updatedSettings
        };
      return _state;
    }
    case 'UPDATE_PROJECT_SETTINGS_REJECTED': {
      return {
        ...state,
        projectEventsError: action.payload.err
      };
    }
    case 'UPDATE_PROJECT_DETAILS_FULFILLED': {
      let _state = { ...state };
      if (_state.active_project)
        _state.active_project = {
          ..._state.active_project,
          ...action.payload.updatedDetails
        };
      return _state;
    }
    case 'UPDATE_PROJECT_DETAILS_REJECTED': {
      return {
        ...state,
        projectEventsError: action.payload.err
      };
    }
    case 'FETCH_PROJECT_SETTINGS_LOADING': {
      return {
        ...state,
        currentProjectSettingsLoading: true
      };
    }
    case 'FETCH_PROJECT_SETTINGS_FULFILLED': {
      return {
        ...state,
        currentProjectSettingsLoading: false,
        currentProjectSettings: action.payload.settings
      };
    }
    case 'FETCH_PROJECT_SETTINGS_V1_FULFILLED': {
      return {
        ...state,
        projectSettingsV1: action.payload.settings
      };
    }
    case 'FETCH_PROJECT_SETTINGS_REJECTED': {
      return {
        ...state,
        currentProjectSettingsLoading: false,
        projectSettingsError: action.payload.err
      };
    }
    case 'ENABLE_FACEBOOK_USER_ID': {
      let fbUserID = action.payload.int_facebook_user_id;

      let _state = { ...state };
      _state.currentProjectSettings = {
        ...state.currentProjectSettings,
        int_facebook_user_id: fbUserID
      };
      return _state;
    }
    case 'ENABLE_SALESFORCE_FULFILLED': {
      let enabledAgentUUID = action.payload.int_salesforce_enabled_agent_uuid;
      if (!enabledAgentUUID || enabledAgentUUID == '') return state;

      let _state = { ...state };
      _state.currentProjectSettings = {
        ...state.currentProjectSettings,
        int_salesforce_enabled_agent_uuid: enabledAgentUUID
      };
      return _state;
    }
    case 'FETCH_PROJECTS_REJECTED': {
      return {
        ...state,
        fetchingProjects: false,
        projectsError: action.payload
      };
    }
    case 'FETCH_PROJECTS_FULFILLED': {
      // Indexed project objects by projectId. Kept projectId on value also intentionally
      // for array of projects from Object.values().

      let projectsWithRoles = [];
      _.toArray(action.payload).map((project, index) => {
        project.map((projectDetails) => {
          projectDetails.role = index + 1;
          projectsWithRoles.push(projectDetails);
        });
      });

      return {
        ...state,
        projects: projectsWithRoles
      };
    }
    case 'CREATE_CONTENT_GROUP': {
      const props = [...state.contentGroup];
      props.push(action.payload);
      return { ...state, contentGroup: props };
    }
    case 'UPDATE_CONTENT_GROUP': {
      const propsToUpdate = [
        ...state.contentGroup.map((prop, i) => {
          if (prop.id === action.payload.id) {
            return action.payload;
          } else {
            return prop;
          }
        })
      ];
      return { ...state, contentGroup: propsToUpdate };
    }
    case 'FETCH_CONTENT_GROUP': {
      return { ...state, contentGroup: action.payload };
    }
    case 'FETCH_BINGADS_FULFILLED': {
      return { ...state, bingAds: action.payload };
    }
    case 'FETCH_BINGADS_REJECTED': {
      return { ...state, bingAds: action.payload };
    }
    case 'DISABLE_BINGADS_FULFILLED': {
      return { ...state, bingAds: {} };
    }
    case 'FETCH_ALERTS': {
      return { ...state, Alerts: action.payload };
    }
    case 'FETCH_EVENT_ALERTS': {
      return { ...state, eventAlerts: action.payload };
    }
    case 'FETCH_SHARED_ALERTS': {
      return { ...state, sharedAlerts: action.payload };
    }
    case 'FETCH_MARKETO_FULFILLED': {
      return { ...state, marketo: action.payload };
    }
    case 'FETCH_MARKETO_REJECTED': {
      return { ...state, marketo: action.payload };
    }
    case 'DISABLE_MARKETO_FULFILLED': {
      return { ...state, marketo: {} };
    }
    case 'FETCH_SLACK_FULFILLED': {
      return { ...state, slack: action.payload };
    }
    case 'FETCH_TEAMS_FULFILLED': {
      return { ...state, teams: action.payload };
    }
    case 'FETCH_SLACK_REJECTED': {
      return { ...state, slack: action.payload };
    }
    case 'FETCH_TEAMS_REJECTED': {
      return { ...state, teams: action.payload };
    }
    case 'DISABLE_SLACK_FULFILLED': {
      return { ...state, slack: {} };
    }
    case 'DISABLE_TEAMS_FULFILLED': {
      return { ...state, teams: {} };
    }
    default:
      return state;
  }
}

// Action creators
export function fetchProjectAction(projects, status = 'success') {
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

export function fetchProjects() {
  return function (dispatch) {
    return new Promise((resolve, reject) => {
      get(dispatch, host + 'v1/projects')
        .then((response) => {
          dispatch({
            type: 'FETCH_PROJECTS_FULFILLED',
            payload: response.data
          });
          resolve(response);
        })
        .catch((err) => {
          dispatch({ type: 'FETCH_PROJECTS_REJECTED', payload: err });
          reject(err);
        });
    });
  };
}

export function fetchDemoProject() {
  return function (dispatch) {
    return new Promise((resolve, reject) => {
      get(dispatch, host + 'v1/demoprojects')
        .then((response) => {
          resolve(response);
        })
        .catch((err) => {
          dispatch({ type: 'FETCH_PROJECTS_REJECTED', payload: err });
          reject(err);
        });
    });
  };
}

export function createProject(projectName) {
  return function (dispatch) {
    return new Promise((resolve, reject) => {
      post(dispatch, host + 'projects?create_dashboard=false', {
        name: projectName
      })
        .then((r) => {
          if (r.ok) {
            dispatch({ type: 'CREATE_PROJECT_FULFILLED', payload: r.data });
            resolve(r);
          } else {
            dispatch({
              type: 'CREATE_PROJECT_REJECTED',
              payload: 'Failed to create project.'
            });
            reject(r);
          }
        })
        .catch((err) => {
          dispatch({ type: 'CREATE_PROJECT_REJECTED', payload: err });
          reject(err);
        });
    });
  };
}

export function createProjectWithTimeZone(data) {
  return function (dispatch) {
    return new Promise((resolve, reject) => {
      post(dispatch, host + 'projects?create_dashboard=false', data)
        .then((r) => {
          if (r.ok) {
            dispatch({
              type: 'CREATE_PROJECT_TIMEZONE_FULFILLED',
              payload: r.data
            });
            resolve(r);
          } else {
            dispatch({
              type: 'CREATE_PROJECT_TIMEZONE_REJECTED',
              payload: 'Failed to create project.'
            });
            reject(r);
          }
        })
        .catch((err) => {
          dispatch({ type: 'CREATE_PROJECT_TIMEZONE_REJECTED', payload: err });
          reject(err);
        });
    });
  };
}

export function fetchProjectSettings(projectId) {
  return function (dispatch) {
    return new Promise((resolve, reject) => {
      dispatch({ type: 'FETCH_PROJECT_SETTINGS_LOADING' });
      get(dispatch, host + 'projects/' + projectId + '/settings')
        .then((r) => {
          if (r.ok) {
            dispatch({
              type: 'FETCH_PROJECT_SETTINGS_FULFILLED',
              payload: {
                currentProjectId: projectId,
                settings: r.data
              }
            });

            resolve(r);
          } else {
            dispatch({
              type: 'FETCH_PROJECT_SETTINGS_REJECTED',
              payload: {
                currentProjectId: projectId,
                settings: {},
                err: 'Failed to get project settings.'
              }
            });

            reject(r);
          }
        })
        .catch((err) => {
          dispatch({
            type: 'FETCH_PROJECT_SETTINGS_REJECTED',
            payload: {
              currentProjectId: projectId,
              settings: {},
              err: err
            }
          });

          reject(err);
        });
    });
  };
}

export function fetchProjectSettingsV1(projectId) {
  return function (dispatch) {
    return new Promise((resolve, reject) => {
      get(dispatch, host + 'projects/' + projectId + '/v1/settings')
        .then((r) => {
          if (r.ok) {
            dispatch({
              type: 'FETCH_PROJECT_SETTINGS_V1_FULFILLED',
              payload: {
                currentProjectId: projectId,
                settings: r.data
              }
            });

            resolve(r);
          } else {
            dispatch({
              type: 'FETCH_PROJECT_SETTINGS_REJECTED',
              payload: {
                currentProjectId: projectId,
                settings: {},
                err: 'Failed to get project settings.'
              }
            });

            reject(r);
          }
        })
        .catch((err) => {
          dispatch({
            type: 'FETCH_PROJECT_SETTINGS_REJECTED',
            payload: {
              currentProjectId: projectId,
              settings: {},
              err: err
            }
          });

          reject(err);
        });
    });
  };
}

export function udpateProjectSettings(projectId, payload) {
  return function (dispatch) {
    return new Promise((resolve, reject) => {
      put(dispatch, host + 'projects/' + projectId + '/settings', payload)
        .then((response) => {
          dispatch({
            type: 'UPDATE_PROJECT_SETTINGS_FULFILLED',
            payload: {
              updatedSettings: payload
            }
          });
          resolve(response);
        })
        .catch((err) => {
          dispatch({
            type: 'UPDATE_PROJECT_SETTINGS_REJECTED',
            payload: {
              updatedSettings: {},
              err: err
            }
          });
          reject(err);
        });
    });
  };
}

export function udpateProjectDetails(projectId, payload) {
  return function (dispatch) {
    return new Promise((resolve, reject) => {
      put(dispatch, host + 'projects/' + projectId, payload)
        .then((response) => {
          dispatch({
            type: 'UPDATE_PROJECT_DETAILS_FULFILLED',
            payload: {
              updatedDetails: response.data
            }
          });
          resolve(response);
        })
        .catch((err) => {
          dispatch({
            type: 'UPDATE_PROJECT_DETAILS_REJECTED',
            payload: {
              updatedDetails: {},
              err: err
            }
          });
          reject(err);
        });
    });
  };
}

export function addFacebookAccessToken(data) {
  return function (dispatch) {
    return new Promise((resolve, reject) => {
      post(dispatch, host + 'integrations/facebook/add_access_token', data)
        .then((r) => {
          if (r.ok) {
            dispatch({ type: 'ENABLE_FACEBOOK_USER_ID', payload: data });
            resolve(r);
          } else {
            reject(r);
          }
        })
        .catch((err) => {
          reject(err);
        });
    });
  };
}

export function addLinkedinAccessToken(data) {
  return function (dispatch) {
    return new Promise((resolve, reject) => {
      post(dispatch, host + 'integrations/linkedin/add_access_token', data)
        .then((r) => {
          if (r.ok) {
            dispatch({ type: 'ENABLE_LINKEDIN_AD_ACCOUNT', payload: data });
            resolve(r);
          } else {
            reject(r);
          }
        })
        .catch((err) => {
          reject(err);
        });
    });
  };
}

export function enableSalesforceIntegration(projectId) {
  return function (dispatch) {
    return new Promise((resolve, reject) => {
      let payload = { project_id: projectId };
      post(dispatch, host + 'integrations/salesforce/enable', payload)
        .then((r) => {
          if (r.ok) {
            dispatch({ type: 'ENABLE_SALESFORCE_FULFILLED', payload: r.data });
            resolve(r);
          } else {
            dispatch({ type: 'ENABLE_SALESFORCE_REJECTED' });
            reject(r);
          }
        })
        .catch((err) => {
          dispatch({ type: 'ENABLE_SALESFORCE_REJECTED', payload: err });
          reject(err);
        });
    });
  };
}

export function fetchSalesforceRedirectURL(projectId, agentUUID) {
  return function (dispatch) {
    return new Promise((resolve, reject) => {
      let payload = { project_id: projectId };

      post(dispatch, host + 'integrations/salesforce/auth', payload)
        .then((r) => {
          if (r.ok) {
            dispatch({
              type: 'FETCH_SALESFORCE_REDIRECT_URL_FULFILLED',
              payload: r.data
            });
            resolve(r);
          } else {
            dispatch({ type: 'FETCH_SALESFORCE_REDIRECT_URL_REJECTED' });
            reject(r);
          }
        })
        .catch((err) => {
          dispatch({
            type: 'FETCH_SALESFORCE_REDIRECT_URL_REJECTED',
            payload: err
          });
          reject(err);
        });
    });
  };
}

export function enableAdwordsIntegration(projectId) {
  return function (dispatch) {
    return new Promise((resolve, reject) => {
      let payload = { project_id: projectId.toString() };
      post(dispatch, host + 'integrations/adwords/enable', payload)
        .then((r) => {
          if (r.ok) {
            dispatch({ type: 'ENABLE_ADWORDS_FULFILLED', payload: r.data });
            resolve(r);
          } else {
            dispatch({ type: 'ENABLE_ADWORDS_REJECTED' });
            reject(r);
          }
        })
        .catch((err) => {
          dispatch({ type: 'ENABLE_ADWORDS_REJECTED', payload: err });
          reject(err);
        });
    });
  };
}
export function enableSearchConsoleIntegration(projectId) {
  return function (dispatch) {
    return new Promise((resolve, reject) => {
      let payload = { project_id: projectId.toString() };
      post(dispatch, host + 'integrations/google_organic/enable', payload)
        .then((r) => {
          if (r.ok) {
            dispatch({ type: 'ENABLE_ADWORDS_FULFILLED', payload: r.data });
            resolve(r);
          } else {
            dispatch({ type: 'ENABLE_ADWORDS_REJECTED' });
            reject(r);
          }
        })
        .catch((err) => {
          dispatch({ type: 'ENABLE_ADWORDS_REJECTED', payload: err });
          reject(err);
        });
    });
  };
}

export function fetchAdwordsCustomerAccounts(payload) {
  return function (dispatch) {
    return new Promise((resolve, reject) => {
      post(dispatch, host + 'adwords/v1/get_customer_accounts', payload)
        .then((r) => {
          if (r.ok) {
            dispatch({
              type: 'FETCH_ADWORDS_CUSTOMER_ACCOUNTS_FULFILLED',
              payload: r.data
            });
            resolve(r.data);
          } else {
            dispatch({ type: 'FETCH_ADWORDS_CUSTOMER_ACCOUNTS_REJECTED' });
            reject(r);
          }
        })
        .catch((err) => {
          dispatch({
            type: 'FETCH_ADWORDS_CUSTOMER_ACCOUNTS_REJECTED',
            payload: err
          });
          reject(err);
        });
    });
  };
}
export function fetchSearchConsoleCustomerAccounts(payload) {
  return function (dispatch) {
    return new Promise((resolve, reject) => {
      post(
        dispatch,
        host + 'google_organic/v1/get_google_organic_urls',
        payload
      )
        .then((r) => {
          if (r.ok) {
            resolve(r.data);
          } else {
            reject(r);
          }
        })
        .catch((err) => {
          reject(err);
        });
    });
  };
}

export function createBingAdsIntegration(projectId) {
  return function (dispatch) {
    return new Promise((resolve, reject) => {
      post(dispatch, host + 'projects/' + projectId + '/v1/bingads')
        .then((r) => {
          if (r.ok) {
            dispatch({ type: 'CREATE_BINGADS_FULFILLED', payload: r.data });
            resolve(r);
          } else {
            dispatch({ type: 'CREATE_BINGADS_REJECTED' });
            reject(r);
          }
        })
        .catch((err) => {
          dispatch({ type: 'CREATE_BINGADS_REJECTED', payload: err });
          reject(err);
        });
    });
  };
}

export function enableBingAdsIntegration(projectId) {
  return function (dispatch) {
    return new Promise((resolve, reject) => {
      put(dispatch, host + 'projects/' + projectId + '/v1/bingads/enable')
        .then((r) => {
          if (r.ok) {
            dispatch({ type: 'ENABLE_BINGADS_FULFILLED', payload: r.data });
            resolve(r);
          } else {
            dispatch({ type: 'ENABLE_BINGADS_REJECTED' });
            reject(r);
          }
        })
        .catch((err) => {
          dispatch({ type: 'ENABLE_BINGADS_REJECTED', payload: err });
          reject(err);
        });
    });
  };
}

export function disableBingAdsIntegration(projectId) {
  return (dispatch) => {
    return new Promise((resolve, reject) => {
      del(dispatch, host + 'projects/' + projectId + '/v1/bingads/disable', {})
        .then((res) => {
          if (res.ok) {
            dispatch({ type: 'DISABLE_BINGADS_FULFILLED', payload: res.data });
            resolve(res);
          } else {
            reject(res);
          }
        })
        .catch((err) => {
          reject(err);
        });
    });
  };
}

export function fetchBingAdsIntegration(projectId) {
  return function (dispatch) {
    return new Promise((resolve, reject) => {
      get(dispatch, host + 'projects/' + projectId + '/v1/bingads', {})
        .then((r) => {
          if (r.ok) {
            dispatch({ type: 'FETCH_BINGADS_FULFILLED', payload: r.data });
            resolve(r);
          } else {
            dispatch({ type: 'FETCH_BINGADS_REJECTED', payload: {} });
            reject(r);
          }
        })
        .catch((err) => {
          dispatch({ type: 'FETCH_BINGADS_REJECTED', payload: {} });
          reject(err);
        });
    });
  };
}

export function createMarketoIntegration(projectId) {
  return function (dispatch) {
    return new Promise((resolve, reject) => {
      post(dispatch, host + 'projects/' + projectId + '/v1/marketo')
        .then((r) => {
          if (r.ok) {
            dispatch({ type: 'CREATE_MARKETO_FULFILLED', payload: r.data });
            resolve(r);
          } else {
            dispatch({ type: 'CREATE_MARKETO_REJECTED' });
            reject(r);
          }
        })
        .catch((err) => {
          dispatch({ type: 'CREATE_MARKETO_REJECTED', payload: err });
          reject(err);
        });
    });
  };
}

export function enableMarketoIntegration(projectId) {
  return function (dispatch) {
    return new Promise((resolve, reject) => {
      put(dispatch, host + 'projects/' + projectId + '/v1/marketo/enable')
        .then((r) => {
          if (r.ok) {
            dispatch({ type: 'ENABLE_MARKETO_FULFILLED', payload: r.data });
            resolve(r);
          } else {
            dispatch({ type: 'ENABLE_MARKETO_REJECTED' });
            reject(r);
          }
        })
        .catch((err) => {
          dispatch({ type: 'ENABLE_MARKETO_REJECTED', payload: err });
          reject(err);
        });
    });
  };
}

export function disableMarketoIntegration(projectId) {
  return (dispatch) => {
    return new Promise((resolve, reject) => {
      del(dispatch, host + 'projects/' + projectId + '/v1/marketo/disable', {})
        .then((res) => {
          if (res.ok) {
            dispatch({ type: 'DISABLE_MARKETO_FULFILLED', payload: res.data });
            resolve(res);
          } else {
            reject(res);
          }
        })
        .catch((err) => {
          reject(err);
        });
    });
  };
}

export function fetchMarketoIntegration(projectId) {
  return function (dispatch) {
    return new Promise((resolve, reject) => {
      get(dispatch, host + 'projects/' + projectId + '/v1/marketo', {})
        .then((r) => {
          if (r.ok) {
            dispatch({ type: 'FETCH_MARKETO_FULFILLED', payload: r.data });
            resolve(r);
          } else {
            dispatch({ type: 'FETCH_MARKETO_REJECTED', payload: {} });
            reject(r);
          }
        })
        .catch((err) => {
          dispatch({ type: 'FETCH_MARKETO_REJECTED', payload: {} });
          reject(err);
        });
    });
  };
}

export function deleteIntegration(projectId, name) {
  return (dispatch) => {
    return new Promise((resolve, reject) => {
      del(dispatch, host + 'integrations/' + projectId + '/' + name)
        .then((res) => {
          resolve(res);
        })
        .catch((err) => {
          reject(err);
        });
    });
  };
}

export function addContentGroup(projectId, payload) {
  return function (dispatch) {
    return new Promise((resolve, reject) => {
      post(
        dispatch,
        host + 'projects/' + projectId + '/v1/contentgroup',
        payload
      )
        .then((r) => {
          if (r.ok) {
            dispatch({ type: 'CREATE_CONTENT_GROUP', payload: r.data });
            resolve(r);
          } else {
            reject(r);
          }
        })
        .catch((err) => {
          reject(err);
        });
    });
  };
}

export function updateContentGroup(projectId, payload) {
  return function (dispatch) {
    return new Promise((resolve, reject) => {
      put(
        dispatch,
        host + 'projects/' + projectId + '/v1/contentgroup/' + payload.id,
        payload
      )
        .then((r) => {
          if (r.ok) {
            dispatch({ type: 'UPDATE_CONTENT_GROUP', payload: r.data });
            resolve(r);
          } else {
            reject(r);
          }
        })
        .catch((err) => {
          reject(err);
        });
    });
  };
}

export function fetchContentGroup(projectId) {
  return function (dispatch) {
    return new Promise((resolve, reject) => {
      get(dispatch, host + 'projects/' + projectId + '/v1/contentgroup', {})
        .then((r) => {
          if (r.ok) {
            dispatch({ type: 'FETCH_CONTENT_GROUP', payload: r.data });
            resolve(r);
          } else {
            reject(r);
          }
        })
        .catch((err) => {
          reject(err);
        });
    });
  };
}

export function deleteContentGroup(projectId, id) {
  return (dispatch) => {
    return new Promise((resolve, reject) => {
      del(dispatch, host + 'projects/' + projectId + '/v1/contentgroup/' + id)
        .then((res) => {
          resolve(res);
        })
        .catch((err) => {
          reject(err);
        });
    });
  };
}

export function getHubspotContact(email) {
  return function (dispatch) {
    return new Promise((resolve, reject) => {
      get(dispatch, host + 'hubspot/getcontact?email=' + email, {})
        .then((r) => {
          if (r.ok) {
            dispatch({ type: 'FETCH_HUBSPOT_CONTACT', payload: r.data });
            resolve(r);
          } else {
            reject(r);
          }
        })
        .catch((err) => {
          reject(err);
        });
    });
  };
}

export function createAlert(projectId, payload, query_id) {
  return function (dispatch) {
    return new Promise((resolve, reject) => {
      post(
        dispatch,
        host + 'projects/' + projectId + '/v1/alerts?query_id=' + query_id,
        payload
      )
        .then((r) => {
          dispatch({ type: 'CREATE_ALERT', payload: r.data });
          resolve(r);
        })
        .catch((err) => {
          reject(err);
        });
    });
  };
}

export function sendAlertNow(
  projectId,
  payload,
  query_id,
  dateFromTo,
  overrideDate
) {
  return function (dispatch) {
    return new Promise((resolve, reject) => {
      post(
        dispatch,
        host +
          'projects/' +
          projectId +
          '/v1/alerts/send_now?query_id=' +
          query_id +
          '&override_date_range=' +
          overrideDate +
          '&from_time=' +
          dateFromTo.from +
          '&to_time=' +
          dateFromTo.to,
        payload
      )
        .then((r) => {
          if (r.ok) {
            dispatch({ type: 'SEND_ALERT', payload: r.data });
            resolve(r);
          } else {
            reject(r);
          }
        })
        .catch((err) => {
          reject(err);
        });
    });
  };
}

export function fetchAlerts(projectId) {
  return function (dispatch) {
    return new Promise((resolve, reject) => {
      get(dispatch, host + 'projects/' + projectId + '/v1/alerts', {})
        .then((r) => {
          if (r.ok) {
            dispatch({ type: 'FETCH_ALERTS', payload: r.data });
            resolve(r);
          } else {
            reject(r);
          }
        })
        .catch((err) => {
          reject(err);
        });
    });
  };
}

export function fetchSharedAlerts(projectId) {
  return function (dispatch) {
    return new Promise((resolve, reject) => {
      get(
        dispatch,
        host + 'projects/' + projectId + '/v1/alerts?saved_queries=true',
        {}
      )
        .then((r) => {
          if (r.ok) {
            dispatch({ type: 'FETCH_SHARED_ALERTS', payload: r.data });
            resolve(r);
          } else {
            reject(r);
          }
        })
        .catch((err) => {
          reject(err);
        });
    });
  };
}

export function deleteAlert(projectId, id) {
  return (dispatch) => {
    return new Promise((resolve, reject) => {
      del(dispatch, host + 'projects/' + projectId + '/v1/alerts/' + id)
        .then((res) => {
          resolve(res);
        })
        .catch((err) => {
          reject(err);
        });
    });
  };
}

export function editAlert(projectId, payload, id) {
  return (dispatch) => {
    return new Promise((resolve, reject) => {
      put(
        dispatch,
        host + 'projects/' + projectId + '/v1/alerts/' + id,
        payload
      )
        .then((res) => {
          resolve(res);
        })
        .catch((err) => {
          reject(err);
        });
    });
  };
}

export function enableSlackIntegration(projectId, redirect_url = '') {
  return function (dispatch) {
    return new Promise((resolve, reject) => {
      post(
        dispatch,
        host +
          'projects/' +
          projectId +
          '/slack/auth?redirect_url=' +
          redirect_url
      )
        .then((r) => {
          if (r.ok) {
            dispatch({ type: 'ENABLE_SLACK_FULFILLED', payload: r.data });
            resolve(r);
          } else {
            dispatch({ type: 'ENABLE_SLACK_REJECTED' });
            reject(r);
          }
        })
        .catch((err) => {
          dispatch({ type: 'ENABLE_SLACK_REJECTED', payload: err });
          reject(err);
        });
    });
  };
}

export function fetchSlackChannels(projectId) {
  return function (dispatch) {
    return new Promise((resolve, reject) => {
      get(dispatch, host + 'projects/' + projectId + '/slack/channels')
        .then((r) => {
          if (r.ok) {
            dispatch({ type: 'FETCH_SLACK_FULFILLED', payload: r.data });
            resolve(r);
          } else {
            dispatch({ type: 'FETCH_SLACK_REJECTED', payload: {} });
            reject(r);
          }
        })
        .catch((err) => {
          dispatch({ type: 'FETCH_SLACK_REJECTED', payload: {} });
          reject(err);
        });
    });
  };
}

export function disableSlackIntegration(projectId) {
  return (dispatch) => {
    return new Promise((resolve, reject) => {
      del(dispatch, host + 'projects/' + projectId + '/slack/delete', {})
        .then((res) => {
          if (res.ok) {
            dispatch({ type: 'DISABLE_SLACK_FULFILLED', payload: res.data });
            resolve(res);
          } else {
            reject(res);
          }
        })
        .catch((err) => {
          reject(err);
        });
    });
  };
}

export function enableTeamsIntegration(projectId) {
  return function (dispatch) {
    return new Promise((resolve, reject) => {
      post(dispatch, host + 'projects/' + projectId + '/teams/auth')
        .then((r) => {
          if (r.ok) {
            dispatch({ type: 'ENABLE_TEAMS_FULFILLED', payload: r.data });
            resolve(r);
          } else {
            dispatch({ type: 'ENABLE_TEAMS_REJECTED' });
            reject(r);
          }
        })
        .catch((err) => {
          dispatch({ type: 'ENABLE_TEAMS_REJECTED', payload: err });
          reject(err);
        });
    });
  };
}

export function fetchTeamsWorkspace(projectId) {
  return function (dispatch) {
    return new Promise((resolve, reject) => {
      get(dispatch, host + 'projects/' + projectId + '/teams/get_teams')
        .then((r) => {
          if (r.ok) {
            resolve(r);
          } else {
            reject(r);
          }
        })
        .catch((err) => {
          reject(err);
        });
    });
  };
}

export function fetchTeamsChannels(projectId, teamId) {
  return function (dispatch) {
    return new Promise((resolve, reject) => {
      get(dispatch, host + 'projects/' + projectId + '/teams/channels?teams_id=' + teamId)
        .then((r) => {
          if (r.ok) {
            dispatch({ type: 'FETCH_TEAMS_FULFILLED', payload: r.data });
            resolve(r);
          } else {
            dispatch({ type: 'FETCH_TEAMS_REJECTED', payload: {} });
            reject(r);
          }
        })
        .catch((err) => {
          dispatch({ type: 'FETCH_TEAMS_REJECTED', payload: {} });
          reject(err);
        });
    });
  };
}

export function disableTeamsIntegration(projectId) {
  return (dispatch) => {
    return new Promise((resolve, reject) => {
      del(dispatch, host + 'projects/' + projectId + '/teams/delete', {})
        .then((res) => {
          if (res.ok) {
            dispatch({ type: 'DISABLE_TEAMS_FULFILLED', payload: res.data });
            resolve(res);
          } else {
            reject(res);
          }
        })
        .catch((err) => {
          reject(err);
        });
    });
  };
}


export function enableHubspotIntegration(projectId) {
  return function (dispatch) {
    return new Promise((resolve, reject) => {
      let payload = { project_id: projectId.toString() };
      post(dispatch, host + 'integrations/hubspot/auth', payload)
        .then((r) => {
          if (r.ok) {
            dispatch({ type: 'ENABLE_HUBSPOT_FULFILLED', payload: r.data });
            resolve(r);
          } else {
            dispatch({ type: 'ENABLE_HUBSPOT_REJECTED' });
            reject(r);
          }
        })
        .catch((err) => {
          dispatch({ type: 'ENABLE_HUBSPOT_REJECTED', payload: err });
          reject(err);
        });
    });
  };
}

export function enableLeadSquaredIntegration(projectId, payload) {
  return function (dispatch) {
    return new Promise((resolve, reject) => {
      put(
        dispatch,
        host + 'projects/' + projectId + '/leadsquaredsettings',
        payload
      )
        .then((r) => {
          if (r.ok) {
            dispatch({
              type: 'ENABLE_LEADSQUARED_INTEGRATION',
              payload: r.data
            });
            resolve(r);
          } else {
            reject(r);
          }
        })
        .catch((err) => {
          dispatch({ type: 'ENABLE_LEADSQUARED_REJECTED', payload: err });
          reject(err);
        });
    });
  };
}

export function disableLeadSquaredIntegration(projectId) {
  return (dispatch) => {
    return new Promise((resolve, reject) => {
      del(
        dispatch,
        host + 'projects/' + projectId + '/leadsquaredsettings/remove',
        {}
      )
        .then((res) => {
          if (res.ok) {
            dispatch({
              type: 'DISABLE_LEADSQUARED_FULFILLED',
              payload: res.data
            });
            resolve(res);
          } else {
            reject(res);
          }
        })
        .catch((err) => {
          reject(err);
        });
    });
  };
}

export function createEventAlert(projectId, payload) {
  return function (dispatch) {
    return new Promise((resolve, reject) => {
      post(
        dispatch,
        host + 'projects/' + projectId + '/v1/eventtriggeralert',
        payload
      )
        .then((r) => {
          dispatch({ type: 'CREATE_EVENT_ALERT', payload: r.data });
          resolve(r);
        })
        .catch((err) => {
          reject(err);
        });
    });
  };
}

export function fetchEventAlerts(projectId) {
  return function (dispatch) {
    return new Promise((resolve, reject) => {
      get(
        dispatch,
        host + 'projects/' + projectId + '/v1/eventtriggeralert',
        {}
      )
        .then((r) => {
          if (r.ok) {
            dispatch({ type: 'FETCH_EVENT_ALERTS', payload: r.data });
            resolve(r);
          } else {
            reject(r);
          }
        })
        .catch((err) => {
          reject(err);
        });
    });
  };
}

export function deleteEventAlert(projectId, id) {
  return (dispatch) => {
    return new Promise((resolve, reject) => {
      del(
        dispatch,
        host + 'projects/' + projectId + '/v1/eventtriggeralert/' + id
      )
        .then((res) => {
          resolve(res);
        })
        .catch((err) => {
          reject(err);
        });
    });
  };
}

export function editEventAlert(projectId, payload, id) {
  return function (dispatch) {
    return new Promise((resolve, reject) => {
      put(
        dispatch,
        host + 'projects/' + projectId + '/v1/eventtriggeralert/' + id,
        payload
      )
        .then((r) => {
          dispatch({ type: 'EDIT_EVENT_ALERT', payload: r.data });
          resolve(r);
        })
        .catch((err) => {
          reject(err);
        });
    });
  };
}

export function uploadList(projectId, payload) {
  return function (dispatch) {
    return new Promise((resolve, reject) => {
      post(
        dispatch,
        host + 'projects/' + projectId + '/uploadlist',
        payload
      )
        .then((r) => {
          resolve(r);
        })
        .catch((err) => {
          reject(err);
        });
    });
  };
}
