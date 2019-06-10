export default function reducer(state={
    agent: {},
    agentError: null,
    isLoggedIn: true,
    billing: {},
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
      case "UPDATE_AGENT_INFO_FULFILLED": {
        return { ...state, agent: action.payload }
      }
      case "UPDATE_AGENT_INFO_REJECTED": {
        return { ...state, agentError: action.payload }
      }
      case "UPDATE_AGENT_PASSWORD_FULFILLED":{
        return state
      }
      case "FETCH_AGENT_BILLING_ACCOUNT_FULFILLED":{
        let billing = {
          billingAccount: action.payload.billing_account,
          projects: action.payload.projects,
          accountAgents: action.payload.agents,
          plan: action.payload.plan,
          availablePlans: action.payload.available_plans
        }
        return {...state, billing:billing}
      }
      case "UPDATE_AGENT_BILLING_ACCOUNT_FULFILLED": {
        let _state = { ...state  };
        let billing = {..._state.billing};
        billing.billingAccount = action.payload.billing_account;
        billing.plan = action.payload.plan;
        return {...state, billing:billing}
      }
    }
    return state
}
