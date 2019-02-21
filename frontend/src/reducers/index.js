import { combineReducers } from "redux"

import factors from "./factorsReducer"
import projects from "./projectsReducer"
import agents from "./agentsReducer"


export default combineReducers({
  projects,
  factors,
  agents,
})
