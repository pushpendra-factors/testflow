import { combineReducers } from "redux"

import dashboards from "./dashboardsReducer";
import factors from "./factorsReducer";
import projects from "./projectsReducer";
import agents from "./agentsReducer";
import reports from "./reportsReducer";
import query from "./queryReducer";

export default combineReducers({
  dashboards,
  projects,
  factors,
  agents,
  reports,
  query
})
