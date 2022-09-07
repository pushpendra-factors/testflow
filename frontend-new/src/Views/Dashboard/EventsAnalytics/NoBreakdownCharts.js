import React, { useState, useMemo, useContext } from 'react';
import cx from 'classnames';
import { useSelector } from 'react-redux';
import {
  formatData,
  getDataInLineChartFormat,
  getDefaultSortProp,
  getDefaultDateSortProp
} from '../../CoreQuery/EventsAnalytics/NoBreakdownCharts/utils';
import NoBreakdownTable from '../../CoreQuery/EventsAnalytics/NoBreakdownCharts/NoBreakdownTable';
import SparkLineChart from '../../../components/SparkLineChart';
import LineChart from '../../../components/HCLineChart';
import { generateColors } from '../../../utils/dataFormatter';
import {
  CHART_TYPE_TABLE,
  CHART_TYPE_SPARKLINES,
  DASHBOARD_WIDGET_AREA_CHART_HEIGHT,
  CHART_TYPE_LINECHART
} from '../../../utils/constants';
import NoDataChart from '../../../components/NoDataChart';
import { DashboardContext } from '../../../contexts/DashboardContext';

function NoBreakdownCharts({
  queries,
  resultState,
  page,
  chartType,
  durationObj,
  arrayMapper,
  unit,
  section
}) {
  const [hiddenEvents, setHiddenEvents] = useState([]);
  const { eventNames } = useSelector((state) => state.coreQuery);
  const { handleEditQuery } = useContext(DashboardContext);
  const appliedColors = generateColors(queries.length);
  const [sorter, setSorter] = useState(getDefaultSortProp(arrayMapper));
  const [dateSorter, setDateSorter] = useState(getDefaultDateSortProp());

  let chartsData = [];

  if (resultState.data && !resultState.data.metrics.rows.length) {
    chartsData = [];
  } else {
    chartsData = formatData(resultState.data, arrayMapper);
  }

  const { categories, data } = useMemo(() => {
    if (chartType === CHART_TYPE_LINECHART) {
      return getDataInLineChartFormat(
        resultState.data,
        arrayMapper,
        eventNames
      );
    }
    return {
      categories: [],
      data: []
    };
  }, [resultState.data, arrayMapper, eventNames, chartType]);

  const visibleSeriesData = useMemo(() => {
    return data.map((elem, index) => {
      const color = appliedColors[index];
      return {
        ...elem,
        color
      };
    });
  }, [data, appliedColors]);

  if (!chartsData.length) {
    return (
      <div className="flex justify-center items-center w-full h-full pt-4 pb-4">
        <NoDataChart />
      </div>
    );
  }

  let chartContent = null;

  if (chartType === CHART_TYPE_SPARKLINES) {
    chartContent = (
      <SparkLineChart
        frequency={durationObj.frequency}
        queries={queries}
        chartsData={chartsData}
        appliedColors={appliedColors}
        arrayMapper={arrayMapper}
        page={page}
        resultState={resultState}
        cardSize={unit.cardSize}
        title={unit.id}
        height={queries.length === 1 && unit.cardSize === 1 ? 180 : 100}
        section={section}
      />
    );
  }

  if (chartType === CHART_TYPE_TABLE) {
    chartContent = (
      <NoBreakdownTable
        data={chartsData}
        events={queries}
        chartType={chartType}
        setHiddenEvents={setHiddenEvents}
        hiddenEvents={hiddenEvents}
        isWidgetModal={false}
        durationObj={durationObj}
        arrayMapper={arrayMapper}
        sorter={sorter}
        setSorter={setSorter}
        dateSorter={dateSorter}
        setDateSorter={setDateSorter}
        responseData={resultState.data}
        section={section}
      />
    );
  }

  if (chartType === CHART_TYPE_LINECHART) {
    chartContent = (
      <LineChart
        frequency={durationObj.frequency}
        categories={categories}
        data={visibleSeriesData}
        height={DASHBOARD_WIDGET_AREA_CHART_HEIGHT}
        legendsPosition="top"
        cardSize={unit.cardSize}
        chartId={`line-${unit.id}`}
      />
    );
  }

  return (
    <div
      className={cx('w-full flex-1', {
        'p-2': chartType !== CHART_TYPE_TABLE
      })}
    >
      {chartContent}
    </div>
  );
}

export default NoBreakdownCharts;
