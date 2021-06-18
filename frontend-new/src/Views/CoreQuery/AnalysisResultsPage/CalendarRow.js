import React, { useCallback, useContext } from 'react';
import FaDatepicker from '../../../components/FaDatepicker';
import ChartTypeDropdown from '../../../components/ChartTypeDropdown';
import { QUERY_TYPE_WEB } from '../../../utils/constants';
import styles from './index.module.scss';
import { Button } from 'antd';
import { SVG, Spiner } from '../../../components/factorsComponents';
import { CoreQueryContext } from '../../../contexts/CoreQueryContext';

function CalendarRow({
  durationObj,
  handleDurationChange,
  handleChartTypeChange,
  chartTypeMenuItems,
  chartType,
  metricsDropdown,
  queryType,
  triggerAttrComparision,
}) {
  const {
    coreQueryState,
    resetComparisonData,
    handleCompareWithClick,
  } = useContext(CoreQueryContext);

  let comparison_supported,
    comparison_data,
    comparison_enabled,
    comparison_duration;

  if (queryType !== 'web') {
    comparison_supported = coreQueryState.comparison_supported;
    comparison_data = coreQueryState.comparison_data;
    comparison_enabled = coreQueryState.comparison_enabled;
    comparison_duration = coreQueryState.comparison_duration;
  }

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

  return (
    <div className='flex justify-between items-center'>
      <div className='flex items-center'>
        {metricsDropdown}
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
        {comparison_supported && renderCompareScenario()}
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
