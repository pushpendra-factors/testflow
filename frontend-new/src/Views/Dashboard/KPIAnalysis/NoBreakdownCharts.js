import React, { useState, useEffect, useCallback, memo, useMemo } from 'react';
import { useSelector } from 'react-redux';
import cx from 'classnames';
import {
  getDefaultDateSortProp,
  getDefaultSortProp,
  formatData,
  formatDataInSeriesFormat
} from '../../CoreQuery/KPIAnalysis/NoBreakdownCharts/utils';
import NoDataChart from '../../../components/NoDataChart';
import {
  generateColors,
  getNewSorterState
} from '../../../utils/dataFormatter';
import { CHART_COLOR_1 } from '../../../constants/color.constants';
import {
  CHART_TYPE_SPARKLINES,
  CHART_TYPE_LINECHART,
  CHART_TYPE_TABLE,
  DASHBOARD_WIDGET_AREA_CHART_HEIGHT,
  CHART_TYPE_BARCHART,
  DASHBOARD_WIDGET_BAR_CHART_HEIGHT
} from '../../../utils/constants';
import ChartHeader from '../../../components/SparkLineChart/ChartHeader';
import SparkChart from '../../../components/SparkLineChart/Chart';
import LineChart from '../../../components/HCLineChart';
import NoBreakdownTable from '../../CoreQuery/KPIAnalysis/NoBreakdownCharts/NoBreakdownTable';
import { getKpiLabel } from '../../CoreQuery/KPIAnalysis/kpiAnalysis.helpers';
import ColumnChart from 'Components/ColumnChart/ColumnChart';

function NoBreakdownCharts({
  kpis,
  responseData,
  chartType,
  section,
  unit,
  arrayMapper,
  durationObj,
  currentEventIndex
}) {
  const { eventNames } = useSelector((state) => state.coreQuery);

  const [sorter, setSorter] = useState(getDefaultSortProp(kpis));
  const [dateSorter, setDateSorter] = useState(getDefaultDateSortProp());
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

  useEffect(() => {
    const aggData = formatData(responseData, kpis);
    const { categories: cats, data: d } = formatDataInSeriesFormat(aggData);
    setAggregateData(aggData);
    setCategories(cats);
    setData(d);
  }, [responseData, kpis]);

  const columnChartSeries = useMemo(() => {
    return [
      {
        data: aggregateData[currentEventIndex]?.dataOverTime?.map(
          (elem) => elem[kpis[currentEventIndex].label]
        )
      }
    ];
  }, [aggregateData, currentEventIndex, kpis]);

  if (!aggregateData.length) {
    return (
      <div className='flex justify-center items-center w-full h-full pt-4 pb-4'>
        <NoDataChart />
      </div>
    );
  }

  let chartContent = null;

  if (chartType === CHART_TYPE_TABLE) {
    chartContent = (
      <NoBreakdownTable
        data={aggregateData}
        seriesData={data}
        section={section}
        chartType={chartType}
        // durationObj={durationObj}
        categories={categories}
        sorter={sorter}
        handleSorting={handleSorting}
        dateSorter={dateSorter}
        handleDateSorting={handleDateSorting}
        kpis={kpis}
      />
    );
  }

  if (chartType === CHART_TYPE_SPARKLINES) {
    if (aggregateData.length === 1) {
      chartContent = (
        <div className='flex items-center justify-center w-full  h-full'>
          <div
            className={`flex items-center justify-center  h-full ${
              unit?.cardSize === 2
                ? 'flex-col w-full'
                : unit?.cardSize === 0
                ? 'w-4/5'
                : 'w-3/5'
            }`}
          >
            <div className={`${unit?.cardSize === 2 ? 'w-full' : 'w-1/2'}`}>
              <ChartHeader
                bgColor={CHART_COLOR_1}
                query={aggregateData[0].name}
                total={aggregateData[0].total}
                metricType={aggregateData[0].metricType}
                eventNames={eventNames}
              />
            </div>
            <div
              className={`flex justify-center items-center ${
                unit.cardSize === 2 ? 'w-full' : 'w-1/2'
              }`}
            >
              <div className={`${unit?.cardSize === 2 ? 'w-3/5' : 'w-full'}`}>
                <SparkChart
                  frequency={durationObj.frequency}
                  page='kpi'
                  event={getKpiLabel(kpis[0])}
                  chartData={aggregateData[0].dataOverTime}
                  chartColor={CHART_COLOR_1}
                  height={unit.cardSize === 1 ? 220 : 100}
                  title={unit.id}
                  metricType={aggregateData[0].metricType}
                  eventTitle={getKpiLabel(kpis[0])}
                />
              </div>
            </div>
          </div>
        </div>
      );
    }

    if (aggregateData.length > 1) {
      const appliedColors = generateColors(aggregateData.length);
      const colors = {};
      arrayMapper.forEach((elem, index) => {
        colors[elem.mapper] = appliedColors[index];
      });
      chartContent = (
        <div
          className={`flex items-center flex-wrap justify-center ${
            unit?.cardSize !== 2 ? 'pt-4' : ''
          } `}
        >
          {aggregateData
            .filter((d) => d.total)
            .slice(0, 3)
            .map((chartData, index) => {
              if (unit.cardSize === 1 || unit.cardSize === 0) {
                return (
                  <div
                    // style={{ minWidth: '300px' }}
                    key={chartData.index}
                    className='w-1/3 px-4 h-full'
                  >
                    <div className='flex flex-col'>
                      <ChartHeader
                        total={chartData.total}
                        query={chartData.name}
                        bgColor={appliedColors[index]}
                        metricType={chartData.metricType}
                        eventNames={eventNames}
                        titleCharCount={unit?.cardSize === 0 ? 16 : null}
                      />
                      <div className='mt-8'>
                        <SparkChart
                          frequency={durationObj.frequency}
                          page='kpi'
                          event={chartData.name}
                          chartData={chartData.dataOverTime}
                          chartColor={appliedColors[index]}
                          height={100}
                          title={unit.id}
                          metricType={chartData.metricType}
                          eventTitle={chartData.name}
                        />
                      </div>
                    </div>
                  </div>
                );
              }
              return (
                <div
                  style={{ minWidth: '300px' }}
                  key={chartData.index}
                  className='w-1/3 mt-6 px-1'
                >
                  <div className='flex flex-col'>
                    <ChartHeader
                      total={chartData.total}
                      query={chartData.name}
                      bgColor={appliedColors[index]}
                      smallFont={true}
                      metricType={chartData.metricType}
                      eventNames={eventNames}
                    />
                  </div>
                </div>
              );
            })}
        </div>
      );
    }
  }

  if (chartType === CHART_TYPE_LINECHART) {
    chartContent = (
      <LineChart
        frequency={durationObj.frequency}
        categories={categories}
        data={data}
        height={DASHBOARD_WIDGET_AREA_CHART_HEIGHT}
        legendsPosition='top'
        cardSize={unit.cardSize}
        chartId={`line-${unit.id}`}
      />
    );
  } else if (chartType === CHART_TYPE_BARCHART) {
    chartContent = (
      <ColumnChart
        categories={categories}
        xAxisType='date-time'
        series={columnChartSeries}
        frequency={durationObj.frequency}
        height={DASHBOARD_WIDGET_BAR_CHART_HEIGHT}
        chartId={`kpi${unit.id}`}
        cardSize={unit.cardSize}
      />
    );
  }

  return (
    <div
      className={cx('w-full flex-1', {
        'px-2': chartType !== CHART_TYPE_TABLE
      })}
    >
      {chartContent}
    </div>
  );
}

export default memo(NoBreakdownCharts);
