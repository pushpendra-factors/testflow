import React, { useState, useEffect } from 'react';
import { formatData, formatDataInHighChartsSeriesFormat } from './utils';
import ChartHeader from '../../../../components/SparkLineChart/ChartHeader';
import SparkChart from '../../../../components/SparkLineChart/Chart';
import { generateColors } from '../../../../utils/dataFormatter';
import { CHART_COLOR_1 } from '../../../../constants/color.constants';
import LineChart from '../../../../components/HCLineChart';
import NoBreakdownTable from './NoBreakdownTable';
import {
  CHART_TYPE_SPARKLINES,
  CHART_TYPE_LINECHART,
  DASHBOARD_MODAL
} from '../../../../utils/constants';
import NoDataChart from '../../../../components/NoDataChart';

function NoBreakdownCharts({
  chartType,
  data,
  arrayMapper,
  section,
  durationObj
}) {
  const [chartsData, setChartsData] = useState([]);
  const [categories, setCategories] = useState([]);
  const [seriesData, setSeriesData] = useState([]);

  useEffect(() => {
    setChartsData(formatData(data, arrayMapper));

    const { categories: cat, seriesData: sd } =
      formatDataInHighChartsSeriesFormat(data, arrayMapper);
    setCategories(cat);
    setSeriesData(sd);
  }, [data, arrayMapper]);

  if (!chartsData.length) {
    return (
      <div className='flex justify-center items-center w-full h-full pt-4 pb-4'>
        <NoDataChart />
      </div>
    );
  }

  const table = (
    <div className='mt-12 w-full'>
      <NoBreakdownTable
        chartType={chartType}
        chartsData={chartsData}
        isWidgetModal={section === DASHBOARD_MODAL}
        frequency={durationObj.frequency}
      />
    </div>
  );

  let chart = null;

  if (chartType === CHART_TYPE_SPARKLINES) {
    if (chartsData.length === 1) {
      chart = (
        <div className='flex items-center justify-center w-full'>
          <div className='w-1/4'>
            <ChartHeader
              bgColor={CHART_COLOR_1}
              query={chartsData[0].name}
              total={chartsData[0].total}
            />
          </div>
          <div className='w-3/4'>
            <SparkChart
              frequency='date'
              page='campaigns'
              event={chartsData[0].mapper}
              chartData={chartsData[0].dataOverTime}
              chartColor={CHART_COLOR_1}
            />
          </div>
        </div>
      );
    }

    if (chartsData.length > 1) {
      const appliedColors = generateColors(chartsData.length);
      chart = (
        <div className='flex items-center flex-wrap justify-center w-full'>
          {chartsData.map((chartData, index) => {
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
                      frequency='date'
                      page='campaigns'
                      event={chartData.mapper}
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
        <LineChart frequency='date' categories={categories} data={seriesData} />
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

export default NoBreakdownCharts;
