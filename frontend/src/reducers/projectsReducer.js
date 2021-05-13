const DEFAULT_PROJECT_STATE = {
  projects: {},
  projectsError: null,
  currentProjectSettings: {},
  currentProjectEventNames: [],
  eventPropertiesMap: {},
  eventPropertyValuesMap: {},
  userProperties: [],
  userPropertyValuesMap: {},
  queryEventPropertiesMap: {},
  queryEventPropertyValuesMap:{},
  fetchingProjects: false,
  fetchedProjects: false,
  filters: [],
  filtersError: null,
  intervals: [],
  defaultModelInterval: null,
  projectAgents: [],
  agents: {},
  viewQuery: {},
  adwordsCustomerAccounts: null,
  channelFilterValues: {},
}

export default function reducer(state=DEFAULT_PROJECT_STATE, action) {
    switch (action.type) {      
      case "CHANGE_PROJECT": {
        return {
          ...DEFAULT_PROJECT_STATE, // reset store to default.
          currentProjectId: action.payload,
          projects: state.projects
        }
      }
      case "CREATE_PROJECT_FULFILLED" : {
        let _state = { ...state  };
        _state.projects = { ..._state.projects };
        _state.projects[action.payload.id] = action.payload        
        // Set currentProjectId to this newly created project        
        _state.currentProjectId = action.payload.id;        
        return _state;
      }
      case "FETCH_PROJECTS": {
        return {...state, fetchingProjects: true}
      }
      case "FETCH_PROJECTS_REJECTED": {
        return {...state, fetchingProjects: false, projectsError: action.payload}
      }
      case "FETCH_PROJECTS_FULFILLED": {
        // Indexed project objects by projectId. Kept projectId on value also intentionally 
        // for array of projects from Object.values().
        let projects = {};
        for (let project of action.payload.projects) {
          projects[project.id] = project;
        }

        // Initial project set.
        let currentProjectId = null;
        if (action.payload.projects.length > 0)
          currentProjectId = action.payload.projects[0].id;

        return {
          ...state,
          fetchingProjects: false,
          fetchedProjects: true,
          projects: projects,
          currentProjectId: currentProjectId,
        }
      }
      case "FETCH_PROJECT_SETTINGS_FULFILLED": {
        return {
          ...state,
          currentProjectSettings: action.payload.settings
        }
      }
      case "FETCH_PROJECT_SETTINGS_REJECTED": {
        return {
          ...state,
          projectSettingsError: action.payload.err
        }
      }
      case "UPDATE_PROJECT_SETTINGS_FULFILLED": {
        let _state = { ...state };
        if (_state.currentProjectSettings)
          _state.currentProjectSettings = { 
            ..._state.currentProjectSettings,
            ...action.payload.updatedSettings // Updates the state of settings only which are updated.
          };
        return _state;
      }
      case "UPDATE_PROJECT_SETTINGS_REJECTED": {
        return {
          ...state,
          projectEventsError: action.payload.err
        }
      }
      case "FETCH_PROJECT_EVENTS_FULFILLED": {
        return {
          ...state,
          currentProjectEventNames: action.payload.eventNames
        }
      }
      case "FETCH_PROJECT_EVENTS_REJECTED": {
        return {
          ...state,
          currentProjectEventNames: action.payload.eventNames,
          projectEventsError: action.payload.err,
        }
      }
      case "UPDATE_PROJECT_EVENTS_REJECTED":{
        return{
          ...state,
          projectEventsError:action.payload.err
        }
      }
      case "FETCH_PROJECT_EVENT_PROPERTIES_FULFILLED": {
        // Only the latest fetch is maintained.
        let eventPropertiesMap = {};
        eventPropertiesMap[action.payload.eventName] = action.payload.eventProperties;
        return {
          ...state,
          eventPropertiesMap: eventPropertiesMap,
        }
      }
      case "FETCH_PROJECT_EVENT_PROPERTIES_REJECTED": {
        return {...state,
                eventPropertiesError: action.payload.err}
      }
      case "FETCH_PROJECT_EVENT_PROPERTY_VALUES_FULFILLED": {
        // Only the latest fetch is maintained.
        var eventPropertyValuesMap = {};
        eventPropertyValuesMap[action.payload.eventName] = {}
        eventPropertyValuesMap[action.payload.eventName][
          action.payload.propertyName] = action.payload.eventPropertyValues;
        return {
          ...state,
          eventPropertyValuesMap: eventPropertyValuesMap,
        }
      }
      case "FETCH_PROJECT_EVENT_PROPERTY_VALUES_REJECTED": {
        return {...state,
                eventPropertyValuesError: action.payload.err}
      }

      case "FETCH_PROJECT_USER_PROPERTIES_FULFILLED": {
        return {...state,
                userProperties: action.payload.userProperties,
              }
      }
      case "FETCH_PROJECT_USER_PROPERTIES_REJECTED": {
        return {...state,
                userPropertiesError: action.payload.err}
      }
      case "FETCH_PROJECT_USER_PROPERTY_VALUES_FULFILLED": {
        // Only the latest fetch is maintained.
        var userPropertyValuesMap = {};
        userPropertyValuesMap[action.payload.propertyName] = action.payload.userPropertyValues;
        return {
          ...state,
          userPropertyValuesMap: userPropertyValuesMap,
        }
      }
      case "FETCH_PROJECT_USER_PROPERTY_VALUES_REJECTED": {
        return {...state,
                userPropertyValuesError: action.payload.err}
      }
      case "FETCH_FILTERS_FULFILLED": {
        return {
          ...state,
          filters:  Array.from(action.payload)
        }
      }
      case "FETCH_FILTERS_REJECTED": {
        return { 
          ...state,
          filtersError: {
            error: action.error
          }
        }
      }
      case "CREATE_FILTER_FULFILLED": {
        let _state = { ...state }
        // Note: state clone uses same ref of pref filters,
        // which won't trigger render.
        _state.filters = [...state.filters];
        _state.filters.push(action.payload);
        return _state
      }
      case "CREATE_FILTER_REJECTED": {
        // no redux state change.
        return state
      }
      case "UPDATE_FILTER_FULFILLED": {
        let _state = { ...state }
        _state.filters = [...state.filters];
        _state.filters[action.payload.storeIndex] = { 
          ..._state.filters[action.payload.storeIndex],
          ...action.payload.data,
        };
        return _state
      }
      case "UPDATE_FILTER_REJECTED": {
        // no redux state change.
        return state
      }
      case "DELETE_FILTER_FULFILLED": {
        let _state = { ...state };
        _state.filters = [...state.filters];
        // payload is the index ref to be deleted.
        _state.filters.splice(action.payload, 1);
        return _state
      }
      case "DELETE_FILTER_REJECTED": {
        // no redux state change.
        return state
      }
      case "FETCH_PROJECT_MODELS_FULFILLED": {
        let nextState = {
          ...state,
          intervals: action.payload
        }
        if (nextState.intervals.length > 0) {
          // default interval is set here.
          nextState.defaultModelInterval = nextState.intervals[0];
        } else {
          nextState.defaultModelInterval = null;
        }
        return nextState
      }
      case "FETCH_PROJECT_MODELS_REJECTED": {
        return {
          ...state,
          intervals: []
        }
      }
      case "FETCH_PROJECT_AGENTS_FULFILLED":{
        return {
          ...state,
          projectAgents: action.payload.project_agent_mappings,
          agents: action.payload.agents,
        }
      }
      case "FETCH_PROJECT_AGENTS_REJECTED":{
        return {
          ...state,
          projectAgents: {},
          agents: {},
        }
      }
      case "PROJECT_AGENT_INVITE_FULFILLED": {
        let nextState = { ...state };
        
        let projectAgentMapping = action.payload.project_agent_mappings[0];
        nextState.projectAgents = [...state.projectAgents];
        nextState.projectAgents.push(projectAgentMapping);
        nextState.agents[projectAgentMapping.agent_uuid] = action.payload.agents[projectAgentMapping.agent_uuid];        
        return nextState
      }
      case "PROJECT_AGENT_INVITE_REJECTED": {
        return {
          ...state
        }
      }
      case "PROJECT_AGENT_REMOVE_FULFILLED": {
        let nextState = { ...state };
        nextState.projectAgents = state.projectAgents.filter((projectAgent)=>{
          return projectAgent.agent_uuid != action.payload.agent_uuid
        })
        return nextState
      }
      case "VIEW_QUERY": {
        return {
          ...state,
          viewQuery: action.payload,
        }
      }
      case "FETCH_ADWORDS_CUSTOMER_ACCOUNTS_FULFILLED": {
        let _state = { ...state }
        _state.adwordsCustomerAccounts = [...action.payload.customer_accounts];
        return _state;
      }
      case "ENABLE_ADWORDS_FULFILLED": {
        let enabledAgentUUID = action.payload.int_adwords_enabled_agent_uuid;
        if (!enabledAgentUUID || enabledAgentUUID == "")
          return state;

        let _state = { ...state };
        _state.currentProjectSettings = {
          ...state.currentProjectSettings,
          int_adwords_enabled_agent_uuid: enabledAgentUUID,
        }
        return _state;
      }
      case "FETCH_GSC_CUSTOMER_ACCOUNTS_FULFILLED": {
        let _state = { ...state }
        _state.gscURLs = [...action.payload.urls];
        return _state;
      }
      case "ENABLE_SALESFORCE_FULFILLED": {
        let enabledAgentUUID = action.payload.int_salesforce_enabled_agent_uuid;
        if (!enabledAgentUUID || enabledAgentUUID == "")
          return state;

        let _state = { ...state };
        _state.currentProjectSettings = {
          ...state.currentProjectSettings,
          int_salesforce_enabled_agent_uuid: enabledAgentUUID,
        }
        return _state;
      }
      case "FETCH_CHANNEL_FILTER_VALUES_FULFILLED": {
        let _state = { ...state };
        if (!_state.channelFilterValues[action.payload.channel]) {
          _state.channelFilterValues[action.payload.channel] = {};
        }
        _state.channelFilterValues[action.payload.channel][action.payload.filter] = action.payload.values;

        return _state;
      }
      case "ENABLE_FACEBOOK_USER_ID": {
        let fbUserID = action.payload.int_facebook_user_id;

        let _state = {...state};
        _state.currentProjectSettings = {
          ...state.currentProjectSettings,
          int_facebook_user_id: fbUserID,
        }
        return _state;
      }
      case "ENABLE_LINKEDIN_AD_ACCOUNT": {
        let linkedinAdAccount = action.payload.int_linkedin_ad_account;
        let linkedinAccessToken = action.payload.int_linkedin_access_token
        let _state = {...state};
        _state.currentProjectSettings = {
          ...state.currentProjectSettings,
          int_linkedin_ad_account: linkedinAdAccount,
          int_linkedin_access_token: linkedinAccessToken,
        }
        return _state;
      }
    }

    return state
}
