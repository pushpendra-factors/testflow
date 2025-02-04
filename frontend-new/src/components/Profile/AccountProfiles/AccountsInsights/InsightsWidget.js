import React, { memo, useCallback, useEffect, useMemo } from 'react';
import { useHistory } from 'react-router-dom';
import { useDispatch, useSelector } from 'react-redux';
import cx from 'classnames';
import { Button } from 'antd';
import {
  selectAccountPayload,
  selectEditInsightsMetricStatus,
  selectInsightsByWidgetGroupId,
  selectInsightsCompareSegmentBySegmentId
} from 'Reducers/accountProfilesView/selectors';
import { fetchInsights } from 'Reducers/accountProfilesView/services';
import { SVG as Svg, Text } from 'Components/factorsComponents';
import ControlledComponent from 'Components/ControlledComponent';
import { selectSegments } from 'Reducers/timelines/selectors';
import { resetEditMetricStatus } from 'Reducers/accountProfilesView/actions';
import { PathUrls } from 'Routes/pathUrls';
import {
  getCompareDate,
  getSegmentName,
  getInsightsDataByKey
} from './accountsInsightsHelpers';
import QueryMetric from './QueryMetric';
import styles from './index.module.scss';

function InsightsWidget({
  widget,
  dateRange,
  onEditMetricClick,
  editWidgetGroupId
}) {
  const dispatch = useDispatch();
  const history = useHistory();
  const activeProject = useSelector((state) => state.global.active_project);
  const accountPayload = useSelector((state) => selectAccountPayload(state));
  const compareDateRange = getCompareDate(dateRange);
  const comparedSegmentId = useSelector((state) =>
    selectInsightsCompareSegmentBySegmentId(state, accountPayload.segment.id)
  );
  const segments = useSelector(selectSegments);
  const editMetricStatus = useSelector(selectEditInsightsMetricStatus);

  const currentProjectSettings = useSelector(
    (state) => state.global.currentProjectSettings
  );

  const isIntegrationDone =
    Boolean(currentProjectSettings.int_hubspot) ||
    Boolean(currentProjectSettings.int_salesforce_enabled_agent_uuid);

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

  const handleIntegrateNowClick = useCallback(() => {
    history.push(PathUrls.SettingsIntegration);
  }, []);

  // fetch insights for selected date
  useEffect(() => {
    if (
      accountPayload?.segment?.id != null &&
      insights.completed !== true &&
      insights.loading !== true &&
      isIntegrationDone === true
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
    accountPayload?.segment?.id,
    isIntegrationDone
  ]);

  // fetch insights for selected date for the compared segment
  useEffect(() => {
    if (
      comparedSegmentId != null &&
      comparedSegmentInsights.completed !== true &&
      comparedSegmentInsights.loading !== true &&
      isIntegrationDone === true &&
      insights.completed === true
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
    comparedSegmentId,
    isIntegrationDone,
    insights.completed
  ]);

  // fetch insights for previous date
  useEffect(() => {
    if (
      accountPayload?.segment?.id != null &&
      compareInsights.completed !== true &&
      compareInsights.loading !== true &&
      Boolean(comparedSegmentId) === false &&
      isIntegrationDone === true &&
      insights.completed === true
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
    accountPayload?.segment?.id,
    isIntegrationDone,
    insights.completed
  ]);

  // fetch for current date and previous date (or compared segment) when metric is edited
  useEffect(() => {
    if (
      editMetricStatus.completed === true &&
      editWidgetGroupId === widget.wid_g_id
    ) {
      dispatch(resetEditMetricStatus());
      dispatch(
        fetchInsights({
          projectId: activeProject.id,
          segmentId: accountPayload?.segment?.id,
          widgetGroupId: widget.wid_g_id,
          dateFrom: dateRange.startDate,
          dateTo: dateRange.endDate
        })
      );
      if (comparedSegmentId != null) {
        dispatch(
          fetchInsights({
            projectId: activeProject.id,
            segmentId: comparedSegmentId,
            widgetGroupId: widget.wid_g_id,
            dateFrom: dateRange.startDate,
            dateTo: dateRange.endDate
          })
        );
      } else {
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
    }
  }, [
    editMetricStatus.completed,
    activeProject.id,
    comparedSegmentId,
    widget,
    dateRange,
    accountPayload?.segment?.id,
    compareDateRange,
    editWidgetGroupId
  ]);

  const comparedSegmentName = useMemo(
    () => getSegmentName(segments, comparedSegmentId),
    [segments, comparedSegmentId]
  );

  const insightsDataByKey = getInsightsDataByKey(insights);

  const compareData =
    comparedSegmentId == null ? compareInsights : comparedSegmentInsights;

  const isLoading =
    insights.loading === true ||
    (insights.completed !== true && insights.error !== true);

  const compareLoading =
    compareData.loading === true ||
    (compareData.completed !== true && compareData.error !== true);

  const compareInsightsDataByKey = getInsightsDataByKey(compareData);

  return (
    <div className='flex flex-col border rounded-lg'>
      <div className='p-4 border-b flex gap-x-3'>
        <ControlledComponent controller={widget.name === 'marketing'}>
          <Svg name='analysis' size={24} color='#73D13D' />
        </ControlledComponent>
        <ControlledComponent controller={widget.name === 'sales'}>
          <Svg name='lightBulbOn' size={24} color='#FFC53D' />
        </ControlledComponent>

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
      <div className={cx('flex py-8', styles['min-h-48'])}>
        <ControlledComponent controller={isIntegrationDone === false}>
          <div className='flex flex-col justify-center items-center gap-y-5'>
            <img alt='no-data' src='../../../../assets/images/disconnect.svg' />
            <div className='flex flex-col justify-center items-center gap-y-2'>
              <Text type='title' extraClass='mb-0' color='character-primary'>
                Connect your CRM for this widget
              </Text>
              <Button
                onClick={handleIntegrateNowClick}
                type='text'
                className={styles.linkButton}
              >
                Integrate now
              </Button>
            </div>
          </div>
        </ControlledComponent>
        <ControlledComponent controller={insights.error !== true}>
          {widget.wids.map((queryMetric, index) => (
            <QueryMetric
              key={queryMetric.id}
              index={index}
              insightsDataByKey={insightsDataByKey}
              compareDateRange={compareDateRange}
              compareInsightsDataByKey={compareInsightsDataByKey}
              comparedSegmentId={comparedSegmentId}
              comparedSegmentName={comparedSegmentName}
              isLoading={isLoading}
              compareLoading={compareLoading}
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
