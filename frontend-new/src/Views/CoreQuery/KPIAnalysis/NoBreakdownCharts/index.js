import React, {
  useState,
  useEffect,
  useCallback,
  forwardRef,
  useImperativeHandle,
  useContext,
  memo,
} from 'react';
import { CoreQueryContext } from '../../../../contexts/CoreQueryContext';
import {
  getDefaultDateSortProp,
  getDefaultSortProp,
  formatData,
  formatDataInSeriesFormat,
} from './utils';
import NoDataChart from '../../../../components/NoDataChart';
import {
  generateColors,
  getNewSorterState,
} from '../../../../utils/dataFormatter';
import {
  CHART_TYPE_SPARKLINES,
  CHART_TYPE_LINECHART,
} from '../../../../utils/constants';
import ChartHeader from '../../../../components/SparkLineChart/ChartHeader';
import SparkChart from '../../../../components/SparkLineChart/Chart';
import LineChart from '../../../../components/HCLineChart';
import NoBreakdownTable from './NoBreakdownTable';
import _ from 'lodash';

const NoBreakdownCharts = forwardRef(
  (
    { queries, responseData, chartType, durationObj, title = 'Kpi', section },
    ref
  ) => {
    const {
      coreQueryState: { savedQuerySettings },
    } = useContext(CoreQueryContext);

    const [sorter, setSorter] = useState(
      savedQuerySettings.sorter && Array.isArray(savedQuerySettings.sorter)
        ? savedQuerySettings.sorter
        : getDefaultSortProp(queries)
    );
    const [dateSorter, setDateSorter] = useState(
      savedQuerySettings.dateSorter &&
        Array.isArray(savedQuerySettings.dateSorter)
        ? savedQuerySettings.dateSorter
        : getDefaultDateSortProp()
    );
    const [aggregateData, setAggregateData] = useState([]);
    const [categories, setCategories] = useState([]);
    const [data, setData] = useState([]);

    const handleSorting = useCallback((prop) => {
      setSorter((currentSorter) => {
        return getNewSorterState(currentSorter, prop);
      });
    }, []);

    const handleDateSorting = useCallback((prop) => {
      setDateSorter((currentSorter) => {
        return getNewSorterState(currentSorter, prop);
      });
    }, []);

    useImperativeHandle(ref, () => {
      return {
        currentSorter: { sorter, dateSorter },
      };
    });

    useEffect(() => {
      const aggData = formatData(responseData, queries);
      const { categories: cats, data: d } = formatDataInSeriesFormat(aggData);
      setAggregateData(aggData);
      setCategories(cats);
      setData(d);
    }, [responseData, queries]);

    if (!aggregateData.length) {
      return (
        <div className='mt-4 flex justify-center items-center w-full h-64 '>
          <NoDataChart />
        </div>
      );
    }

    let chart = null;
    const table = (
      <div className='mt-12 w-full'>
        <NoBreakdownTable
          data={aggregateData}
          seriesData={data}
          section={section}
          chartType={chartType}
          frequency={durationObj.frequency}
          categories={categories}
          sorter={sorter}
          handleSorting={handleSorting}
          dateSorter={dateSorter}
          handleDateSorting={handleDateSorting}
          queries={queries}
        />
      </div>
    );

    if (chartType === CHART_TYPE_SPARKLINES) {
      if (aggregateData.length === 1) {
        chart = (
          <div className='flex items-center justify-center w-full'>
            <div className='w-1/4'>
              <ChartHeader
                bgColor='#4D7DB4'
                query={aggregateData[0].name}
                total={_.round(aggregateData[0].total, 1)}
              />
            </div>
            <div className='w-3/4'>
              <SparkChart
                frequency={durationObj.frequency}
                page='kpi'
                event={aggregateData[0].name}
                chartData={aggregateData[0].dataOverTime}
                chartColor='#4D7DB4'
              />
            </div>
          </div>
        );
      }

      if (aggregateData.length > 1) {
        const appliedColors = generateColors(aggregateData.length);
        chart = (
          <div className='flex items-center flex-wrap justify-center w-full'>
            {aggregateData
              .filter((d) => d.total)
              .map((chartData, index) => {
                return (
                  <div
                    style={{ minWidth: '300px' }}
                    key={chartData.index}
                    className='w-1/3 mt-4 px-4'
                  >
                    <div className='flex flex-col'>
                      <ChartHeader
                        total={chartData.total}
                        query={chartData.name}
                        bgColor={appliedColors[index]}
                      />
                      <div className='mt-8'>
                        <SparkChart
                          frequency={durationObj.frequency}
                          page='kpi'
                          event={chartData.name}
                          chartData={chartData.dataOverTime}
                          chartColor={appliedColors[index]}
                        />
                      </div>
                    </div>
                  </div>
                );
              })}
          </div>
        );
      }
    } else if (chartType === CHART_TYPE_LINECHART) {
      chart = (
        <div className='w-full'>
          <LineChart
            frequency={durationObj.frequency}
            categories={categories}
            data={data}
            showAllLegends={true}
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

export default memo(NoBreakdownCharts);
