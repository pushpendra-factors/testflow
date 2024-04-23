import { createSelector } from 'reselect';
import { generateInsightsKey } from './index';

export const selectAccountPayload = (state) =>
  state.accountProfilesView.accountPayload;

export const selectSegmentModalState = (state) =>
  state.accountProfilesView.showSegmentModal;

export const selectActiveTab = (state) => state.accountProfilesView.activeTab;
export const selectInsightsConfig = (state) =>
  state.accountProfilesView.insightsConfig;

export const selectInsightsByWidgetGroupId = createSelector(
  (state) => state.accountProfilesView,
  (state, segmentId, widgetGroupId, dateFrom, dateTo) => ({
    segmentId,
    widgetGroupId,
    dateFrom,
    dateTo
  }),
  (accountProfilesView, { segmentId, widgetGroupId, dateFrom, dateTo }) => {
    if (segmentId != null) {
      const key = generateInsightsKey({
        widgetGroupId,
        segmentId,
        dateFrom,
        dateTo
      });
      return accountProfilesView.insights[key] ?? {};
    }
    return {};
  }
);

export const selectInsightsCompareSegmentBySegmentId = (state, segmentId) =>
  state.accountProfilesView.insightsCompareConfig[segmentId];

export const selectEditInsightsMetricStatus = (state) =>
  state.accountProfilesView.editInsightsMetric;
