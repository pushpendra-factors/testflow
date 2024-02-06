import React, { useCallback, useEffect, useMemo, useState } from 'react';
import { Button } from 'antd';
import { Text, SVG } from 'Components/factorsComponents';
import { connect, useDispatch, useSelector } from 'react-redux';
import { setReportFilterPayloadAction } from 'Views/PreBuildDashboard/state/services';
import CardLayout from 'Components/CardLayout';
import cx from 'classnames';
import ControlledComponent from 'Components/ControlledComponent';
import FilterWrapper from 'Components/GlobalFilter/FilterWrapper';
import { cloneDeep, isEqual } from 'lodash';
import styles from './index.module.scss';

function Filter({ handleFilterChange }) {
  const dispatch = useDispatch();
  const filtersData = useSelector(
    (state) => state.preBuildDashboardConfig.reportFilters
  );
  const predefinedProperty = useSelector(
    (state) => state.preBuildDashboardConfig.config.data.result
  );
  const activeProject = useSelector((state) => state.global.active_project);
  const [appliedFilters, setAppliedFilters] = useState([]);
  const [filterDD, setFilterDD] = useState(false);
  const [selectedFilters, setSelectedFilters] = useState([]);
  const [filtersExpanded, setFiltersExpanded] = useState(false);

  const mainFilterProps = useMemo(() => {
    const props = {};
    props.user = predefinedProperty?.pr;
    return props;
  }, [predefinedProperty]);

  useEffect(() => {
    if (filtersData) {
      setAppliedFilters(filtersData);
      setSelectedFilters(filtersData);
    }
  }, [filtersData]);

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

  const showFilterDropdown = useCallback(() => {
    setFilterDD(true);
  }, []);

  const handleCloseFilter = useCallback(() => {
    setFilterDD(false);
  }, []);

  const handleInsertFilter = useCallback(
    (filter, index) => {
      if (selectedFilters.length === index) {
        setSelectedFilters([...selectedFilters, filter]);
      } else {
        setSelectedFilters([
          ...selectedFilters.slice(0, index),
          filter,
          ...selectedFilters.slice(index + 1)
        ]);
      }
    },
    [selectedFilters, setSelectedFilters]
  );

  const handleDeleteFilter = useCallback(
    (filterIndex) => {
      if (filterIndex === selectedFilters.length) {
        setFilterDD(false);
        return;
      }
      setSelectedFilters(
        selectedFilters.filter((_, index) => index !== filterIndex)
      );
    },
    [setSelectedFilters, selectedFilters]
  );

  const toggleFilters = useCallback(() => {
    setFiltersExpanded((curr) => !curr);
  }, [dispatch, setFiltersExpanded]);

  const handleClearFilters = useCallback(() => {
    toggleFilters();
    clearFilters();
    setAppliedFilters([]);
    setSelectedFilters([]);
  }, [toggleFilters]);

  const applyFilters = useCallback(() => {
    setAppliedFilters(cloneDeep(selectedFilters));
    setFiltersExpanded(false);
    setFilters(selectedFilters);
  }, [selectedFilters, activeProject.id]);

  const checkApplyButtonState = () => {
    if (selectedFilters.length > 0) {
      return {
        applyButtonDisabled: false
      };
    }
    const areFiltersEqual = isEqual(selectedFilters, appliedFilters);
    const applyButtonDisabled = areFiltersEqual === true;
    return { applyButtonDisabled };
  };

  const { applyButtonDisabled } = useMemo(
    () => checkApplyButtonState(),
    [appliedFilters, selectedFilters]
  );

  const showClearAllButton = useMemo(
    () => appliedFilters.length === 0,
    [appliedFilters.length]
  );

  const footerActionsBtn = () => (
    <div className='flex items-center justify-between'>
      <Button
        disabled={showClearAllButton}
        type='text'
        icon={
          <SVG
            name='times_circle'
            size={16}
            color={`${showClearAllButton ? '#00000040' : 'grey'}`}
            extraClass='-mt-1'
          />
        }
        onClick={handleClearFilters}
      >
        Clear filters
      </Button>
      <Button
        disabled={applyButtonDisabled}
        onClick={applyFilters}
        type='primary'
      >
        Apply filters
      </Button>
    </div>
  );

  const filterChildren = () => (
    <>
      <div className={cx('px-6 pb-1', styles['section-title-container'])}>
        <Text
          type='title'
          color='character-secondary'
          extraClass='mb-0'
          weight='medium'
        >
          Filter data
        </Text>
      </div>
      <div className='px-6'>
        <ControlledComponent controller={selectedFilters.length > 0}>
          {selectedFilters.map((filter, index) => (
            <FilterWrapper
              key={index}
              viewMode={false}
              projectID={activeProject?.id}
              filter={filter}
              index={index}
              filterProps={mainFilterProps}
              minEntriesPerGroup={3}
              insertFilter={handleInsertFilter}
              closeFilter={handleCloseFilter}
              deleteFilter={handleDeleteFilter}
              profileType='predefined'
            />
          ))}
        </ControlledComponent>

        <ControlledComponent controller={filterDD === true}>
          <FilterWrapper
            viewMode={false}
            projectID={activeProject?.id}
            index={selectedFilters.length}
            filterProps={mainFilterProps}
            minEntriesPerGroup={3}
            insertFilter={handleInsertFilter}
            closeFilter={handleCloseFilter}
            deleteFilter={handleDeleteFilter}
            profileType='predefined'
          />
        </ControlledComponent>

        <Button
          className={cx(
            'flex items-center col-gap-2',
            styles['add-filter-button']
          )}
          type='text'
          onClick={showFilterDropdown}
        >
          <SVG name='plus' color='#00000073' />
          <Text
            type='title'
            color='character-title'
            extraClass='mb-0'
            weight='medium'
          >
            Add filter
          </Text>
        </Button>
      </div>
    </>
  );

  const footerActions = <div>{footerActionsBtn()}</div>;
  const children = <>{filterChildren()}</>;

  const renderPropertyFilter = () => (
    <CardLayout footerActions={footerActions}>{children}</CardLayout>
  );

  const changeBtnText = () => {
    if (filtersExpanded === false && appliedFilters.length > 0) {
      return ['View', 1];
    }
    if (filtersExpanded === false && appliedFilters.length === 0) {
      return ['Add', 2];
    }
    if (filtersExpanded === true) {
      return ['Hide', 3];
    }
    return ['Add', 2];
  };

  return (
    <>
      <div className='flex justify-end mb-2'>
        <Button
          className={cx(
            'flex items-center justify-center',
            styles['filter-button']
          )}
          onClick={toggleFilters}
        >
          <SVG
            size={18}
            name={`${changeBtnText()?.[1] === 3 ? 'CaretUp' : 'filter'}`}
            extraClass={`mr-1 ${changeBtnText()?.[1] === 3 && '-mt-1'}`}
            color='#8C8C8C'
          />
          <Text
            type='title'
            extraClass='mb-0'
            weight='medium'
            color='character-primary'
          >
            {changeBtnText()?.[0]}{' '}
            {changeBtnText()?.[1] === 1 && appliedFilters.length} filter
            {changeBtnText()?.[1] === 1 && '(s)'}
          </Text>
        </Button>
      </div>
      {filtersExpanded === true && (
        <div className='mt-4'>{renderPropertyFilter()}</div>
      )}
    </>
  );
}

const mapStateToProps = () => ({});

export default connect(mapStateToProps)(Filter);
