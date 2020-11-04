/* eslint-disable */
import { get, post, put } from '../utils/request';
var host = BUILD_CONFIG.backend_host;
host = (host[host.length - 1] === '/') ? host : host + '/';

const agentsSample = [
  {
      "uuid": "6e3803e0-7d63-429f-ba5c-358c6e7d215f",
      "email": "janani@factors.ai",
      "first_name": "Janani",
      "last_name": "Somaskandan",
      "is_email_verified": true,
      "last_logged_in": "2020-10-05T10:11:57.375816+05:30",
      "phone": "123456789",
      "project_id": 14128,
      "role": 2,
      "invited_by": null,
      "created_at": "2020-10-23T10:43:47.105715+05:30",
      "updated_at": "2020-10-23T10:43:47.105715+05:30"
  },
  {
      "uuid": "858388fc-52a1-49e5-bce6-5e6ec793917c",
      "email": "dinesh@factors.ai",
      "first_name": "",
      "last_name": "",
      "is_email_verified": false,
      "last_logged_in": null,
      "phone": "",
      "project_id": 14128,
      "role": 1,
      "invited_by": "6e3803e0-7d63-429f-ba5c-358c6e7d215f",
      "created_at": "2020-10-28T10:51:31.958635+05:30",
      "updated_at": "2020-10-28T10:51:31.958635+05:30"
  },
  {
      "uuid": "765a333b-73c0-49a5-8d14-5afc4d1dd7eb",
      "email": "baliga@factors.ai",
      "first_name": "Vishnu",
      "last_name": "Baliga",
      "is_email_verified": false,
      "last_logged_in": null,
      "phone": "",
      "project_id": 14128,
      "role": 2,
      "invited_by": "6e3803e0-7d63-429f-ba5c-358c6e7d215f",
      "created_at": "2020-10-28T10:51:31.958635+05:30",
      "updated_at": "2020-10-28T10:51:31.958635+05:30"
  }
];

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
    case "FETCH_PROJECTS_REJECTED": {
      return {...state, fetchingProjects: false, projectsError: action.payload}
    }
    case "FETCH_PROJECTS_FULFILLED": {
      // Indexed project objects by projectId. Kept projectId on value also intentionally 
      // for array of projects from Object.values().
      let projects = {};
      for (let project of action.payload.projects) {
        projects[project.id] = project;
      } 

      return {
        ...state, 
        projects: projects, 
      }
    }
    case "PROJECT_AGENT_INVITE_FULFILLED": {
      let nextState = { ...state };  
      let projectAgentMapping = action.payload.project_agent_mappings[0]; 
      nextState.agents = [...state.agents]; 
      nextState.agents.push(projectAgentMapping);  
      nextState.agents[projectAgentMapping.agent_uuid] = action.payload.agents[projectAgentMapping.agent_uuid];         
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
      nextState.agents = state.agents.filter((projectAgent)=>{ 
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

export function fetchProjects() {
  return function(dispatch) {
    return new Promise((resolve,reject) => {
      get(dispatch, host + "projects")
        .then((response)=>{        
          dispatch({type:"FETCH_PROJECTS_FULFILLED", payload: response.data});
          resolve(response)
        }).catch((err)=>{        
          dispatch({type:"FETCH_PROJECTS_REJECTED", payload: err});
          reject(err);
        });
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
          if (r.status == 403) {
            dispatch({
              type: "PROJECT_AGENT_REMOVE_REJECTED",
              error: r
            });
            reject(r.data.error);
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
