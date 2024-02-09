import React, { useState, useEffect } from 'react';
import { keys } from 'lodash';
import { Button } from 'antd';
import {
  getEventPropertyValues,
  getGroupPropertyValues,
  getPredefinedPropertyValues,
  getUserPropertyValues
} from 'Reducers/coreQuery/middleware';
import { bindActionCreators } from 'redux';
import { connect, useSelector } from 'react-redux';
import { PropTextFormat } from 'Utils/dataFormatter';
import { DEFAULT_OPERATOR_PROPS } from 'Components/FaFilterSelect/utils';
import getGroupIcon from 'Utils/getGroupIcon';
import { selectActivePreDashboard } from 'Reducers/dashboard/selectors';
import { SVG, Text } from 'Components/factorsComponents';
import FaFilterSelect from '../../FaFilterSelect';
import { CustomGroupDisplayNames } from './utils';

function FilterWrapper({
  projectID,
  event,
  index,
  filterProps,
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
  getPredefinedPropertyValues,
  propertyValuesMap,
  minEntriesPerGroup,
  operatorsMap = DEFAULT_OPERATOR_PROPS,
  groups,
  profileType = ''
}) {
  const [newFilterState, setNewFilterState] = useState({
    props: [],
    operator: '',
    values: []
  });
  const [filterDropDownOptions, setFiltDD] = useState({});
  const [extraProps, setExtraProps] = useState({});
  const activeDashboard = useSelector((state) =>
    selectActivePreDashboard(state)
  );

  function transformProperty(data) {
    const result = {};
    data.forEach((item) => {
      const [displayName, name] = item;
      result[name] = displayName;
    });
    return result;
  }

  useEffect(() => {
    if (profileType === 'predefined') {
      const output = transformProperty(filterProps?.user);
      const extraProp = {
        displayNames: output
      };
      setExtraProps(extraProp);
    }
  }, [filterProps, profileType]);

  useEffect(() => {
    if (
      filter &&
      (filter.props[2] === 'categorical' || filter.props[1] === 'categorical')
    ) {
      setValuesByProps(filter.props);
      setNewFilterState(filter);
    }
  }, [filter]);

  useEffect(() => {
    const formatFilterOptions = (key, values) => {
      if (!Array.isArray(values)) {
        const categorisedProperties = values;
        return keys(categorisedProperties).map((category) => {
          const label = PropTextFormat(category);
          const icon = getGroupIcon(category);

          return {
            key,
            label,
            icon,
            propertyType: key,
            values: categorisedProperties[category]
          };
        });
      }
      const label =
        CustomGroupDisplayNames[key] ||
        (groups?.all_groups?.[key]
          ? groups.all_groups[key]
          : PropTextFormat(key));

      const propertyType = ['user', 'event'].includes(key)
        ? key
        : ['button_click', 'page_view'].includes(key)
          ? 'event'
          : 'group';

      const icon = getGroupIcon(label);

      return [
        {
          key,
          label,
          icon,
          propertyType,
          values
        }
      ];
    };

    const formattedFilterDDOptions = { ...filterDropDownOptions, props: [] };

    Object.keys(filterProps || {}).forEach((propertyKey) => {
      formattedFilterDDOptions.props.push(
        ...formatFilterOptions(propertyKey, filterProps[propertyKey])
      );
    });

    formattedFilterDDOptions.operator = operatorsMap;
    setFiltDD(formattedFilterDDOptions);
  }, [filterProps]);

  useEffect(() => {
    const [groupName, propertyName, propertyType, entity] =
      newFilterState.props;
    const propGrp = groupName || groupName !== '' ? groupName : entity;

    const payload =
      newFilterState?.props?.length === 3
        ? newFilterState?.props[0]
        : newFilterState?.props[1];

    if (profileType === 'predefined' && payload) {
      getPredefinedPropertyValues(
        projectID,
        payload,
        activeDashboard?.inter_id
      );
    } else if (propertyType === 'categorical') {
      if (['user', 'user_g'].includes(propGrp)) {
        getUserPropertyValues(projectID, propertyName);
      } else if (['event', 'page_view', 'button_click'].includes(propGrp)) {
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
    const payload = props?.length === 3 ? props[0] : props[1];
    if (profileType === 'predefined' && payload) {
      getPredefinedPropertyValues(
        projectID,
        payload,
        activeDashboard?.inter_id
      );
    } else if (propertyType === 'categorical') {
      if (['user', 'user_g'].includes(groupName)) {
        getUserPropertyValues(projectID, propertyName);
      } else if (['event', 'page_view', 'button_click'].includes(groupName)) {
        getEventPropertyValues(projectID, event.label, propertyName);
      } else if (
        !['group', 'user', 'user_g'].includes(groupName) &&
        ['group', 'user', 'user_g'].includes(entity)
      ) {
        getGroupPropertyValues(projectID, groupName, propertyName);
      }
    }
  };

  const renderFilterContent = () => (
    <FaFilterSelect
      viewMode={viewMode}
      propOpts={filterDropDownOptions.props}
      operatorOpts={filterDropDownOptions.operator}
      valueOpts={propertyValuesMap.data}
      applyFilter={applyFilter}
      refValue={refValue}
      setValuesByProps={setValuesByProps}
      filter={filter}
      caller={caller}
      dropdownPlacement={dropdownPlacement}
      dropdownMaxHeight={dropdownMaxHeight}
      showInList={showInList}
      valueOptsLoading={propertyValuesMap.loading}
      extraProps={extraProps}
    />
  );

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
      extraProps={extraProps}
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
          type='title'
          extraClass={`m-0 ${
            caller === 'profiles'
              ? 'mx-3'
              : index >= 1
                ? 'mr-16 ml-10'
                : 'mx-10'
          } ${filterPrefix?.split(' ')?.length ? 'whitespace-no-wrap' : ''}`}
          weight='thin'
        >
          {index >= 1 ? 'and' : filterPrefix}
        </Text>
      )}
      {showOr && (
        <Text
          level={8}
          type='title'
          extraClass={`m-0 ${caller === 'profiles' ? 'mx-3' : 'mx-2'}`}
          weight='thin'
        >
          or
        </Text>
      )}
      <div className='relative flex'>
        {filter ? renderFilterContent() : filterSelComp()}
      </div>
      {delFilter && !viewMode && (
        <Button
          type='text'
          onClick={delFilter}
          size='small'
          className='fa-btn--custom filter-buttons-margin btn-right-round filter-remove-button'
        >
          <SVG name={delIcon} />
        </Button>
      )}
    </div>
  );
}

const mapStateToProps = (state) => ({
  propertyValuesMap: state.coreQuery.propertyValuesMap,
  groups: state.coreQuery.groups
});

const mapDispatchToProps = (dispatch) =>
  bindActionCreators(
    {
      getEventPropertyValues,
      getGroupPropertyValues,
      getUserPropertyValues,
      getPredefinedPropertyValues
    },
    dispatch
  );

export default connect(mapStateToProps, mapDispatchToProps)(FilterWrapper);
