export default function reducer(state={}, action) {

    switch (action.type) {
      case "FETCH_PROJECT_DASHBOARDS_FULFILLED": {
        return { ...state, dashboards: action.payload }
      }
    }
    return state
}
