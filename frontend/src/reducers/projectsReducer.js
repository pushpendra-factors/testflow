export default function reducer(state={
    projects: [],
    eventPropertiesMap: {},
    eventPropertyValuesMap: {},
    userProperties: [],
    userPropertyValuesMap: {},
    fetchingProjects: false,
    fetchedProjects: false,
    projectsError: null,
    filters: [],
    filtersError: null,
    intervals: [],
    defaultModelInterval: null
  }, action) {

    switch (action.type) {
      case "CHANGE_PROJECT": {
        return {...state, currentProjectId: action.payload }
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
        for (let project of action.payload) {
          projects[project.id] = project;
        }

        // Initial project set.
        let currentProjectId = null;
        if (action.payload.length > 0)
          currentProjectId = action.payload[0].id;

        return {
          ...state,
          fetchingProjects: false,
          fetchedProjects: true,
          projects: projects,
          currentProjectId: currentProjectId
        }
      }
      case "FETCH_CURRENT_PROJECT_SETTINGS_FULFILLED": {
        return {
          ...state,
          currentProjectSettings: action.payload.settings
        }
      }
      case "FETCH_CURRENT_PROJECT_SETTINGS_REJECTED": {
        return {
          ...state,
          projectSettingsError: action.payload.err
        }
      }
      case "UPDATE_CURRENT_PROJECT_SETTINGS_FULFILLED": {
        let _state = { ...state };
        if (_state.currentProjectSettings)
          _state.currentProjectSettings = { 
            ..._state.currentProjectSettings,
            ...action.payload.updatedSettings // Updates the state of settings only which are updated.
          };
        return _state;
      }
      case "UPDATE_CURRENT_PROJECT_SETTINGS_REJECTED": {
        return {
          ...state,
          projectEventsError: action.payload.err
        }
      }
      case "FETCH_CURRENT_PROJECT_EVENTS_FULFILLED": {
        return {...state,
                currentProjectId: action.payload.currentProjectId,
                currentProjectEventNames: action.payload.currentProjectEventNames
              }
      }
      case "FETCH_CURRENT_PROJECT_EVENTS_REJECTED": {
        return {...state,
                currentProjectId: action.payload.currentProjectId,
                currentProjectEventNames: action.payload.currentProjectEventNames,
                projectEventsError: action.payload.err}
      }
      case "FETCH_CURRENT_PROJECT_EVENT_PROPERTIES_FULFILLED": {
        // Only the latest fetch is maintained.
        var eventPropertiesMap = {};
        eventPropertiesMap[action.payload.eventName] = action.payload.eventProperties;
        return {...state,
                eventPropertiesMap: eventPropertiesMap,
              }
      }
      case "FETCH_CURRENT_PROJECT_EVENT_PROPERTIES_REJECTED": {
        return {...state,
                eventPropertiesError: action.payload.err}
      }
      case "FETCH_CURRENT_PROJECT_EVENT_PROPERTY_VALUES_FULFILLED": {
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
      case "FETCH_CURRENT_PROJECT_EVENT_PROPERTY_VALUES_REJECTED": {
        return {...state,
                eventPropertyValuesError: action.payload.err}
      }

      case "FETCH_CURRENT_PROJECT_USER_PROPERTIES_FULFILLED": {
        return {...state,
                userProperties: action.payload.userProperties,
              }
      }
      case "FETCH_CURRENT_PROJECT_USER_PROPERTIES_REJECTED": {
        return {...state,
                userPropertiesError: action.payload.err}
      }
      case "FETCH_CURRENT_PROJECT_USER_PROPERTY_VALUES_FULFILLED": {
        // Only the latest fetch is maintained.
        var userPropertyValuesMap = {};
        userPropertyValuesMap[action.payload.propertyName] = action.payload.userPropertyValues;
        return {
          ...state,
          userPropertyValuesMap: userPropertyValuesMap,
        }
      }
      case "FETCH_CURRENT_PROJECT_USER_PROPERTY_VALUES_REJECTED": {
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
        return {
          ...state,
          intervals: action.payload.intervals,
          defaultModelInterval: action.payload.default_interval,
        }
      }
    }
    return state
}
