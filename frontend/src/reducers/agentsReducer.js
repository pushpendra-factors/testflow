export default function reducer(state={
    agent: {},
    agentError: null,
    isLoggedIn: true
  }, action) {

    switch (action.type) {
      case "AGENT_LOGIN_FULFILLED": {
        return { ...state, isLoggedIn: true }
      }
      case "AGENT_LOGIN_REJECTED": {
        return { ...state, isLoggedIn: false }
      }
      case "AGENT_LOGOUT_FULFILLED": {
        return { ...state, isLoggedIn: false }
      }
      case "AGENT_LOGIN_FORCE": {
        return { ...state, isLoggedIn: false }
      }
      case "FETCH_AGENT_INFO_FULFILLED": {
        return { ...state, agent: action.payload }
      }
      case "FETCH_AGENT_INFO_REJECTED": {
        return { ...state, agentError: action.payload }
      }
    }
    return state
}
