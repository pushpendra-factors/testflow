import React from 'react';
import cx from 'classnames';
import { Popover, Tooltip } from 'antd';
import {
  Text,
  Number as NumFormat,
  SVG as Svg
} from 'Components/factorsComponents';
import ControlledComponent from 'Components/ControlledComponent';
import ComparePercent from 'Components/ComparePercent/ComparePercent';
import { getFormattedMetricValue } from './accountsInsightsHelpers';
import styles from './index.module.scss';

const CompareDurationTooltip = ({ title }) => (
  <Tooltip title={title}>
    <span>
      <Svg
        extraClass='cursor-pointer'
        size={12}
        name='infoCircle'
        color='#8c8c8c'
      />
    </span>
  </Tooltip>
);

const getPopoverContent = (metricName) => {
  let text =
    'Average of time between deal create date and deal close date for all deals that were closed in the selected time range.';
  if (metricName === 'Marketing qualified leads') {
    text = 'Count of marketing qualified leads from this segment.';
  }
  if (metricName === 'Sales qualified leads') {
    text = 'Count of sales qualified leads from this segment.';
  }
  if (metricName === 'Opportunity Created') {
    text =
      'Count of all deals associated with accounts in this segment, created in the selected time range. ';
  }
  if (metricName === 'Pipeline Created') {
    text =
      'Sum of deal amount of all deals associated with accounts in this segment, created in the selected time range.';
  }
  if (metricName === 'Average Deal Size') {
    text =
      'Average of deal amount of all deals associated with accounts in this segment, created in the selected time range.';
  }
  if (metricName === 'Revenue Booked') {
    text =
      'Sum of deal amount of all closed won deals associated with accounts in this segment, closed in the selected time range. ';
  }
  if (metricName === 'Close Rate (%)') {
    text =
      '% of deals that were marked as closed won out of all the deals that were created in the selected time range. ';
  }
  return (
    <div className={styles.metricDescriptionText}>
      <Text type='title' extraClass='mb-0' color='character-primary' level={8}>
        {text}
      </Text>
    </div>
  );
};

function QueryMetric({
  queryMetric,
  index,
  totalWidgets = 4,
  insightsDataByKey,
  showComparisonData,
  compareInsightsDataByKey,
  comparedSegmentId,
  comparedSegmentName,
  compareDateRange,
  onEditMetricClick
}) {
  const handleEditMetric = () => {
    onEditMetricClick(queryMetric);
  };

  const compareText =
    comparedSegmentId == null ? (
      `in prev. period`
    ) : (
      <Tooltip title={comparedSegmentName}>
        <span>in {comparedSegmentName}</span>
      </Tooltip>
    );

  const tooltipTitle = `${compareDateRange.startDate.format(
    'MMMM DD, YYYY'
  )} - ${compareDateRange.endDate.format('MMMM DD, YYYY')}`;

  return (
    <div
      key={queryMetric.id}
      className={cx(
        'flex w-1/4 px-4 items-center justify-center flex-col gap-y-4',
        styles['metric-container'],
        {
          'border-r': index !== totalWidgets - 1
        }
      )}
    >
      <div className='flex flex-col items-center w-full'>
        <div className='flex items-center justify-between w-full'>
          <div className='w-6' />
          <Popover
            trigger='hover'
            placement='topRight'
            title={
              <Text
                weight='medium'
                type='title'
                color='character-title'
                extraClass='mb-0'
              >
                {queryMetric.d_name}
              </Text>
            }
            content={getPopoverContent(queryMetric.d_name)}
          >
            <Text
              type='title'
              level={7}
              weight='medium'
              color='character-primary'
              extraClass='mb-0'
            >
              {queryMetric.d_name}
            </Text>
          </Popover>
          <div
            onClick={handleEditMetric}
            className={cx('invisible', styles['edit-button'])}
          >
            <Svg name='pencil' color='currentColor' />
          </div>
        </div>
        <Text
          extraClass='mb-0'
          type='title'
          level={2}
          weight='bold'
          color='character-primary'
        >
          <ControlledComponent
            controller={insightsDataByKey[queryMetric.q_me] != null}
          >
            <ControlledComponent
              controller={Boolean(queryMetric.q_me_ty) === true}
            >
              {getFormattedMetricValue(
                insightsDataByKey[queryMetric.q_me]?.[0],
                queryMetric.q_me_ty
              )}
            </ControlledComponent>
            <ControlledComponent
              controller={Boolean(queryMetric.q_me_ty) === false}
            >
              <NumFormat
                number={insightsDataByKey[queryMetric.q_me]?.[0]}
                shortHand
              />
            </ControlledComponent>
          </ControlledComponent>
          <ControlledComponent
            controller={insightsDataByKey[queryMetric.q_me] == null}
          >
            {getFormattedMetricValue(0, queryMetric.q_me_ty)}
          </ControlledComponent>
        </Text>
      </div>
      <ControlledComponent controller={showComparisonData}>
        <ControlledComponent
          controller={compareInsightsDataByKey[queryMetric.q_me] != null}
        >
          <div className='flex flex-col items-center w-full'>
            <ComparePercent
              value={
                insightsDataByKey[queryMetric.q_me] != null &&
                compareInsightsDataByKey[queryMetric.q_me] != null
                  ? ((insightsDataByKey[queryMetric.q_me][0] -
                      compareInsightsDataByKey[queryMetric.q_me][0]) /
                      compareInsightsDataByKey[queryMetric.q_me][0]) *
                    100 // (((new-old)/old) * 100)
                  : 0
              }
            />
            <div
              className={cx('flex gap-x-1 items-center justify-center w-full')}
            >
              <Text
                type='title'
                level={8}
                extraClass={cx('mb-0 truncate', {
                  [styles['max-w-100']]: comparedSegmentId != null
                })}
                color='character-secondary'
              >
                <ControlledComponent
                  controller={Boolean(queryMetric.q_me_ty) === true}
                >
                  <span className='font-bold'>
                    {getFormattedMetricValue(
                      compareInsightsDataByKey[queryMetric.q_me]?.[0],
                      queryMetric.q_me_ty
                    )}{' '}
                  </span>
                </ControlledComponent>
                <ControlledComponent
                  controller={Boolean(queryMetric.q_me_ty) === false}
                >
                  <span className='font-bold'>
                    <NumFormat
                      number={compareInsightsDataByKey[queryMetric.q_me]?.[0]}
                      shortHand
                    />{' '}
                  </span>
                </ControlledComponent>
                {compareText}
              </Text>
              <ControlledComponent controller={comparedSegmentId == null}>
                <CompareDurationTooltip title={tooltipTitle} />
              </ControlledComponent>
            </div>
          </div>
        </ControlledComponent>
        <ControlledComponent
          controller={compareInsightsDataByKey[queryMetric.q_me] == null}
        >
          <div className='flex flex-col items-center w-full'>
            <ComparePercent value={0} />
            <div
              className={cx('flex gap-x-1 items-center justify-center w-full')}
            >
              <Text
                type='title'
                level={8}
                extraClass={cx('mb-0 truncate', {
                  [styles['max-w-100']]: comparedSegmentId != null
                })}
                color='character-secondary'
              >
                <span className='font-bold'>
                  {getFormattedMetricValue(0, queryMetric.q_me_ty)}{' '}
                </span>
                {compareText}
              </Text>
              <ControlledComponent controller={comparedSegmentId == null}>
                <CompareDurationTooltip title={tooltipTitle} />
              </ControlledComponent>
            </div>
          </div>
        </ControlledComponent>
      </ControlledComponent>
    </div>
  );
}

export default QueryMetric;
