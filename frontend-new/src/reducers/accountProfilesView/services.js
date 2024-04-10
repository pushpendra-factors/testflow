import { get, getHostUrl, post } from 'Utils/request';
import { EMPTY_ARRAY } from 'Utils/global';
import MomentTz from 'Components/MomentTz';
import {
  ACCOUNTS_INSIGHTS_CONFIG_LOADING,
  ACCOUNTS_INSIGHTS_CONFIG_SUCCESS,
  ACCOUNTS_INSIGHTS_CONFIG_ERROR,
  ACCOUNTS_INSIGHTS_LOADING,
  ACCOUNTS_INSIGHTS_SUCCESS,
  ACCOUNTS_INSIGHTS_ERROR
} from './types';

const host = getHostUrl();

export const fetchInsightsConfig = (projectId) =>
  async function (dispatch) {
    try {
      const url = `${host}projects/${projectId}/segments/analytics/config`;
      dispatch({ type: ACCOUNTS_INSIGHTS_CONFIG_LOADING });
      const res = await get(null, url);
      dispatch({
        type: ACCOUNTS_INSIGHTS_CONFIG_SUCCESS,
        payload: res.data?.result ?? EMPTY_ARRAY
      });
    } catch (err) {
      console.log(err);
      dispatch({
        type: ACCOUNTS_INSIGHTS_CONFIG_ERROR
      });
    }
  };

export const fetchInsights = ({
  projectId,
  segmentId,
  widgetGroupId,
  dateFrom,
  dateTo
}) =>
  async function (dispatch) {
    try {
      const url = `${host}projects/${projectId}/segments/${segmentId}/analytics/widget_group/${widgetGroupId}/query`;
      dispatch({
        type: ACCOUNTS_INSIGHTS_LOADING,
        payload: { widgetGroupId, segmentId, dateFrom, dateTo }
      });
      const res = await post(null, url, {
        fr: MomentTz(dateFrom).utc().unix(),
        to: MomentTz(dateTo).utc().unix(),
        tz: localStorage.getItem('project_timeZone') || 'Asia/Kolkata'
      });
      dispatch({
        type: ACCOUNTS_INSIGHTS_SUCCESS,
        payload: {
          data: res.data.result ?? res.data,
          widgetGroupId,
          segmentId,
          dateFrom,
          dateTo
        }
      });
    } catch (err) {
      console.log(err);
      dispatch({
        type: ACCOUNTS_INSIGHTS_ERROR,
        payload: { widgetGroupId, segmentId, dateFrom, dateTo }
      });
    }
  };
