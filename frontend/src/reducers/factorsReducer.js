export default function reducer(state={
    factors: {},
    fetchingFactors: false,
    fetchedFactors: false,
    factorsError: null,
  }, action) {

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
    }
    return state
}
