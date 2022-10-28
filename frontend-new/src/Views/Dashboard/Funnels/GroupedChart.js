import React, { useEffect, useState, useContext } from 'react';
import cx from 'classnames';
import { formatData } from '../../CoreQuery/FunnelsResultPage/utils';
import Chart from '../../CoreQuery/FunnelsResultPage/GroupedChart/Chart';
import FunnelsResultTable from '../../CoreQuery/FunnelsResultPage/FunnelsResultTable';
import { DashboardContext } from '../../../contexts/DashboardContext';
import {
  MAX_ALLOWED_VISIBLE_PROPERTIES,
  CHART_TYPE_BARCHART,
  CHART_TYPE_SCATTER_PLOT,
  DASHBOARD_WIDGET_SECTION,
  DASHBOARD_WIDGET_SCATTERPLOT_CHART_HEIGHT,
  CHART_TYPE_TABLE
} from '../../../utils/constants';
import NoDataChart from '../../../components/NoDataChart';
import FunnelsScatterPlot from '../../CoreQuery/FunnelsResultPage/GroupedChart/FunnelsScatterPlot';

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
  const { handleEditQuery } = useContext(DashboardContext);
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
      <div className="flex justify-center items-center w-full h-full pt-4 pb-4">
        <NoDataChart />
      </div>
    );
  }

  let chartContent = null;

  if (chartType === CHART_TYPE_BARCHART) {
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
  } else if (chartType === CHART_TYPE_SCATTER_PLOT) {
    chartContent = (
      <div className="mt-2">
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
      />
    );
  }

  return (
    <div
      className={cx('w-full flex-1', {
        'p-2 flex items-center': chartType !== CHART_TYPE_TABLE,
        'overflow-scroll': chartType === CHART_TYPE_TABLE
      })}
    >
      {chartContent}
    </div>
  );
}

export default GroupedChart;
