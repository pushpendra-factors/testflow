import React, { useState, useEffect, useCallback } from 'react';
import {
  getHorizontalBarChartColumns,
  getDataInHorizontalBarChartFormat,
} from './utils';
import DataTable from '../../../../components/DataTable';

const HorizontalBarChartTable = ({
  data,
  queries,
  groupAnalysis,
  cardSize = 1,
  isDashboardWidget = false,
}) => {
  const [columns, setColumns] = useState([]);
  const [chartData, setChartData] = useState([]);

  useEffect(() => {
    setColumns(getHorizontalBarChartColumns());
  }, []);

  const formatData = useCallback(async () => {
    setChartData(
      getDataInHorizontalBarChartFormat(
        data,
        queries,
        groupAnalysis,
        cardSize,
        isDashboardWidget
      )
    );
  }, [data, cardSize, queries, groupAnalysis, isDashboardWidget]);

  useEffect(() => {
    formatData();
  }, [formatData]);

  return (
    <DataTable
      renderSearch={false}
      isWidgetModal={false}
      tableData={chartData}
      columns={columns}
      ignoreDocumentClick={true}
      isPaginationEnabled={false}
    />
  );
};

export default HorizontalBarChartTable;
