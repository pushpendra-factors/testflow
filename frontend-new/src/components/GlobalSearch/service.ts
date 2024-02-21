import logger from 'Utils/logger';
import { getHostUrl, post } from 'Utils/request';

const host = getHostUrl();

export const getQueryFromTextPrompt = async (
  projectId: string,
  prompt: string,
  kpiConfig: any
) => {
  try {
    if (!projectId) {
      throw new Error('Invalid parameters passed');
    }
    const url = `${host}chat`;
    return post(null, url, { prompt, pid: projectId, kpi_config: kpiConfig });
  } catch (error) {
    logger.error(error);
    return null;
  }
};

export interface TextPromptAPIResponse {
  status: number;
  ok: boolean;
  data?: {
    payload: any;
    url: string;
  };
}
