export default function reducer(state={
    projects: [],
    eventPropertiesMap: {},
    eventPropertyValuesMap: {},
    fetchingProjects: false,
    fetchedProjects: false,
    projectsError: null,
  }, action) {

    switch (action.type) {
      case "FETCH_PROJECTS": {
        return {...state, fetchingProjects: true}
      }
      case "FETCH_PROJECTS_REJECTED": {
        return {...state, fetchingProjects: false, projectsError: action.payload}
      }
      case "FETCH_PROJECTS_FULFILLED": {
        return {
          ...state,
          fetchingProjects: false,
          fetchedProjects: true,
          projects: action.payload,
        }
      }
      case "FETCH_CURRENT_PROJECT_EVENTS_FULFILLED": {
        return {...state,
                currentProject: action.payload.currentProject,
                currentProjectEventNames: action.payload.currentProjectEventNames
              }
      }
      case "FETCH_CURRENT_PROJECT_EVENTS_REJECTED": {
        return {...state,
                currentProject: action.payload.currentProject,
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
        return {...state,
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
