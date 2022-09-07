import React, {
  useState,
  useEffect,
  useCallback,
  useContext,
  memo
} from 'react';
import { useSelector } from 'react-redux';
import { values } from 'lodash';
import cx from 'classnames';
import { DashboardContext } from '../../../contexts/DashboardContext';
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
import {
  CHART_TYPE_SPARKLINES,
  CHART_TYPE_LINECHART,
  CHART_TYPE_TABLE,
  DASHBOARD_WIDGET_AREA_CHART_HEIGHT
} from '../../../utils/constants';
import ChartHeader from '../../../components/SparkLineChart/ChartHeader';
import SparkChart from '../../../components/SparkLineChart/Chart';
import LineChart from '../../../components/HCLineChart';
import {
  Text,
  Number as NumFormat
} from '../../../components/factorsComponents';
import NoBreakdownTable from '../../CoreQuery/KPIAnalysis/NoBreakdownCharts/NoBreakdownTable';
import TopLegends from '../../../components/GroupedBarChart/TopLegends';
import { getKpiLabel } from '../../CoreQuery/KPIAnalysis/kpiAnalysis.helpers';

const NoBreakdownCharts = ({
  kpis,
  responseData,
  chartType,
  section,
  unit,
  arrayMapper,
  durationObj
}) => {
  const { handleEditQuery } = useContext(DashboardContext);
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

  if (!aggregateData.length) {
    return (
      <div className="flex justify-center items-center w-full h-full pt-4 pb-4">
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
        <div
          className={`flex items-center flex-wrap justify-center ${
            unit.cardSize !== 1 ? 'flex-col' : ''
          }`}
        >
          <div className={unit.cardSize === 1 ? 'w-1/4' : 'w-full'}>
            <ChartHeader
              bgColor="#4D7DB4"
              query={aggregateData[0].name}
              total={aggregateData[0].total}
              metricType={aggregateData[0].metricType}
              eventNames={eventNames}
            />
          </div>
          <div className={unit.cardSize === 1 ? 'w-3/4' : 'w-full'}>
            <SparkChart
              frequency={durationObj.frequency}
              page="kpi"
              event={getKpiLabel(kpis[0])}
              chartData={aggregateData[0].dataOverTime}
              chartColor="#4D7DB4"
              height={unit.cardSize === 1 ? 220 : 100}
              title={unit.id}
              metricType={aggregateData[0].metricType}
              eventTitle={getKpiLabel(kpis[0])}
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
            <TopLegends
              cardSize={unit.cardSize}
              legends={aggregateData.map((c) => c.name)}
              colors={values(colors)}
            />
          ) : null}
          {aggregateData
            .filter((d) => d.total)
            .slice(0, 3)
            .map((chartData, index) => {
              if (unit.cardSize === 0) {
                return (
                  <div className="flex items-center w-full justify-center">
                    <Text
                      extraClass="flex items-center w-1/4 justify-center"
                      type={'title'}
                      level={3}
                      weight={'bold'}
                    >
                      <NumFormat
                        shortHand={chartData.total > 1000}
                        number={chartData.total}
                      />
                    </Text>
                    <div className="w-2/3">
                      <SparkChart
                        frequency={durationObj.frequency}
                        page="kpi"
                        event={chartData.name}
                        chartData={chartData.dataOverTime}
                        chartColor={appliedColors[index]}
                        height={40}
                        title={unit.id}
                        metricType={chartData.metricType}
                        eventTitle={chartData.name}
                      />
                    </div>
                  </div>
                );
              } else if (unit.cardSize === 1) {
                return (
                  <div
                    style={{ minWidth: '300px' }}
                    key={chartData.index}
                    className="w-1/3 mt-4 px-4"
                  >
                    <div className="flex flex-col">
                      <ChartHeader
                        total={chartData.total}
                        query={chartData.name}
                        bgColor={appliedColors[index]}
                        metricType={chartData.metricType}
                        eventNames={eventNames}
                      />
                      <div className="mt-8">
                        <SparkChart
                          frequency={durationObj.frequency}
                          page="kpi"
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
                        metricType={chartData.metricType}
                        eventNames={eventNames}
                      />
                    </div>
                  </div>
                );
              }
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
        legendsPosition="top"
        cardSize={unit.cardSize}
        chartId={`line-${unit.id}`}
      />
    );
  }

  return (
    <div
      className={cx('w-full flex-1', {
        'p-2': chartType !== CHART_TYPE_TABLE
      })}
    >
      {chartContent}
    </div>
  );
};

export default memo(NoBreakdownCharts);
