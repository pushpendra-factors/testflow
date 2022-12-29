import React from 'react';
import PropTypes from 'prop-types';
import cx from 'classnames';
import { Tooltip } from 'antd';

import { Number as NumFormat, SVG, Text } from '../factorsComponents';
import { METRIC_CHART_TITLE_CHAR_COUNT } from '../../constants/charts.constants';
import LegendsCircle from '../../styles/components/LegendsCircle';
import styles from './index.module.scss';
import { CHART_COLOR_1 } from '../../constants/color.constants';
import ControlledComponent from '../ControlledComponent';

function MetricChart({
  value,
  headerTitle,
  iconColor,
  valueType,
  showComparison = false,
  compareValue = 0
}) {
  const percentChange = showComparison
    ? ((value - compareValue) / compareValue) * 100
    : 0;

  const changeIcon = showComparison ? (
    <SVG
      color={percentChange >= 0 ? '#5ACA89' : '#FF0000'}
      name={percentChange >= 0 ? 'arrowLift' : 'arrowDown'}
      size={16}
    />
  ) : null;

  return (
    <div className={cx('flex flex-col items-center justify-center row-gap-4 w-full')}>
      <Tooltip title={headerTitle}>
        <div className={'flex items-center col-gap-1 justify-center w-full'}>
          <LegendsCircle color={iconColor} />
          <Text
            color='grey-2'
            type='title'
            level={7}
            extraClass={'text-with-no-margin'}
          >
            {headerTitle.length > METRIC_CHART_TITLE_CHAR_COUNT
              ? headerTitle.slice(0, METRIC_CHART_TITLE_CHAR_COUNT) + '...'
              : headerTitle}
          </Text>
        </div>
      </Tooltip>

      <Text
        weight='bold'
        type='title'
        level={1}
        color='grey-2'
        extraClass={cx('text-with-no-margin', styles.count)}
      >
        <NumFormat shortHand={value > 1000} number={value} />
        <ControlledComponent controller={valueType === 'percentage'}>
          %
        </ControlledComponent>
      </Text>

      <ControlledComponent controller={showComparison}>
        <div className='flex flex-col row-gap-1 items-center'>
          <div className='flex col-gap-1 items-center'>
            {changeIcon}
            <Text
              level={7}
              type='title'
              color={percentChange < 0 ? 'red' : 'green'}
              extraClass={'text-with-no-margin'}
            >
              <NumFormat number={Math.abs(percentChange)} />%
            </Text>
          </div>
          <Text
            type='title'
            level={8}
            color='grey'
            extraClass={'text-with-no-margin'}
          >
            <NumFormat number={compareValue} shortHand={compareValue > 1000} />
            <ControlledComponent controller={valueType === 'percentage'}>
              %
            </ControlledComponent>
            &nbsp;in prev. period
          </Text>
        </div>
      </ControlledComponent>
    </div>
  );
}

export default MetricChart;

MetricChart.propTypes = {
  value: PropTypes.number,
  headerTitle: PropTypes.string,
  iconColor: PropTypes.string,
  valueType: PropTypes.oneOf(['numerical', 'percentage']),
  showComparison: PropTypes.bool,
  compareValue: PropTypes.number
};

MetricChart.defaultProps = {
  value: 0,
  headerTitle: 'Metric Chart',
  iconColor: CHART_COLOR_1,
  valueType: 'numerical',
  showComparison: false,
  compareValue: 0
};
