import { getHostURL } from "../util";
import {get, post} from "./request.js";
var host = getHostURL();

export function fetchAgentInfo(){
  return function(dispatch) {
    return new Promise((resolve,reject) => {
      get(dispatch, host + "agents/info")
        .then((response) => {        
          resolve(dispatch({type:"FETCH_AGENT_INFO_FULFILLED", payload: response.data}));
        })
        .catch(() => {       
          reject(dispatch({
            type:"FETCH_AGENT_INFO_REJECTED", 
            payload: 'Failed to fetch agent info',
          }));
        });
    });
  }
}

export function login(email, password) {
    return function(dispatch) {
      return new Promise((resolve, reject) => {
        post(dispatch, host + "agents/signin", {
          "email":email,
          "password":password,
        })       
        .then((r) => {
            dispatch({
              type: "AGENT_LOGIN_FULFILLED",
              payload: r.data
            });

            resolve(r.data);
          })
          .catch((r) => {
            dispatch({
              type: "AGENT_LOGIN_REJECTED",
              payload: null
            });

            if(r.status && r.status == 401) reject("Invalid email or password");
            else reject("Login failed. Please try again.");
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

  export function signup(email){
    return function(dispatch){
      return new Promise((resolve, reject) => {
        dispatch({type: "AGENT_SIGNUP"});

        post(dispatch, host+"accounts/signup", { email: email })
          .then(() => {
            resolve(dispatch({
              type: "AGENT_SIGNUP_FULFILLED",
              payload: {}
            }));
          })
          .catch(() => {
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

  