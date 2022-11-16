import React from 'react';
import cx from 'classnames';
import { Tooltip } from 'antd';
import { useSelector } from 'react-redux';

import { Number as NumFormat, SVG, Text } from '../factorsComponents';
import { SPARK_LINE_CHART_TITLE_CHAR_COUNT } from '../../constants/charts.constants';
import { getFormattedKpiValue } from '../../Views/CoreQuery/KPIAnalysis/kpiAnalysis.helpers';
import LegendsCircle from '../../styles/components/LegendsCircle';
import { getEventDisplayName } from '../../Views/CoreQuery/EventsAnalytics/eventsAnalytics.helpers';
import styles from './index.module.scss';
import ControlledComponent from '../ControlledComponent/ControlledComponent';

function ChartHeader({
  total,
  query,
  bgColor,
  smallFont = false,
  metricType = null,
  eventNames,
  compareTotal,
  comparisonEnabled,
  titleCharCount
}) {
  const queryName = getEventDisplayName({ event: query, eventNames });

  const TitleCharCount = titleCharCount || SPARK_LINE_CHART_TITLE_CHAR_COUNT;

  const percentChange = comparisonEnabled
    ? ((total - compareTotal) / compareTotal) * 100
    : 0;

  const changeIcon = comparisonEnabled ? (
    <SVG
      color={percentChange > 0 ? '#5ACA89' : '#FF0000'}
      name={percentChange > 0 ? 'arrowLift' : 'arrowDown'}
      size={16}
    />
  ) : null;

  return (
    <div className={cx('flex flex-col items-center justify-center row-gap-2')}>
      <Tooltip title={queryName}>
        <div className={'flex items-center col-gap-1'}>
          <LegendsCircle color={bgColor} />
          <div className={styles.eventText}>
            {queryName.length > TitleCharCount
              ? queryName.slice(0, TitleCharCount) + '...'
              : queryName}
          </div>
        </div>
      </Tooltip>

      <ControlledComponent controller={!smallFont}>
        <Text weight='bold' type='title' level={2} color='grey-2'>
          {metricType ? (
            getFormattedKpiValue({ value: total, metricType })
          ) : (
            <NumFormat shortHand={total > 10000} number={total} />
          )}
        </Text>
      </ControlledComponent>

      <ControlledComponent controller={smallFont}>
        <Text weight='bold' type='title' level={3} color='grey-2'>
          {metricType ? (
            getFormattedKpiValue({ value: total, metricType })
          ) : (
            <NumFormat shortHand={total > 10000} number={total} />
          )}
        </Text>
      </ControlledComponent>

      {comparisonEnabled && (
        <div className='flex flex-col row-gap-1 items-center'>
          <div className='flex col-gap-1 items-center'>
            {changeIcon}
            <Text
              level={7}
              type='title'
              color={percentChange < 0 ? 'red' : 'green'}
            >
              <NumFormat number={Math.abs(percentChange)} />%
            </Text>
          </div>
          <Text type='title' level={8} color='grey'>
            <NumFormat number={compareTotal} shortHand={true} /> in prev. period
          </Text>
        </div>
      )}
    </div>
  );
}

export default ChartHeader;
