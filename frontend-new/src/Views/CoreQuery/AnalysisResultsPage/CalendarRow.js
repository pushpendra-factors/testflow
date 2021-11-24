import React, { useCallback, useContext } from 'react';
import FaDatepicker from '../../../components/FaDatepicker';
import ChartTypeDropdown from '../../../components/ChartTypeDropdown';
import {
  QUERY_TYPE_WEB,
  QUERY_TYPE_EVENT,
  QUERY_TYPE_CAMPAIGN,
  REPORT_SECTION,
  QUERY_TYPE_PROFILE,
  QUERY_TYPE_KPI,
} from '../../../utils/constants';
import styles from './index.module.scss';
import { Button } from 'antd';
import { SVG, Spiner } from '../../../components/factorsComponents';
import { CoreQueryContext } from '../../../contexts/CoreQueryContext';
import { isSeriesChart } from '../../../utils/dataFormatter';
import GranularityOptions from '../../../components/GranularityOptions';

function CalendarRow({
  durationObj,
  handleDurationChange,
  handleChartTypeChange,
  chartTypeMenuItems,
  chartType,
  metricsDropdown,
  queryType,
  triggerAttrComparision,
  handleGranularityChange,
  section,
}) {
  const {
    coreQueryState: {
      comparison_supported,
      comparison_data,
      comparison_enabled,
      comparison_duration,
    },
    resetComparisonData,
    handleCompareWithClick,
  } = useContext(CoreQueryContext);

  const setDateRange = useCallback(
    (range) => {
      handleDurationChange(range);
    },
    [handleDurationChange]
  );

  const setCompareDateRange = useCallback(
    (range) => {
      handleDurationChange(range, true);
    },
    [handleDurationChange]
  );

  if (queryType === QUERY_TYPE_WEB) {
    return null;
  }

  const renderCompareScenario = () => {
    return (
      <div
        className={`${styles.comparisonCalendar} flex items-center ml-4 mr-4`}
      >
        {/* {!isCompareCalendar && (
          <Button
            icon={<SVG name='compare' />}
            onClick={() => showCompareCalendar(true)}
          ></Button>
        )} */}

        {comparison_enabled && (
          <>
            <div>vs</div>
            <div className='calendar ml-4'>
              <FaDatepicker
                customPicker
                presetRange
                monthPicker
                placement='topRight'
                range={
                  comparison_duration && {
                    startDate: comparison_duration.from,
                    endDate: comparison_duration.to,
                  }
                }
                onSelect={setCompareDateRange}
              />
            </div>
          </>
        )}

        {comparison_enabled ? (
          <Button
            className={`ml-2`}
            icon={<SVG name='times' />}
            onClick={resetComparisonData}
          ></Button>
        ) : null}

        {comparison_data.loading && <Spiner size={`small`}></Spiner>}
      </div>
    );
  };

  let granularity = null;

  if (
    (queryType === QUERY_TYPE_EVENT ||
      queryType === QUERY_TYPE_CAMPAIGN ||
      queryType === QUERY_TYPE_KPI) &&
    section === REPORT_SECTION &&
    isSeriesChart(chartType)
  ) {
    granularity = (
      <>
        <div className='mx-2'>as</div>
        <GranularityOptions
          onClick={handleGranularityChange}
          durationObj={durationObj}
          queryType={queryType}
        />
      </>
    );
  }

  let calendarWidget = null;

  if (queryType === QUERY_TYPE_PROFILE) {
    calendarWidget = (
      <Button
        className={``}
        size={'large'}
        type={'default'}
        style={{ pointerEvents: 'none' }}
      >
        <SVG name={`calendar`} size={20} extraClass={`mr-1`}></SVG>All Time
      </Button>
    );
  } else {
    calendarWidget = (
      <div className='calendar'>
        <FaDatepicker
          customPicker
          presetRange
          monthPicker
          range={{
            startDate: durationObj.from,
            endDate: durationObj.to,
          }}
          placement='topRight'
          onSelect={setDateRange}
          comparison_supported={comparison_supported}
          handleCompareWithClick={handleCompareWithClick}
        />
      </div>
    );
  }

  return (
    <div className='flex justify-between items-center'>
      <div className='flex items-center'>
        {metricsDropdown}
        {calendarWidget}
        {comparison_supported && renderCompareScenario()}
        {granularity}
      </div>
      <div className='flex items-center'>
        {chartTypeMenuItems.length ? (
          <ChartTypeDropdown
            chartType={chartType}
            onClick={handleChartTypeChange}
            menuItems={chartTypeMenuItems}
          />
        ) : null}
      </div>
    </div>
  );
}

export default CalendarRow;
