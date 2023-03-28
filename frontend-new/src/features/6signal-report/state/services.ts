import logger from 'Utils/logger';
import { getHostUrl, post, get } from 'Utils/request';

const host = getHostUrl();
export const getSixSignalReportData = async (
  projectId: string,
  from: number,
  to: number,
  timezone: string
) => {
  try {
    if (!projectId || !from || !to || !timezone) {
      throw new Error('Invalid parameters passed');
    }
    const url = `${host}projects/${projectId}/v1/sixsignal`;
    return post(null, url, {
      six_signal_query_group: [
        {
          fr: from,
          to: to,
          tz: timezone
        }
      ]
    });
  } catch (error) {
    logger.error(error);
    return null;
  }
};

export const shareSixSignalReport = async (
  projectId: string,
  from: number,
  to: number,
  timezone: string
) => {
  try {
    if (!projectId || !from || !to || !timezone) {
      throw new Error('Invalid parameters passed');
    }
    const url = `${host}projects/${projectId}/sixsignal/share`;
    return post(null, url, {
      six_signal_query: {
        fr: from,
        to: to,
        tz: timezone
      },
      entity_type: 4,
      share_type: 1
    });
  } catch (error) {
    logger.error('Error in sharing report', error);
    return null;
  }
};

export const getSixSignalReportPublicData = async (
  projectId: string,
  queryId: string
) => {
  try {
    if (!queryId || !projectId) {
      throw new Error('Invalid parameters passed');
    }
    const url = `${host}projects/${projectId}/v1/sixsignal/publicreport?query_id=${queryId}`;
    return get(null, url);
  } catch (error) {
    logger.error(error);
    return null;
  }
};

export const shareSixSignalReportToEmails = async (
  emails: string[],
  shareUrl: string,
  domain: string,
  from: number,
  to: number,
  timezone: string
) => {
  try {
    if (!emails || !shareUrl || !domain || !from || !to || !timezone) {
      throw new Error('Invalid parameters passed');
    }
    const url = `${host}projects/2/sixsignal/email`;
    return post(null, url, {
      email_ids: emails,
      url: shareUrl,
      domain: domain,
      fr: from,
      to: to,
      tz: timezone
    });
  } catch (error) {
    logger.error(error);
    return null;
  }
};
