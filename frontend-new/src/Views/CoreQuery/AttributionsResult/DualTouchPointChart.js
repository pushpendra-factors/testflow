import React, { useMemo } from 'react';
import GroupedBarChart from '../../../components/GroupedBarChart';
import { ATTRIBUTION_METHODOLOGY, REPORT_SECTION } from '../../../utils/constants';

const DualTouchPointChart = ({
  attribution_method,
  attribution_method_compare,
  currMetricsValue,
  chartsData,
  visibleIndices,
  data,
  event
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
    return ['#4D7DB4', '#4CBCBD'];
  };

  let legends, tooltipTitle;
  if (currMetricsValue) {
    legends = [
      `Cost Per Conversion (${attributionMethodsMapper[attribution_method]})`,
      `Cost Per Conversion (${attributionMethodsMapper[attribution_method_compare]})`,
    ];
    tooltipTitle = 'Cost Per Conversion';
  } else {
    legends = [
      `Conversions as Unique users (${attributionMethodsMapper[attribution_method]})`,
      `Conversions as Unique users (${attributionMethodsMapper[attribution_method_compare]})`,
    ];
    tooltipTitle = 'Conversions';
  }
  
  return (
    <GroupedBarChart
      colors={getColors()}
      chartData={chartsData}
      visibleIndices={visibleIndices}
      responseRows={data.rows}
      responseHeaders={data.headers}
      method1={attribution_method}
      method2={attribution_method_compare}
      event={event}
      section={REPORT_SECTION}
      allValues={allValues}
      legends={legends}
      tooltipTitle={tooltipTitle}
      attributionMethodsMapper={attributionMethodsMapper}
    />
  );
};

export default DualTouchPointChart;
