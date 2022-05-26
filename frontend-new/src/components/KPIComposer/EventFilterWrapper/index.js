/* eslint-disable */
import React, { useState, useEffect } from 'react';
import { useSelector, connect } from 'react-redux';
import styles from './index.module.scss';
import { DateRangePicker } from 'react-date-range';
import { Input, Button, Result } from 'antd';
import MomentTz from 'Components/MomentTz';
import { SVG, Text } from 'factorsComponents';
import { DEFAULT_DATE_RANGE } from '../DateRangeSelector/utils';
import { DEFAULT_OPERATOR_PROPS } from 'Components/FaFilterSelect/utils';

import {
  fetchEventPropertyValues,
  fetchUserPropertyValues,
  fetchChannelObjPropertyValues,
} from '../../../reducers/coreQuery/services';
import FaFilterSelectKPI from '../FaFilterSelectKPI';
import { fetchKPIFilterValues } from 'Reducers/kpi';
import _ from 'lodash';

const defaultOpProps = DEFAULT_OPERATOR_PROPS;

function EventFilterWrapper({
  index,
  blockType = 'event',
  filterType = 'analytics',
  typeProps,
  filterProps,
  activeProject,
  operatorProps = defaultOpProps,
  event,
  filter,
  delIcon = 'remove',
  propsConstants = ['user', 'event'],
  extraClass,
  delBtnClass,
  deleteFilter,
  insertFilter,
  closeFilter,
  fetchKPIFilterValues,
  KPI_config,
  selectedMainCategory,
}) {
  const [filterTypeState, setFilterTypeState] = useState('props');
  const [groupCollapseState, setGroupCollapse] = useState({});
  const [searchTerm, setSearchTerm] = useState('');
  const [newFilterState, setNewFilterState] = useState({
    props: [],
    operator: '',
    values: [],
  });

  const [dropDownValues, setDropDownValues] = useState({});
  const [selectedRngState, setSelectedRngState] = useState([
    { ...DEFAULT_DATE_RANGE },
  ]);

  const placeHolder = {
    props: 'Choose a property',
    operator: 'Choose an operator',
    values: 'Choose values',
  };

  const [filterDropDownOptions, setFiltDD] = useState({
    props: [
      {
        label: ' ',
        icon: 'mouseclick',
      },
    ],
    operator: operatorProps,
  });

  const { userPropNames } = useSelector((state) => state.coreQuery);

  useEffect(() => { 
    if (filter) { 
      setValuesByProps(filter.props);
      setNewFilterState(filter);

      if (filter && filter?.extra) {
        let filterData = {}; 
        if (selectedMainCategory?.category == 'channels') {
          filterData = {
            category: selectedMainCategory?.category,
            object_type: filter?.extra[3],
            property_name: filter?.extra[1],
            display_category: selectedMainCategory?.group,
            entity: 'event',
          };
        } else {
          filterData = {
            category: selectedMainCategory?.category,
            object_type: selectedMainCategory?.group,
            property_name: filter?.extra[1],
            entity: filter?.extra[3] ? filter?.extra[3] : filter?.extra[2],
          };
        } 
        fetchKPIFilterValues(activeProject.id, filterData)
          .then((res) => {
            const ddValues = Object.assign({}, dropDownValues);
            ddValues[filter?.extra[0]] = [...res.data, '$none'];
            setDropDownValues(ddValues);
          })
          .catch((err) => {
            const ddValues = Object.assign({}, dropDownValues);
            ddValues[filter?.extra[0]] = ['$none'];
            setDropDownValues(ddValues);
          });
      } 


    }
  }, [filter]);

  useEffect(() => {
    const filterDD = Object.assign({}, filterDropDownOptions);
    const propState = [];
    Object.keys(filterProps).forEach((k, i) => {
      propState.push({
        label: k,
        icon: k === 'event' ? 'mouseclick' : k,
        values: filterProps[k],
      });
    });
    let KPIlist = KPI_config || [];
    let selGroup = KPIlist.find((item) => {
      return item.display_category == event?.group;
    });
    let DDvalues = selGroup?.properties?.map((item) => {
      if (item == null) return;
      let ddName = item.display_name ? item.display_name : item.name;
      let ddtype =
        selGroup?.category == 'channels'
          ? item.object_type
          : item.entity
          ? item.entity
          : item.object_type;
      return [ddName, item.name, item.data_type, ddtype];
    });

    // filterDD.props = propState;
    filterDD.props = [
      {
        icon: 'user',
        label: 'user',
        values: DDvalues,
      },
    ]; 
    setFiltDD(filterDD);
  }, [filterProps]);
  
 


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
    if (props && props[3]) {
      let filterData = {};

      if (selectedMainCategory?.category == 'channels') {
        filterData = {
          category: selectedMainCategory?.category,
          object_type: props[3] ? props[3] : selectedMainCategory?.filters[0]?.extra[1] ,
          property_name: props[1],
          display_category: selectedMainCategory?.group,
          entity: 'event',
        };
      } else {
        filterData = {
          category: selectedMainCategory?.category,
          object_type: selectedMainCategory?.group,
          property_name: props[1],
          entity: props[3] ? props[3] : props[2],
        };
      }
      fetchKPIFilterValues(activeProject.id, filterData)
        .then((res) => {
          const ddValues = Object.assign({}, dropDownValues);
          ddValues[props[0]] = [...res.data, '$none'];
          setDropDownValues(ddValues);
        })
        .catch((err) => {
          const ddValues = Object.assign({}, dropDownValues);
          ddValues[props[0]] = ['$none'];
          setDropDownValues(ddValues);
        });
    } 
  };

  const renderFilterContent = () => {
    return (
      <FaFilterSelectKPI
        propOpts={filterDropDownOptions.props}
        operatorOpts={filterDropDownOptions.operator}
        valueOpts={dropDownValues}
        applyFilter={applyFilter}
        setValuesByProps={setValuesByProps}
        filter={filter}
      />
    );
  }; 

  return (
    <div className={`flex items-center relative w-full`}>
      {
        <Text level={8} type={'title'} extraClass={'m-0 mr-2'} weight={'thin'}>
          {index >= 1 ? 'and' : 'Filter by'}
        </Text>
      }
      <div className={`relative flex`}>
        {/* {filter ? renderFilterContent() : filterSelComp()} */}
        {renderFilterContent()}
      </div>
      {delFilter && (
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
  KPI_config: state.kpi?.config,
});

export default connect(mapStateToProps, { fetchKPIFilterValues })(
  EventFilterWrapper
);
