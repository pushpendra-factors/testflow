import React, { memo, useEffect, useMemo } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import cx from 'classnames';
import { Skeleton, Tooltip } from 'antd';
import {
  Number as NumFormat,
  SVG as Svg,
  Text
} from 'Components/factorsComponents';
import {
  selectAccountPayload,
  selectInsightsByWidgetGroupId,
  selectInsightsCompareSegmentBySegmentId
} from 'Reducers/accountProfilesView/selectors';
import { fetchInsights } from 'Reducers/accountProfilesView/services';
import ControlledComponent from 'Components/ControlledComponent';
import ComparePercent from 'Components/ComparePercent/ComparePercent';
import { selectSegments } from 'Reducers/timelines/selectors';
import {
  getInsightsDataByKey,
  getSegmentName
} from './accountsInsightsHelpers';
import styles from './index.module.scss';

function InputSkeleton() {
  return (
    <Skeleton.Input
      className={styles['overview-input-skeleton']}
      size='small'
      active
    />
  );
}

function ButtonSkeleton() {
  return (
    <Skeleton.Button className={styles['overview-button-skeleton']} active />
  );
}

function InsightsSkeleton() {
  return (
    <div className='flex flex-col gap-y-2 justify-center items-center'>
      <InputSkeleton />
      <ButtonSkeleton />
    </div>
  );
}

function CompareInsightsSkeleton() {
  return (
    <div className='flex flex-col gap-y-2 justify-center items-center'>
      <InputSkeleton />
    </div>
  );
}

function SegmentKpisOverview({ widget, dateRange }) {
  const dispatch = useDispatch();
  const activeProject = useSelector((state) => state.global.active_project);
  const accountPayload = useSelector((state) => selectAccountPayload(state));
  const segments = useSelector(selectSegments);
  const comparedSegmentId = useSelector((state) =>
    selectInsightsCompareSegmentBySegmentId(state, accountPayload.segment.id)
  );

  const insights = useSelector((state) =>
    selectInsightsByWidgetGroupId(
      state,
      accountPayload.segment.id,
      widget.wid_g_id,
      dateRange.startDate,
      dateRange.endDate
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

  // fetch insights for selected date
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
    accountPayload?.segment?.id,
    dateRange
  ]);

  // fetch insights for selected date for the compared segment
  useEffect(() => {
    if (
      comparedSegmentId != null &&
      comparedSegmentInsights.completed !== true &&
      comparedSegmentInsights.loading !== true &&
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
    insights.completed
  ]);

  const isLoading =
    insights.loading === true ||
    (insights.completed !== true && insights.error !== true);

  const isCompleted = insights.completed === true;

  // const isLoading = true;
  // const isCompleted = false;

  const comparedDataLoading =
    comparedSegmentInsights.loading === true ||
    (comparedSegmentInsights.completed !== true &&
      comparedSegmentInsights.error !== true);

  const isComparedSegmentDataLoaded =
    comparedSegmentInsights.completed === true;

  const showComparisonData = comparedSegmentId != null;

  const insightsDataByKey = getInsightsDataByKey(insights);
  const compareInsightsDataByKey = getInsightsDataByKey(
    comparedSegmentInsights
  );

  const comparedSegmentName = useMemo(
    () => getSegmentName(segments, comparedSegmentId),
    [segments, comparedSegmentId]
  );

  return (
    <div
      className={cx(
        'flex flex-col gap-y-4 border rounded-lg py-8 items-center',
        {
          'items-center justify-center': isLoading,
          [styles['min-h-48']]: isLoading
        }
      )}
    >
      <div className='flex items-center w-full'>
        {widget.wids.map((queryMetric, index) => (
          <div
            key={queryMetric.id}
            className={cx('flex flex-1 items-center flex-col gap-y-2', {
              'border-r': index === 0
            })}
          >
            <ControlledComponent controller={isLoading}>
              <InsightsSkeleton />
            </ControlledComponent>
            <ControlledComponent controller={isCompleted}>
              <div className='flex gap-x-3 items-center'>
                <Svg
                  name={index === 0 ? 'buildings' : 'fireFlameCurved'}
                  size={24}
                  color={index === 0 ? '#1890FF' : '#FA8C16'}
                />
                <Text
                  type='title'
                  color='character-primary'
                  extraClass='mb-0'
                  level={6}
                  weight='medium'
                >
                  {queryMetric.d_name}
                </Text>
              </div>
              <Text
                extraClass='mb-0'
                type='title'
                level={2}
                weight='bold'
                color='character-primary'
              >
                <NumFormat
                  shortHand
                  number={insightsDataByKey[queryMetric.q_me]}
                />
              </Text>
              <ControlledComponent
                controller={showComparisonData && comparedDataLoading}
              >
                <CompareInsightsSkeleton />
              </ControlledComponent>
              <ControlledComponent
                controller={showComparisonData && isComparedSegmentDataLoaded}
              >
                <>
                  <ComparePercent
                    value={
                      Boolean(compareInsightsDataByKey[queryMetric.q_me]) &&
                      Boolean(insightsDataByKey[queryMetric.q_me])
                        ? ((insightsDataByKey[queryMetric.q_me] -
                            compareInsightsDataByKey[queryMetric.q_me]) /
                            compareInsightsDataByKey[queryMetric.q_me]) *
                          100
                        : 0
                    }
                  />
                  <Text
                    extraClass='mb-0'
                    type='title'
                    level={8}
                    color='character-secondary'
                  >
                    <span className='font-bold'>
                      <NumFormat
                        shortHand
                        number={compareInsightsDataByKey[queryMetric.q_me]}
                      />
                    </span>{' '}
                    <Tooltip title={comparedSegmentName}>
                      <span>in {comparedSegmentName}</span>
                    </Tooltip>
                  </Text>
                </>
              </ControlledComponent>
            </ControlledComponent>
          </div>
        ))}
      </div>
      <ControlledComponent controller={isCompleted}>
        <Text
          type='title'
          extraClass='mb-0 text-center'
          color='character-secondary'
          level={8}
        >
          These metrics show the current state and are independent of the date
          range
        </Text>
      </ControlledComponent>
    </div>
  );
}

export default memo(SegmentKpisOverview);
