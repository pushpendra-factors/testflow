import React from 'react';
import cx from 'classnames';
import { Tooltip } from 'antd';
import { useSelector } from 'react-redux';

import { Number as NumFormat } from '../factorsComponents';
import { SPARK_LINE_CHART_TITLE_CHAR_COUNT } from '../../constants/charts.constants';
import { getFormattedKpiValue } from '../../Views/CoreQuery/KPIAnalysis/kpiAnalysis.helpers';
import LegendsCircle from '../../styles/components/LegendsCircle';
import { getEventDisplayName } from '../../Views/CoreQuery/EventsAnalytics/eventsAnalytics.helpers';
import styles from './index.module.scss';

function ChartHeader({
  total,
  query,
  bgColor,
  smallFont = false,
  metricType = null
}) {
  const { eventNames } = useSelector((state) => state.coreQuery);

  const queryName = getEventDisplayName({ event: query, eventNames });

  return (
    <div
      className={cx(
        'flex flex-col items-center justify-center',
        { 'row-gap-2': smallFont },
        { 'row-gap-4': !smallFont }
      )}
    >
      <Tooltip title={queryName}>
        <div className={'flex items-center col-gap-1'}>
          <LegendsCircle color={bgColor} />
          <div className={styles.eventText}>
            {queryName.length > SPARK_LINE_CHART_TITLE_CHAR_COUNT
              ? queryName.slice(0, SPARK_LINE_CHART_TITLE_CHAR_COUNT) + '...'
              : queryName}
          </div>
        </div>
      </Tooltip>

      <div
        className={`${smallFont ? styles.smallerTotalText : styles.totalText}`}
      >
        {metricType ? (
          getFormattedKpiValue({ value: total, metricType })
        ) : (
          <NumFormat shortHand={total > 10000} number={total} />
        )}
      </div>
    </div>
  );
}

export default ChartHeader;
