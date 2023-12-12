import logger from 'Utils/logger';
import { getHostUrl, get, post } from 'Utils/request';

const host = getHostUrl();

export const getSubscriptionDetails = async (projectId: string) => {
  try {
    if (!projectId) {
      throw new Error('Invalid parameters passed');
    }
    const url = `${host}projects/${projectId}/billing/subscription`;
    return get(null, url);
  } catch (error) {
    logger.error(error);
    return null;
  }
};

export const getPlansDetails = async (projectId: string) => {
  try {
    if (!projectId) {
      throw new Error('Invalid parameters passed');
    }
    const url = `${host}projects/${projectId}/billing/pricing`;
    return get(null, url);
  } catch (error) {
    logger.error(error);
    return null;
  }
};

export const upgradePlan = async (
  projectId: string,
  planId?: string,
  addons?: { addon_id: string; quantity: number }[]
) => {
  try {
    if (!projectId) {
      throw new Error('Invalid parameters passed');
    }

    const url = `${host}projects/${projectId}/billing/upgrade`;
    const body = {};
    if (planId) {
      body.updated_plan_id = planId;
    }
    if (addons) {
      body.add_ons = addons;
    }
    return post(null, url, body);
  } catch (error) {
    logger.error(error);
    return null;
  }
};

export const getInvoices = async (projectId: string) => {
  try {
    if (!projectId) {
      throw new Error('Invalid parameters passed');
    }
    const url = `${host}projects/${projectId}/billing/invoices`;
    return get(null, url);
  } catch (error) {
    logger.error(error);
    return null;
  }
};

export const downloadInvoice = async (projectId: string, invoiceId: string) => {
  try {
    if (!projectId || !invoiceId) {
      throw new Error('Invalid parameters passed');
    }
    const url = `${host}projects/${projectId}/billing/invoice/download?invoice_id=${invoiceId}`;
    return get(null, url);
  } catch (error) {
    logger.error(error);
    return null;
  }
};
