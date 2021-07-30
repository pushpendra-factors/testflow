import React, { useState, useContext, useMemo, useCallback } from 'react';
import moment from 'moment';
import ReactDOMServer from 'react-dom/server';
import { formatData } from '../../CoreQuery/AttributionsResult/utils';
import AttributionTable from '../../CoreQuery/AttributionsResult/AttributionTable';
import {
  CHART_TYPE_BARCHART,
  CHART_TYPE_TABLE,
  DASHBOARD_WIDGET_BARLINE_CHART_HEIGHT,
  MAX_ALLOWED_VISIBLE_PROPERTIES,
  BAR_COUNT,
  ATTRIBUTION_METHODOLOGY,
} from '../../../utils/constants';
import { DashboardContext } from '../../../contexts/DashboardContext';
import HCBarLineChart from '../../../components/HCBarLineChart';
import {
  Text,
  Number as NumFormat,
} from '../../../components/factorsComponents';
import chartStyles from '../../../components/HCBarLineChart/styles.module.scss';

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
  attr_dimensions,
  durationObj,
}) {
  const aggregateData = useMemo(() => {
    return formatData(data, touchpoint, event, attr_dimensions);
  }, [data, touchpoint, event, attr_dimensions]);

  const [visibleIndices, setVisibleIndices] = useState(
    Array.from(Array(MAX_ALLOWED_VISIBLE_PROPERTIES).keys())
  );

  const {
    attributionMetrics,
    setAttributionMetrics,
    handleEditQuery,
  } = useContext(DashboardContext);

  const chartData = useMemo(() => {
    if (!aggregateData.categories.length) {
      return {
        ...aggregateData,
      };
    }
    return {
      categories: aggregateData.categories
        .filter((_, index) => visibleIndices.includes(index))
        .slice(0, BAR_COUNT[unit.cardSize]),
      series: aggregateData.series.map((s) => {
        return {
          ...s,
          data: s.data
            .filter((_, index) => visibleIndices.includes(index))
            .slice(0, BAR_COUNT[unit.cardSize]),
        };
      }),
    };
  }, [aggregateData, visibleIndices, unit.cardSize]);

  const generateTooltip = useCallback(
    (category) => {
      const categoryIdx = chartData.categories.findIndex((d) => d === category);
      const conversionIdx = 0;
      const costIdx = 1;
      return ReactDOMServer.renderToString(
        <>
          <Text
            color='grey-6'
            weight='normal'
            type='title'
            extraClass={`text-sm mb-0 ${chartStyles.categoryBottomBorder}`}
          >
            {category}
          </Text>
          <span className='flex items-center mt-3'>
            <Text
              color='grey'
              type='title'
              weight='bold'
              extraClass='text-sm mb-0'
            >
              Conversions
            </Text>
          </span>
          <span className='flex justify-between items-center mt-3'>
            <span
              className={`flex flex-col justify-center items-start pl-2 ${chartStyles.leftBlueBar}`}
            >
              <Text
                color='grey'
                type='title'
                weight='normal'
                extraClass='text-sm mb-0'
              >
                {moment(durationObj.from).format('MMM DD')}
                {' - '}
                {moment(durationObj.to).format('MMM DD')}
              </Text>
              <Text
                color='grey-6'
                type='title'
                weight='bold'
                extraClass='text-base mb-0'
              >
                <NumFormat
                  number={chartData.series[conversionIdx].data[categoryIdx]}
                />
              </Text>
            </span>
          </span>
          <span className='flex items-center mt-3'>
            <Text
              color='grey'
              type='title'
              weight='bold'
              extraClass='text-sm mb-0'
            >
              Cost per conversion
            </Text>
          </span>
          <span className='flex justify-between items-center mt-3'>
            <span
              className={`flex flex-col justify-center items-start pl-2 ${chartStyles.leftRedBar}`}
            >
              <Text
                color='grey'
                type='title'
                weight='normal'
                extraClass='text-sm mb-0'
              >
                {moment(durationObj.from).format('MMM DD')}
                {' - '}
                {moment(durationObj.to).format('MMM DD')}
              </Text>
              <Text
                color='grey-6'
                type='title'
                weight='bold'
                extraClass='text-base mb-0'
              >
                <NumFormat
                  number={chartData.series[costIdx].data[categoryIdx]}
                />
              </Text>
            </span>
          </span>
        </>
      );
    },
    [chartData.categories, chartData.series, durationObj]
  );

  const attributionMethodsMapper = useMemo(() => {
    const mapper = {};
    ATTRIBUTION_METHODOLOGY.forEach((am) => {
      mapper[am.value] = am.text;
    });
    return mapper;
  }, []);

  if (!chartData.categories.length) {
    return null;
  }

  let chartContent = null;

  const legends = [
    `Conversions as Unique users (${attributionMethodsMapper[attribution_method]})`,
    'Cost per conversion',
  ];

  if (chartType === CHART_TYPE_BARCHART) {
    chartContent = (
      <HCBarLineChart
        height={DASHBOARD_WIDGET_BARLINE_CHART_HEIGHT}
        legendsPosition='top'
        cardSize={unit.cardSize}
        chartId={`barLine-${unit.id}`}
        legends={legends}
        categories={chartData.categories}
        series={chartData.series}
        generateTooltip={generateTooltip}
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
        maxAllowedVisibleProperties={MAX_ALLOWED_VISIBLE_PROPERTIES}
        attribution_method={attribution_method}
        attributionMetrics={attributionMetrics}
        setAttributionMetrics={setAttributionMetrics}
        section={section}
        attr_dimensions={attr_dimensions}
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
