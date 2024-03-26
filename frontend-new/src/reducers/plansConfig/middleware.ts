import { Dispatch } from 'redux';
import logger from 'Utils/logger';
import { ADDITIONAL_ACCOUNTS_ADDON_ID } from 'Constants/plans.constants';
import {
  getDifferentialPricing,
  getPlansDetails,
  getSubscriptionDetails
} from './services';
import {
  AddonsStateInterface,
  DifferentialPricingAPIResponse,
  PlansConfigActionType,
  PlansDetailAPIResponse,
  PlansDetailStateInterface,
  SubscriptionDetailsAPIResponse
} from './types';

export const fetchPlansDetail =
  (projectId: string) => async (dispatch: Dispatch) => {
    try {
      dispatch({ type: PlansConfigActionType.SET_PLANS_DETAIL_LOADING });

      const response = (await getPlansDetails(
        projectId
      )) as PlansDetailAPIResponse;
      if (response?.data) {
        const addons = response.data.filter((data) => data.type === 'addon');
        const plans = response.data.filter((data) => data.type === 'plan');
        let stateAddons: AddonsStateInterface[] = [];
        let statePlans: PlansDetailStateInterface[] = [];

        if (addons && addons.length) {
          stateAddons = addons.map((addon) => ({
            name: addon.name,
            id: addon.id,
            price: addon.price
          }));
        }
        if (plans && plans.length) {
          // collecting monthly and yearly prices of all plans
          const reducerPlanObj: { [key: string]: PlansDetailStateInterface } =
            {};
          plans.forEach((plan) => {
            if (plan?.external_name) {
              if (!reducerPlanObj?.[plan.external_name]) {
                const obj: PlansDetailStateInterface = {
                  name: plan.external_name,
                  terms: [
                    {
                      name: plan.name,
                      period: plan.period_unit,
                      id: plan.id,
                      price: plan.price
                    }
                  ]
                };

                reducerPlanObj[plan.external_name] = obj;
              } else if (reducerPlanObj[plan.external_name]) {
                reducerPlanObj[plan.external_name]?.terms?.push({
                  name: plan.name,
                  period: plan.period_unit,
                  id: plan.id,
                  price: plan.price
                });
              }
            }
          });
          statePlans = Object.values(reducerPlanObj) || [];
        }

        dispatch({
          type: PlansConfigActionType.SET_PLANS_CONFIG_DETAILS,
          payload: {
            plansDetail: statePlans,
            addOnsDetail: stateAddons
          }
        });
      }
    } catch (error) {
      logger.error('Error in fetching feature config', error);
      dispatch({ type: PlansConfigActionType.SET_PLANS_DETAIL_ERROR });
    }
  };

export const fetchCurrentSubscriptionDetail =
  (projectId: string) => async (dispatch: Dispatch) => {
    try {
      dispatch({ type: PlansConfigActionType.SET_CURRENT_PLAN_LOADING });

      const response = (await getSubscriptionDetails(
        projectId
      )) as SubscriptionDetailsAPIResponse;
      if (response?.data) {
        const addons = response.data?.subscription_details?.filter(
          (data) => data.type === 'addon'
        );
        const plans = response.data?.subscription_details?.filter(
          (data) => data.type === 'plan'
        );
        const stateCurrentPlanConfig = {
          renews_on: response.data.renews_on,
          status: response.data.status,
          period: response.data?.period_unit,
          billingPeriod: response.data?.billing_period
        };
        if (plans && plans?.length > 0) {
          const firstPlan = plans[0];
          const externalName = firstPlan?.id?.split('-')?.[0] || '';
          stateCurrentPlanConfig.plan = {
            id: firstPlan.id || '',
            amount: firstPlan.amount || '',
            name: firstPlan.id || '',
            externalName,
            quantity: firstPlan.quantity
          };
        }
        if (addons && addons?.length > 0) {
          stateCurrentPlanConfig.addons = addons.map((addon) => ({
            id: addon?.id,
            amount: addon?.amount,
            quantity: addon?.quantity
          }));
        }
        dispatch({
          type: PlansConfigActionType.SET_CURRENT_PLAN_DETAILS,
          payload: stateCurrentPlanConfig
        });
      }
    } catch (error) {
      logger.error('Error in fetching feature config', error);
      dispatch({ type: PlansConfigActionType.SET_CURRENT_PLAN_ERROR });
    }
  };

export const fetchDifferentialPricing =
  (projectId: string) => async (dispatch: Dispatch) => {
    try {
      dispatch({
        type: PlansConfigActionType.SET_DIFFERENTIAL_PRICING_LOADING
      });
      const response = (await getDifferentialPricing(
        projectId
      )) as DifferentialPricingAPIResponse;
      if (response?.data) {
        const filteredData = response?.data.filter((data) => {
          if (data?.item_price_id === ADDITIONAL_ACCOUNTS_ADDON_ID) return true;
          return false;
        });
        dispatch({
          type: PlansConfigActionType.SET_DIFFERENTIAL_PRICING_DETAILS,
          payload: filteredData
        });
      }
    } catch (error) {
      logger.error('Error in fetching feature config', error);
      dispatch({ type: PlansConfigActionType.SET_DIFFERENTIAL_PRICING_ERROR });
    }
  };
