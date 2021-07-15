import React, { useEffect, useState, useMemo } from 'react';
import {
  generateEventsData,
  generateGroups,
  generateGroupedChartsData,
} from '../utils';
import Chart from './Chart';
import FunnelsResultTable from '../FunnelsResultTable';
import { DASHBOARD_MODAL } from '../../../../utils/constants';
import NoDataChart from '../../../../components/NoDataChart';

function GroupedChart({
  resultState,
  queries,
  breakdown,
  isWidgetModal,
  arrayMapper,
  section,
}) {
  const [groups, setGroups] = useState([]);
  const maxAllowedVisibleProperties = 5;

  useEffect(() => {
    const formattedGroups = generateGroups(
      resultState.data,
      maxAllowedVisibleProperties,
      resultState.data.meta?.query?.gbp[0]?.grn
    );
    setGroups(formattedGroups);
  }, [resultState.data]);

  const chartData = useMemo(() => {
    if (groups.length) {
      return generateGroupedChartsData(
        resultState.data,
        queries,
        groups,
        arrayMapper,
        resultState.data.meta?.query?.gbp
      );
    }
  }, [resultState.data, groups, arrayMapper, queries]);

  const eventsData = useMemo(() => {
    if (groups.length) {
      return generateEventsData(resultState.data, queries, arrayMapper);
    }
  }, [resultState.data, queries, arrayMapper, groups]);

  if (!groups.length) {
    return (
      <div className='mt-4 flex justify-center items-center w-full h-full '>
        <NoDataChart />
      </div>
    );
  }

  return (
    <div className='flex items-center justify-center flex-col'>
      <Chart
        isWidgetModal={isWidgetModal}
        chartData={chartData}
        groups={groups.filter((elem) => elem.is_visible)}
        eventsData={eventsData}
        arrayMapper={arrayMapper}
        section={section}
        durations={resultState.data.meta}
      />

      <div className='mt-12 w-full'>
        <FunnelsResultTable
          breakdown={breakdown}
          queries={queries}
          groups={groups}
          setGroups={setGroups}
          chartData={eventsData}
          arrayMapper={arrayMapper}
          maxAllowedVisibleProperties={maxAllowedVisibleProperties}
          isWidgetModal={section === DASHBOARD_MODAL}
          durations={resultState.data.meta}
          resultData={resultState.data}
        />
      </div>
    </div>
  );
}

export default GroupedChart;
