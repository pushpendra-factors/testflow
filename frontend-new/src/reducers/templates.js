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
        case 'FETCH_TEMPLATE_CONFIG_FULFILLED': {
          return { ...state, config: action.payload };
        } 
        case 'FETCH_TEMPLATE_INSIGHT_FULFILLED': {
          return { ...state, insight: action.payload };
        } 
        
      }
      return state;
    }

    
    

export function fetchTemplateConfig(projectID, templateID) {
  return function(dispatch) {
    return new Promise((resolve,reject) => {
      get(dispatch, host + "projects/"+projectID+"/v1/templates/"+templateID+"/config")
        .then((response)=>{        
          dispatch({type:"FETCH_TEMPLATE_CONFIG_FULFILLED", payload: response.data});
          resolve(response)
        }).catch((err)=>{        
          dispatch({type:"FETCH_TEMPLATE_CONFIG_REJECTED", payload: err});
          reject(err);
        });
    });
  }
}
 
export function fetchTemplateInsights(projectID, data) {
    return function(dispatch) {
      return new Promise((resolve,reject) => { 
        post(dispatch, host + "projects/"+projectID+"/v1/templates/1/query", data)
          .then((response)=>{        
            dispatch({type:"FETCH_TEMPLATE_INSIGHT_FULFILLED", payload: response.data});
            resolve(response)
          }).catch((err)=>{        
            dispatch({type:"FETCH_TEMPLATE_INSIGHT_REJECTED", payload: err});
            reject(err);
          });
      });
    }
  }
