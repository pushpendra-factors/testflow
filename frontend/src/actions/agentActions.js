import { getHostURL } from "../util";
import {get, post} from "./request.js";
var host = getHostURL();

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
              error: r
            })
            reject({body: r.data, status: r.status});
          });
      })
    }
  }

  export function signout(){
    return function(dispatch){
      return new Promise((resolve, reject)=>{
        get(dispatch, host + "agents/signout").then((response)=> {
          resolve(dispatch({
            type: "AGENT_LOGOUT_FULFILLED",
          }));
        });
      })
    }
  }

  export function signup(email){
    return function(dispatch){
      return new Promise((resolve, reject) => {
        dispatch({type: "AGENT_SIGNUP"});
        post(dispatch, host+"accounts/signup",{email: email})
        .then((response) => {
          resolve(dispatch({
            type: "AGENT_SIGNUP_FULFILLED",
            payload: {}
          }));
        })
        .catch((err)=>{
          reject(dispatch({type: "AGENT_SIGNUP_REJECTED", payload: err}));
        })
      });
    }
  }

  export function verify(firstName, lastName, password, token){
    return function(dispatch){
      return new Promise((resolve, reject) => {
        dispatch({type: "AGENT_VERIFY"});
        post(dispatch, host+"agents/verify?token="+token,{
          first_name: firstName,
          last_name:lastName,
          password: password}
        )
        .then((response) => {
          resolve(dispatch({
            type: "AGENT_VERIFY_FULFILLED",
            payload: {}
          }));
        })
        .catch((err)=>{
          reject(dispatch({type: "AGENT_VERIFY_REJECTED", payload: err}));
        })
      });
    }
  }