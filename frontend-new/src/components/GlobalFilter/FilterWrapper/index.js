import React, { useState, useEffect } from 'react';
import { SVG, Text } from 'factorsComponents';
import FaFilterSelect from '../../FaFilterSelect';
import { Button } from 'antd';
import {
  getEventPropertyValues,
  getGroupPropertyValues,
  getUserPropertyValues
} from 'Reducers/coreQuery/middleware';
import { bindActionCreators } from 'redux';
import { connect } from 'react-redux';
import { DEFAULT_OPERATOR_PROPS } from 'Components/FaFilterSelect/utils';
import getGroupIcon from 'Utils/getGroupIcon';
import { getPropertyGroupLabel } from './utils';
import { PropTextFormat } from 'Utils/dataFormatter';

function FilterWrapper({
  projectID,
  event,
  index,
  filterProps,
  groupName,
  filter,
  deleteFilter,
  insertFilter,
  closeFilter,
  refValue,
  showOr,
  caller,
  viewMode,
  dropdownPlacement,
  dropdownMaxHeight,
  showInList = false,
  delIcon = 'remove',
  hasPrefix = false,
  filterPrefix = 'Filter by',
  getEventPropertyValues,
  getGroupPropertyValues,
  getUserPropertyValues,
  propertyValuesMap,
  minEntriesPerGroup,
  operatorsMap = DEFAULT_OPERATOR_PROPS
}) {
  const [newFilterState, setNewFilterState] = useState({
    props: [],
    operator: '',
    values: []
  });
  const [filterDropDownOptions, setFiltDD] = useState({});

  useEffect(() => {
    if (filter && filter.props[2] === 'categorical') {
      setValuesByProps(filter.props);
      setNewFilterState(filter);
    }
  }, [filter]);

  useEffect(() => {
    const formattedFilterDDOptions = { ...filterDropDownOptions, props: [] };
    Object.keys(filterProps || {})?.forEach((propertyKey) => {
      if (!Array.isArray(filterProps[propertyKey])) {
        const propertyGroups = filterProps[propertyKey];
        if (propertyGroups) {
          Object.keys(propertyGroups)?.forEach((groupKey) => {
            const label = PropTextFormat(groupKey);
            const icon = getGroupIcon(groupKey);
            const values = propertyGroups?.[groupKey];
            formattedFilterDDOptions.props.push({
              key: propertyKey,
              label,
              icon,
              propertyType: propertyKey,
              values
            });
          });
        }
      } else {
        const label = getPropertyGroupLabel(propertyKey);
        const propertyType = ['user', 'event'].includes(propertyKey)
          ? propertyKey
          : ['button_click', 'page_view'].includes(propertyKey)
          ? 'event'
          : 'group'; //'button_click', 'page_view' custom types used in pathanalysis
        const values = filterProps[propertyKey];
        const icon = getGroupIcon(label);
        formattedFilterDDOptions.props.push({
          key: propertyKey,
          label,
          icon,
          propertyType: propertyType,
          values
        });
      }
    });

    formattedFilterDDOptions.operator = operatorsMap;
    setFiltDD(formattedFilterDDOptions);
  }, [filterProps]);

  const renderFilterContent = () => {
    return (
      <FaFilterSelect
        viewMode={viewMode}
        propOpts={filterDropDownOptions.props}
        operatorOpts={filterDropDownOptions.operator}
        valueOpts={propertyValuesMap.data}
        applyFilter={applyFilter}
        setValuesByProps={setValuesByProps}
        filter={filter}
        refValue={refValue}
        caller={caller}
        dropdownPlacement={dropdownPlacement}
        dropdownMaxHeight={dropdownMaxHeight}
        showInList={showInList}
        valueOptsLoading={propertyValuesMap.loading}
      />
    );
  };

  useEffect(() => {
    const [groupName, propertyName, propertyType, entity] =
      newFilterState.props;
    const propGrp = groupName || groupName !== '' ? groupName : entity;
    if (propertyType === 'categorical') {
      if (['user', 'user_g'].includes(propGrp)) {
        getUserPropertyValues(projectID, propertyName);
      } else if (propGrp === 'event') {
        getEventPropertyValues(projectID, event.label, propertyName);
      } else if (
        !['group', 'user', 'user_g'].includes(propGrp) &&
        ['group', 'user', 'user_g'].includes(entity)
      ) {
        getGroupPropertyValues(projectID, groupName, propertyName);
      }
    }
  }, [newFilterState?.props]);

  const delFilter = () => {
    deleteFilter(index);
  };

  const applyFilter = (filterState) => {
    if (filterState) {
      insertFilter(filterState, index);
      closeFilter();
    }
  };
  const setValuesByProps = (props) => {
    const [groupName, propertyName, propertyType, entity] = props;
    if (propertyType === 'categorical') {
      if (['user', 'user_g'].includes(groupName)) {
        getUserPropertyValues(projectID, propertyName);
      } else if (groupName === 'event') {
        getEventPropertyValues(projectID, event.label, propertyName);
      } else if (
        !['group', 'user', 'user_g'].includes(groupName) &&
        ['group', 'user', 'user_g'].includes(entity)
      ) {
        getGroupPropertyValues(projectID, groupName, propertyName);
      }
    }
  };

  const filterSelComp = () => (
    <FaFilterSelect
      viewMode={viewMode}
      propOpts={filterDropDownOptions.props}
      operatorOpts={filterDropDownOptions.operator}
      valueOpts={propertyValuesMap.data}
      applyFilter={applyFilter}
      refValue={refValue}
      setValuesByProps={setValuesByProps}
      caller={caller}
      dropdownPlacement={dropdownPlacement}
      dropdownMaxHeight={dropdownMaxHeight}
      showInList={showInList}
      minEntriesPerGroup={minEntriesPerGroup}
      valueOptsLoading={propertyValuesMap.loading}
    />
  );

  return (
    <div
      className={`flex items-center relative ${
        caller === 'profiles' ? 'mb-2' : 'mb-2'
      }`}
    >
      {!showOr && hasPrefix && (
        <Text
          level={8}
          type={'title'}
          extraClass={`m-0 ${
            caller === 'profiles'
              ? 'mx-3'
              : index >= 1
              ? 'mr-16 ml-10'
              : 'mx-10'
          } ${filterPrefix?.split(' ')?.length ? 'whitespace-no-wrap' : ''}`}
          weight={'thin'}
        >
          {index >= 1 ? 'and' : filterPrefix}
        </Text>
      )}
      {showOr && (
        <Text
          level={8}
          type={'title'}
          extraClass={`m-0 ${caller === 'profiles' ? 'mx-3' : 'mx-2'}`}
          weight={'thin'}
        >
          or
        </Text>
      )}
      <div className={`relative flex`}>
        {filter ? renderFilterContent() : filterSelComp()}
      </div>
      {delFilter && !viewMode && (
        <Button
          type='text'
          onClick={delFilter}
          size={'small'}
          className={`fa-btn--custom filter-buttons-margin btn-right-round filter-remove-button`}
        >
          <SVG name={delIcon} />
        </Button>
      )}
    </div>
  );
}

const mapStateToProps = (state) => ({
  propertyValuesMap: state.coreQuery.propertyValuesMap
});

const mapDispatchToProps = (dispatch) =>
  bindActionCreators(
    { getEventPropertyValues, getGroupPropertyValues, getUserPropertyValues },
    dispatch
  );

export default connect(mapStateToProps, mapDispatchToProps)(FilterWrapper);
