import React, { useContext, useState, useEffect } from 'react';
import {
  formatData,
  formatDataInHighChartsSeriesFormat
} from '../../CoreQuery/CampaignAnalytics/NoBreakdownCharts/utils';
import ChartHeader from '../../../components/SparkLineChart/ChartHeader';
import SparkChart from '../../../components/SparkLineChart/Chart';
import { generateColors, isSeriesChart } from '../../../utils/dataFormatter';
import LineChart from '../../../components/HCLineChart';
import NoBreakdownTable from '../../CoreQuery/CampaignAnalytics/NoBreakdownCharts/NoBreakdownTable';
import {
  CHART_TYPE_SPARKLINES,
  CHART_TYPE_LINECHART,
  CHART_TYPE_TABLE,
  DASHBOARD_WIDGET_AREA_CHART_HEIGHT
} from '../../../utils/constants';
import NoDataChart from '../../../components/NoDataChart';
import {
  Text,
  Number as NumFormat
} from '../../../components/factorsComponents';
import { DashboardContext } from '../../../contexts/DashboardContext';

function NoBreakdownCharts({
  chartType,
  data,
  arrayMapper,
  isWidgetModal,
  unit,
  durationObj
}) {
  const { handleEditQuery } = useContext(DashboardContext);
  const [chartsData, setChartsData] = useState([]);
  const [categories, setCategories] = useState([]);
  const [seriesData, setSeriesData] = useState([]);

  useEffect(() => {
    setChartsData(formatData(data, arrayMapper));

    const { categories: cat, seriesData: sd } = isSeriesChart(chartType)
      ? formatDataInHighChartsSeriesFormat(data, arrayMapper)
      : { categories: [], seriesData: [] };
    setCategories(cat);
    setSeriesData(sd);
  }, [data, arrayMapper, chartType]);

  if (!chartsData.length) {
    return (
      <div className="flex justify-center items-center w-full h-full pt-4 pb-4">
        <NoDataChart />
      </div>
    );
  }

  let tableContent = null;

  // if (chartType === CHART_TYPE_TABLE) {
  //   tableContent = (
  //     <div
  //       onClick={handleEditQuery}
  //       style={{ color: '#5949BC' }}
  //       className="mt-3 font-medium text-base cursor-pointer flex justify-end item-center"
  //     >
  //       Show More &rarr;
  //     </div>
  //   );
  // }

  let chartContent = null;

  if (chartType === CHART_TYPE_SPARKLINES) {
    if (chartsData.length === 1) {
      chartContent = (
        <div
          className={`flex items-center flex-wrap justify-center ${
            unit.cardSize !== 1 ? 'flex-col' : ''
          }`}
        >
          <div className={unit.cardSize === 1 ? 'w-1/4' : 'w-full'}>
            <ChartHeader
              bgColor="#4D7DB4"
              query={chartsData[0].name}
              total={chartsData[0].total}
            />
          </div>
          <div className={unit.cardSize === 1 ? 'w-3/4' : 'w-full'}>
            <SparkChart
              frequency="date"
              page="campaigns"
              event={chartsData[0].mapper}
              chartData={chartsData[0].dataOverTime}
              chartColor="#4D7DB4"
              height={unit.cardSize === 1 ? 180 : 100}
              title={unit.id}
            />
          </div>
        </div>
      );
    }

    if (chartsData.length > 1) {
      const appliedColors = generateColors(chartsData.length);
      const colors = {};
      arrayMapper.forEach((elem, index) => {
        colors[elem.mapper] = appliedColors[index];
      });
      chartContent = (
        <div
          className={`flex items-center flex-wrap justify-center ${
            !unit.cardSize ? 'flex-col' : ''
          }`}
        >
          {chartsData.slice(0, 3).map((chartData, index) => {
            if (unit.cardSize === 0) {
              return (
                <div className="flex items-center w-full justify-center">
                  <Text
                    extraClass="flex items-center w-1/4 justify-center"
                    type={'title'}
                    level={3}
                    weight={'bold'}
                  >
                    <NumFormat shortHand={true} number={chartData.total} />
                  </Text>
                  <div className="w-2/3">
                    <SparkChart
                      frequency="date"
                      page="campaigns"
                      event={chartData.mapper}
                      chartData={chartData.dataOverTime}
                      chartColor={appliedColors[index]}
                      height={40}
                      title={unit.id}
                    />
                  </div>
                </div>
              );
            } else if (unit.cardSize === 1) {
              return (
                <div
                  style={{ minWidth: '300px' }}
                  key={chartData.index}
                  className="w-1/3 mt-4 px-1"
                >
                  <div className="flex flex-col">
                    <ChartHeader
                      total={chartData.total}
                      query={chartData.name}
                      bgColor={appliedColors[index]}
                    />
                    <div className="mt-4">
                      <SparkChart
                        frequency="date"
                        page="campaigns"
                        event={chartData.mapper}
                        chartData={chartData.dataOverTime}
                        chartColor={appliedColors[index]}
                        height={100}
                        title={unit.id}
                      />
                    </div>
                  </div>
                </div>
              );
            } else {
              return (
                <div
                  style={{ minWidth: '300px' }}
                  key={chartData.index}
                  className="w-1/3 mt-6 px-1"
                >
                  <div className="flex flex-col">
                    <ChartHeader
                      total={chartData.total}
                      query={chartData.name}
                      bgColor={appliedColors[index]}
                      smallFont={true}
                    />
                  </div>
                </div>
              );
            }
          })}
        </div>
      );
    }
  } else if (chartType === CHART_TYPE_LINECHART) {
    chartContent = (
      <LineChart
        frequency="date"
        categories={categories}
        data={seriesData}
        height={DASHBOARD_WIDGET_AREA_CHART_HEIGHT}
        legendsPosition="top"
        cardSize={unit.cardSize}
        chartId={`line-${unit.id}`}
      />
    );
  } else {
    chartContent = (
      <NoBreakdownTable
        chartType={chartType}
        chartsData={chartsData}
        isWidgetModal={isWidgetModal}
        frequency={durationObj.frequency}
      />
    );
  }

  return (
    <div className={'w-full'}>
      {chartContent}
      {tableContent}
    </div>
  );
}

export default NoBreakdownCharts;
