import React, { useCallback, useMemo, memo } from 'react';
import moment from 'moment';
import ReactDOMServer from 'react-dom/server';
import HCBarLineChart from '../../../components/HCBarLineChart';
import { ATTRIBUTION_METHODOLOGY } from '../../../utils/constants';
import {
  Text,
  Number as NumFormat
} from '../../../components/factorsComponents';
import chartStyles from '../../../components/HCBarLineChart/styles.module.scss';

const SingleTouchPointChart = ({
  aggregateData,
  durationObj,
  attribution_method,
  comparison_data = {},
  comparison_duration = {},
  height,
  cardSize = 1,
  legendsPosition,
  chartId = 'barLineChart'
}) => {
  const { legends: chartLegends, categories, series } = aggregateData;

  const generateTooltip = useCallback(
    (category) => {
      const categoryIdx = categories.findIndex((d) => d === category);
      const conversionIdx = 0;
      const compareConversionIdx = comparison_data.data ? 1 : null;
      const costIdx = comparison_data.data ? 2 : 1;
      const compareCostIdx = comparison_data.data ? 3 : null;
      return ReactDOMServer.renderToString(
        <>
          <Text
            color="grey-6"
            weight="normal"
            type="title"
            extraClass={`text-sm mb-0 ${chartStyles.categoryBottomBorder}`}
          >
            {category}
          </Text>
          <span className="flex items-center mt-3">
            <Text
              color="grey"
              type="title"
              weight="bold"
              extraClass="text-sm mb-0"
            >
              {`${chartLegends[0]} (${attributionMethodsMapper[attribution_method]})`}
            </Text>
          </span>
          <span className="flex justify-between items-center mt-3">
            <span
              className={`flex flex-col justify-center items-start pl-2 ${
                comparison_data.data ? 'w-1/2' : ''
              } ${chartStyles.leftBlueBar}`}
            >
              <Text
                color="grey"
                type="title"
                weight="normal"
                extraClass="text-sm mb-0"
              >
                {moment(durationObj.from).format('MMM DD')}
                {' - '}
                {moment(durationObj.to).format('MMM DD')}
              </Text>
              <Text
                color="grey-6"
                type="title"
                weight="bold"
                extraClass="text-base mb-0"
              >
                <NumFormat number={series[conversionIdx].data[categoryIdx]} />
              </Text>
            </span>
            {comparison_data.data ? (
              <span
                className={
                  'flex flex-col justify-center items-start ml-2 w-1/2'
                }
              >
                <span className={chartStyles.leftDashedBlueBar}></span>
                <Text
                  color="grey"
                  type="title"
                  weight="normal"
                  extraClass="text-sm mb-0 ml-4"
                >
                  {moment(comparison_duration.from).format('MMM DD')}
                  {' - '}
                  {moment(comparison_duration.to).format('MMM DD')}
                </Text>
                <Text
                  color="grey-6"
                  type="title"
                  weight="bold"
                  extraClass="text-base mb-0 ml-4"
                >
                  <NumFormat
                    number={series[compareConversionIdx].data[categoryIdx]}
                  />
                </Text>
              </span>
            ) : null}
          </span>
          <span className="flex items-center mt-3">
            <Text
              color="grey"
              type="title"
              weight="bold"
              extraClass="text-sm mb-0"
            >
              {chartLegends[1]}
            </Text>
          </span>
          <span className="flex justify-between items-center mt-3">
            <span
              className={`flex flex-col justify-center items-start pl-2 ${
                comparison_data.data ? 'w-1/2' : ''
              } ${chartStyles.leftRedBar}`}
            >
              <Text
                color="grey"
                type="title"
                weight="normal"
                extraClass="text-sm mb-0"
              >
                {moment(durationObj.from).format('MMM DD')}
                {' - '}
                {moment(durationObj.to).format('MMM DD')}
              </Text>
              <Text
                color="grey-6"
                type="title"
                weight="bold"
                extraClass="text-base mb-0"
              >
                <NumFormat number={series[costIdx].data[categoryIdx]} />
              </Text>
            </span>
            {comparison_data.data ? (
              <span
                className={`flex flex-col justify-center items-start ml-2 pl-2 w-1/2 ${chartStyles.leftDashedRedBar}`}
              >
                <Text
                  color="grey"
                  type="title"
                  weight="normal"
                  extraClass="text-sm mb-0"
                >
                  {moment(comparison_duration.from).format('MMM DD')}
                  {' - '}
                  {moment(comparison_duration.to).format('MMM DD')}
                </Text>
                <Text
                  color="grey-6"
                  type="title"
                  weight="bold"
                  extraClass="text-base mb-0"
                >
                  <NumFormat
                    number={series[compareCostIdx].data[categoryIdx]}
                  />
                </Text>
              </span>
            ) : null}
          </span>
        </>
      );
    },
    [categories, series, comparison_data, durationObj, comparison_duration]
  );

  const attributionMethodsMapper = useMemo(() => {
    const mapper = {};
    ATTRIBUTION_METHODOLOGY.forEach((am) => {
      mapper[am.value] = am.text;
    });
    return mapper;
  }, []);

  const legends = useMemo(() => {
    return [
      `${chartLegends[0]} (${attributionMethodsMapper[attribution_method]})`,
      chartLegends[1]
    ];
  }, [attributionMethodsMapper, attribution_method, chartLegends]);

  return (
    <HCBarLineChart
      series={series}
      categories={categories}
      legends={legends}
      generateTooltip={generateTooltip}
      height={height}
      cardSize={cardSize}
      legendsPosition={legendsPosition}
      chartId={chartId}
    />
  );
};

export default memo(SingleTouchPointChart);
