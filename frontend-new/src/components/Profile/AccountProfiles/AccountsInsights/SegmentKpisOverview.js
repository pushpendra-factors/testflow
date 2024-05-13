import React, { memo, useEffect, useMemo } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import cx from 'classnames';
import { Spin, Tooltip } from 'antd';
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

  const isLoading =
    insights.loading === true ||
    comparedSegmentInsights.loading === true ||
    (insights.completed !== true && insights.error !== true);

  const isComparedSegmentDataLoaded =
    comparedSegmentId != null
      ? comparedSegmentInsights.completed === true
      : true;

  const isCompleted =
    !isLoading && insights.completed && isComparedSegmentDataLoaded;

  const showComparisonData = comparedSegmentInsights.completed === true;

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
      <ControlledComponent controller={isLoading}>
        <Spin size='small' />
      </ControlledComponent>
      <ControlledComponent controller={isCompleted}>
        <div className='flex items-center w-full'>
          {widget.wids.map((queryMetric, index) => (
            <div
              key={queryMetric.id}
              className={cx('flex flex-1 items-center flex-col gap-y-2', {
                'border-r': index === 0
              })}
            >
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
              <ControlledComponent controller={showComparisonData}>
                <>
                  <ComparePercent value={3} />
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
            </div>
          ))}
        </div>
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
