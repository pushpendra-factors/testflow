import React, {
  useEffect,
  useState,
  useContext,
  forwardRef,
  useImperativeHandle,
  memo,
  useMemo
} from 'react';
import cx from 'classnames';
import { formatData, getVisibleData } from '../utils';
import FunnelChart from './Chart';
import FunnelsResultTable from '../FunnelsResultTable';
import NoDataChart from '../../../../components/NoDataChart';
import { CoreQueryContext } from '../../../../contexts/CoreQueryContext';
import FunnelsScatterPlot from './FunnelsScatterPlot';
import {
  CHART_TYPE_BARCHART,
  CHART_TYPE_FUNNEL_CHART,
  CHART_TYPE_HORIZONTAL_BAR_CHART,
  CHART_TYPE_METRIC_CHART,
  MAX_ALLOWED_VISIBLE_PROPERTIES,
  METRIC_TYPES
} from '../../../../utils/constants';
import MetricChart from 'Components/MetricChart/MetricChart';
import { generateColors } from 'Utils/dataFormatter';
import HorizontalBarChart from 'Components/HorizontalBarChart';
import ColumnChart from '../../../../components/ColumnChart/ColumnChart';
import { EMPTY_ARRAY } from 'Utils/global';
import {
  getColumChartSeries,
  getCompareGroupsByName,
  getHorizontalBarChartSeries,
  getValueFromPercentString
} from './groupedChart.helpers';

const colors = generateColors(MAX_ALLOWED_VISIBLE_PROPERTIES);

const GroupedChartComponent = forwardRef(
  (
    {
      resultState,
      queries,
      breakdown,
      arrayMapper,
      section,
      chartType,
      tableConfig,
      tableConfigPopoverContent,
      savedQuerySettings,
      comparisonData,
      durationObj,
      comparisonDuration
    },
    ref
  ) => {
    const [visibleProperties, setVisibleProperties] = useState([]);
    const [sorter, setSorter] = useState(
      savedQuerySettings.sorter && Array.isArray(savedQuerySettings.sorter)
        ? savedQuerySettings.sorter
        : EMPTY_ARRAY
    );

    useImperativeHandle(ref, () => {
      return {
        currentSorter: { sorter }
      };
    });

    const { groups, eventsData } = useMemo(() => {
      const { groups: appliedGroups, events } = formatData(
        resultState.data,
        arrayMapper
      );
      return { groups: appliedGroups, eventsData: events };
    }, [arrayMapper, resultState.data]);

    const { compareGroups } = useMemo(() => {
      if (comparisonData.data == null) {
        return { compareGroups: null };
      }
      const { groups: appliedGroups } = formatData(
        comparisonData.data,
        arrayMapper
      );
      return { compareGroups: appliedGroups };
    }, [arrayMapper, comparisonData.data]);

    useEffect(() => {
      setVisibleProperties(getVisibleData(groups, sorter));
    }, [groups, sorter]);

    const horizontalBarChartSeries = useMemo(() => {
      return getHorizontalBarChartSeries({
        visibleProperties,
        chartType,
        compareGroups
      });
    }, [visibleProperties, compareGroups, chartType]);

    const columnChartSeries = useMemo(() => {
      return getColumChartSeries({
        visibleProperties,
        chartType,
        compareGroups
      });
    }, [visibleProperties, compareGroups, chartType]);

    const chartCategories = useMemo(() => {
      return visibleProperties.map((v) => v.name);
    }, [visibleProperties]);

    if (!visibleProperties.length) {
      return (
        <div className='flex justify-center items-center w-full h-full pt-4 pb-4'>
          <NoDataChart />
        </div>
      );
    }

    let chart = null;

    if (chartType === CHART_TYPE_HORIZONTAL_BAR_CHART) {
      chart = (
        <div className='w-full'>
          <HorizontalBarChart
            categories={chartCategories}
            hideXAxis={true}
            series={horizontalBarChartSeries}
            comparisonApplied={comparisonData.data != null}
          />
        </div>
      );
    } else if (chartType === CHART_TYPE_BARCHART) {
      chart = (
        <div className='w-full'>
          <ColumnChart
            categories={chartCategories}
            multiColored
            valueMetricType={METRIC_TYPES.percentType}
            comparisonApplied={comparisonData.data != null}
            series={columnChartSeries}
          />
        </div>
      );
    } else if (chartType === CHART_TYPE_FUNNEL_CHART) {
      chart = (
        <FunnelChart
          groups={visibleProperties}
          eventsData={eventsData}
          arrayMapper={arrayMapper}
          section={section}
          durations={resultState.data.meta}
        />
      );
    } else if (chartType === CHART_TYPE_METRIC_CHART) {
      const compareGroupsByName =
        compareGroups != null ? getCompareGroupsByName({ compareGroups }) : {};
      chart = (
        <div
          className={cx(
            'grid w-full col-gap-2 row-gap-12',
            { 'grid-flow-col': visibleProperties.length < 3 },
            { 'grid-cols-3': visibleProperties.length >= 3 }
          )}
        >
          {visibleProperties.map((elem, index) => {
            const compareGroup = compareGroupsByName[elem.name];
            const value = getValueFromPercentString(elem.value);
            const compareValue =
              compareGroup != null
                ? getValueFromPercentString(compareGroup.value)
                : 0;
            return (
              <MetricChart
                key={colors[index]}
                value={value}
                iconColor={colors[index]}
                headerTitle={elem.name}
                valueType='percentage'
                compareValue={compareValue}
                showComparison={compareGroups != null}
              />
            );
          })}
        </div>
      );
    } else {
      chart = (
        <div className='w-full'>
          <FunnelsScatterPlot
            visibleProperties={visibleProperties}
            arrayMapper={arrayMapper}
            section={section}
          />
        </div>
      );
    }

    return (
      <div className='flex items-center justify-center flex-col'>
        {chart}
        <div className='mt-12 w-full'>
          <FunnelsResultTable
            breakdown={breakdown}
            queries={queries}
            groups={groups}
            visibleProperties={visibleProperties}
            setVisibleProperties={setVisibleProperties}
            chartData={eventsData}
            arrayMapper={arrayMapper}
            resultData={resultState.data}
            sorter={sorter}
            setSorter={setSorter}
            tableConfig={tableConfig}
            tableConfigPopoverContent={tableConfigPopoverContent}
            comparisonChartData={compareGroups}
            isBreakdownApplied={true}
            durationObj={durationObj}
            comparison_duration={comparisonDuration}
          />
        </div>
      </div>
    );
  }
);

const GroupedChartMemoized = memo(GroupedChartComponent);

function GroupedChart(props) {
  const { renderedCompRef, ...rest } = props;
  const {
    coreQueryState: {
      savedQuerySettings,
      comparison_data: comparisonData,
      comparison_duration: comparisonDuration
    }
  } = useContext(CoreQueryContext);

  return (
    <GroupedChartMemoized
      ref={renderedCompRef}
      savedQuerySettings={savedQuerySettings}
      comparisonData={comparisonData}
      comparisonDuration={comparisonDuration}
      {...rest}
    />
  );
}

export default memo(GroupedChart);
