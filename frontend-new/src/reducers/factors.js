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