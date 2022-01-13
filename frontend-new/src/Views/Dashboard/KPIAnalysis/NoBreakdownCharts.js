import React, {
  useState,
  useEffect,
  useCallback,
  useContext,
  memo,
} from 'react';
import { DashboardContext } from '../../../contexts/DashboardContext';
import {
  getDefaultDateSortProp,
  getDefaultSortProp,
  formatData,
  formatDataInSeriesFormat,
} from '../../CoreQuery/KPIAnalysis/NoBreakdownCharts/utils';
import NoDataChart from '../../../components/NoDataChart';
import {
  generateColors,
  getNewSorterState,
} from '../../../utils/dataFormatter';
import {
  CHART_TYPE_SPARKLINES,
  CHART_TYPE_LINECHART,
  CHART_TYPE_TABLE,
  DASHBOARD_WIDGET_AREA_CHART_HEIGHT,
} from '../../../utils/constants';
import ChartHeader from '../../../components/SparkLineChart/ChartHeader';
import SparkChart from '../../../components/SparkLineChart/Chart';
import LineChart from '../../../components/HCLineChart';
import {
  Text,
  Number as NumFormat,
} from '../../../components/factorsComponents';
import NoBreakdownTable from '../../CoreQuery/KPIAnalysis/NoBreakdownCharts/NoBreakdownTable';
import DashboardWidgetLegends from '../../../components/DashboardWidgetLegends';

const NoBreakdownCharts = ({
  kpis,
  responseData,
  chartType,
  section,
  unit,
  arrayMapper,
}) => {
  const { handleEditQuery } = useContext(DashboardContext);

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

  if (!aggregateData.length) {
    return (
      <div className='mt-4 flex justify-center items-center w-full h-64 '>
        <NoDataChart />
      </div>
    );
  }

  let tableContent = null;
  let chartContent = null;

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
        <div
          className={`flex items-center flex-wrap justify-center ${
            unit.cardSize !== 1 ? 'flex-col' : ''
          }`}
        >
          <div className={unit.cardSize === 1 ? 'w-1/4' : 'w-full'}>
            <ChartHeader
              bgColor='#4D7DB4'
              query={aggregateData[0].name}
              total={aggregateData[0].total}
            />
          </div>
          <div className={unit.cardSize === 1 ? 'w-3/4' : 'w-full'}>
            <SparkChart
              frequency='date'
              page='campaigns'
              event={aggregateData[0].name}
              chartData={aggregateData[0].dataOverTime}
              chartColor='#4D7DB4'
              height={unit.cardSize === 1 ? 180 : 100}
              title={unit.id}
            />
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
            !unit.cardSize ? 'flex-col' : ''
          }`}
        >
          {!unit.cardSize ? (
            <DashboardWidgetLegends
              arrayMapper={arrayMapper}
              cardSize={unit.cardSize}
              colors={colors}
              legends={aggregateData.map((c) => c.name)}
            />
          ) : null}
          {aggregateData
            .filter((d) => d.total)
            .slice(0, 3)
            .map((chartData, index) => {
              if (unit.cardSize === 0) {
                return (
                  <div className='flex items-center w-full justify-center'>
                    <Text
                      extraClass='flex items-center w-1/4 justify-center'
                      type={'title'}
                      level={3}
                      weight={'bold'}
                    >
                      <NumFormat shortHand={true} number={chartData.total} />
                    </Text>
                    <div className='w-2/3'>
                      <SparkChart
                        frequency='date'
                        page='kpi'
                        event={chartData.name}
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
                          page='kpi'
                          event={chartData.name}
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
                    className='w-1/3 mt-6 px-1'
                  >
                    <div className='flex flex-col'>
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
        frequency={'date'}
        categories={categories}
        data={data}
        height={DASHBOARD_WIDGET_AREA_CHART_HEIGHT}
        legendsPosition='top'
        cardSize={unit.cardSize}
        chartId={`line-${unit.id}`}
      />
    );
  }

  return (
    <div className={`w-full px-6 flex flex-1 flex-col  justify-center`}>
      {chartContent}
      {tableContent}
    </div>
  );
};

export default memo(NoBreakdownCharts);
