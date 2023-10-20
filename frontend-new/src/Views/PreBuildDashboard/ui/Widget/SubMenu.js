import React, { useCallback, useState } from 'react';
import { Button, Tooltip } from 'antd';
import { Text, SVG } from 'Components/factorsComponents';
import FaDatepicker from 'Components/FaDatepicker';
import { connect, useDispatch, useSelector } from 'react-redux';
import PropertyFilter from 'Components/Profile/MyComponents/PropertyFilter';
import { setFilterPayloadAction } from 'Views/PreBuildDashboard/state/services';

const SubMenu = ({config, durationObj, handleDurationChange, activeDashboard}) => {

    const dispatch = useDispatch();
    const filtersData = useSelector((state) => state.preBuildDashboardConfig.filters);

    const setFilterPayload = useCallback(
      (payload) => {
        dispatch(setFilterPayloadAction(payload));
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
            <PropertyFilter
              profileType='predefined'
              filters={filtersData}
              setFilters={setFilters}
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
          <div className='flex justify-between'>
            {renderPropertyFilter()}
          </div>
          <div className='flex items-center justify-between'>
            {filtersData?.length ? renderClearFilterButton() : null}
          </div>
        </div>
      );

  return (
    <div className={'flex items-center px-0'}>
         <div className={'flex justify-between items-center mr-2'}>
            <Text type={'title'} level={7} extraClass={'m-0 mr-2'}>
              Data from
            </Text>
            <FaDatepicker
              customPicker
              nowPicker={activeDashboard?.name === 'Website Aggregation' ? true : false}
              presetRange
              quarterPicker
              range={{
                startDate: durationObj.from,
                endDate: durationObj.to
              }}
              placement='bottomLeft'
              onSelect={handleDurationChange}
              buttonSize={'default'}
              className={'datepicker-minWidth'}
            />
          </div>
        {renderActions()}
    </div>
  );
}

const mapStateToProps = () => {
  return {
    // activeDashboard: state.dashboard.activeDashboard
  };
};

export default connect(mapStateToProps)(SubMenu);
