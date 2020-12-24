/* eslint-disable */

  import { get, post, put } from '../utils/request';
  var host = BUILD_CONFIG.backend_host;
      host = (host[host.length - 1] === '/') ? host : host + '/';


      const inititalState = {
        loading: false,
        error: false, 
      };

      export default function reducer(state= inititalState, action) {
        switch (action.type) {
          case 'FETCH_GOALS_FULFILLED': {
            return { ...state, goals: action.payload };
          } 
          case 'FETCH_GOALS_FULFILLED': {
            return { ...state, error: action.payload };
          } 
          case 'FETCH_TRACKED_EVENTS_FULFILLED': {
            return { ...state, tracked_events: action.payload };
          } 
          case 'FETCH_TRACKED_EVENTS_REJECTED': {
            return { ...state, error: action.payload };
          } 
          case 'FETCH_TRACKED_USER_PROPERTIES_FULFILLED': {
            return { ...state, tracked_user_property: action.payload };
          } 
          case 'FETCH_TRACKED_USER_PROPERTIES_REJECTED': {
            return { ...state, error: action.payload };
          } 
          case 'FETCH_GOAL_INSIGHTS_FULFILLED': {
            return { ...state, goal_insights: action.payload };
          } 
          case 'FETCH_GOAL_INSIGHTS_REJECTED': {
            return { ...state, error: action.payload };
          } 
          case 'FETCH_FACTORS_MODELS_FULFILLED': {
            return { ...state, factors_models: action.payload };
          } 
          case 'FETCH_FACTORS_MODELS_REJECTED': {
            return { ...state, factors_models: action.payload };
          } 
          case 'SAVE_GOAL_INSIGHT_RULES_FULFILLED': {
            return { ...state, factors_insight_rules: action.payload };
          } 
          
        }
        return state;
      }

      
      

  export function fetchFactorsGoals(projectID) {
    return function(dispatch) {
      return new Promise((resolve,reject) => {
        get(dispatch, host + "projects/"+projectID+"/v1/factors/goals")
          .then((response)=>{        
            dispatch({type:"FETCH_GOALS_FULFILLED", payload: response.data});
            resolve(response)
          }).catch((err)=>{        
            dispatch({type:"FETCH_GOALS_REJECTED", payload: err});
            reject(err);
          });
      });
    }
  }
  
  export function fetchFactorsTrackedEvents(projectID) {
    return function(dispatch) {
      return new Promise((resolve,reject) => {
        get(dispatch, host + "projects/"+projectID+"/v1/factors/tracked_event")
          .then((response)=>{        
            dispatch({type:"FETCH_TRACKED_EVENTS_FULFILLED", payload: response.data});
            resolve(response)
          }).catch((err)=>{        
            dispatch({type:"FETCH_TRACKED_EVENTS_REJECTED", payload: err});
            reject(err);
          });
      });
    }
  }
  export function fetchFactorsTrackedUserProperties(projectID) {
    return function(dispatch) {
      return new Promise((resolve,reject) => {
        get(dispatch, host + "projects/"+projectID+"/v1/factors/tracked_user_property")
          .then((response)=>{        
            dispatch({type:"FETCH_TRACKED_USER_PROPERTIES_FULFILLED", payload: response.data});
            resolve(response)
          }).catch((err)=>{        
            dispatch({type:"FETCH_TRACKED_USER_PROPERTIES_REJECTED", payload: err});
            reject(err);
          });
      });
    }
  }
  export function fetchGoalInsights(projectID, isJourney=false, data, modelId) {
    return function(dispatch) {
      return new Promise((resolve,reject) => {
        const insightsUrl = `/v1/factor?type=${isJourney ? 'journey' : 'singleevent'}&model_id=${modelId}`;
        post(dispatch, host + "projects/"+projectID+ insightsUrl, data)
          .then((response)=>{        
            dispatch({type:"FETCH_GOAL_INSIGHTS_FULFILLED", payload: response.data});
            resolve(response)
          }).catch((err)=>{        
            dispatch({type:"FETCH_GOAL_INSIGHTS_REJECTED", payload: err});
            reject(err);
          });
      });
    }
  }
  export function fetchFactorsModels(projectID) {
    return function(dispatch) {
      return new Promise((resolve,reject) => {
        get(dispatch, host + "projects/"+projectID+"/models")
          .then((response)=>{        
            dispatch({type:"FETCH_FACTORS_MODELS_FULFILLED", payload: response.data});
            resolve(response)
          }).catch((err)=>{        
            dispatch({type:"FETCH_FACTORS_MODELS_REJECTED", payload: err});
            reject(err);
          });
      });
    }
  }

export function saveGoalInsightRules(data) {
    return function(dispatch) {
      dispatch({type:"SAVE_GOAL_INSIGHT_RULES_FULFILLED", payload: data}); 
    }
}

export function saveGoalInsights(projectID, data) {
    return function(dispatch) {
      return new Promise((resolve,reject) => { 
        post(dispatch, host + "projects/"+projectID+`/v1/factors/goals`, data)
          .then((response)=>{        
            dispatch({type:"SAVE_GOAL_INSIGHTS_FULFILLED", payload: response.data});
            resolve(response)
          }).catch((err)=>{        
            dispatch({type:"SAVE_GOAL_INSIGHTS_REJECTED", payload: err});
            reject(err);
          });
      });
    }
  }