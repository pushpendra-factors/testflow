import React, { useState, useMemo, memo, useCallback } from 'react';
import { useSelector } from 'react-redux';
import ReactDOMServer from 'react-dom/server';
import moment from 'moment';
import ScatterPlotChart from '../../../components/ScatterPlotChart';
import { getScatterPlotChartData, getAxisMetricOptions } from './utils';
import { REPORT_SECTION, DATE_FORMATS } from '../../../utils/constants';
import AxisOptionsContainer from '../../../components/AxisOptionsContainer';
import chartStyles from '../../../components/ScatterPlotChart/styles.module.scss';
import {
  Text,
  Number as NumFormat
} from '../../../components/factorsComponents';
import LegendsCircle from '../../../styles/components/LegendsCircle';
import {
  CHART_COLOR_1,
  CHART_COLOR_8
} from '../../../constants/color.constants';

function AttributionsScatterPlot({
  data,
  attr_dimensions,
  content_groups,
  selectedTouchpoint,
  visibleIndices,
  attribution_method,
  attribution_method_compare,
  section,
  linkedEvents,
  comparison_data = {},
  durationObj,
  comparison_duration = {},
  cardSize = 1,
  height,
  chartId = 'scatterPlot'
}) {
  const { eventNames } = useSelector((state) => state.coreQuery);
  const [xAxisMetric, setXAxisMetric] = useState('Conversion');
  const [yAxisMetric, setYAxisMetric] = useState(
    attribution_method_compare ? 'Conversion(compare)' : 'Cost Per Conversion'
  );

  const chartData = useMemo(() => {
    return getScatterPlotChartData(
      selectedTouchpoint,
      attr_dimensions,
      content_groups,
      data,
      visibleIndices,
      xAxisMetric,
      yAxisMetric,
      !!comparison_data.data
    );
  }, [
    selectedTouchpoint,
    attr_dimensions,
    content_groups,
    data,
    visibleIndices,
    xAxisMetric,
    yAxisMetric,
    comparison_data.data
  ]);

  const options = useMemo(() => {
    return getAxisMetricOptions(
      selectedTouchpoint,
      linkedEvents,
      attribution_method,
      attribution_method_compare,
      eventNames
    );
  }, [
    selectedTouchpoint,
    linkedEvents,
    attribution_method,
    attribution_method_compare,
    eventNames
  ]);

  const xAxisTitle = options.find((elem) => elem.value === xAxisMetric).title;
  const yAxisTitle = options.find((elem) => elem.value === yAxisMetric).title;

  const handleXAxisOptionChange = useCallback((e) => {
    setXAxisMetric(e.key);
  }, []);

  const handleYAxisOptionChange = useCallback((e) => {
    setYAxisMetric(e.key);
  }, []);

  const generateTooltip = useCallback(
    (categoryIdx) => {
      const category = chartData.categories[categoryIdx];
      const categoryData = data.find((d) => d.category === category);
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
              {xAxisTitle}
            </Text>
          </span>
          <span className='flex justify-between items-center mt-1'>
            <span
              className={`flex flex-col justify-center items-start ${
                comparison_data.data ? 'w-1/2' : ''
              }`}
            >
              {!!comparison_data.data && (
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
              )}
              <Text
                color='grey-6'
                type='title'
                weight='bold'
                extraClass='text-base mb-0'
              >
                <NumFormat
                  number={
                    comparison_data.data
                      ? categoryData[xAxisMetric].value
                      : categoryData[xAxisMetric]
                  }
                />
              </Text>
            </span>
            {comparison_data.data ? (
              <span
                className={
                  'flex flex-col justify-center items-start ml-2 w-1/2'
                }
              >
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
                  <NumFormat number={categoryData[xAxisMetric].compare_value} />
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
              {yAxisTitle}
            </Text>
          </span>
          <span className={'flex justify-between items-center mt-1'}>
            <span
              className={`flex flex-col justify-center items-start ${
                comparison_data.data ? 'w-1/2' : ''
              }`}
            >
              {!!comparison_data.data && (
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
              )}
              <Text
                color='grey-6'
                type='title'
                weight='bold'
                extraClass='text-base mb-0'
              >
                <NumFormat
                  number={
                    comparison_data.data
                      ? categoryData[yAxisMetric].value
                      : categoryData[yAxisMetric]
                  }
                />
              </Text>
            </span>
            {comparison_data.data ? (
              <span
                className={
                  'flex flex-col justify-center items-start ml-2 w-1/2'
                }
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
                  <NumFormat number={categoryData[yAxisMetric].compare_value} />
                </Text>
              </span>
            ) : null}
          </span>
        </>
      );
    },
    [
      chartData,
      comparison_data.data,
      durationObj,
      comparison_duration,
      data,
      xAxisTitle,
      yAxisTitle
    ]
  );

  return (
    <>
      <ScatterPlotChart
        series={chartData.series}
        xAxisTitle={xAxisTitle}
        yAxisTitle={yAxisTitle}
        generateTooltip={generateTooltip}
        cardSize={cardSize}
        height={height}
        chartId={chartId}
      />
      {!!comparison_data.data && (
        <div className='flex items-center justify-center'>
          <div className='mr-2'>
            <LegendsCircle color={CHART_COLOR_1} />
          </div>
          <div className='mr-4'>
            {moment(durationObj.from).format(DATE_FORMATS.date)} to{' '}
            {moment(durationObj.to).format(DATE_FORMATS.date)}
          </div>
          <div className='mr-2'>
            <LegendsCircle color={CHART_COLOR_8} />
          </div>
          <div>
            {moment(comparison_duration.from).format(DATE_FORMATS.date)} to{' '}
            {moment(comparison_duration.to).format(DATE_FORMATS.date)}
          </div>
        </div>
      )}
      {section === REPORT_SECTION && (
        <AxisOptionsContainer
          xAxisOptions={options}
          yAxisOptions={options}
          onXAxisOptionChange={handleXAxisOptionChange}
          onYAxisOptionChange={handleYAxisOptionChange}
          xAxisMetric={xAxisTitle}
          yAxisMetric={yAxisTitle}
          visiblePointsCount={visibleIndices.length}
          xAxisValue={xAxisMetric}
          yAxisValue={yAxisMetric}
        />
      )}
    </>
  );
}

export default memo(AttributionsScatterPlot);
