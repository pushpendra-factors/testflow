import { FUNNEL_RESULTS_AVAILABLE, FUNNEL_RESULTS_UNAVAILABLE } from "./types"

const defaultState = {
    is_funnel_results_visible: false,
    funnel_events: []
}

export default function (state = defaultState, action) {
    switch (action.type) {
        case FUNNEL_RESULTS_AVAILABLE:
            return { ...state, is_funnel_results_visible: true, funnel_events: action.payload }
        case FUNNEL_RESULTS_UNAVAILABLE:
            return { ...state, is_funnel_results_visible: false, funnel_events: [] }
        default:
            return state
    }
}