export default function reducer(state={
  dashboards: [],
  units: [],
}, action) {

  switch (action.type) { 
    case "FETCH_DASHBOARDS_FULFILLED": {
      return { ...state, dashboards: action.payload }
    }
    case "FETCH_DASHBOARD_UNITS_FULFILLED": {
      return {
        ...state,
        units: action.payload
      }
    }
  }
  return state
}
