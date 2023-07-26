import {
  FeatureConfigActionType,
  FeatureConfigActions,
  FeatureConfigState
} from './types';

const defaultFeatureConfigState: FeatureConfigState = {
  activeFeatures: [],
  loading: true,
  error: false
};

export default function FeatureConfigReducer(
  state: FeatureConfigState = defaultFeatureConfigState,
  action: FeatureConfigActions
): FeatureConfigState {
  switch (action.type) {
    case FeatureConfigActionType.UPDATE_FEATURE_CONFIG:
      return {
        ...state,
        ...action.payload,
        loading: false
      };
    case FeatureConfigActionType.RESET_FEATURE_CONFIG:
      return {
        ...defaultFeatureConfigState,
        loading: false
      };
    case FeatureConfigActionType.SET_FEATURE_CONFIG_ERROR:
      return {
        ...defaultFeatureConfigState,
        error: true,
        loading: false
      };
    case FeatureConfigActionType.SET_FEATURE_CONFIG_LOADING:
      return {
        ...state,
        loading: true
      };
    default:
      return state;
  }
}
