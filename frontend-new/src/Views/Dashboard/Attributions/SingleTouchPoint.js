import React, { useState, useEffect, useContext } from 'react';
import { formatData } from '../../CoreQuery/AttributionsResult/utils';
import AttributionTable from '../../CoreQuery/AttributionsResult/AttributionTable';
import BarLineChart from '../../../components/BarLineChart';
import {
  CHART_TYPE_BARCHART,
  CHART_TYPE_TABLE,
  DASHBOARD_WIDGET_BARLINE_CHART_HEIGHT,
} from '../../../utils/constants';
import { DashboardContext } from '../../../contexts/DashboardContext';

function SingleTouchPoint({
  data,
  isWidgetModal,
  event,
  attribution_method,
  touchpoint,
  linkedEvents,
  chartType,
  resultState,
  unit,
  section,
}) {
  const maxAllowedVisibleProperties = 5;
  const [chartsData, setChartsData] = useState([]);
  const [visibleIndices, setVisibleIndices] = useState(
    Array.from(Array(maxAllowedVisibleProperties).keys())
  );
  const {
    attributionMetrics,
    setAttributionMetrics,
    handleEditQuery,
  } = useContext(DashboardContext);

  useEffect(() => {
    const formattedData = formatData(data, event, visibleIndices, touchpoint);
    setChartsData(formattedData);
  }, [data, event, visibleIndices, touchpoint]);

  if (!chartsData.length) {
    return null;
  }

  let chartContent = null;

  const legends = [
    `Conversions as Unique users (${attribution_method})`,
    'Cost per conversion',
  ];

  if (chartType === CHART_TYPE_BARCHART) {
    chartContent = (
      <BarLineChart
        responseRows={data.rows}
        responseHeaders={data.headers}
        visibleIndices={visibleIndices}
        title={unit.id}
        chartData={chartsData}
        section={section}
        height={DASHBOARD_WIDGET_BARLINE_CHART_HEIGHT}
        cardSize={unit.cardSize}
        legends={legends}
      />
    );
  } else {
    chartContent = (
      <AttributionTable
        linkedEvents={linkedEvents}
        touchpoint={touchpoint}
        event={event}
        data={data}
        isWidgetModal={isWidgetModal}
        visibleIndices={visibleIndices}
        setVisibleIndices={setVisibleIndices}
        maxAllowedVisibleProperties={maxAllowedVisibleProperties}
        attribution_method={attribution_method}
        attributionMetrics={attributionMetrics}
        setAttributionMetrics={setAttributionMetrics}
        section={section}
      />
    );
  }

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

  return (
    <div className={`w-full px-6 flex flex-1 flex-col  justify-center`}>
      {chartContent}
      {tableContent}
    </div>
  );
}

export default SingleTouchPoint;
