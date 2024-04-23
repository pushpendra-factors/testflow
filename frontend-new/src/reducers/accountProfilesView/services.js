import { notification } from 'antd';
import { get, getHostUrl, put, post } from 'Utils/request';
import { EMPTY_ARRAY } from 'Utils/global';
import MomentTz from 'Components/MomentTz';
import {
  ACCOUNTS_INSIGHTS_CONFIG_LOADING,
  ACCOUNTS_INSIGHTS_CONFIG_SUCCESS,
  ACCOUNTS_INSIGHTS_CONFIG_ERROR,
  ACCOUNTS_INSIGHTS_LOADING,
  ACCOUNTS_INSIGHTS_SUCCESS,
  ACCOUNTS_INSIGHTS_ERROR,
  EDIT_INSIGHTS_METRIC_LOADING,
  EDIT_INSIGHTS_METRIC_SUCCESS,
  EDIT_INSIGHTS_METRIC_ERROR
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

export const updateInsightsQueryMetric = ({
  projectId,
  widgetId,
  widgetGroupId,
  metric,
  metricName
}) =>
  async function (dispatch) {
    try {
      const url = `${host}projects/${projectId}/segments/analytics/widget_group/${widgetGroupId}/widgets/${widgetId}`;
      dispatch({
        type: EDIT_INSIGHTS_METRIC_LOADING
      });
      const requestBody = {};
      if (metric != null) {
        requestBody.q_me = metric;
      }
      if (metricName != null) {
        requestBody.d_name = metricName;
      }
      await put(null, url, requestBody);
      dispatch({
        type: EDIT_INSIGHTS_METRIC_SUCCESS,
        payload: { widgetId, widgetGroupId, metric, metricName }
      });
      notification.success({
        message: 'Success',
        description:
          metric != null
            ? 'Metric updated successfully'
            : 'Metric name updated successfully',
        duration: 2
      });
    } catch (err) {
      console.log(err);
      dispatch({
        type: EDIT_INSIGHTS_METRIC_ERROR
      });
    }
  };
