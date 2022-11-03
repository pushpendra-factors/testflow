import React, { useMemo } from 'react';
import GroupedBarChart from '../../../components/GroupedBarChart';
import {
  ATTRIBUTION_METHODOLOGY,
  REPORT_SECTION
} from '../../../utils/constants';
import {
  CHART_COLOR_1,
  CHART_COLOR_3
} from '../../../constants/color.constants';

const DualTouchPointChart = ({
  attribution_method,
  attribution_method_compare,
  currMetricsValue,
  chartsData,
  visibleIndices,
  data,
  event,
  cardSize = 1,
  chartId,
  height,
  section = REPORT_SECTION
}) => {
  const attributionMethodsMapper = useMemo(() => {
    const mapper = {};
    ATTRIBUTION_METHODOLOGY.forEach((am) => {
      mapper[am.value] = am.text;
    });
    return mapper;
  }, []);

  const allValues = [];

  chartsData.forEach((cd) => {
    allValues.push(cd[attribution_method]);
    allValues.push(allValues.push(cd[attribution_method_compare]));
  });

  const getColors = () => {
    return [CHART_COLOR_1, CHART_COLOR_3];
  };

  let legends, tooltipTitle;
  if (currMetricsValue) {
    legends = [
      `Cost Per Conversion (${attributionMethodsMapper[attribution_method]})`,
      `Cost Per Conversion (${attributionMethodsMapper[attribution_method_compare]})`
    ];
    tooltipTitle = 'Cost Per Conversion';
  } else {
    legends = [
      `Conversions as Unique users (${attributionMethodsMapper[attribution_method]})`,
      `Conversions as Unique users (${attributionMethodsMapper[attribution_method_compare]})`
    ];
    tooltipTitle = 'Conversions';
  }

  return (
    <GroupedBarChart
      colors={getColors()}
      chartData={chartsData}
      visibleIndices={visibleIndices}
      metricsData={data}
      method1={attribution_method}
      method2={attribution_method_compare}
      event={event}
      section={section}
      allValues={allValues}
      legends={legends}
      tooltipTitle={tooltipTitle}
      attributionMethodsMapper={attributionMethodsMapper}
      cardSize={cardSize}
      title={chartId}
      height={height}
    />
  );
};

export default DualTouchPointChart;
