import React, {
  useState,
  useMemo,
  forwardRef,
  useContext,
  useImperativeHandle,
} from 'react';
import {
  formatData,
  getDataInLineChartFormat,
  getDefaultSortProp,
  getDefaultDateSortProp,
} from './utils';
import NoBreakdownTable from './NoBreakdownTable';
import SparkLineChart from '../../../../components/SparkLineChart';
import LineChart from '../../../../components/HCLineChart';
import { generateColors } from '../../../../utils/dataFormatter';
import {
  DASHBOARD_MODAL,
  CHART_TYPE_SPARKLINES,
  CHART_TYPE_LINECHART,
} from '../../../../utils/constants';
import { useSelector } from 'react-redux';
import { CoreQueryContext } from '../../../../contexts/CoreQueryContext';

const NoBreakdownCharts = forwardRef(
  (
    {
      queries,
      resultState,
      page,
      chartType,
      durationObj,
      arrayMapper,
      section,
    },
    ref
  ) => {
    const {
      coreQueryState: { savedQuerySettings },
    } = useContext(CoreQueryContext);

    const [sorter, setSorter] = useState(
      savedQuerySettings.sorter && Array.isArray(savedQuerySettings.sorter)
        ? savedQuerySettings.sorter
        : getDefaultSortProp(arrayMapper)
    );

    const [dateSorter, setDateSorter] = useState(
      savedQuerySettings.dateSorter &&
        Array.isArray(savedQuerySettings.dateSorter)
        ? savedQuerySettings.dateSorter
        : getDefaultDateSortProp()
    );

    useImperativeHandle(ref, () => {
      return {
        currentSorter: { sorter, dateSorter },
      };
    });

    const [hiddenEvents, setHiddenEvents] = useState([]);
    const { eventNames } = useSelector((state) => state.coreQuery);
    const appliedColors = useMemo(() => {
      return generateColors(queries.length);
    }, [queries]);

    const chartsData = useMemo(() => {
      return formatData(resultState.data, arrayMapper, queries.length);
    }, [resultState.data, arrayMapper, queries.length]);

    const { categories, data } = useMemo(() => {
      return getDataInLineChartFormat(
        resultState.data,
        arrayMapper,
        eventNames
      );
    }, [resultState.data, arrayMapper, eventNames]);

    const visibleSeriesData = useMemo(() => {
      return data
        .filter(
          (elem) => hiddenEvents.findIndex((he) => he === elem.name) === -1
        )
        .map((elem, index) => {
          const color = appliedColors[index];
          return {
            ...elem,
            color,
          };
        });
    }, [data, hiddenEvents, appliedColors]);

    if (!chartsData.length) {
      return null;
    }

    let chart = null;

    const table = (
      <div className='mt-12 w-full'>
        <NoBreakdownTable
          isWidgetModal={section === DASHBOARD_MODAL}
          data={chartsData}
          events={queries}
          chartType={chartType}
          setHiddenEvents={setHiddenEvents}
          hiddenEvents={hiddenEvents}
          durationObj={durationObj}
          arrayMapper={arrayMapper}
          sorter={sorter}
          setSorter={setSorter}
          dateSorter={dateSorter}
          setDateSorter={setDateSorter}
          responseData={resultState.data}
        />
      </div>
    );

    if (chartType === CHART_TYPE_SPARKLINES) {
      chart = (
        <SparkLineChart
          frequency={durationObj.frequency}
          queries={queries}
          chartsData={chartsData}
          appliedColors={appliedColors}
          arrayMapper={arrayMapper}
          page={page}
          resultState={resultState}
          section={section}
        />
      );
    } else if (chartType === CHART_TYPE_LINECHART) {
      chart = (
        <div className='w-full'>
          <LineChart
            frequency={durationObj.frequency}
            categories={categories}
            data={visibleSeriesData}
          />
        </div>
      );
    }

    return (
      <div className='flex items-center justify-center flex-col'>
        {chart}
        {table}
      </div>
    );
  }
);

export default NoBreakdownCharts;
