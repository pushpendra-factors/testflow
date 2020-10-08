/* eslint-disable */

import {
  FUNNEL_RESULTS_AVAILABLE, FUNNEL_RESULTS_UNAVAILABLE, SET_PROJECTS, SET_ACTIVE_PROJECT, CREATE_PROJECT_FULFILLED, FETCH_PROJECTS_REJECTED
} from './types';
import { get, post } from '../utils/request';

const defaultState = {
  is_funnel_results_visible: false,
  funnel_events: [],
  projects: [],
  active_project: {},
  fetchingProjects: null,
  projectsError: null
};

export default function (state = defaultState, action) {
  switch (action.type) {
    case FUNNEL_RESULTS_AVAILABLE:
      return { ...state, is_funnel_results_visible: true, funnel_events: action.payload };
    case FUNNEL_RESULTS_UNAVAILABLE:
      return { ...state, is_funnel_results_visible: false, funnel_events: [] };
    case SET_PROJECTS:
      return { ...state, projects: action.payload };
    case SET_ACTIVE_PROJECT:
      return { ...state, active_project: action.payload };
    case CREATE_PROJECT_FULFILLED : {
        let _state = { ...state  };
        _state.projects = [..._state.projects, action.payload ];     
        // Set currentProjectId to this newly created project        
        _state.active_project = action.payload;        
        return _state;
      };
    case FETCH_PROJECTS_REJECTED: {
        return {...state, fetchingProjects: false, projectsError: action.payload}
      }
    default:
      return state;
  }
}

// Action creators
export function fetchProjectAction(projects, status = 'success') {
  return { type: SET_PROJECTS, payload: projects };
}

export const setActiveProject = project => {
  return { type: SET_ACTIVE_PROJECT, payload: project };
};

// Service Call

export function fetchProjects(projects) {
  return function (dispatch) {
    let host = BUILD_CONFIG.backend_host;
    host = (host[host.length - 1] === '/') ? host : host + '/';
    return new Promise((resolve, reject) => {
      get(dispatch, host + 'projects', {})
        .then((response) => {
          dispatch(setActiveProject(response.data.projects[0]));
          resolve(dispatch(fetchProjectAction(response.data.projects)));
        }).catch((err) => {
          resolve(dispatch(fetchProjectAction([])));
        });
    });
  };
}


export function createProject(projectName){
  console.log('createProject called',projectName)
  return function(dispatch){
    let host = BUILD_CONFIG.backend_host;
    host = (host[host.length - 1] === '/') ? host : host + '/';
    return new Promise((resolve, reject) => {
      post(dispatch, host + "projects", { name: projectName })
        .then((r) => {
          if (r.ok) {
            dispatch({ type: "CREATE_PROJECT_FULFILLED", payload: r.data })
            resolve(r);
          } else {
            dispatch({ type:"CREATE_PROJECT_REJECTED", payload: "Failed to create project." });
            reject(r);
          }
        })
        .catch((err) => {
          dispatch({ type:"CREATE_PROJECT_REJECTED", payload: err });
          reject(err);
        })
    })
  }
}