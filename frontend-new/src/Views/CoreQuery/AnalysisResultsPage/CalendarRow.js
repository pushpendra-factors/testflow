import React, { useCallback } from "react";
import FaDatepicker from "../../../components/FaDatepicker";
import ChartTypeDropdown from "../../../components/ChartTypeDropdown";
import { QUERY_TYPE_WEB } from "../../../utils/constants";

function CalendarRow({
  durationObj,
  handleDurationChange,
  handleChartTypeChange,
  chartTypeMenuItems,
  chartType,
  metricsDropdown,
  queryType
}) {
  const setDateRange = useCallback(
    (range) => {
      handleDurationChange(range);
    },
    [handleDurationChange]
  );

  if(queryType === QUERY_TYPE_WEB) {
    return null;
  }


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
