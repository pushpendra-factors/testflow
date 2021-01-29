import React, { useCallback } from "react";
import FaDatepicker from "../../../components/FaDatepicker";
import ChartTypeDropdown from "../../../components/ChartTypeDropdown";

function CalendarRow({
  durationObj,
  handleDurationChange,
  handleChartTypeChange,
  chartTypeMenuItems,
  chartType,
  metricsDropdown
}) {
  const setDateRange = useCallback(
    (range) => {
      handleDurationChange(range);
    },
    [handleDurationChange]
  );


  return (
    <div className="flex justify-between items-center">
      <div className="flex items-center">
        {metricsDropdown}
        <div className="calendar">
          <FaDatepicker
            customPicker
            presetRange
            monthPicker
            quarterPicker
            range={{
              startDate: durationObj.from,
              endDate: durationObj.to,
            }}
            placement="topRight"
            onSelect={setDateRange}
          />
        </div>
      </div>
      <div className="flex items-center">
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
