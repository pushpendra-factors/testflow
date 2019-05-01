export default function reducer(state={
  dashboards: [],
  dashboardUnits: [],
}, action) {

  switch (action.type) { 
    case "FETCH_DASHBOARDS_FULFILLED": {
      return { ...state, dashboards: action.payload }
    }
    case "FETCH_DASHBOARD_UNITS_FULFILLED": {
      return {
        ...state,
        dashboardUnits: action.payload
      }
    }
  }
  return state
}
