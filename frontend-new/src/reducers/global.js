/* eslint-disable */

import {
  FUNNEL_RESULTS_AVAILABLE, FUNNEL_RESULTS_UNAVAILABLE, SET_PROJECTS, SET_ACTIVE_PROJECT
} from './types';
import { get } from '../request';

const defaultState = {
  is_funnel_results_visible: false,
  funnel_events: [],
  projects: [],
  active_project: {}
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
      get(host + 'projects', {})
        .then((response) => {
          dispatch(setActiveProject(response.data.projects[1]));
          resolve(dispatch(fetchProjectAction(response.data.projects)));
        }).catch((err) => {
          resolve(dispatch(fetchProjectAction([])));
        });
    });
  };
}
