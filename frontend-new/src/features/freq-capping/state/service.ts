import logger from 'Utils/logger';
import { getHostUrl, get, post, put, del } from 'Utils/request';
import { FrequencyCap } from '../types';

const host = getHostUrl();

export const getLinkedinFreqCapRuleConfig = async (projectId: string) => {
    try {
        if (!projectId) {
            throw new Error('Invalid parameters passed');
        }
        const url = `${host}projects/${projectId}/v1/linkedin_capping/rules/config`;
        return get(null, url);
    } catch (error) {
        logger.error(error);
        return null;
    }
};

export const getLinkedinFreqCapRules = async (projectId: string) => {
    try {
        if (!projectId) {
            throw new Error('Invalid parameters passed');
        }
        const url = `${host}projects/${projectId}/v1/linkedin_capping/rules`;
        return get(null, url);
    } catch (error) {
        logger.error(error);
        return null;
    }
};

export const publishLinkedinFreqCapRules = async (
    projectId: string,
    freqRule: FrequencyCap
) => {
    try {
        if (!projectId) {
            throw new Error('Invalid parameters passed');
        }
        const url = `${host}projects/${projectId}/v1/linkedin_capping/rules`;
        return post(null, url, freqRule);
    } catch (error) {
        logger.error(error);
        return null;
    }
};

export const updateLinkedinFreqCapRules = async (
    projectId: string,
    freqRule: FrequencyCap
) => {
    try {
        if (!projectId) {
            throw new Error('Invalid parameters passed');
        }
        const url = `${host}projects/${projectId}/v1/linkedin_capping/rules/${freqRule.id}`;
        return put(null, url, freqRule);
    } catch (error) {
        logger.error(error);
        return null;
    }
};

export const deleteLinkedinFreqCapRules = async (
    projectId: string,
    ruleId: string
) => {
    try {
        if (!projectId) {
            throw new Error('Invalid parameters passed');
        }
        const url = `${host}projects/${projectId}/v1/linkedin_capping/rules/${ruleId}`;
        return del(null, url);
    } catch (error) {
        logger.error(error);
        return null;
    }
};
