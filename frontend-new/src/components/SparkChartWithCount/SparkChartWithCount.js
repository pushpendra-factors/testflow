import React from 'react';
import PropTypes from 'prop-types';
import cx from 'classnames';
import { METRIC_TYPES } from '../../utils/constants';
import SparkChart from '../SparkLineChart/Chart';
import ChartHeader from './ChartHeader';

const SparkChartWithCount = ({
  total,
  compareTotal,
  event,
  metricType,
  chartColor,
  comparisonEnabled,
  smallFont,
  alignment,
  headerTitle,
  ...rest
}) => {
  return (
    <div
      className={cx('flex items-center justify-center w-full', {
        'flex-col': alignment === 'vertical'
      })}
    >
      <div
        className={cx(
          { 'w-1/4': alignment === 'horizontal' },
          { 'w-full': alignment === 'vertical' }
        )}
      >
        <ChartHeader
          bgColor={chartColor}
          total={total}
          metricType={metricType}
          compareTotal={compareTotal}
          comparisonEnabled={comparisonEnabled}
          smallFont={smallFont}
          headerTitle={headerTitle}
        />
      </div>
      <div
        className={cx(
          { 'w-3/4': alignment === 'horizontal' },
          { 'w-full': alignment === 'vertical' }
        )}
      >
        <SparkChart
          event={event}
          chartColor={chartColor}
          metricType={metricType}
          comparisonEnabled={comparisonEnabled}
          {...rest}
        />
      </div>
    </div>
  );
};

export default SparkChartWithCount;

SparkChartWithCount.propTypes = {
  total: PropTypes.number,
  compareTotal: PropTypes.number,
  title: PropTypes.string,
  headerTitle: PropTypes.string,
  chartColor: PropTypes.string,
  event: PropTypes.string,
  frequency: PropTypes.string,
  height: PropTypes.number,
  metricType: PropTypes.oneOf([
    METRIC_TYPES.dateType,
    METRIC_TYPES.percentType
  ]),
  page: PropTypes.string,
  chartData: PropTypes.arrayOf(
    PropTypes.shape({
      date: PropTypes.instanceOf(Date)
    })
  ),
  comparisonEnabled: PropTypes.bool,
  smallFont: PropTypes.bool,
  alignment: PropTypes.oneOf(['horizontal', 'vertical'])
};

SparkChartWithCount.defaultProps = {
  total: 0,
  compareTotal: 0,
  headerTitle: '',
  title: 'Chart',
  chartColor: '#4D7DB4',
  event: 'event',
  frequency: 'date',
  height: 180,
  metricType: undefined,
  page: 'KPI',
  chartData: [],
  comparisonEnabled: false,
  smallFont: false,
  alignment: 'horizontal'
};
