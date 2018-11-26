export default function reducer(state={
    projects: [],
    eventPropertiesMap: {},
    eventPropertyValuesMap: {},
    fetchingProjects: false,
    fetchedProjects: false,
    projectsError: null,
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
    }
    return state
}
