import React from 'react';
import cx from 'classnames';
import { Tooltip } from 'antd';
import { Number as NumFormat, SVG, Text } from '../factorsComponents';
import { SPARK_LINE_CHART_TITLE_CHAR_COUNT } from '../../constants/charts.constants';
import { getFormattedKpiValue } from '../../Views/CoreQuery/KPIAnalysis/kpiAnalysis.helpers';
import LegendsCircle from '../../styles/components/LegendsCircle';
import ControlledComponent from '../ControlledComponent/ControlledComponent';

function ChartHeader({
  total,
  bgColor,
  smallFont = false,
  metricType = null,
  compareTotal,
  comparisonApplied,
  headerTitle
}) {
  const percentChange = comparisonApplied
    ? ((total - compareTotal) / compareTotal) * 100
    : 0;

  const changeIcon = comparisonApplied ? (
    <SVG
      color={percentChange > 0 ? '#5ACA89' : '#FF0000'}
      name={percentChange > 0 ? 'arrowLift' : 'arrowDown'}
      size={16}
    />
  ) : null;

  return (
    <div className={cx('flex flex-col items-center justify-center row-gap-2')}>
      <Tooltip title={headerTitle}>
        <div className={'flex items-center col-gap-1'}>
          <LegendsCircle color={bgColor} />
          <Text color="grey-8" type="title" level={7}>
            {headerTitle.length > SPARK_LINE_CHART_TITLE_CHAR_COUNT
              ? headerTitle.slice(0, SPARK_LINE_CHART_TITLE_CHAR_COUNT) + '...'
              : headerTitle}
          </Text>
        </div>
      </Tooltip>

      <ControlledComponent controller={!smallFont}>
        <Text weight="bold" type="title" level={2} color="grey-2">
          {metricType ? (
            getFormattedKpiValue({ value: total, metricType })
          ) : (
            <NumFormat shortHand={total > 1000} number={total} />
          )}
        </Text>
      </ControlledComponent>

      <ControlledComponent controller={smallFont}>
        <Text weight="bold" type="title" level={3} color="grey-2">
          {metricType ? (
            getFormattedKpiValue({ value: total, metricType })
          ) : (
            <NumFormat shortHand={total > 1000} number={total} />
          )}
        </Text>
      </ControlledComponent>

      {comparisonApplied && (
        <div className="flex flex-col row-gap-1 items-center">
          <div className="flex col-gap-1 items-center">
            {changeIcon}
            <Text
              level={7}
              type="title"
              color={percentChange < 0 ? 'red' : 'green'}
            >
              <NumFormat number={Math.abs(percentChange)} />%
            </Text>
          </div>
          <Text type="title" level={8} color="grey">
            {metricType ? (
              getFormattedKpiValue({ value: compareTotal, metricType })
            ) : (
              <NumFormat
                number={compareTotal}
                shortHand={compareTotal > 1000}
              />
            )}{' '}
            in prev. period
          </Text>
        </div>
      )}
    </div>
  );
}

export default ChartHeader;
