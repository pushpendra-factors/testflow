import React, {
  useState,
  useMemo,
  forwardRef,
  useContext,
  useImperativeHandle,
  memo
} from 'react';
import { useSelector } from 'react-redux';
import { get } from 'lodash';
import {
  formatData,
  getDataInLineChartFormat,
  getDefaultSortProp,
  getDefaultDateSortProp
} from './utils';
import NoBreakdownTable from './NoBreakdownTable';
import LineChart from '../../../../components/HCLineChart';
import { generateColors } from '../../../../utils/dataFormatter';
import {
  DASHBOARD_MODAL,
  CHART_TYPE_SPARKLINES,
  CHART_TYPE_LINECHART
} from '../../../../utils/constants';
import { CoreQueryContext } from '../../../../contexts/CoreQueryContext';
import SparkChartWithCount from '../../../../components/SparkChartWithCount/SparkChartWithCount';

const NoBreakdownChartsComponent = forwardRef(
  (
    {
      queries,
      resultState,
      page,
      chartType,
      durationObj,
      arrayMapper,
      section,
      savedQuerySettings,
      comparisonData
    },
    ref
  ) => {
    const [sorter, setSorter] = useState(
      savedQuerySettings.sorter && Array.isArray(savedQuerySettings.sorter)
        ? savedQuerySettings.sorter
        : getDefaultSortProp(arrayMapper)
    );

    const [dateSorter, setDateSorter] = useState(
      savedQuerySettings.dateSorter &&
        Array.isArray(savedQuerySettings.dateSorter)
        ? savedQuerySettings.dateSorter
        : getDefaultDateSortProp()
    );

    useImperativeHandle(ref, () => {
      return {
        currentSorter: { sorter, dateSorter }
      };
    });

    const [hiddenEvents, setHiddenEvents] = useState([]);
    const { eventNames } = useSelector((state) => state.coreQuery);
    const appliedColors = useMemo(() => {
      return generateColors(queries.length);
    }, [queries]);

    const chartsData = useMemo(() => {
      return formatData(resultState.data, arrayMapper, comparisonData.data);
    }, [resultState.data, arrayMapper, comparisonData.data]);

    const { categories, data, compareCategories } = useMemo(() => {
      return getDataInLineChartFormat(
        resultState.data,
        arrayMapper,
        eventNames,
        comparisonData.data
      );
    }, [resultState.data, arrayMapper, eventNames, comparisonData.data]);

    const visibleSeriesData = useMemo(() => {
      return data
        .filter(
          (elem) => hiddenEvents.findIndex((he) => he === elem.name) === -1
        )
        .map((elem, index) => {
          const color = appliedColors[index];
          return {
            ...elem,
            color
          };
        });
    }, [data, hiddenEvents, appliedColors]);

    if (!chartsData.length) {
      return null;
    }

    let chart = null;

    const table = (
      <div className="mt-12 w-full">
        <NoBreakdownTable
          isWidgetModal={section === DASHBOARD_MODAL}
          data={chartsData}
          events={queries}
          chartType={chartType}
          setHiddenEvents={setHiddenEvents}
          hiddenEvents={hiddenEvents}
          durationObj={durationObj}
          arrayMapper={arrayMapper}
          sorter={sorter}
          setSorter={setSorter}
          dateSorter={dateSorter}
          setDateSorter={setDateSorter}
          responseData={resultState.data}
          comparisonApplied={!!comparisonData.data}
        />
      </div>
    );

    if (chartType === CHART_TYPE_SPARKLINES) {
      if (queries.length === 1) {
        chart = (
          <div className="flex items-center justify-center w-full">
            <SparkChartWithCount
              total={get(resultState, 'data.metrics.rows.0.2', 0)}
              compareTotal={get(comparisonData, 'data.metrics.rows.0.2', 0)}
              event={arrayMapper[0].mapper}
              compareKey={`${arrayMapper[0].mapper} - compareValue`}
              frequency={durationObj.frequency}
              chartData={chartsData}
              comparisonApplied={!!comparisonData.data}
              headerTitle={arrayMapper[0].displayName}
            />
          </div>
        );
      }

      if (queries.length > 1) {
        const appliedColors = generateColors(queries.length);
        chart = (
          <div className="flex items-center flex-wrap justify-center w-full">
            {queries.map((_, index) => {
              return (
                <div
                  style={{ minWidth: '300px' }}
                  key={arrayMapper[index].mapper}
                  className="w-1/3 mt-4 px-4"
                >
                  <SparkChartWithCount
                    total={get(resultState, `data.metrics.rows.${index}.2`, 0)}
                    compareTotal={get(
                      comparisonData,
                      `data.metrics.rows.${index}.2`,
                      0
                    )}
                    event={arrayMapper[index].mapper}
                    compareKey={`${arrayMapper[index].mapper} - compareValue`}
                    frequency={durationObj.frequency}
                    chartData={chartsData}
                    chartColor={appliedColors[index]}
                    alignment="vertical"
                    comparisonApplied={!!comparisonData.data}
                    headerTitle={arrayMapper[index].displayName}
                  />
                </div>
              );
            })}
          </div>
        );
      }
    } else if (chartType === CHART_TYPE_LINECHART) {
      chart = (
        <div className="w-full">
          <LineChart
            frequency={durationObj.frequency}
            categories={categories}
            data={visibleSeriesData}
            comparisonApplied={!!comparisonData.data}
            compareCategories={compareCategories}
          />
        </div>
      );
    }

    return (
      <div className="flex items-center justify-center flex-col">
        {chart}
        {table}
      </div>
    );
  }
);

const NoBreakdownChartsMemoized = memo(NoBreakdownChartsComponent);

const NoBreakdownCharts = (props) => {
  const {
    coreQueryState: { savedQuerySettings, comparison_data: comparisonData }
  } = useContext(CoreQueryContext);

  return (
    <NoBreakdownChartsMemoized
      savedQuerySettings={savedQuerySettings}
      comparisonData={comparisonData}
      {...props}
    />
  );
};

export default NoBreakdownCharts;
