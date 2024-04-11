import React from 'react';
import cx from 'classnames';
import { Tooltip } from 'antd';
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

function QueryMetric({
  queryMetric,
  index,
  totalWidgets = 4,
  insightsDataByKey,
  showComparisonData,
  compareInsightsDataByKey,
  comparedSegmentId,
  comparedSegmentName,
  compareDateRange
}) {
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
      className={cx('flex w-1/4 items-center justify-center flex-col gap-y-4', {
        'border-r': index !== totalWidgets - 1
      })}
    >
      <div className='flex flex-col items-center'>
        <Text
          type='title'
          level={7}
          weight='medium'
          color='character-primary'
          extraClass='mb-0'
        >
          {queryMetric.d_name}
        </Text>
        <Text
          extraClass='mb-0'
          type='title'
          level={1}
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
            <div className={cx('flex gap-x-1 items-center justify-center')}>
              <Text
                type='title'
                level={8}
                extraClass={cx('mb-0 truncate', {
                  [styles['max-w-80']]: comparedSegmentId != null
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
            <div className={cx('flex gap-x-1 items-center justify-center')}>
              <Text
                type='title'
                level={8}
                extraClass={cx('mb-0 truncate', {
                  [styles['max-w-80']]: comparedSegmentId != null
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
