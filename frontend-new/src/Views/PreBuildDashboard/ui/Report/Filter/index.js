import React, { useCallback } from 'react';
import { Button } from 'antd';
import { SVG } from 'Components/factorsComponents';
import { connect, useDispatch, useSelector } from 'react-redux';
import { setReportFilterPayloadAction } from 'Views/PreBuildDashboard/state/services';
import GlobalFilter from 'Components/GlobalFilter';

function Filter({ handleFilterChange }) {
  const dispatch = useDispatch();
  const filtersData = useSelector(
    (state) => state.preBuildDashboardConfig.reportFilters
  );

  const setFilterPayload = useCallback(
    (payload) => {
      dispatch(setReportFilterPayloadAction(payload));
      handleFilterChange(payload);
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
        <div className='ml-2 -mt-2'>{renderActions()}</div>
      </div>
    </div>
  );
}

const mapStateToProps = () => ({
  // activeDashboard: state.dashboard.activeDashboard
});

export default connect(mapStateToProps)(Filter);
