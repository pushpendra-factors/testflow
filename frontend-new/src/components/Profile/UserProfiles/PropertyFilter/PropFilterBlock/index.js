import React, { useState, useEffect } from 'react';
import { Button } from 'antd';
import { SVG } from '../../../../factorsComponents';
import { fetchUserPropertyValues } from 'Reducers/coreQuery/services';
import { DEFAULT_OPERATOR_PROPS } from '../../../../FaFilterSelect/utils';
import FAFilterSelect from '../../../../FaFilterSelect';

export default function PropFilterBlock({
  index,
  filterProps,
  activeProject,
  operatorProps = DEFAULT_OPERATOR_PROPS,
  filter,
  deleteFilter,
  insertFilter,
  closeFilter,
}) {
  const [newFilterState, setNewFilterState] = useState({
    props: [],
    operator: '',
    values: [],
  });
  const [dropDownValues, setDropDownValues] = useState({});
  const [filterDropDownOptions, setFiltDD] = useState({
    props: [
      {
        label: 'User Properties',
        icon: 'user',
      },
    ],
    operator: operatorProps,
  });

  useEffect(() => {
    if (filter) {
      setValuesByProps(filter.props);
      setNewFilterState(filter);
    }
  }, [filter]);

  useEffect(() => {
    const filterDD = Object.assign({}, filterDropDownOptions);
    const propState = [];
    Object.keys(filterProps).forEach((k) => {
      propState.push({
        label: k,
        icon: k,
        values: filterProps[k],
      });
    });
    filterDD.props = propState;
    setFiltDD(filterDD);
  }, [filterProps]);

  const renderFilterContent = () => {
    return (
      <FAFilterSelect
        propOpts={filterDropDownOptions.props}
        operatorOpts={filterDropDownOptions.operator}
        valueOpts={dropDownValues}
        applyFilter={applyFilter}
        setValuesByProps={setValuesByProps}
        filter={filter}
      ></FAFilterSelect>
    );
  };

  useEffect(() => {
    if (newFilterState.props[1] === 'categorical') {
      if (newFilterState.props[2] === 'user') {
        if (!dropDownValues[newFilterState.props[0]]) {
          fetchUserPropertyValues(activeProject.id, newFilterState.props[0])
            .then((res) => {
              const ddValues = Object.assign({}, dropDownValues);
              ddValues[newFilterState.props[0]] = [...res.data, '$none'];
              setDropDownValues(ddValues);
            })
            .catch(() => {
              const ddValues = Object.assign({}, dropDownValues);
              ddValues[newFilterState.props[0]] = ['$none'];
              setDropDownValues(ddValues);
            });
        }
      }
    }
  }, [newFilterState]);

  const delFilter = () => {
    deleteFilter(index);
  };

  const applyFilter = (filterState) => {
    if (filterState) {
      insertFilter(filterState, index);
      closeFilter();
    }
  };

  const propOpByPayload = (props, index) => {
    if (props.length === 4) {
      return props[index + 1];
    } else {
      return props[index];
    }
  };

  const setValuesByProps = (props) => {
    if (propOpByPayload(props, 1) === 'categorical') {
      if (
        propOpByPayload(props, 2) === 'user' ||
        propOpByPayload(props, 2) === 'user_g'
      ) {
        if (!dropDownValues[props[0]]) {
          fetchUserPropertyValues(activeProject.id, propOpByPayload(props, 0))
            .then((res) => {
              const ddValues = Object.assign({}, dropDownValues);
              ddValues[propOpByPayload(props, 0)] = [...res.data, '$none'];
              setDropDownValues(ddValues);
            })
            .catch(() => {
              const ddValues = Object.assign({}, dropDownValues);
              ddValues[propOpByPayload(props, 0)] = ['$none'];
              setDropDownValues(ddValues);
            });
        }
      }
    }
  };

  const filterSelComp = () => {
    return (
      <FAFilterSelect
        propOpts={filterDropDownOptions.props}
        operatorOpts={filterDropDownOptions.operator}
        valueOpts={dropDownValues}
        applyFilter={applyFilter}
        setValuesByProps={setValuesByProps}
      ></FAFilterSelect>
    );
  };

  return (
    <div className={`flex items-center relative`}>
      <div className={`relative flex`}>
        {filter ? renderFilterContent() : filterSelComp()}
      </div>
      {delFilter && (
        <Button
          type='text'
          onClick={delFilter}
          size={'small'}
          className={`fa-btn--custom filter-buttons-margin btn-right-round filter-remove-button`}
        >
          <SVG name='remove' />
        </Button>
      )}
    </div>
  );
}
