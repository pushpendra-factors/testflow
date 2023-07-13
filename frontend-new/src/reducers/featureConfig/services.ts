import logger from 'Utils/logger';
import { getHostUrl, get, post } from 'Utils/request';

const host = getHostUrl();
export const getFeatureConfigData = async (projectId: string) => {
  try {
    if (!projectId) {
      throw new Error('Invalid parameters passed');
    }
    const url = `${host}projects/${projectId}/v1/features`;
    return get(null, url);
  } catch (error) {
    logger.error(error);
    return null;
  }
};

export const changePlanType = async (projectId: string, planType: string) => {
  try {
    if (!projectId) {
      throw new Error('Invalid parameters passed');
    }
    const url = `${host}projects/${projectId}/v1/plan`;
    return post(null, url, {
      plan_type: planType
    });
  } catch (error) {
    logger.error(error);
    return null;
  }
};

export const updatePlanConfig = async (
  projectId: string,
  accountLimit: number,
  mtuLimit: number,
  activatedFeatures: string[]
) => {
  try {
    if (!projectId || !accountLimit) {
      throw new Error('Invalid parameters passed');
    }
    const url = `${host}projects/${projectId}/v1/features/update`;
    const reqBody = {
      account_limit: accountLimit,
      // mtu_limit: mtuLimit,
      activated_features: activatedFeatures
    };
    return post(null, url, reqBody);
  } catch (error) {}
};
