import React, { useEffect, useState } from 'react';
import { formatData } from '../utils';
import Chart from './Chart';
import FunnelsResultTable from '../FunnelsResultTable';
import {
  DASHBOARD_MODAL,
  MAX_ALLOWED_VISIBLE_PROPERTIES,
} from '../../../../utils/constants';
import NoDataChart from '../../../../components/NoDataChart';

function GroupedChart({
  resultState,
  queries,
  breakdown,
  isWidgetModal,
  arrayMapper,
  section,
}) {
  const [visibleProperties, setVisibleProperties] = useState([]);
  const [eventsData, setEventsData] = useState([]);
  const [groups, setGroups] = useState([]);

  useEffect(() => {
    const { groups: appliedGroups, events } = formatData(resultState.data, arrayMapper);
    setGroups(appliedGroups);
    setEventsData(events);
    setVisibleProperties([...appliedGroups.slice(0, MAX_ALLOWED_VISIBLE_PROPERTIES)]);
  }, [resultState.data, arrayMapper]);

  if (!visibleProperties.length) {
    return (
      <div className='mt-4 flex justify-center items-center w-full h-full'>
        <NoDataChart />
      </div>
    );
  }

  return (
    <div className='flex items-center justify-center flex-col'>
      <Chart
        isWidgetModal={isWidgetModal}
        groups={visibleProperties}
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
          visibleProperties={visibleProperties}
          setVisibleProperties={setVisibleProperties}
          chartData={eventsData}
          arrayMapper={arrayMapper}
          maxAllowedVisibleProperties={MAX_ALLOWED_VISIBLE_PROPERTIES}
          isWidgetModal={section === DASHBOARD_MODAL}
          durations={resultState.data.meta}
          resultData={resultState.data}
        />
      </div>
    </div>
  );
}

export default GroupedChart;
