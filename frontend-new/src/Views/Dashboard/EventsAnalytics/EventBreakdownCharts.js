import React, { useState, useEffect, useContext } from 'react';
import { useSelector } from 'react-redux';
import cx from 'classnames';
import {
  formatData,
  getDefaultSortProp
} from '../../CoreQuery/EventsAnalytics/EventBreakdown/utils';
import BarChart from '../../../components/BarChart';
import EventBreakdownTable from '../../CoreQuery/EventsAnalytics/EventBreakdown/EventBreakdownTable';
import ChartHeader from '../../../components/SparkLineChart/ChartHeader';
import {
  CHART_TYPE_TABLE,
  CHART_TYPE_BARCHART,
  CHART_TYPE_SPARKLINES,
  DASHBOARD_WIDGET_BAR_CHART_HEIGHT,
  MAX_ALLOWED_VISIBLE_PROPERTIES
} from '../../../utils/constants';
import NoDataChart from '../../../components/NoDataChart';
import { DashboardContext } from '../../../contexts/DashboardContext';
import { CHART_COLOR_1 } from '../../../constants/color.constants';

function EventBreakdownCharts({
  resultState,
  breakdown,
  section,
  chartType,
  unit,
  durationObj
}) {
  const [chartsData, setChartsData] = useState([]);
  const [visibleProperties, setVisibleProperties] = useState([]);
  const { handleEditQuery } = useContext(DashboardContext);
  const [sorter, setSorter] = useState(getDefaultSortProp());
  const { eventNames } = useSelector((state) => state.coreQuery);

  useEffect(() => {
    const formattedData = formatData(resultState.data);
    setChartsData(formattedData);
    setVisibleProperties([
      ...formattedData.slice(0, MAX_ALLOWED_VISIBLE_PROPERTIES)
    ]);
  }, [resultState.data]);

  if (!chartsData.length) {
    return (
      <div className='flex justify-center items-center w-full h-full pt-4 pb-4'>
        <NoDataChart />
      </div>
    );
  }

  let chartContent = null;

  if (chartType === CHART_TYPE_BARCHART) {
    chartContent = (
      <BarChart
        height={DASHBOARD_WIDGET_BAR_CHART_HEIGHT}
        title={unit.id}
        section={section}
        chartData={visibleProperties}
        cardSize={unit.cardSize}
      />
    );
  } else if (chartType === CHART_TYPE_SPARKLINES) {
    chartContent = (
      <ChartHeader
        total={resultState.data.rows[0]}
        query={'Count'}
        bgColor={CHART_COLOR_1}
        eventNames={eventNames}
      />
    );
  } else {
    chartContent = (
      <EventBreakdownTable
        data={chartsData}
        breakdown={breakdown}
        setVisibleProperties={setVisibleProperties}
        visibleProperties={visibleProperties}
        sorter={sorter}
        setSorter={setSorter}
        durationObj={durationObj}
        resultState={resultState}
      />
    );
  }

  return (
    <div
      className={cx('w-full flex-1', {
        'h-full flex justify-center items-center':
          chartType !== CHART_TYPE_TABLE
      })}
    >
      {chartContent}
    </div>
  );
}

export default EventBreakdownCharts;
