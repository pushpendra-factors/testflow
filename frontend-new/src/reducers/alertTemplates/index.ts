import { SET_ALERT_TEMPLATES } from "Reducers/types";
import { get, getHostUrl } from "Utils/request";
import { Dispatch } from "redux";

export const defaultState = {
  isLoaded: false,
  isLoading: false,
  data: []
}
const host = getHostUrl();

export function fetchAlertTemplates(){
  return async function (dispatch: Dispatch<any>) {
  
      const url = host + 'common/alert_templates';
      const res = await get(null, url);
      dispatch({ type: SET_ALERT_TEMPLATES, payload: res.data });
  }
}

export default function (state = defaultState, action: {type: string, payload: any}) {
  switch (action.type) {
    case SET_ALERT_TEMPLATES:
      return {  isLoaded: true, data: action.payload };
    default:
      return state;
  }
}
