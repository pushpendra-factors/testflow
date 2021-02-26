import React, { useState, useEffect, useCallback } from "react";
import FaDatepicker from "../../../components/FaDatepicker";
import ChartTypeDropdown from "../../../components/ChartTypeDropdown";
import { QUERY_TYPE_ATTRIBUTION, QUERY_TYPE_WEB } from "../../../utils/constants";
import styles from './index.module.scss';
import { Button } from "antd";
import { SVG, Spiner} from "../../../components/factorsComponents";

function CalendarRow({
  durationObj,
  handleDurationChange,
  handleChartTypeChange,
  chartTypeMenuItems,
  chartType,
  metricsDropdown,
  queryType,
  triggerAttrComparision,
  cmprResultState
}) {

  const [comparisionEnabled, enableComparision] = useState(false);
  const [isCompareCalendar, showCompareCalendar] = useState(false);
  const [selComparRange, setSelComparRange] = useState();

  useEffect(() => {
    if(queryType === QUERY_TYPE_ATTRIBUTION) {
      enableComparision(true);
    }
  }, [queryType])

  const setDateRange = useCallback(
    (range) => {
      handleDurationChange(range);
    },
    [handleDurationChange]
  );

  if(queryType === QUERY_TYPE_WEB) {
    return null;
  }

  const delComparision = () => {
    showCompareCalendar(false);
    triggerAttrComparision(null);
  }

  const raiseComparisionRange = (rnge) => {
    const dtRange = {};
        if (rnge && rnge.startDate) {
            if (Array.isArray(rnge.startDate)) {
                dtRange.from = rnge.startDate[0]
                dtRange.to = rnge.startDate[1];
            } else {
                dtRange.from = rnge.startDate;
                dtRange.to = rnge.endDate;
            }
        }
    setSelComparRange(dtRange);
    triggerAttrComparision(dtRange);
  }

  const renderCompareScenario = () => {
    return (
      <div className={`${styles.comparisonCalendar} flex items-center ml-4 mr-4`}>
        {!isCompareCalendar && <Button icon={<SVG name="compare" />} onClick={() => showCompareCalendar(true)}></Button>}

        

        {(isCompareCalendar) &&
          <>
            <SVG name="compare"/>
            <div className="calendar ml-4">
              <FaDatepicker
                customPicker
                presetRange
                monthPicker
                placement="topRight"
                range={selComparRange && {
                  startDate: selComparRange.from,
                  endDate: selComparRange.to,
                }}
                onSelect={raiseComparisionRange} 
              />
            </div>
          </> 
        }

        {cmprResultState && cmprResultState.data && 
          <Button className={`ml-2`} icon={<SVG name="times" />} onClick={() => delComparision()}></Button>}

        {cmprResultState && cmprResultState.loading && <Spiner size={`small`}></Spiner>}


      </div>
    )
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
            range={{
              startDate: durationObj.from,
              endDate: durationObj.to,
            }}
            placement="topRight"
            onSelect={setDateRange} 
          />
        </div>
        {comparisionEnabled && renderCompareScenario()}
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
