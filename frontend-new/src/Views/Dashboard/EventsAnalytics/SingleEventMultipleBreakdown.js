import React, { useState, useEffect, useMemo, useContext } from 'react';
import {
  formatData,
  formatDataInStackedAreaFormat,
} from '../../CoreQuery/EventsAnalytics/SingleEventMultipleBreakdown/utils';
import BarChart from '../../../components/BarChart';
import LineChart from '../../../components/HCLineChart';
import SingleEventMultipleBreakdownTable from '../../CoreQuery/EventsAnalytics/SingleEventMultipleBreakdown/SingleEventMultipleBreakdownTable';
import { generateColors } from '../../../utils/dataFormatter';
import {
  CHART_TYPE_TABLE,
  CHART_TYPE_BARCHART,
  DASHBOARD_WIDGET_BAR_CHART_HEIGHT,
  CHART_TYPE_STACKED_AREA,
  DASHBOARD_WIDGET_AREA_CHART_HEIGHT,
  CHART_TYPE_STACKED_BAR,
  MAX_ALLOWED_VISIBLE_PROPERTIES,
} from '../../../utils/constants';
import StackedAreaChart from '../../../components/StackedAreaChart';
import StackedBarChart from '../../../components/StackedBarChart';
import { DashboardContext } from '../../../contexts/DashboardContext';
import NoDataChart from '../../../components/NoDataChart';

function SingleEventMultipleBreakdown({
  resultState,
  page,
  chartType,
  breakdown,
  queries,
  unit,
  durationObj,
  section,
}) {
  const [visibleProperties, setVisibleProperties] = useState([]);
  const { handleEditQuery } = useContext(DashboardContext);

  const aggregateData = useMemo(() => {
    return formatData(resultState.data);
  }, [resultState.data]);

  const { categories, data } = useMemo(() => {
    if (chartType === CHART_TYPE_BARCHART) {
      return {
        categories: [],
        data: [],
      };
    }
    return formatDataInStackedAreaFormat(resultState.data, aggregateData);
  }, [resultState.data, aggregateData, chartType]);

  const visibleSeriesData = useMemo(() => {
    const colors = generateColors(visibleProperties.length);
    return data
      .filter(
        (elem) =>
          visibleProperties.findIndex((vp) => vp.index === elem.index) > -1
      )
      .map((elem, index) => {
        const color = colors[index];
        return {
          ...elem,
          color,
        };
      });
  }, [data, visibleProperties]);

  useEffect(() => {
    setVisibleProperties([
      ...aggregateData.slice(0, MAX_ALLOWED_VISIBLE_PROPERTIES),
    ]);
  }, [aggregateData]);

  if (!visibleProperties.length) {
    return <NoDataChart />;
  }

  let chartContent = null;

  let tableContent = null;

  if (chartType === CHART_TYPE_TABLE) {
    tableContent = (
      <div
        onClick={handleEditQuery}
        style={{ color: '#5949BC' }}
        className='mt-3 font-medium text-base cursor-pointer flex justify-end item-center'
      >
        Show More &rarr;
      </div>
    );
  }

  if (chartType === CHART_TYPE_BARCHART) {
    chartContent = (
      <div className='flex mt-4'>
        <BarChart
          chartData={visibleProperties}
          height={DASHBOARD_WIDGET_BAR_CHART_HEIGHT}
          title={unit.id}
          cardSize={unit.cardSize}
          section={section}
          queries={queries}
        />
      </div>
    );
  } else if (chartType === CHART_TYPE_TABLE) {
    chartContent = (
      <SingleEventMultipleBreakdownTable
        data={aggregateData}
        seriesData={data}
        breakdown={breakdown}
        events={queries}
        chartType={chartType}
        setVisibleProperties={setVisibleProperties}
        visibleProperties={visibleProperties}
        page={page}
        durationObj={durationObj}
        categories={categories}
        section={section}
      />
    );
  } else if (chartType === CHART_TYPE_STACKED_AREA) {
    chartContent = (
      <StackedAreaChart
        frequency={durationObj.frequency}
        categories={categories}
        data={visibleSeriesData}
        height={DASHBOARD_WIDGET_AREA_CHART_HEIGHT}
        legendsPosition='top'
        cardSize={unit.cardSize}
        chartId={`area-${unit.id}`}
      />
    );
  } else if (chartType === CHART_TYPE_STACKED_BAR) {
    chartContent = (
      <StackedBarChart
        frequency={durationObj.frequency}
        categories={categories}
        data={visibleSeriesData}
        height={DASHBOARD_WIDGET_AREA_CHART_HEIGHT}
        legendsPosition='top'
        cardSize={unit.cardSize}
        chartId={`bar-${unit.id}`}
      />
    );
  } else {
    chartContent = (
      <LineChart
        frequency={durationObj.frequency}
        categories={categories}
        data={visibleSeriesData}
        height={DASHBOARD_WIDGET_AREA_CHART_HEIGHT}
        legendsPosition='top'
        cardSize={unit.cardSize}
        chartId={`line-${unit.id}`}
      />
    );
  }

  return (
    <div className={`w-full px-6 flex flex-1 flex-col  justify-center`}>
      {chartContent}
      {tableContent}
    </div>
  );
}

export default SingleEventMultipleBreakdown;
