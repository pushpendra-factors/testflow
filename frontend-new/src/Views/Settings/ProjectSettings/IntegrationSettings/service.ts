import logger from 'Utils/logger';
import { getHostUrl, get } from 'Utils/request';

const host = getHostUrl();

export const getIntegrationStatus = async (projectId: string) => {
  try {
    if (!projectId) {
      throw new Error('Invalid parameters passed');
    }
    const url = `${host}projects/${projectId}/integrations_status`;
    return get(null, url);
  } catch (error) {
    logger.error(error);
    return null;
  }
};
