import { SET_ACTIVE_PROJECT } from 'Reducers/types';
import {
  CREATE_SMART_PROPERTY,
  FETCH_CLICKABLE_ELEMENTS,
  FETCH_PROPERTY_CONFIG,
  FETCH_SMART_PROPERTIES,
  TOGGLE_CLICKABLE_ELEMENT,
  UPDATE_SMART_PROPERTY,
  FETCH_PROPERTY_MAPPING
} from './actions';

const defaultState = {
  smartProperties: [],
  propertyConfig: {},
  clickableElements: []
};

export default function (state = defaultState, action) {
  switch (action.type) {
    case FETCH_SMART_PROPERTIES:
      return { ...state, smartProperties: action.payload };
    case FETCH_PROPERTY_MAPPING:
      return { ...state, propertyMapping: action.payload };
    case FETCH_PROPERTY_CONFIG:
      return { ...state, propertyConfig: action.payload };
    case CREATE_SMART_PROPERTY:
      const props = [...state.smartProperties];
      props.push(action.payload);
      return { ...state, propertyConfig: props };
    case UPDATE_SMART_PROPERTY:
      const propsToUpdate = [
        ...state.smartProperties.map((prop, i) => {
          if (prop.id === action.payload.id) {
            return action.payload;
          } else {
            return action.payload;
          }
        })
      ];
      return { ...state, propertyConfig: propsToUpdate };
    case FETCH_CLICKABLE_ELEMENTS:
      return { ...state, clickableElements: action.payload };
    case TOGGLE_CLICKABLE_ELEMENT:
      let nextState = { ...state };
      for (let i = 0; i < nextState.clickableElements.length; i++) {
        if (nextState.clickableElements[i].id == action.payload.id) {
          nextState.clickableElements[i].enabled = action.payload.enabled;
        }
      }
      return nextState;
    case SET_ACTIVE_PROJECT:
      return {
        ...defaultState
      };
    default:
      return state;
  }
}
