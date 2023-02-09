import React, {
  useEffect,
  useState,
  useContext,
  forwardRef,
  useImperativeHandle
} from 'react';
import { formatData, getVisibleData } from '../utils';
import BarChart from './Chart';
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

const colors = generateColors(MAX_ALLOWED_VISIBLE_PROPERTIES);

const GroupedChart = forwardRef(
  (
    {
      resultState,
      queries,
      breakdown,
      isWidgetModal,
      arrayMapper,
      section,
      chartType,
      tableConfig,
      tableConfigPopoverContent
    },
    ref
  ) => {
    const {
      coreQueryState: { savedQuerySettings }
    } = useContext(CoreQueryContext);
    const [visibleProperties, setVisibleProperties] = useState([]);
    const [sorter, setSorter] = useState(
      savedQuerySettings.sorter && Array.isArray(savedQuerySettings.sorter)
        ? savedQuerySettings.sorter
        : []
    );
    const [eventsData, setEventsData] = useState([]);
    const [groups, setGroups] = useState([]);

    useImperativeHandle(ref, () => {
      return {
        currentSorter: { sorter }
      };
    });

    useEffect(() => {
      const { groups: appliedGroups, events } = formatData(
        resultState.data,
        arrayMapper
      );
      setGroups(appliedGroups);
      setEventsData(events);
    }, [resultState.data, arrayMapper]);

    useEffect(() => {
      setVisibleProperties(getVisibleData(groups, sorter));
    }, [groups, sorter]);

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
            categories={visibleProperties.map((v) => v.name)}
            hideXAxis={true}
            series={[
              {
                name: 'OG',
                data: visibleProperties.map((v, index) => {
                  return {
                    y: parseInt(v.value.split('%')[0]),
                    color: colors[index],
                    metricType: METRIC_TYPES.percentType
                  };
                })
              }
            ]}
          />
        </div>
      );
    } else if (chartType === CHART_TYPE_BARCHART) {
      chart = (
        <div className='w-full'>
          <ColumnChart
            categories={visibleProperties.map((v) => v.name)}
            multiColored
            valueMetricType={METRIC_TYPES.percentType}
            series={[
              {
                name: 'OG',
                data: visibleProperties.map((v, index) =>
                  parseInt(v.value.split('%')[0])
                )
              }
            ]}
          />
        </div>
      );
    } else if (chartType === CHART_TYPE_FUNNEL_CHART) {
      chart = (
        <BarChart
          isWidgetModal={isWidgetModal}
          groups={visibleProperties}
          eventsData={eventsData}
          arrayMapper={arrayMapper}
          section={section}
          durations={resultState.data.meta}
        />
      );
    } else if (chartType === CHART_TYPE_METRIC_CHART) {
      chart = (
        <div className='grid grid-cols-3 w-full col-gap-2 row-gap-12'>
          {visibleProperties.map((elem, index) => {
            return (
              <MetricChart
                key={colors[index]}
                value={elem.value}
                iconColor={colors[index]}
                headerTitle={elem.name}
                valueType='percentage'
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
          />
        </div>
      </div>
    );
  }
);

export default GroupedChart;
