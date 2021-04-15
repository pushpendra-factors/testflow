import { CREATE_SMART_PROPERTY, FETCH_PROPERTY_CONFIG, FETCH_SMART_PROPERTIES, UPDATE_SMART_PROPERTY } from "./actions";

const defaultState = {
    smartProperties: [],
    propertyConfig: {}
}

export default function (state = defaultState, action) { 
    switch (action.type) {
        case FETCH_SMART_PROPERTIES:
          return { ...state, smartProperties: action.payload };
        case FETCH_PROPERTY_CONFIG:
            return { ...state, propertyConfig: action.payload };
        case CREATE_SMART_PROPERTY:
            const props = [...state.smartProperties];
            props.push(action.payload);
            return { ...state, propertyConfig: props}
        case UPDATE_SMART_PROPERTY:
            const propsToUpdate = [...state.smartProperties.map((prop, i) => {
                if(prop.id === action.payload.id) {
                    return action.payload;
                } else {
                    return action.payload;
                }
            })];
            return { ...state, propertyConfig: propsToUpdate}
        default:
            return state;
    }
}