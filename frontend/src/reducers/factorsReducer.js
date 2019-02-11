const DEFAULT_FACTORS_STATE = {
  factors: {},
  fetchingFactors: false,
  fetchedFactors: false,
  factorsError: null,
}

export default function reducer(state=DEFAULT_FACTORS_STATE, action) {
    switch (action.type) {
      case "FETCH_FACTORS": {
        return {...state, fetchingFactors: true}
      }
      case "FETCH_FACTORS_REJECTED": {
        return {...state, fetchingFactors: false, projectsError: action.payload}
      }
      case "FETCH_FACTORS_FULFILLED": {
        return {
          ...state,
          fetchingFactors: false,
          fetchedFactors: true,
          factors: action.payload,
        }
      }
      case "RESET_FACTORS": {
        return {
          ...DEFAULT_FACTORS_STATE
        };
      }
    }
    return state
}
