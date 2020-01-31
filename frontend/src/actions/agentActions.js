import { getHostURL, getAdwordsHostURL } from "../util";
import {get, post, put} from "./request.js";
var host = getHostURL();

export function setLoginToken(token="") {
  return function(dispatch) {
    if (token == "") return;
    window.FACTORS_AI_LOGIN_TOKEN = token;
    dispatch({ type: "AGENT_LOGIN_FULFILLED" });
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

export function updateAgentPassword(params){
  return function(dispatch){
    return new Promise((resolve, reject)=> {
      put(dispatch, host + "agents/updatepassword", params)
      .then((response) => {
        dispatch({type:"UPDATE_AGENT_PASSWORD_FULFILLED", 
            payload: response.data});
          resolve(response);
      })
      .catch((err) => {
        dispatch({type:"UPDATE_AGENT_PASSWORD_REJECTED", 
            payload: 'Failed to update agent password'});
          reject(err);
      })
    });
  }
}



export function login(email, password) {
    return function(dispatch) {
      return new Promise((resolve, reject) => {
        let invalidMsg = "Invalid email or password";
        let loginFailMsg = "Login failed. Please try again.";

        post(dispatch, host + "agents/signin", {
          "email":email,
          "password":password,
        })       
        .then((r) => {
          if (r.ok) {
            dispatch({
              type: "AGENT_LOGIN_FULFILLED",
              payload: r.data
            });

            resolve(r.data);
          } else {
            dispatch({
              type: "AGENT_LOGIN_REJECTED",
              payload: null
            });

            if (r.status == 404) reject(invalidMsg);
            else reject(loginFailMsg);
          }
        })
        .catch((r) => {
          dispatch({
            type: "AGENT_LOGIN_REJECTED",
            payload: null
          });

          if(r.status && r.status == 401) reject(invalidMsg);
          else reject(loginFailMsg);
        });
      })
    }
  }

  export function signout(){
    return function(dispatch){
      return new Promise((resolve, reject) => {
        get(dispatch, host + "agents/signout")
          .then(() => {
            resolve(dispatch({
              type: "AGENT_LOGOUT_FULFILLED",
            }));
          })
          .catch(() => {
            reject("Sign out failed");
          });
      })
    }
  }

  export function signup(email, phone, planCode){
    return function(dispatch){
      return new Promise((resolve, reject) => {
        dispatch({type: "AGENT_SIGNUP"});

        post(dispatch, host+"accounts/signup", { email: email, phone: phone, plan_code: planCode })
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

  export function activate(firstName, lastName, password, token){
    return function(dispatch){
      return new Promise((resolve, reject) => {
        dispatch({type: "AGENT_VERIFY"});

        post(dispatch, host+"agents/activate?token="+token, {
          first_name: firstName,
          last_name:lastName,
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

export function fetchAgentBillingAccount(){
  return function(dispatch){
    return get(dispatch, host + "agents/billing")
      .then((r) => {
        dispatch({type: "FETCH_AGENT_BILLING_ACCOUNT_FULFILLED", payload: r.data });
      })
      .catch((r) => {
        if (r.status) {
          // use this pattern for error handling. 
          // decided to use redux store.
          dispatch({type: "FETCH_AGENT_BILLING_ACCOUNT_REJECTED", payload: r.data, code: r.status });        
        } else {
          // network error. Idea: Use a global error component for this.
          console.log("network error");
        }
      });
  }
}

export function updateBillingAccount(params){
  return function(dispatch){
    return put(dispatch, host + "agents/billing", params)
      .then((r) => {
        dispatch({type: "UPDATE_AGENT_BILLING_ACCOUNT_FULFILLED", payload: r.data });
      })
      .catch((r) => {
        if (r.status) {
          // use this pattern for error handling. 
          // decided to use redux store.
          dispatch({type: "UPDATE_AGENT_BILLING_ACCOUNT_REJECTED", payload: r.data, code: r.status });        
        } else {
          // network error. Idea: Use a global error component for this.
          console.log("network error");
        }
      });
  }
}