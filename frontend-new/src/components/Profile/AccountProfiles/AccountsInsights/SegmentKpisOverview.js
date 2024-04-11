import React, { memo, useEffect } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import cx from 'classnames';
import { Spin } from 'antd';
import {
  Number as NumFormat,
  SVG as Svg,
  Text
} from 'Components/factorsComponents';
import {
  selectAccountPayload,
  selectInsightsByWidgetGroupId
} from 'Reducers/accountProfilesView/selectors';
import { fetchInsights } from 'Reducers/accountProfilesView/services';
import ControlledComponent from 'Components/ControlledComponent';
import { getInsightsDataByKey } from './accountsInsightsHelpers';
import styles from './index.module.scss';

function SegmentKpisOverview({ widget, dateRange }) {
  const dispatch = useDispatch();
  const activeProject = useSelector((state) => state.global.active_project);
  const accountPayload = useSelector((state) => selectAccountPayload(state));

  const insights = useSelector((state) =>
    selectInsightsByWidgetGroupId(
      state,
      accountPayload.segment.id,
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

  const isLoading =
    insights.loading === true ||
    (insights.completed !== true && insights.error !== true);

  const insightsDataByKey = getInsightsDataByKey(insights);

  return (
    <div
      className={cx(
        'flex flex-col gap-y-2 border rounded-lg py-8 px-4 items-center',
        styles['min-h-48'],
        { 'items-center justify-center': isLoading }
      )}
    >
      <ControlledComponent controller={isLoading}>
        <Spin size='small' />
      </ControlledComponent>
      <ControlledComponent controller={insights.completed === true}>
        <div className='flex items-center w-full'>
          {widget.wids.map((queryMetric, index) => (
            <div
              key={queryMetric.id}
              className='flex flex-1 items-center border-r flex-col gap-y-2'
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
                level={1}
                weight='bold'
                color='character-primary'
              >
                <NumFormat
                  shortHand
                  number={insightsDataByKey[queryMetric.q_me]}
                />
              </Text>
            </div>
          ))}
        </div>
        <Text
          type='title'
          extraClass='mb-0 text-center'
          color='character-secondary'
          level={8}
        >
          These numbers are real-time
        </Text>
      </ControlledComponent>
    </div>
  );
}

export default memo(SegmentKpisOverview);
