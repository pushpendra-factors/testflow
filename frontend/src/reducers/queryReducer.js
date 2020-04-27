const DEFAULT_QUERY_STATE = {
    data: null,
    topError: null
  }

export default function reducer(state=DEFAULT_QUERY_STATE, action){
    switch (action.type) {
        case "FETCH_DUMMY_QUERY_FULFILLED": {
          return {...state, data: action.payload}
        }
        case "FETCH_DUMMY_QUERY_REJECTED": {
          return {...state,topError: action.payload.error}
        }
      }
      return state
}