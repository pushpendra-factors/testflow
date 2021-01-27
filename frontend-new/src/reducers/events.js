import { get, post, del } from '../utils/request';
var host = BUILD_CONFIG.backend_host;
    host = (host[host.length - 1] === '/') ? host : host + '/';


    const inititalState = {
        loading: false,
        error: false, 
      };

      export default function reducer(state= inititalState, action) {
        switch (action.type) { 
          case 'FETCH_SMART_EVENTS_FULFILLED': {
            return { ...state, smart_events: action.payload };
          } 
          case 'FETCH_SMART_EVENTS_REJECTED': {
            return { ...state, error: action.payload };
          } 
          case 'FETCH_OBJECTPROPERTIESBYSOURCE_FULFILLED': {
            return { ...state, objPropertiesSource: action.payload };
          } 
          case 'FETCH_OBJECTPROPERTIESBYSOURCE_REJECTED': {
            return { ...state, error: action.payload };
          } 
          case 'SAVE_SMART_EVENTS_FULFILLED': {
            return { ...state };
          } 
          case 'SAVE_SMART_EVENTS_REJECTED': {
            return { ...state, error: action.payload };
          } 
          
        }
        return state;
      }


export function fetchSmartEvents(projectID) {
        return function(dispatch) {
          return new Promise((resolve,reject) => {
            get(dispatch, host + "projects/"+projectID+'/v1/smart_event')
              .then((response)=>{        
                dispatch({type:"FETCH_SMART_EVENTS_FULFILLED", payload: response.data});
                resolve(response)
              }).catch((err)=>{        
                dispatch({type:"FETCH_SMART_EVENTS_REJECTED", payload: err});
                reject(err);
              });
          });
        }
} 

export function saveSmartEvents(projectID,data) {
  return function(dispatch) {
    return new Promise((resolve,reject) => {
      post(dispatch, host + "projects/"+projectID+'/v1/smart_event?type=crm',data)
      .then((response)=>{        
        dispatch({type:"SAVE_SMART_EVENTS_FULFILLED", payload: response.data});
        resolve(response)
      }).catch((err)=>{        
        dispatch({type:"SAVE_SMART_EVENTS_REJECTED", payload: err});
        reject(err);
      });
    });
  }
}

export function fetchObjectPropertiesbySource(projectID,source,dataObj) {
        return function(dispatch) {
          return new Promise((resolve,reject) => {
            get(dispatch, host + "projects/"+projectID+'/v1/crm/'+source+'/'+dataObj+'/properties')
              .then((response)=>{        
                dispatch({type:"FETCH_OBJECTPROPERTIESBYSOURCE_FULFILLED", payload: response.data});
                resolve(response)
              }).catch((err)=>{        
                dispatch({type:"FETCH_OBJECTPROPERTIESBYSOURCE_REJECTED", payload: err});
                reject(err);
              });
          });
        }
} 
    