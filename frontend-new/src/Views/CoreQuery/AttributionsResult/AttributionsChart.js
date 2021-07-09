import React, { useState, useContext, useMemo, useCallback } from 'react';
import { formatData } from './utils';
import chartStyles from '../../../components/HCBarLineChart/styles.module.scss';
import moment from 'moment';
import ReactDOMServer from 'react-dom/server';
import AttributionTable from './AttributionTable';
import {
  DASHBOARD_MODAL,
  MAX_ALLOWED_VISIBLE_PROPERTIES,
} from '../../../utils/constants';
import { CoreQueryContext } from '../../../contexts/CoreQueryContext';
import HCBarLineChart from '../../../components/HCBarLineChart';
import {
  Text,
  Number as NumFormat,
} from '../../../components/factorsComponents';

function AttributionsChart({
  data,
  event,
  attribution_method,
  touchpoint,
  linkedEvents,
  section,
  durationObj,
  attr_dimensions,
}) {
  const {
    coreQueryState: { comparison_data, comparison_duration },
  } = useContext(CoreQueryContext);

  const aggregateData = useMemo(() => {
    return formatData(
      data,
      touchpoint,
      event,
      attr_dimensions,
      comparison_data.data
    );
  }, [data, touchpoint, event, attr_dimensions, comparison_data.data]);

  const [visibleIndices, setVisibleIndices] = useState(
    Array.from(Array(MAX_ALLOWED_VISIBLE_PROPERTIES).keys())
  );
  const { attributionMetrics, setAttributionMetrics } = useContext(
    CoreQueryContext
  );

  const chartData = useMemo(() => {
    if (!aggregateData.categories.length) {
      return {
        ...aggregateData,
      };
    }
    return {
      categories: aggregateData.categories.filter((_, index) =>
        visibleIndices.includes(index)
      ),
      series: aggregateData.series.map((s) => {
        return {
          ...s,
          data: s.data.filter((_, index) => visibleIndices.includes(index)),
        };
      }),
    };
  }, [aggregateData, visibleIndices]);

  const generateTooltip = useCallback(
    (category) => {
      const categoryIdx = chartData.categories.findIndex((d) => d === category);
      const conversionIdx = 0;
      const compareConversionIdx = comparison_data.data ? 1 : null;
      const costIdx = comparison_data.data ? 2 : 1;
      const compareCostIdx = comparison_data.data ? 3 : null;
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
              Opportunities
            </Text>
          </span>
          <span className='flex justify-between items-center mt-3'>
            <span
              className={`flex flex-col justify-center items-start pl-2 ${
                comparison_data.data ? 'w-1/2' : ''
              } ${chartStyles.leftBlueBar}`}
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
            {comparison_data.data ? (
              <span
                className={`flex flex-col justify-center items-start ml-2 w-1/2`}
              >
                <span className={chartStyles.leftDashedBlueBar}></span>
                <Text
                  color='grey'
                  type='title'
                  weight='normal'
                  extraClass='text-sm mb-0 ml-4'
                >
                  {moment(comparison_duration.from).format('MMM DD')}
                  {' - '}
                  {moment(comparison_duration.to).format('MMM DD')}
                </Text>
                <Text
                  color='grey-6'
                  type='title'
                  weight='bold'
                  extraClass='text-base mb-0 ml-4'
                >
                  <NumFormat
                    number={
                      chartData.series[compareConversionIdx].data[categoryIdx]
                    }
                  />
                </Text>
              </span>
            ) : null}
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
              className={`flex flex-col justify-center items-start pl-2 ${
                comparison_data.data ? 'w-1/2' : ''
              } ${chartStyles.leftRedBar}`}
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
            {comparison_data.data ? (
              <span
                className={`flex flex-col justify-center items-start ml-2 pl-2 w-1/2 ${chartStyles.leftDashedRedBar}`}
              >
                <Text
                  color='grey'
                  type='title'
                  weight='normal'
                  extraClass='text-sm mb-0'
                >
                  {moment(comparison_duration.from).format('MMM DD')}
                  {' - '}
                  {moment(comparison_duration.to).format('MMM DD')}
                </Text>
                <Text
                  color='grey-6'
                  type='title'
                  weight='bold'
                  extraClass='text-base mb-0'
                >
                  <NumFormat
                    number={chartData.series[compareCostIdx].data[categoryIdx]}
                  />
                </Text>
              </span>
            ) : null}
          </span>
        </>
      );
    },
    [
      chartData.categories,
      chartData.series,
      comparison_data.data,
      durationObj,
      comparison_duration,
    ]
  );

  if (!chartData.categories.length) {
    return null;
  }

  const legends = [
    `Conversions as Unique users (${attribution_method})`,
    'Cost per conversion',
  ];

  return (
    <div className='flex items-center justify-center flex-col'>
      <div className='w-full'>
        <HCBarLineChart
          series={chartData.series}
          categories={chartData.categories}
          legends={legends}
          generateTooltip={generateTooltip}
        />
      </div>
      <div className='mt-12 w-full'>
        <AttributionTable
          linkedEvents={linkedEvents}
          touchpoint={touchpoint}
          event={event}
          data={data}
          comparison_data={comparison_data.data}
          durationObj={durationObj}
          cmprDuration={comparison_duration}
          isWidgetModal={section === DASHBOARD_MODAL}
          visibleIndices={visibleIndices}
          setVisibleIndices={setVisibleIndices}
          maxAllowedVisibleProperties={MAX_ALLOWED_VISIBLE_PROPERTIES}
          attribution_method={attribution_method}
          attributionMetrics={attributionMetrics}
          setAttributionMetrics={setAttributionMetrics}
          section={section}
          attr_dimensions={attr_dimensions}
        />
      </div>
    </div>
  );
}

export default AttributionsChart;
