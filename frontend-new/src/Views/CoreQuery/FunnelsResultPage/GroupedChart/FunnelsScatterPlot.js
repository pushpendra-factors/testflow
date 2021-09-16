import React, { useState, useMemo, useCallback, memo } from 'react';
import ReactDOMServer from 'react-dom/server';
import { getScatterPlotChartData, getAxisMetricOptions } from '../utils';
import ScatterPlotChart from '../../../../components/ScatterPlotChart';
import AxisOptionsContainer from '../../../../components/AxisOptionsContainer';
import { REPORT_SECTION } from '../../../../utils/constants';
import { Text } from '../../../../components/factorsComponents';
import chartStyles from '../../../../components/ScatterPlotChart/styles.module.scss';

const FunnelsScatterPlot = ({
  arrayMapper,
  visibleProperties,
  section,
  height = null,
  cardSize = 1,
  chartId = 'funnels-scatterPlot',
}) => {
  const [xAxisMetric, setXAxisMetric] = useState('Conversion');
  const [yAxisMetric, setYAxisMetric] = useState('Conversion Time');

  const chartData = useMemo(() => {
    return getScatterPlotChartData(visibleProperties, xAxisMetric, yAxisMetric);
  }, [xAxisMetric, yAxisMetric, visibleProperties]);

  const options = useMemo(() => {
    return getAxisMetricOptions(arrayMapper);
  }, [arrayMapper]);

  const handleXAxisOptionChange = useCallback((e) => {
    setXAxisMetric(e.key);
  }, []);

  const handleYAxisOptionChange = useCallback((e) => {
    setYAxisMetric(e.key);
  }, []);

  const xAxisTitle = options.find((elem) => elem.value === xAxisMetric).title;
  const yAxisTitle = options.find((elem) => elem.value === yAxisMetric).title;

  const generateTooltip = useCallback(
    (idx) => {
      const categoryData = visibleProperties[idx];
      const category = categoryData.name;
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
            <span className={`flex flex-col justify-center items-start`}>
              <Text
                color='grey-6'
                type='title'
                weight='bold'
                extraClass='text-base mb-0'
              >
                {categoryData[xAxisMetric].value || categoryData[xAxisMetric]}
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
              {yAxisTitle}
            </Text>
          </span>
          <span className='flex justify-between items-center mt-1'>
            <span className={`flex flex-col justify-center items-start`}>
              <Text
                color='grey-6'
                type='title'
                weight='bold'
                extraClass='text-base mb-0'
              >
                {categoryData[yAxisMetric].value || categoryData[yAxisMetric]}
              </Text>
            </span>
          </span>
        </>
      );
    },
    [visibleProperties, xAxisTitle, yAxisTitle, xAxisMetric, yAxisMetric]
  );

  return (
    <>
      <ScatterPlotChart
        series={chartData.series}
        xAxisTitle={xAxisTitle}
        yAxisTitle={yAxisTitle}
        generateTooltip={generateTooltip}
        height={height}
        cardSize={cardSize}
        chartId={chartId}
      />
      {section === REPORT_SECTION && (
        <AxisOptionsContainer
          xAxisOptions={options}
          yAxisOptions={options}
          onXAxisOptionChange={handleXAxisOptionChange}
          onYAxisOptionChange={handleYAxisOptionChange}
          xAxisMetric={xAxisTitle}
          yAxisMetric={yAxisTitle}
          visiblePointsCount={visibleProperties.length}
          xAxisValue={xAxisMetric}
          yAxisValue={yAxisMetric}
        />
      )}
    </>
  );
};

export default memo(FunnelsScatterPlot);
