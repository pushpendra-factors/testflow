import { Dispatch } from 'redux';
import { changePlanType, getFeatureConfigData } from './services';
import { FeatureConfigActionType, FeatureConfigApiResponse } from './types';
import logger from 'Utils/logger';

export const fetchFeatureConfig = (projectId: string) => {
  return async (dispatch: Dispatch) => {
    try {
      dispatch({ type: FeatureConfigActionType.SET_FEATURE_CONFIG_LOADING });

      const response = (await getFeatureConfigData(
        projectId
      )) as FeatureConfigApiResponse;
      if (response?.data) {
        dispatch({
          type: FeatureConfigActionType.UPDATE_FEATURE_CONFIG,
          payload: {
            activeFeatures: response?.data?.plan?.feature_list,
            addOns: response?.data?.add_ons,
            lastRenewedOn: response?.data?.last_renewed_on,
            plan: {
              id: response?.data?.plan?.id,
              name: response?.data?.plan?.name,
              display_name: response?.data?.display_name
            },
            sixSignalInfo: response?.data?.six_signal_info
          }
        });
      }
    } catch (error) {
      logger.error('Error in fetching feature config', error);
      dispatch({ type: FeatureConfigActionType.SET_FEATURE_CONFIG_ERROR });
    }
  };
};

export const updatePlan = (projectId: string, planName: string) => {
  return async (dispatch: Dispatch) => {
    try {
      dispatch({ type: FeatureConfigActionType.SET_FEATURE_CONFIG_LOADING });

      const response = (await changePlanType(
        projectId,
        planName
      )) as FeatureConfigApiResponse;
      return response;
    } catch (error) {
      logger.error('Error in fetching feature config', error);
      dispatch({ type: FeatureConfigActionType.SET_FEATURE_CONFIG_ERROR });
    }
  };
};
