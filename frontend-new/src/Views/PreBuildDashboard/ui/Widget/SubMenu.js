import React, { useCallback, useState } from 'react';
import { Button, Tooltip } from 'antd';
import { Text, SVG } from 'Components/factorsComponents';
import FaDatepicker from 'Components/FaDatepicker';
import { connect, useDispatch, useSelector } from 'react-redux';
import {
  setFilterPayloadAction,
  setReportFilterPayloadAction
} from 'Views/PreBuildDashboard/state/services';
import GlobalFilter from 'Components/GlobalFilter';

function SubMenu({
  config,
  durationObj,
  handleDurationChange,
  activeDashboard
}) {
  const dispatch = useDispatch();
  const filtersData = useSelector(
    (state) => state.preBuildDashboardConfig.filters
  );

  const setFilterPayload = useCallback(
    (payload) => {
      dispatch(setFilterPayloadAction(payload));
      dispatch(setReportFilterPayloadAction(payload));
    },
    [dispatch]
  );

  const setFilters = (filters) => {
    setFilterPayload(filters);
  };

  const clearFilters = () => {
    setFilterPayload([]);
  };

  const renderPropertyFilter = () => (
    <div key={0} className='max-w-3xl'>
      <GlobalFilter
        profileType='predefined'
        filters={filtersData}
        setGlobalFilters={setFilters}
      />
    </div>
  );

  const renderClearFilterButton = () => (
    <Button
      className='dropdown-btn large mr-2'
      type='text'
      icon={<SVG name='times_circle' size={16} />}
      onClick={clearFilters}
    >
      Clear Filters
    </Button>
  );

  const renderActions = () => (
    <div className='flex justify-between items-start px-0 mt-2'>
      <div className='flex justify-between'>{renderPropertyFilter()}</div>
      <div className='flex items-center justify-between'>
        {filtersData?.length ? renderClearFilterButton() : null}
      </div>
    </div>
  );

  return (
    <div className='flex items-center px-0'>
      <div className='flex justify-between'>
        <FaDatepicker
          customPicker
          nowPicker={activeDashboard?.name === 'Website Aggregation'}
          presetRange
          quarterPicker
          range={{
            startDate: durationObj.from,
            endDate: durationObj.to
          }}
          placement='bottomLeft'
          onSelect={handleDurationChange}
          buttonSize='default'
          className='datepicker-minWidth'
        />
        <div className='ml-2 -mt-2'>{renderActions()}</div>
      </div>
    </div>
  );
}

const mapStateToProps = () => ({
  // activeDashboard: state.dashboard.activeDashboard
});

export default connect(mapStateToProps)(SubMenu);
