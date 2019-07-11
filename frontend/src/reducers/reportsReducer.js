export default function reducer(state={
    reports_list: [],
    report: null
  }, action) {
  
    switch (action.type) { 
      case "FETCH_REPORTS_FULFILLED": {
        return { ...state, reports_list: action.payload.reports }
      }
      case "FETCH_REPORTS_REJECTED": {
        return { ...state, reports_list: action.payload.reports }
      }
      case "FETCH_REPORT_FULFILLED": {
        return { ...state, report: action.payload.report }
      }
      case "FETCH_REPORT_REJECTED": {
        return { ...state, report: action.payload.report }
      }
    }
    return state
}