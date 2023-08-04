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
import { PropTextFormat } from 'Utils/dataFormatter';
import { DEFAULT_OPERATOR_PROPS } from 'Components/FaFilterSelect/utils';
import getGroupIcon from 'Utils/getGroupIcon';
import startCase from 'lodash/startCase';

function FilterWrapper({
  viewMode,
  index,
  groupName,
  filterProps,
  refValue,
  projectID,
  event,
  filter,
  delIcon = 'remove',
  deleteFilter,
  insertFilter,
  closeFilter,
  showOr,
  hasPrefix = false,
  filterPrefix = 'Filter by',
  caller,
  dropdownPlacement,
  dropdownMaxHeight,
  showInList = false,
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
    if (filter && filter.props[1] === 'categorical') {
      setValuesByProps(filter.props);
      setNewFilterState(filter);
    }
  }, [filter]);

  useEffect(() => {
    const filterDD = { ...filterDropDownOptions, props: [] };
    Object.keys(filterProps)?.forEach((key) => {       
      if (!Array.isArray(filterProps[key])) {
        const groups = filterProps[key];
        if (groups) {
          Object.keys(groups)?.forEach((groupKey) => {
            const label = startCase(groupKey);
            const icon = getGroupIcon(groupKey);
            const values = groups?.[groupKey];
            filterDD.props.push({
              label,
              icon,
              propertyType: key,
              values
            });
          });
        }
      } else {
        const label = `${PropTextFormat(key)} Properties`;
        const icon = ['user', 'event'].includes(key) ? key : ['button_click', 'page_view'].includes(key) ? 'event' : 'group'; //'button_click', 'page_view' custom types used in pathanalysis
        const values = filterProps[key];
        filterDD.props.push({
          label,
          icon,
          propertyType: icon,
          values
        });
      }
    });
    filterDD.operator = operatorsMap;
    setFiltDD(filterDD);
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
    if (newFilterState.props[1] === 'categorical') {
      if (
        newFilterState.props[2] === 'user' ||
        newFilterState.props[2] === 'user_g'
      ) {
        getUserPropertyValues(projectID, newFilterState.props[0]);
      } else if (newFilterState.props[2] === 'event') {
        getEventPropertyValues(projectID, event.label, newFilterState.props[0]);
      } else if (newFilterState.props[2] === 'group') {
        let group = groupName;
        if (groupName === 'All') {
          if (newFilterState.props[0].toLowerCase().includes('hubspot'))
            group = '$hubspot_company';
          if (newFilterState.props[0].toLowerCase().includes('salesforce'))
            group = '$salesforce_account';
          if (newFilterState.props[0].toLowerCase().includes('6signal'))
            group = '$6signal';
          if (newFilterState.props[0].toLowerCase().includes('$li_'))
            group = '$linkedin_company';
          if (newFilterState.props[0].toLowerCase().includes('$g2_'))
            group = '$g2';
        }
        getGroupPropertyValues(projectID, group, newFilterState.props[0]);
      }
    }
  }, [
    newFilterState.props[0],
    newFilterState.props[1],
    newFilterState.props[2]
  ]);

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
    if (props[2] === 'categorical') {
      if (props[3] === 'user' || props[3] === 'user_g') {
        getUserPropertyValues(projectID, props[1]);
      } else if (props[3] === 'event') {
        getEventPropertyValues(projectID, event.label, props[1]);
      } else if (props[3] === 'group') {
        let group = groupName;
        if (groupName === 'All') {
          if (props[1].toLowerCase().includes('hubspot'))
            group = '$hubspot_company';
          if (props[1].toLowerCase().includes('salesforce'))
            group = '$salesforce_account';
          if (props[1].toLowerCase().includes('6signal')) group = '$6signal';
          if (props[1].toLowerCase().includes('$li_'))
            group = '$linkedin_company';
          if (props[1].toLowerCase().includes('$g2_')) group = '$g2';
        }
        getGroupPropertyValues(projectID, group, props[1]);
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
