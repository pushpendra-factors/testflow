/* eslint-disable */

import { get, post, del, getHostUrl } from '../utils/request';
var host = getHostUrl();
    host = (host[host.length - 1] === '/') ? host : host + '/';


    const inititalState = {
      loading: false,
      error: false,
      metadata: {},
      weekly_insights: {},
      active_insight: {},
    };

    export default function reducer(state= inititalState, action) {
      switch (action.type) {
        
        case 'RESET_WEEKLY_INSIGHTS': {
          return { ...state, weekly_insights: inititalState.weekly_insights };
        } 
        case 'SET_ACTIVE_INSIGHT': {
          return { ...state, active_insight: action.payload };
        } 
        case 'FETCH_WEEKLY_INSIGHTS_FULLFILLED': {
          return { ...state, weekly_insights: action.payload };
        } 
        case 'FETCH_WEEKLY_INSIGHTS_REJECTED': {
          return { ...state };
        }  
        case 'FETCH_WEEKLY_INSIGHTS_METADATA_FULLFILLED': {
          return { ...state, metadata: action.payload };
        } 
        case 'FETCH_WEEKLY_INSIGHTS_METADATA_REJECTED': {
          return { ...state };
        }   
        
      }
      return state;
    }

    
    
  

export function fetchWeeklyIngishts(projectID, dashboardID, baseTime, startTime, isDashboard=true) {
  const queryURL = isDashboard ? 'dashboard_unit_id' : 'query_id';
  return function(dispatch) {
    return new Promise((resolve,reject) => {
      get(dispatch, host + "projects/"+projectID+"/insights?"+queryURL+"="+dashboardID+"&base_start_time="+baseTime+"&comp_start_time="+startTime+"&insights_type=w&number_of_records=11")        
        .then((response)=>{        
          dispatch({type:"FETCH_WEEKLY_INSIGHTS_FULLFILLED", payload: response.data});
          resolve(response)
        }).catch((err)=>{        
          dispatch({type:"FETCH_WEEKLY_INSIGHTS_REJECTED", payload: err});
          reject(err);
        });
    });
  }
}
export function fetchWeeklyIngishtsMetaData(projectID) {
  return function(dispatch) {
    return new Promise((resolve,reject) => {
      get(dispatch, host + "projects/"+projectID+"/weekly_insights_metadata")        
        .then((response)=>{        
          dispatch({type:"FETCH_WEEKLY_INSIGHTS_METADATA_FULLFILLED", payload: response.data});
          resolve(response)
        }).catch((err)=>{        
          dispatch({type:"FETCH_WEEKLY_INSIGHTS_METADATA_REJECTED", payload: err});
          reject(err);
        });
    });
  }
}
export function updateInsightFeedback(projectID,data) {
  return function(dispatch) {
    return new Promise((resolve,reject) => { 
      post(dispatch, host + "projects/"+projectID+`/feedback`, data)    
        .then((response)=>{        
          dispatch({type:"UPDATE_INSIGHT_FEEDBACK_FULLFILLED", payload: response.data});
          resolve(response)
        }).catch((err)=>{        
          dispatch({type:"UPDATE_INSIGHT_FEEDBACK__REJECTED", payload: err});
          reject(err);
        });
    });
  }
}
