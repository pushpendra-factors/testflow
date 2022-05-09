import React, {
  useState,
  useEffect,
  useContext,
  useCallback,
  useImperativeHandle,
  forwardRef,
} from 'react';

import BarChart from 'Components/BarChart';
import NoDataChart from 'Components/NoDataChart';
import PivotTable from 'Components/PivotTable';

import {
  CHART_TYPE_BARCHART,
  CHART_TYPE_HORIZONTAL_BAR_CHART,
  CHART_TYPE_PIVOT_CHART,
  QUERY_TYPE_PROFILE,
} from 'Utils/constants';
import { getNewSorterState } from 'Utils/dataFormatter';

import { formatData, defaultSortProp, getVisibleData } from './utils';
import { CoreQueryContext } from '../../../../contexts/CoreQueryContext';
import BreakdownTable from './BreakdownTable';
import HorizontalBarChartTable from './HorizontalBarChartTable';

const BreakdownCharts = forwardRef(
  (
    {
      chartType,
      breakdown,
      data,
      title = 'Profile-chart',
      currentEventIndex,
      section,
      queries,
      groupAnalysis,
    },
    ref
  ) => {
    const {
      coreQueryState: { savedQuerySettings },
    } = useContext(CoreQueryContext);
    const [sorter, setSorter] = useState(
      savedQuerySettings.sorter && Array.isArray(savedQuerySettings.sorter)
        ? savedQuerySettings.sorter
        : defaultSortProp()
    );
    const [visibleProperties, setVisibleProperties] = useState([]);
    const [aggregateData, setAggregateData] = useState([]);

    const handleSorting = useCallback((prop) => {
      setSorter((currentSorter) => {
        return getNewSorterState(currentSorter, prop);
      });
    }, []);

    useImperativeHandle(ref, () => {
      return {
        currentSorter: { sorter },
      };
    });

    useEffect(() => {
      const aggData = formatData(data, breakdown, queries, currentEventIndex);
      setAggregateData(aggData);
    }, [data, breakdown, queries, currentEventIndex]);

    useEffect(() => {
      setVisibleProperties(getVisibleData(aggregateData, sorter));
    }, [aggregateData, sorter]);

    if (!aggregateData.length) {
      return (
        <div className='mt-4 flex justify-center items-center w-full h-64 '>
          <NoDataChart />
        </div>
      );
    }

    let chart = null;

    if (chartType === CHART_TYPE_BARCHART) {
      chart = (
        <BarChart
          section={section}
          title={title}
          chartData={visibleProperties}
        />
      );
    }

    if (chartType === CHART_TYPE_HORIZONTAL_BAR_CHART) {
      chart = (
        <div className='w-full'>
          <HorizontalBarChartTable
            aggregateData={aggregateData}
            breakdown={breakdown}
          />
        </div>
      );
    }

    if (chartType === CHART_TYPE_PIVOT_CHART) {
      chart = (
        <div className='w-full'>
          <PivotTable
            data={aggregateData}
            breakdown={breakdown}
            metrics={queries}
            queryType={QUERY_TYPE_PROFILE}
            groupAnalysis={groupAnalysis}
          />
        </div>
      );
    }

    const table = (
      <div className='mt-12 w-full'>
        <BreakdownTable
          aggregateData={aggregateData}
          sorter={sorter}
          breakdown={breakdown}
          currentEventIndex={currentEventIndex}
          chartType={chartType}
          sorter={sorter}
          handleSorting={handleSorting}
          visibleProperties={visibleProperties}
          isWidgetModal={false}
          setVisibleProperties={setVisibleProperties}
          section={section}
          queries={queries}
          groupAnalysis={groupAnalysis}
        />
      </div>
    );

    return (
      <div className='flex items-center justify-center flex-col'>
        {chart}
        {table}
      </div>
    );
  }
);

export default BreakdownCharts;
