import logger from 'Utils/logger';
import { getHostUrl, post } from 'Utils/request';
import { FactorsDeAnonymisationProvider } from '../ui/types';

const host = getHostUrl();

export const setFactorsDeAnonymisationProvider = async (
  projectId: string,
  provider: FactorsDeAnonymisationProvider
) => {
  try {
    if (!projectId || !provider) {
      throw new Error('Invalid parameters passed');
    }
    const url = `${host}projects/${projectId}/factors_deanon/provider/${provider}/enable`;
    return post(null, url);
  } catch (error) {
    logger.error(error);
    return null;
  }
};
