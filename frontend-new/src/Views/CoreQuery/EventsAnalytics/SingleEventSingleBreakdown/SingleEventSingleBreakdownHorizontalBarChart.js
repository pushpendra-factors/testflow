import React, { memo, useMemo } from 'react';
import { useSelector } from 'react-redux';
import {
  getDataInHorizontalBarChartFormat,
  getHorizontalBarChartColumns
} from './utils';
import DataTable from '../../../../components/DataTable';

function SingleEventSingleBreakdownHorizontalBarChart({
  aggregateData,
  breakdown,
  cardSize = 1,
  isDashboardWidget = false,
  comparisonApplied = false
}) {
  const {
    userPropNames,
    eventPropertiesDisplayNames: eventPropertiesDisplayNamesState
  } = useSelector((state) => state.coreQuery);
  const { data: eventPropertiesDisplayNames } =
    eventPropertiesDisplayNamesState;
  const columns = useMemo(
    () =>
      getHorizontalBarChartColumns(
        breakdown,
        userPropNames,
        eventPropertiesDisplayNames
      ),
    [breakdown, userPropNames, eventPropertiesDisplayNames]
  );

  const data = useMemo(
    () =>
      getDataInHorizontalBarChartFormat(
        aggregateData,
        breakdown,
        cardSize,
        isDashboardWidget,
        comparisonApplied
      ),
    [aggregateData, breakdown, cardSize, isDashboardWidget, comparisonApplied]
  );

  return (
    <DataTable
      renderSearch={false}
      isWidgetModal={false}
      tableData={data}
      columns={columns}
      ignoreDocumentClick
      pagination={false}
      isPaginationEnabled={false}
    />
  );
}

export default memo(SingleEventSingleBreakdownHorizontalBarChart);
