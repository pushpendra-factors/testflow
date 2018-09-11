import { combineReducers } from "redux"

import factors from "./factorsReducer"
import projects from "./projectsReducer"


export default combineReducers({
  projects,
  factors,
})
