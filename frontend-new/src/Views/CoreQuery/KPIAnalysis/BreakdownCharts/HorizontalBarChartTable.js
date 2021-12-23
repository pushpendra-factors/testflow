import React, { useState, useEffect, useCallback } from 'react';
import { useSelector } from 'react-redux';
import { Wait } from '../../../../utils/dataFormatter';
import {
  getHorizontalBarChartColumns,
  getDataInHorizontalBarChartFormat,
} from './utils';
import { Spin } from 'antd';
import DataTable from '../../../../components/DataTable';

const HorizontalBarChartTable = ({
  aggregateData,
  breakdown,
  cardSize = 1,
  isDashboardWidget = false,
}) => {
  const [loading, setLoading] = useState(true);
  const { userPropNames, eventPropNames } = useSelector(
    (state) => state.coreQuery
  );
  const [columns, setColumns] = useState([]);
  const [data, setData] = useState([]);

  useEffect(() => {
    setColumns(
      getHorizontalBarChartColumns(
        breakdown,
        userPropNames,
        eventPropNames,
        cardSize
      )
    );
  }, [breakdown, userPropNames, eventPropNames, cardSize]);

  const formatDataAfterDelay = useCallback(async () => {
    setLoading(true);
    await Wait(500);
    setData(
      getDataInHorizontalBarChartFormat(
        aggregateData,
        breakdown,
        cardSize,
        isDashboardWidget
      )
    );
    setLoading(false);
  }, [aggregateData, breakdown, cardSize, isDashboardWidget]);

  useEffect(() => {
    formatDataAfterDelay();
  }, [formatDataAfterDelay]);

  return (
    <>
      {loading ? (
        <div className='h-64 flex items-center justify-center w-full'>
          <Spin size='small' />
        </div>
      ) : (
        <DataTable
          renderSearch={false}
          isWidgetModal={false}
          tableData={data}
          columns={columns}
          ignoreDocumentClick={true}
          isPaginationEnabled={data.length > 100}
          defaultPageSize={100}
        />
      )}
    </>
  );
};

export default HorizontalBarChartTable;
