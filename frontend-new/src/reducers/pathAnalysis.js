/* eslint-disable */

import { get, post, del, getHostUrl } from '../utils/request';
var host = getHostUrl();
    host = (host[host.length - 1] === '/') ? host : host + '/';


    const inititalState = {
      loading: false,
      error: false, 
    };

    export default function reducer(state= inititalState, action) {
      switch (action.type) {
       
        case 'FETCH_SAVED_PATH_FULFILLED': {
          return { ...state, savedQuery: action.payload };
        } 
        case 'FETCH_SAVED_PATHINSIGHTS_FULFILLED': {
          return { ...state, activeInsights: action.payload };
        } 
        case 'SAVEDPATH_REMOVED_FULFILLED': {
          return { ...state};
        } 
        case 'SET_ACTVIVE_INSIGHTS_FULFILLED': {
          return { ...state, activeQuery: action.payload};
        } 
        case 'RESET_ACTVIVE_INSIGHTS': {
          return { ...state, activeQuery: null, activeInsights: null};
        } 
        case 'PATHQUERY_CREATED_FULFILLED': {
          return { ...state};
        } 
        
      }
      return state;
    }

     
export function fetchSavedPathAnalysis(projectID) {
  return function(dispatch) {
    return new Promise((resolve,reject) => {
      get(dispatch, host + "projects/"+projectID+"/v1/pathanalysis")
        .then((response)=>{        
          dispatch({type:"FETCH_SAVED_PATH_FULFILLED", payload: response.data});
          resolve(response)
        }).catch((err)=>{        
          dispatch({type:"FETCH_SAVED_PATH_REJECTED", payload: err});
          reject(err);
        });
    });
  }
}
export function fetchPathAnalysisInsights(projectID,query_id) {
  return function(dispatch) {
    return new Promise((resolve,reject) => {
      get(dispatch, host + "projects/"+projectID+"/v1/pathanalysis/"+ query_id+"?version=2")
        .then((response)=>{        
          dispatch({type:"FETCH_SAVED_PATHINSIGHTS_FULFILLED", payload: response.data});
          resolve(response)
        }).catch((err)=>{        
          dispatch({type:"FETCH_SAVED_PATHINSIGHTS_REJECTED", payload: err});
          reject(err);
        });
    });
  }
}

export function setActiveInsightQuery(data) { 
  return function(dispatch) {
    if(data){
      dispatch({type:"SET_ACTVIVE_INSIGHTS_FULFILLED", payload: data});
    }
    else{
      dispatch({type:"SET_ACTVIVE_INSIGHTS_REJECTED", payload: null});
    } 
  }
}

export function removeSavedQuery(projectID, data) { 
    return function(dispatch) {
      return new Promise((resolve,reject) => { 
        del(dispatch, host + "projects/"+projectID+`/v1/pathanalysis/`+ data)
          .then((response)=>{        
            dispatch({type:"SAVEDPATH_REMOVED_FULFILLED", payload: response.data});
            resolve(response)
          }).catch((err)=>{        
            dispatch({type:"SAVEDPATH_REMOVED_REJECTED", payload: err});
            reject(err);
          });
      });
    }
  }



  export function createPathPathAnalysisQuery(projectID, data) {
    return function (dispatch) {
      return new Promise((resolve, reject) => {
        post( dispatch, host + 'projects/' + projectID + `/v1/pathanalysis`, data )
          .then((response) => {
            dispatch({
              type: 'PATHQUERY_CREATED_FULFILLED',
              payload: response.data,
            });
            resolve(response);
          })
          .catch((err) => {
            dispatch({ type: 'PATHQUERY_CREATED_REJECTED', payload: err });
            reject(err);
          });
      });
    };
  }