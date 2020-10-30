import React, { useEffect, useState } from 'react';
import { generateEventsData, generateGroups, generateGroupedChartsData } from '../utils';
import Header from '../../../AppLayout/Header';
import EventsInfo from '../EventsInfo';
import Chart from './Chart';
import FunnelsResultTable from '../FunnelsResultTable';
import ResultsHeader from '../../ResultsHeader';

function GroupedChart({
  resultState, queries, setDrawerVisible, breakdown, eventsMapper, reverseEventsMapper, requestQuery, setShowResult, querySaved, setQuerySaved
}) {
  const [groups, setGroups] = useState([]);
  const maxAllowedVisibleProperties = 5;

  useEffect(() => {
    const formattedGroups = generateGroups(resultState.data, maxAllowedVisibleProperties);
    setGroups(formattedGroups);
  }, [queries, resultState.data]);

  if (!groups.length) {
    return null;
  }

  const chartData = generateGroupedChartsData(resultState.data, queries, groups, eventsMapper);
  const eventsData = generateEventsData(resultState.data, queries, eventsMapper);

  return (
    <>
      <Header>
        <ResultsHeader
          setShowResult={setShowResult}
          requestQuery={requestQuery}
          querySaved={querySaved}
          setQuerySaved={setQuerySaved}
        />
        <div className="py-4">
          <EventsInfo setDrawerVisible={setDrawerVisible} queries={queries} />
        </div>
      </Header>

      <div className="mt-40 mb-8 fa-container">

        <Chart
          chartData={chartData}
          groups={groups.filter(elem => elem.is_visible)}
          eventsData={eventsData}
          eventsMapper={eventsMapper}
          reverseEventsMapper={reverseEventsMapper}
        />

        <div className="mt-8">
          <FunnelsResultTable
            breakdown={breakdown}
            queries={queries}
            groups={groups}
            setGroups={setGroups}
            chartData={eventsData}
            eventsMapper={eventsMapper}
            maxAllowedVisibleProperties={maxAllowedVisibleProperties}
          />
        </div>
      </div>
    </>
  );
}

export default GroupedChart;
