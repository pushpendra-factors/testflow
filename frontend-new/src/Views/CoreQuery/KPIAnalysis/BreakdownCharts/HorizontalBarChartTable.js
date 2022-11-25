import React, { useState, useEffect, useCallback } from 'react';
import { useSelector } from 'react-redux';
import { Spin } from 'antd';
import { Wait } from '../../../../utils/dataFormatter';
import {
  getHorizontalBarChartColumns,
  getDataInHorizontalBarChartFormat
} from './utils';
import DataTable from '../../../../components/DataTable';

function HorizontalBarChartTable({
  aggregateData,
  breakdown,
  cardSize = 1,
  isDashboardWidget = false,
  comparisonApplied = false
}) {
  const [loading, setLoading] = useState(true);
  const {
    userPropNames,
    eventPropertiesDisplayNames: eventPropertiesDisplayNamesState
  } = useSelector((state) => state.coreQuery);
  const { data: eventPropertiesDisplayNames } =
    eventPropertiesDisplayNamesState;
  const [columns, setColumns] = useState([]);
  const [data, setData] = useState([]);

  useEffect(() => {
    setColumns(
      getHorizontalBarChartColumns(
        breakdown,
        userPropNames,
        eventPropertiesDisplayNames,
        cardSize
      )
    );
  }, [breakdown, userPropNames, eventPropertiesDisplayNames, cardSize]);

  const formatDataAfterDelay = useCallback(async () => {
    setLoading(true);
    await Wait(500);
    setData(
      getDataInHorizontalBarChartFormat(
        aggregateData,
        breakdown,
        cardSize,
        isDashboardWidget,
        comparisonApplied
      )
    );
    setLoading(false);
  }, [aggregateData, breakdown, cardSize, isDashboardWidget, comparisonApplied]);

  useEffect(() => {
    formatDataAfterDelay();
  }, [formatDataAfterDelay]);

  if (loading) {
    return (
      <div className='h-64 flex items-center justify-center w-full'>
        <Spin size='small' />
      </div>
    );
  }

  return (
    <DataTable
      renderSearch={false}
      isWidgetModal={false}
      tableData={data}
      columns={columns}
      ignoreDocumentClick
      isPaginationEnabled={data.length > 100}
      defaultPageSize={100}
    />
  );
}

export default HorizontalBarChartTable;
