import { applyMiddleware, createStore } from "redux"
import logger from "redux-logger"
import thunk from "redux-thunk"
import promise from "redux-promise-middleware"
import reducer from "./reducers"

var middlewares = [
    promise(),
    thunk
];

if(process.env.NODE_ENV !== "production") middlewares.push(logger());
export default createStore(reducer, applyMiddleware(...middlewares))
