/* eslint-disable */
import { get, post } from '../utils/request';
var host = BUILD_CONFIG.backend_host;
host = (host[host.length - 1] === '/') ? host : host + '/';

export default function reducer(state = {
  agent: {},
  agentError: null,
  isLoggedIn: true,
  billing: {}
}, action) {
  switch (action.type) {
    case 'AGENT_LOGIN_FULFILLED': {
      return { ...state, isLoggedIn: true };
    }
    case 'AGENT_LOGIN_REJECTED': {
      return { ...state, isLoggedIn: false };
    }
    case 'AGENT_LOGOUT_FULFILLED': {
      return { ...state, isLoggedIn: false };
    }
    case 'AGENT_LOGIN_FORCE': {
      return { ...state, isLoggedIn: false };
    }
    case 'FETCH_AGENT_INFO_FULFILLED': {
      return { ...state, agent_details: action.payload };
    }
    case 'FETCH_AGENT_INFO_REJECTED': {
      return { ...state, agentError: action.payload };
    }
    case 'UPDATE_AGENT_INFO_FULFILLED': {
      return { ...state, agent: action.payload };
    }
    case 'UPDATE_AGENT_INFO_REJECTED': {
      return { ...state, agentError: action.payload };
    }
    case 'UPDATE_AGENT_PASSWORD_FULFILLED': {
      return state;
    }
    case "PROJECT_AGENT_INVITE_FULFILLED": {
      let nextState = { ...state };
      
      let projectAgentMapping = action.payload.project_agent_mappings[0];
      nextState.projectAgents = [...state.projectAgents];
      nextState.projectAgents.push(projectAgentMapping);
      nextState.agents[projectAgentMapping.agent_uuid] = action.payload.agents[projectAgentMapping.agent_uuid];        
      return nextState
    }
    case "PROJECT_AGENT_INVITE_REJECTED": {
      return {
        ...state
      }
    }
    case "PROJECT_AGENT_REMOVE_FULFILLED": {
      let nextState = { ...state };
      nextState.projectAgents = state.projectAgents.filter((projectAgent)=>{
        return projectAgent.agent_uuid != action.payload.agent_uuid
      })
      return nextState
    }
  }
  return state;
}

export function login(email, password) {
  return function (dispatch) {
    return new Promise((resolve, reject) => {
      const invalidMsg = 'Invalid email or password';
      const loginFailMsg = 'Login failed. Please try again.';

      post(dispatch,host + 'agents/signin', {
        email,
        password
      })
        .then((r) => {
          if (r.ok) {
            dispatch({
              type: 'AGENT_LOGIN_FULFILLED',
              payload: r.data
            });

            resolve(r.data);
          } else {
            dispatch({
              type: 'AGENT_LOGIN_REJECTED',
              payload: null
            });

            if (r.status == 404) reject(invalidMsg);
            else reject(loginFailMsg);
          }
        })
        .catch((r) => {
          dispatch({
            type: 'AGENT_LOGIN_REJECTED',
            payload: null
          });

          if (r.status && r.status == 401) reject(invalidMsg);
          else reject(loginFailMsg);
        });
    });
  };
}

export function signout() {
  return function (dispatch) {
    return new Promise((resolve, reject) => {
      get(dispatch, host + 'agents/signout')
        .then(() => {
          resolve(dispatch({
            type: 'AGENT_LOGOUT_FULFILLED'
          }));
        })
        .catch(() => {
          reject('Sign out failed');
        });
    });
  };
}

// export function signup(email, phone, planCode){
//   return function(dispatch){
//     return new Promise((resolve, reject) => {
//       dispatch({type: "AGENT_SIGNUP"});

//       post(dispatch, host+"accounts/signup", { email: email, phone: phone, plan_code: planCode })
//         .then((r) => {
//           // status 302 for duplicate email
//           if(r.status != 302)
//           {
//             dispatch({
//               type: "AGENT_SIGNUP_FULFILLED",
//               payload: {}
//             });
//             resolve(r);
//           }
//           else
//           {
//             dispatch({type: "AGENT_SIGNUP_REJECTED", payload: null});
//             reject("Email already exists. Try logging in.")
//           }
//         })
//         .catch( () => {
//           dispatch({type: "AGENT_SIGNUP_REJECTED", payload: null});
//           reject("Sign up failed. Please try again.");
//         });
//     });
//   }
// }


export function fetchAgentInfo(){
  return function(dispatch) {
    return new Promise((resolve,reject) => {
      get(dispatch, host + "agents/info")
        .then((response) => {        
          dispatch({type:"FETCH_AGENT_INFO_FULFILLED", 
            payload: response.data});
          resolve(response);
        })
        .catch((err) => {
          dispatch({type:"FETCH_AGENT_INFO_REJECTED", 
            payload: 'Failed to fetch agent info'});
          reject(err);
        });
    });
  }
} 


export function projectAgentInvite(projectId, emailId){
  return function(dispatch){
    let payload = {"email":emailId};
    return new Promise((resolve, reject) => {
      post(dispatch, host + "projects/" + projectId + "/agents/invite", payload)
      .then((r) => {
        if (r.ok && r.status && r.status == 201){
          dispatch({type: "PROJECT_AGENT_INVITE_FULFILLED", payload: r.data });
          resolve(r.data);
        }else if (r.status && r.status == 409){
          dispatch({type: "PROJECT_AGENT_INVITE_REJECTED", payload: r.data.error }); 
          reject("User Seats limit reached");
        }
        else {
          dispatch({type: "PROJECT_AGENT_INVITE_REJECTED", payload: r.data.error });
          reject(r.data.error);
        }
      })
      .catch((r) => {
        dispatch({type: "PROJECT_AGENT_INVITE_REJECTED", payload: r.data.error });
      });
    });
  }
}

export function projectAgentRemove(projectId, agentUUID){
  return function(dispatch){
    return new Promise((resolve, reject) => {
      put(dispatch, host + "projects/" + projectId +"/agents/remove", {"agent_uuid":agentUUID})
        .then((r) => {
          dispatch({
            type: "PROJECT_AGENT_REMOVE_FULFILLED",
            payload: r.data
          });
          resolve(r.data);
        })
        .catch((r) => {
          dispatch({
            type: "PROJECT_AGENT_REMOVE_REJECTED",
            error: r
          });
          reject({body: r.data, status: r.status});
        });
    })
  }
}