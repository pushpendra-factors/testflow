/* eslint-disable */
import _ from 'lodash';
import { get, post, put, getHostUrl } from '../utils/request';

var host = getHostUrl();
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
    case 'UPDATE_AGENT_ROLE_FULFILLED': {
      return state;
    }
    case "PROJECT_AGENT_INVITE_FULFILLED": {
      let nextState = { ...state };  
      let projectAgentMapping = action.payload.project_agent_mappings[0]; 
      nextState.agents = [...state.agents]; 
      nextState.agents.push(projectAgentMapping);  
      nextState.agents[projectAgentMapping.agent_uuid] = action.payload.agents[projectAgentMapping.agent_uuid];         
      return nextState
    }
    case "PROJECT_AGENT_BATCH_INVITE_FULFILLED": {
      let nextState = { ...state }; 
      for(let i = 0; i < action.payload.project_agent_mappings.length; i++) {
        let projectAgentMapping = action.payload.project_agent_mappings[i]; 
        nextState.agents = [...state.agents]; 
        nextState.agents.push(projectAgentMapping);  
        nextState.agents[projectAgentMapping.agent_uuid] = action.payload.agents[projectAgentMapping.agent_uuid]; 
    } 
      return nextState
    }
    case "PROJECT_AGENT_INVITE_REJECTED": {
      return {
        ...state
      }
    }
    case "PROJECT_AGENT_REMOVE_REJECTED": {
      return {
        ...state
      }
    }
    case "PROJECT_AGENT_REMOVE_FULFILLED": {
      let nextState = { ...state }; 
      nextState.projects = state.projects.filter((projectAgent)=>{ 
        return projectAgent.uuid != action.payload.agent_uuid
      })  
      return nextState
    }
    case "FETCH_PROJECT_AGENTS_FULFILLED":{
      return {
        ...state, 
        agents: action.payload,
      }
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

export function signup(data){
  return function(dispatch){
    return new Promise((resolve, reject) => {
      dispatch({type: "AGENT_SIGNUP"});

      post(dispatch, host+"accounts/signup", data)
        .then((r) => {
          // status 302 for duplicate email
          if(r.status != 302)
          {
            dispatch({
              type: "AGENT_SIGNUP_FULFILLED",
              payload: {}
            });
            resolve(r);
          }
          else
          {
            dispatch({type: "AGENT_SIGNUP_REJECTED", payload: null});
            reject("Email already exists. Try logging in.")
          }
        })
        .catch( () => {
          dispatch({type: "AGENT_SIGNUP_REJECTED", payload: null});
          reject("Sign up failed. Please try again.");
        });
    });
  }
}

export function activate(password, token){
  return function(dispatch){
    return new Promise((resolve, reject) => {
      dispatch({type: "AGENT_VERIFY"});

      post(dispatch, host+"agents/activate?token="+token, { 
        password: password
      })
      .then(() => {
        resolve(dispatch({
          type: "AGENT_VERIFY_FULFILLED",
          payload: {}
        }));
      })
      .catch(() => {
        dispatch({type: "AGENT_VERIFY_REJECTED", payload: null});
        
        reject("Activation failed. Please try again.");
      })
    });
  }
}


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
export function updateAgentInfo(params){
  return function(dispatch){
    return new Promise((resolve, reject)=> {
      put(dispatch, host + "agents/info", params)
      .then((response) => {
        dispatch({type:"UPDATE_AGENT_INFO_FULFILLED", 
            payload: response.data});
          resolve(response);
      })
      .catch((err) => {
        dispatch({type:"UPDATE_AGENT_INFO_REJECTED", 
            payload: 'Failed to update agent info'});
          reject(err);
      })
    });
  }
}

export function fetchProjectAgents(projectId){
  return function(dispatch){
    return new Promise((resolve,reject) => {
       get(dispatch, host + "projects/" + projectId + "/v1/agents")
        .then((r) => {
          dispatch({type: "FETCH_PROJECT_AGENTS_FULFILLED", payload: r.data }); 
          resolve(r);
        })
        .catch((r) => {
          if (r.status) {
            // use this pattern for error handling. 
            // decided to use redux store.
            dispatch({type: "FETCH_PROJECT_AGENTS_REJECTED", payload: r.data, code: r.status });        
          } else {
            // network error. Idea: Use a global error component for this.
            console.log("network error");
          }
          reject(r);
        }); 
    })
  }
}


export function projectAgentInvite(projectId, payload){
  return function(dispatch){ 
    return new Promise((resolve, reject) => {
      post(dispatch, host + "projects/" + projectId + "/agents/invite", payload)
      .then((r) => { 
        if (r.ok && r.status && r.status == 201){ 
          dispatch({type: "PROJECT_AGENT_INVITE_FULFILLED", payload: r.data });
          resolve(r.data);
        }else if (r.status && r.status == 409){ 
          dispatch({type: "PROJECT_AGENT_INVITE_REJECTED", payload: r }); 
          reject("User Seats limit reached");
        }
        else { 
          dispatch({type: "PROJECT_AGENT_INVITE_REJECTED", payload: r });
          reject(r);
        }
      })
      .catch((r) => { 
        dispatch({type: "PROJECT_AGENT_INVITE_REJECTED", payload: r });
      });
    });
  }
}

export function projectAgentBatchInvite(projectId, payload){
  return function(dispatch){ 
    return new Promise((resolve, reject) => {
      post(dispatch, host + "projects/" + projectId + "/agents/batchinvite", payload)
      .then((r) => { 
        if (r.ok && r.status && r.status == 201){ 
          dispatch({type: "PROJECT_AGENT_BATCH_INVITE_FULFILLED", payload: r.data });
          resolve(r.data);
        }else if (r.status && r.status == 409){ 
          dispatch({type: "PROJECT_AGENT_INVITE_REJECTED", payload: r.data.failed_to_invite_agent_idx }); 
          reject("User Seats limit reached");
        }
        else { 
          dispatch({type: "PROJECT_AGENT_INVITE_REJECTED", payload: r.data.failed_to_invite_agent_idx });
          reject(r.data.error);
        }
      })
      .catch((r) => { 
        dispatch({type: "PROJECT_AGENT_INVITE_REJECTED", payload: r.failed_to_invite_agent_idx});
      });
    });
  }
}

export function projectAgentRemove(projectId, agentUUID){
  return function(dispatch){
    return new Promise((resolve, reject) => {
      put(dispatch, host + "projects/" + projectId +"/agents/remove", {"agent_uuid":agentUUID})
        .then((r) => { 
          if (r.status == 403) {
            dispatch({
              type: "PROJECT_AGENT_REMOVE_REJECTED",
              error: r
            });
            reject(r.data.error);
          }
          if (r.status == 202) {
            dispatch({
              type: "PROJECT_AGENT_REMOVE_FULFILLED",
              payload: r.data
            });
            resolve(r.data);
          }

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
          reject(r);
        });
    })
  }
}

export function updateAgentPassword(params){
  return function(dispatch){
    return new Promise((resolve, reject)=> {
      put(dispatch, host + "agents/updatepassword", params)
      .then((response) => {
        if(response.ok){
          dispatch({type:"UPDATE_AGENT_PASSWORD_FULFILLED", 
              payload: response.data});
              resolve(response); 
        }
        else{
          dispatch({type:"UPDATE_AGENT_PASSWORD_REJECTED", 
            payload: 'Failed to update agent password'});
            reject(response.data); 
        }
      })
      .catch((err) => {
        dispatch({type:"UPDATE_AGENT_PASSWORD_REJECTED", 
          payload: 'Failed to update agent password'});
          reject(err);
      })
    });
  }
}
export function updateAgentRole(projectId,uuid,role){
  return function(dispatch){
    return new Promise((resolve, reject)=> {
      put(dispatch, host + "projects/" + projectId +"/agents/update", {"agent_uuid":uuid,role,"role":role})
      .then((response) => {
        if(response.ok){
          dispatch({type:"UPDATE_AGENT_ROLE_FULFILLED", 
              payload: response.data});
              resolve(response); 
        }
        else{
          dispatch({type:"UPDATE_AGENT_ROLE_REJECTED", 
            payload: 'Failed to update agent ROLE'});
            reject(response.data); 
        }
      })
      .catch((err) => {
        dispatch({type:"UPDATE_AGENT_ROLE_REJECTED", 
          payload: 'Failed to update agent ROLE'});
          reject(err);
      })
    });
  }
}

export function forgotPassword(email){
  return function(dispatch){
    return new Promise((resolve, reject) => {
      dispatch({type: "AGENT_FORGOT_PASSWORD"});

      post(dispatch, host+"agents/forgotpassword", { email: email })
        .then(() => {
          resolve(dispatch({
            type: "AGENT_FORGOT_PASSWORD_FULFILLED",
            payload: {}
          }));
        })
        .catch(() => {
          dispatch({type: "AGENT_FORGOT_PASSWORD_REJECTED", payload: null});

          reject("Failed sending the email. Please try again.")
        });
    });
  }
} 

export function setPassword(password, token){
  return function(dispatch){
    return new Promise((resolve, reject) => {
      dispatch({type: "AGENT_SET_PASSWORD"});

      post(dispatch, host+"agents/setpassword?token="+token, {
        password: password
      })
        .then(() => {
          resolve(dispatch({
            type: "AGENT_SET_PASSWORD_FULFILLED",
            payload: {}
          }));
        })
        .catch(() => {
          dispatch({type: "AGENT_SET_PASSWORD_REJECTED", payload: null});

          reject("Reset password failed. Please try again.");
        });
    });
  }
}