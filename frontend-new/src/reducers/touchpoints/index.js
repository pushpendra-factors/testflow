import { OTPService } from "./services";

//Actions
const FETCH_OTP_LIST = 'FETCH_OTP_LIST';


// State
const initialState = {
    touchpoints: [],
    error: false
};

export default function (state = initialState, action) {
    switch (action.type) {
        case FETCH_OTP_LIST:
            return { ...state, touchpoints: action.payload };
        default:
            return state;
    }
}

export const getTouchPoints = (projectId, payload) => (dispatch) => {
    return new Promise((resolve) => {
      getTouchPoints(projectId, payload)
        .then((response) => {
          resolve(
            dispatch({
              type: 'FETCH_OTP_DATA',
              payload: data
            })
          );
        })
    });
  };

