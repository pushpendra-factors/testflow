import React, { useState, useEffect, useContext } from 'react';
import { formatData } from './utils';
import BarLineChart from '../../../components/BarLineChart';
import AttributionTable from './AttributionTable';
import { DASHBOARD_MODAL } from '../../../utils/constants';
import { CoreQueryContext } from '../../../contexts/CoreQueryContext';

function AttributionsChart({
  data,
  event,
  attribution_method,
  touchpoint,
  linkedEvents,
  section,
  data2,
  durationObj,
  cmprDuration,
}) {
  const maxAllowedVisibleProperties = 5;
  const [chartsData, setChartsData] = useState([]);
  const [visibleIndices, setVisibleIndices] = useState(
    Array.from(Array(maxAllowedVisibleProperties).keys())
  );
  const { attributionMetrics, setAttributionMetrics } = useContext(
    CoreQueryContext
  );

  useEffect(() => {
    const formattedData = formatData(data, event, visibleIndices, touchpoint);
    setChartsData(formattedData);
  }, [data, event, visibleIndices, touchpoint]);

  if (!chartsData.length) {
    return null;
  }

  const legends = [
    `Conversions as Unique users (${attribution_method})`,
    'Cost per conversion',
  ];

  return (
    <div className='flex items-center justify-center flex-col'>
      <BarLineChart
        responseRows={data.rows}
        responseHeaders={data.headers}
        chartData={chartsData}
        visibleIndices={visibleIndices}
        section={section}
        legends={legends}
      />
      <div className='mt-12 w-full'>
        <AttributionTable
          linkedEvents={linkedEvents}
          touchpoint={touchpoint}
          event={event}
          data={data}
          data2={data2}
          durationObj={durationObj}
          cmprDuration={cmprDuration}
          isWidgetModal={section === DASHBOARD_MODAL}
          visibleIndices={visibleIndices}
          setVisibleIndices={setVisibleIndices}
          maxAllowedVisibleProperties={maxAllowedVisibleProperties}
          attribution_method={attribution_method}
          attributionMetrics={attributionMetrics}
          setAttributionMetrics={setAttributionMetrics}
          section={section}
        />
      </div>
    </div>
  );
}

export default AttributionsChart;
