import React, { memo, useCallback, useEffect, useMemo } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import cx from 'classnames';
import { Spin } from 'antd';
import {
  selectAccountPayload,
  selectInsightsByWidgetGroupId,
  selectInsightsCompareSegmentBySegmentId
} from 'Reducers/accountProfilesView/selectors';
import { fetchInsights } from 'Reducers/accountProfilesView/services';
import { Text } from 'Components/factorsComponents';
import ControlledComponent from 'Components/ControlledComponent';
import { selectSegments } from 'Reducers/timelines/selectors';
import {
  getCompareDate,
  getSegmentName,
  getInsightsDataByKey
} from './accountsInsightsHelpers';
import QueryMetric from './QueryMetric';
import styles from './index.module.scss';

function InsightsWidget({ widget, dateRange, onEditMetricClick }) {
  const dispatch = useDispatch();
  const activeProject = useSelector((state) => state.global.active_project);
  const accountPayload = useSelector((state) => selectAccountPayload(state));
  const compareDateRange = getCompareDate(dateRange);
  const comparedSegmentId = useSelector((state) =>
    selectInsightsCompareSegmentBySegmentId(state, accountPayload.segment.id)
  );
  const segments = useSelector(selectSegments);

  const insights = useSelector((state) =>
    selectInsightsByWidgetGroupId(
      state,
      accountPayload.segment.id,
      widget.wid_g_id,
      dateRange.startDate,
      dateRange.endDate
    )
  );

  const compareInsights = useSelector((state) =>
    selectInsightsByWidgetGroupId(
      state,
      accountPayload.segment.id,
      widget.wid_g_id,
      compareDateRange.startDate,
      compareDateRange.endDate
    )
  );

  const comparedSegmentInsights = useSelector((state) =>
    selectInsightsByWidgetGroupId(
      state,
      comparedSegmentId,
      widget.wid_g_id,
      dateRange.startDate,
      dateRange.endDate
    )
  );

  const handleEditMetric = useCallback(
    (wid) => {
      onEditMetricClick(wid, widget.wid_g_id);
    },
    [onEditMetricClick]
  );

  useEffect(() => {
    if (
      accountPayload?.segment?.id != null &&
      insights.completed !== true &&
      insights.loading !== true
    ) {
      dispatch(
        fetchInsights({
          projectId: activeProject.id,
          segmentId: accountPayload?.segment?.id,
          widgetGroupId: widget.wid_g_id,
          dateFrom: dateRange.startDate,
          dateTo: dateRange.endDate
        })
      );
    }
  }, [
    activeProject.id,
    widget.wid_g_id,
    insights.completed,
    insights.loading,
    accountPayload?.segment?.id
  ]);

  useEffect(() => {
    if (
      comparedSegmentId != null &&
      comparedSegmentInsights.completed !== true &&
      comparedSegmentInsights.loading !== true
    ) {
      dispatch(
        fetchInsights({
          projectId: activeProject.id,
          segmentId: comparedSegmentId,
          widgetGroupId: widget.wid_g_id,
          dateFrom: dateRange.startDate,
          dateTo: dateRange.endDate
        })
      );
    }
  }, [
    activeProject.id,
    widget.wid_g_id,
    comparedSegmentInsights.completed,
    comparedSegmentInsights.loading,
    comparedSegmentId
  ]);

  useEffect(() => {
    if (
      accountPayload?.segment?.id != null &&
      compareInsights.completed !== true &&
      compareInsights.loading !== true &&
      Boolean(comparedSegmentId) === false
    ) {
      dispatch(
        fetchInsights({
          projectId: activeProject.id,
          segmentId: accountPayload?.segment?.id,
          widgetGroupId: widget.wid_g_id,
          dateFrom: compareDateRange.startDate,
          dateTo: compareDateRange.endDate
        })
      );
    }
  }, [
    activeProject.id,
    widget.wid_g_id,
    compareInsights.completed,
    compareInsights.loading,
    accountPayload?.segment?.id
  ]);

  const comparedSegmentName = useMemo(
    () => getSegmentName(segments, comparedSegmentId),
    [segments, comparedSegmentId]
  );

  const isLoading =
    insights.loading === true ||
    (insights.completed !== true && insights.error !== true);

  const insightsDataByKey = getInsightsDataByKey(insights);

  const compareData =
    comparedSegmentId == null ? compareInsights : comparedSegmentInsights;

  const compareInsightsDataByKey = getInsightsDataByKey(compareData);

  const showComparisonData = compareData.completed === true;

  return (
    <div className='flex flex-col border rounded-lg'>
      <div className='p-4 border-b flex gap-x-3'>
        <Text
          level={6}
          extraClass='mb-0'
          color='character-primary'
          weight='bold'
          type='title'
        >
          {widget.wid_g_d_name}
        </Text>
      </div>
      <div
        className={cx('p-4 flex', styles['min-h-48'], {
          'items-center justify-center': isLoading
        })}
      >
        <ControlledComponent controller={isLoading}>
          <Spin size='small' />
        </ControlledComponent>
        <ControlledComponent controller={insights.completed === true}>
          {widget.wids.map((queryMetric, index) => (
            <QueryMetric
              key={queryMetric.id}
              index={index}
              insightsDataByKey={insightsDataByKey}
              showComparisonData={showComparisonData}
              compareDateRange={compareDateRange}
              compareInsightsDataByKey={compareInsightsDataByKey}
              comparedSegmentId={comparedSegmentId}
              comparedSegmentName={comparedSegmentName}
              totalWidgets={4}
              queryMetric={queryMetric}
              onEditMetricClick={handleEditMetric}
            />
          ))}
        </ControlledComponent>
      </div>
    </div>
  );
}

export default memo(InsightsWidget);
