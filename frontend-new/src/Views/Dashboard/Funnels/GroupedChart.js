import React, { useEffect, useState } from 'react';
import cx from 'classnames';
import { formatData } from '../../CoreQuery/FunnelsResultPage/utils';
import Chart from '../../CoreQuery/FunnelsResultPage/GroupedChart/Chart';
import FunnelsResultTable from '../../CoreQuery/FunnelsResultPage/FunnelsResultTable';
import {
  MAX_ALLOWED_VISIBLE_PROPERTIES,
  CHART_TYPE_BARCHART,
  CHART_TYPE_SCATTER_PLOT,
  DASHBOARD_WIDGET_SECTION,
  DASHBOARD_WIDGET_SCATTERPLOT_CHART_HEIGHT,
  CHART_TYPE_TABLE,
  CHART_TYPE_METRIC_CHART,
  CHART_TYPE_FUNNEL_CHART,
  METRIC_TYPES,
  DASHBOARD_WIDGET_BAR_CHART_HEIGHT,
  CHART_TYPE_HORIZONTAL_BAR_CHART
} from '../../../utils/constants';
import NoDataChart from '../../../components/NoDataChart';
import FunnelsScatterPlot from '../../CoreQuery/FunnelsResultPage/GroupedChart/FunnelsScatterPlot';
import MetricChart from 'Components/MetricChart/MetricChart';
import { generateColors } from 'Utils/dataFormatter';
import ColumnChart from '../../../components/ColumnChart/ColumnChart';
import HorizontalBarChart from '../../../components/HorizontalBarChart';

const cardSizeToMetricCount = {
  0: 2,
  1: 3,
  2: 1
};

const colors = generateColors(MAX_ALLOWED_VISIBLE_PROPERTIES);

function GroupedChart({
  resultState,
  queries,
  arrayMapper,
  breakdown,
  chartType,
  unit,
  section
}) {
  const [visibleProperties, setVisibleProperties] = useState([]);
  const [eventsData, setEventsData] = useState([]);
  const [groups, setGroups] = useState([]);
  // const { handleEditQuery } = useContext(DashboardContext);
  const [sorter, setSorter] = useState([]);

  useEffect(() => {
    const { groups: appliedGroups, events } = formatData(
      {
        ...resultState.data,
        rows: resultState.data.rows.slice(0, MAX_ALLOWED_VISIBLE_PROPERTIES)
      },
      arrayMapper
    );
    setGroups(appliedGroups);
    setEventsData(events);
    setVisibleProperties([
      ...appliedGroups.slice(0, MAX_ALLOWED_VISIBLE_PROPERTIES)
    ]);
  }, [resultState.data, arrayMapper]);

  if (!groups.length) {
    return (
      <div className='flex justify-center items-center w-full h-full pt-4 pb-4'>
        <NoDataChart />
      </div>
    );
  }

  let chartContent = null;

  if (chartType === CHART_TYPE_FUNNEL_CHART) {
    chartContent = (
      <Chart
        groups={visibleProperties}
        eventsData={eventsData}
        title={unit.id}
        arrayMapper={arrayMapper}
        height={225}
        section={section}
        cardSize={unit.cardSize}
        durations={resultState.data.meta}
      />
    );
  } else if (chartType === CHART_TYPE_BARCHART) {
    chartContent = (
      <ColumnChart
        categories={visibleProperties.map((v) => v.name)}
        multiColored
        valueMetricType={METRIC_TYPES.percentType}
        height={DASHBOARD_WIDGET_BAR_CHART_HEIGHT}
        cardSize={unit.cardSize}
        series={[
          {
            name: 'OG',
            data: visibleProperties.map((v, index) =>
              Number(v.value.split('%')[0])
            )
          }
        ]}
      />
    );
  } else if (chartType === CHART_TYPE_HORIZONTAL_BAR_CHART) {
    chartContent = (
      <HorizontalBarChart
        height={DASHBOARD_WIDGET_BAR_CHART_HEIGHT}
        categories={visibleProperties.slice(0, 5).map((v) => v.name)}
        hideXAxis={true}
        series={[
          {
            name: 'OG',
            data: visibleProperties.slice(0, 5).map((v, index) => {
              return {
                y: Number(v.value.split('%')[0]),
                color: colors[index],
                metricType: METRIC_TYPES.percentType
              };
            })
          }
        ]}
      />
    );
  } else if (chartType === CHART_TYPE_SCATTER_PLOT) {
    chartContent = (
      <div className='mt-2'>
        <FunnelsScatterPlot
          visibleProperties={visibleProperties}
          arrayMapper={arrayMapper}
          section={DASHBOARD_WIDGET_SECTION}
          height={DASHBOARD_WIDGET_SCATTERPLOT_CHART_HEIGHT}
          cardSize={unit.cardSize}
          chartId={`funnels-scatterPlot-${unit.id}`}
        />
      </div>
    );
  } else if (chartType === CHART_TYPE_METRIC_CHART) {
    chartContent = (
      <div className='flex justify-between w-full col-gap-2'>
        {visibleProperties
          .slice(0, cardSizeToMetricCount[unit.cardSize])
          .map((elem, index) => {
            return (
              <MetricChart
                key={colors[index]}
                value={elem.value}
                iconColor={colors[index]}
                headerTitle={elem.name}
                valueType='percentage'
              />
            );
          })}
      </div>
    );
  } else {
    chartContent = (
      <FunnelsResultTable
        breakdown={breakdown}
        queries={queries}
        visibleProperties={visibleProperties}
        setVisibleProperties={setVisibleProperties}
        groups={groups}
        setGroups={setGroups}
        chartData={eventsData}
        arrayMapper={arrayMapper}
        durations={resultState.data.meta}
        resultData={resultState.data}
        sorter={sorter}
        setSorter={setSorter}
        isBreakdownApplied={true}
      />
    );
  }

  return (
    <div
      className={cx('w-full flex-1', {
        'px-2 flex items-center': chartType !== CHART_TYPE_TABLE
      })}
    >
      {chartContent}
    </div>
  );
}

export default GroupedChart;
