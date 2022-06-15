import React from 'react';
import styles from './index.module.scss';
import { Number as NumFormat } from '../factorsComponents';
import { useSelector } from 'react-redux';
import { SPARK_LINE_CHART_TITLE_CHAR_COUNT } from '../../constants/charts.constants';
import { displayQueryName } from './sparkLineChart.helpers';
import { getFormattedKpiValue } from '../../Views/CoreQuery/KPIAnalysis/kpiAnalysis.helpers';

function ChartHeader({
  total,
  query,
  bgColor,
  smallFont = false,
  metricType = null
}) {
  const { eventNames } = useSelector((state) => state.coreQuery);

  const queryName = displayQueryName({ query, eventNames });

  return (
    <div className="flex flex-col items-center justify-center">
      <div className={`flex items-center ${smallFont ? 'mb-2' : 'mb-4'}`}>
        <div
          style={{ backgroundColor: bgColor }}
          className={`mr-1 ${styles.eventCircle}`}
        ></div>
        <div className={styles.eventText}>
          {queryName.length > SPARK_LINE_CHART_TITLE_CHAR_COUNT
            ? queryName.slice(0, SPARK_LINE_CHART_TITLE_CHAR_COUNT) + '...'
            : queryName}
        </div>
      </div>
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
