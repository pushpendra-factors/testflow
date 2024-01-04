import {
  PlansConfigActionType,
  PlansConfigActions,
  PlansConfigState
} from './types';

const defaultFeatureConfigState: PlansConfigState = {
  plansConfig: {
    plansDetail: [],
    addOnsDetail: [],
    loading: false,
    error: false
  },
  currentPlanDetail: {
    loading: false,
    error: false
  },
  differentialPricing: {
    loading: false,
    error: false,
    data: []
  }
};

export default function PlansConfigReducer(
  state: PlansConfigState = defaultFeatureConfigState,
  action: PlansConfigActions
): PlansConfigState {
  switch (action.type) {
    case PlansConfigActionType.SET_PLANS_DETAIL_LOADING:
      return {
        ...state,
        plansConfig: { ...state.plansConfig, loading: true }
      };
    case PlansConfigActionType.SET_PLANS_DETAIL_ERROR:
      return {
        ...state,
        plansConfig: { ...state.plansConfig, error: true }
      };
    case PlansConfigActionType.SET_PLANS_CONFIG_DETAILS:
      return {
        ...state,
        plansConfig: {
          loading: false,
          error: false,
          plansDetail: action.payload.plansDetail,
          addOnsDetail: action.payload.addOnsDetail
        }
      };
    case PlansConfigActionType.SET_CURRENT_PLAN_DETAILS:
      return {
        ...state,
        currentPlanDetail: {
          loading: false,
          error: false,
          ...action.payload
        }
      };
    case PlansConfigActionType.SET_CURRENT_PLAN_ERROR:
      return {
        ...state,
        currentPlanDetail: {
          ...state.currentPlanDetail,
          error: true
        }
      };

    case PlansConfigActionType.SET_CURRENT_PLAN_LOADING:
      return {
        ...state,
        currentPlanDetail: {
          ...state.currentPlanDetail,
          loading: true
        }
      };

    case PlansConfigActionType.SET_DIFFERENTIAL_PRICING_DETAILS:
      return {
        ...state,
        differentialPricing: {
          loading: false,
          error: false,
          data: action.payload
        }
      };
    case PlansConfigActionType.SET_DIFFERENTIAL_PRICING_ERROR:
      return {
        ...state,
        differentialPricing: {
          ...state.differentialPricing,
          error: true
        }
      };

    case PlansConfigActionType.SET_DIFFERENTIAL_PRICING_LOADING:
      return {
        ...state,
        differentialPricing: {
          ...state.differentialPricing,
          loading: true
        }
      };

    default:
      return state;
  }
}
