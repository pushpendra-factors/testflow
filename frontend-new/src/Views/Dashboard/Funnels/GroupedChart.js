import React, { useEffect, useState, useContext } from 'react';
import { formatData } from '../../CoreQuery/FunnelsResultPage/utils';
import Chart from '../../CoreQuery/FunnelsResultPage/GroupedChart/Chart';
import FunnelsResultTable from '../../CoreQuery/FunnelsResultPage/FunnelsResultTable';
import { DashboardContext } from '../../../contexts/DashboardContext';
import { MAX_ALLOWED_VISIBLE_PROPERTIES } from '../../../utils/constants';
import NoDataChart from '../../../components/NoDataChart';

function GroupedChart({
  resultState,
  queries,
  arrayMapper,
  breakdown,
  chartType,
  unit,
  section,
}) {
  const [visibleProperties, setVisibleProperties] = useState([]);
  const [eventsData, setEventsData] = useState([]);
  const [groups, setGroups] = useState([]);
  const { handleEditQuery } = useContext(DashboardContext);

  useEffect(() => {
    const { groups: appliedGroups, events } = formatData(
      {
        ...resultState.data,
        rows: resultState.data.rows.slice(0, MAX_ALLOWED_VISIBLE_PROPERTIES),
      },
      arrayMapper
    );
    setGroups(appliedGroups);
    setEventsData(events);
    setVisibleProperties([
      ...appliedGroups.slice(0, MAX_ALLOWED_VISIBLE_PROPERTIES),
    ]);
  }, [resultState.data, arrayMapper]);

  if (!groups.length) {
    return (
      <div className='mt-4 flex justify-center items-center w-full h-full '>
        <NoDataChart />
      </div>
    );
  }

  let chartContent = null;

  if (chartType === 'barchart') {
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
      />
    );
  }

  let tableContent = null;

  if (chartType === 'table') {
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

  return (
    <div className={`w-full px-6 flex flex-1 flex-col  justify-center`}>
      {chartContent}
      {tableContent}
    </div>
  );
}

export default GroupedChart;
