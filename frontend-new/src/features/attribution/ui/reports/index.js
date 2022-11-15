import React, { useEffect, useMemo, useState } from 'react';
import { Button, Spin } from 'antd';
import { useDispatch, useSelector } from 'react-redux';
import { SVG, Text } from 'Components/factorsComponents';
import FaDatepicker from 'Components/FaDatepicker';
import SortableCards from './SortableCards';
import { fetchAttributionActiveUnits } from 'Attribution/state/services';
import { QUERY_TYPE_ATTRIBUTION } from 'Utils/constants';
import NoReports from './NoReports';
import { useHistory } from 'react-router-dom';

function Reports() {
  const dispatch = useDispatch();
  const history = useHistory();
  const { activeDashboard } = useSelector((state) => state.dashboard);
  const { active_project } = useSelector((state) => state.global);
  const { attributionDashboardUnits } = useSelector(
    (state) => state.attributionDashboard
  );
  const { data: savedQueries, loading: savedQueriesLoading } = useSelector(
    (state) => state.queries
  );

  useEffect(() => {
    dispatch(
      fetchAttributionActiveUnits(active_project.id, activeDashboard.id)
    );
  }, [activeDashboard]);

  const activeUnits = useMemo(
    () =>
      attributionDashboardUnits.data.filter(
        (elem) =>
          savedQueries.findIndex(
            (sq) =>
              sq.id === elem.query_id && sq.query.cl === QUERY_TYPE_ATTRIBUTION
          ) > -1
      ),
    [attributionDashboardUnits?.data, savedQueries]
  );

  if (attributionDashboardUnits?.loading || savedQueriesLoading) {
    return (
      <div className='flex items-center justify-center h-full w-full'>
        <div className='w-full h-64 flex items-center justify-center'>
          <Spin size='large' />
        </div>
      </div>
    );
  }

  if (!activeUnits || activeUnits?.length <= 0) {
    return <NoReports />;
  }

  return (
    <div className='flex flex-col items-center mt-16'>
      <div className='flex w-full justify-between items-center px-8 my-4'>
        <div className='flex items-center gap-4'>
          <Text
            type='title'
            level={6}
            weight='bold'
            color='black'
            extraClass='m-0'
          >
            Attribution Reports
          </Text>
          <FaDatepicker
            customPicker
            presetRange
            quarterPicker
            monthPicker
            buttonSize='large'
            placement='bottomRight'
            className='mr-2'
            onSelect={() => console.log('date selecter')}
          />
        </div>
        <div className='flex items-center gap-2'>
          <Button
            type='primary'
            size='large'
            onClick={() => history.push('/attribution/report')}
          >
            <SVG name='plus' color='white' className='w-full' /> Add Report
          </Button>
          <Button
            type='text'
            size='large'
            className='ml-1'
            style={{ padding: '4px 6px' }}
          >
            <SVG name='more' size={24} />
          </Button>
        </div>
      </div>
      <div className='w-full px-8 mt-2 flex flex-col'>
        {/* sortable cards */}
        <SortableCards activeUnits={activeUnits} />
      </div>
    </div>
  );
}

export default Reports;
