import React, { useState, useCallback, useContext } from 'react';
import { Button, DatePicker } from 'antd';
import MomentTz from 'Components/MomentTz';
import ControlledComponent from 'Components/ControlledComponent';
import ChartConfigPopover from './ChartConfigPopover';
import FaDatepicker from '../../../../components/FaDatepicker';
import ChartTypeDropdown from '../../../../components/ChartTypeDropdown';
import {
  QUERY_TYPE_WEB,
  QUERY_TYPE_EVENT,
  QUERY_TYPE_CAMPAIGN,
  REPORT_SECTION,
  QUERY_TYPE_PROFILE,
  QUERY_TYPE_KPI
} from '../../../../utils/constants';
import { SVG, Text, Spiner } from '../../../../components/factorsComponents';
import { CoreQueryContext } from '../../../../contexts/CoreQueryContext';
import { isSeriesChart } from '../../../../utils/dataFormatter';
import GranularityOptions from '../../../../components/GranularityOptions';
import ComparisonDatePicker from '../../../../components/ComparisonDatePicker/ComparisonDatePicker';
import { getCompareRange } from './calendarRow.helpers';
import OptionsPopover from '../../AttributionsResult/OptionsPopover';
import { getKpiLabel } from '../../KPIAnalysis/kpiAnalysis.helpers';

function CalendarRow({
  durationObj,
  handleDurationChange,
  handleChartTypeChange,
  chartTypeMenuItems,
  chartType,
  metricsDropdown,
  queryType,
  // triggerAttrComparision,
  handleGranularityChange,
  section,
  setSecondAxisKpiIndices,
  secondAxisKpiIndices = [],
  showChartConfigOptions = true,
  kpis = []
}) {
  const {
    coreQueryState: {
      comparison_supported: comparisonSupported,
      comparison_data: comparisonData,
      comparison_duration: comparisonDuration
    },
    resetComparisonData,
    handleCompareWithClick
  } = useContext(CoreQueryContext);

  const [showDatePicker, setShowDatePicker] = useState(false);

  const setDateRange = useCallback(
    (range) => {
      if (comparisonDuration) {
        resetComparisonData();
      }
      handleDurationChange(range);
    },
    [handleDurationChange, resetComparisonData, comparisonDuration]
  );

  const setDateSince = (date) => {
    const setDate = {
      startDate: date
    };
    if (comparisonDuration) {
      resetComparisonData();
    }
    handleDurationChange(setDate);
    setShowDatePicker(false);
  };

  const handleCompareDateChange = useCallback(
    ({ value, isPreset = true, customRangeType }) => {
      const compareRange = getCompareRange({
        selectedValue: value,
        isPreset,
        customRangeType,
        durationObj
      });
      handleDurationChange(compareRange, true);
    },
    [durationObj, handleDurationChange]
  );

  const handleKpiSecondAxisConfigChange = useCallback(
    (option) => {
      setSecondAxisKpiIndices((curr) => {
        if (!curr.includes(option.value)) {
          if (curr.length === kpis.length - 1) {
            return curr;
          }
          return [...curr, option.value];
        }
        return curr.filter((c) => c !== option.value);
      });
    },
    [setSecondAxisKpiIndices, kpis]
  );

  if (queryType === QUERY_TYPE_WEB) {
    return null;
  }

  const renderCompareScenario = () => {
    if (!comparisonSupported) {
      return null;
    }
    return (
      <div className='flex items-center col-gap-2'>
        <div className='calendar'>
          <ComparisonDatePicker
            placement='bottomLeft'
            value={comparisonDuration}
            onChange={handleCompareDateChange}
            onRemoveClick={resetComparisonData}
          />
        </div>
        {comparisonData.loading && <Spiner size='small' />}
      </div>
    );
  };

  let granularity = null;

  if (
    (queryType === QUERY_TYPE_EVENT ||
      queryType === QUERY_TYPE_CAMPAIGN ||
      queryType === QUERY_TYPE_KPI) &&
    section === REPORT_SECTION &&
    durationObj.frequency &&
    isSeriesChart(chartType)
  ) {
    granularity = (
      <>
        <div>as</div>
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
      <div className='flex items-center'>
        <Text type='title' level={7} weight='bold' extraClass='m-0 mr-2'>
          Created Since
        </Text>
        <div className='fa-custom-datepicker'>
          {!showDatePicker ? (
            <Button
              onClick={() => {
                setShowDatePicker(true);
              }}
            >
              <SVG name='calendar' size={16} extraClass='mr-1' />
              {MomentTz(durationObj.from).format('MMM DD, YYYY')}
            </Button>
          ) : (
            <Button>
              <SVG name='calendar' size={16} extraClass='mr-1' />
              <DatePicker
                format='MMM DD YYYY'
                style={{ width: '96px' }}
                disabledDate={(d) => !d || d.isAfter(MomentTz())}
                dropdownClassName='fa-custom-datepicker--datepicker'
                size='small'
                suffixIcon={null}
                showToday={false}
                bordered={false}
                autoFocus
                allowClear={false}
                placement='bottomLeft'
                open
                onOpenChange={() => {
                  setShowDatePicker(false);
                }}
                onChange={setDateSince}
              />
            </Button>
          )}
        </div>
      </div>
    );
  } else {
    calendarWidget = (
      <div className='calendar'>
        <FaDatepicker
          customPicker
          presetRange
          monthPicker
          quarterPicker
          range={{
            startDate: durationObj.from,
            endDate: durationObj.to
          }}
          placement='bottomLeft'
          onSelect={setDateRange}
          comparison_supported={comparisonSupported}
          handleCompareWithClick={handleCompareWithClick}
        />
      </div>
    );
  }

  const KpiSecondAxisConfig = (
    <div className='flex flex-col row-gap-1'>
      <div className='border-b pb-2'>
        <Text color='grey-6' type='title'>
          Enable Secondary Y Axis for:
        </Text>
      </div>
      <OptionsPopover
        options={kpis.map((kpi, index) => ({
          title: getKpiLabel(kpi),
          enabled: secondAxisKpiIndices.includes(index),
          value: index
        }))}
        onChange={handleKpiSecondAxisConfigChange}
      />
    </div>
  );

  return (
    <div className='flex justify-between items-center'>
      <div className='flex items-center col-gap-2'>
        {metricsDropdown}
        {calendarWidget}
        {renderCompareScenario()}
        {granularity}
      </div>
      <div className='flex items-center col-gap-2'>
        <ControlledComponent controller={showChartConfigOptions}>
          <ChartConfigPopover>{KpiSecondAxisConfig}</ChartConfigPopover>
        </ControlledComponent>
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
