import { defaultState } from "./constants";

const { TOGGLE_GLOBAL_SEARCH } = require("Reducers/types");

export default function(state=defaultState, action){
    switch(action.type){
        case TOGGLE_GLOBAL_SEARCH : return {...state, visible: !state.visible};
        default: return state;
    }
}