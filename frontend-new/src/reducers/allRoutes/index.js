import { defaultState } from "./constants";

const { UPDATE_ALL_ROUTES} = require("Reducers/types");

export default function(state=defaultState, action){
    switch(action.type){
        case UPDATE_ALL_ROUTES : return {...state, data: new Set([...state.data,...action.payload])};
        default: return state;
    }
}