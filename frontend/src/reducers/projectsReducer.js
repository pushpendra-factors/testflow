export default function reducer(state={
    projects: [],
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
    }
    return state
}
