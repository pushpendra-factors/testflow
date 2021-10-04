import React, { memo, useMemo } from 'react';
import {
  getDataInHorizontalBarChartFormat,
  getHorizontalBarChartColumns,
} from './utils';
import { useSelector } from 'react-redux';
import DataTable from '../../../../components/DataTable';

const SingleEventSingleBreakdownHorizontalBarChart = ({
  aggregateData,
  breakdown,
  cardSize = 1,
  isDashboardWidget = false,
}) => {
  const { userPropNames, eventPropNames } = useSelector(
    (state) => state.coreQuery
  );
  const columns = useMemo(() => {
    return getHorizontalBarChartColumns(
      breakdown,
      userPropNames,
      eventPropNames
    );
  }, [breakdown, userPropNames, eventPropNames]);

  const data = useMemo(() => {
    return getDataInHorizontalBarChartFormat(
      aggregateData,
      breakdown,
      cardSize,
      isDashboardWidget,
    );
  }, [aggregateData, breakdown, cardSize, isDashboardWidget]);

  return (
    <DataTable
      renderSearch={false}
      isWidgetModal={false}
      tableData={data}
      columns={columns}
      ignoreDocumentClick={true}
      pagination={false}
      isPaginationEnabled={false}
    />
  );
};

export default memo(SingleEventSingleBreakdownHorizontalBarChart);
