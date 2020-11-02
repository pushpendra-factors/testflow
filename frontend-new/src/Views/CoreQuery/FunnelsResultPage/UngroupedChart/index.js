import React, { useEffect, useState } from 'react';
import { generateUngroupedChartsData } from '../utils';
import Header from '../../../AppLayout/Header';
import EventsInfo from '../EventsInfo';
import Chart from './Chart';
import FunnelsResultTable from '../FunnelsResultTable';
import ResultsHeader from '../../ResultsHeader';

function UngroupedChart({
  resultState, queries, setDrawerVisible, eventsMapper, requestQuery, setShowResult, querySaved, setQuerySaved
}) {
  const [chartData, setChartData] = useState([]);

  useEffect(() => {
    const formattedData = generateUngroupedChartsData(resultState.data, queries);
    setChartData(formattedData);
  }, [queries, resultState.data]);

  if (!chartData.length) {
    return null;
  }

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
        />

        <div className="mt-8">
          <FunnelsResultTable
            chartData={chartData}
            breakdown={[]}
            queries={queries}
            groups={[]}
            eventsMapper={eventsMapper}
          />
        </div>
      </div>
    </>
  );
}

export default UngroupedChart;
