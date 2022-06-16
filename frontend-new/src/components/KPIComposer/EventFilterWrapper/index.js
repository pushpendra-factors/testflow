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
  refValue,
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
  showOr,
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
            category: event?.category, //use event instead of selectedMainCategory since it is in induvidual level
            object_type: filter?.extra[3],
            property_name: filter?.extra[1],
            display_category: selectedMainCategory?.group,
            entity: 'event',
          };
        } else {
          filterData = {
            category: event?.category, //use event instead of selectedMainCategory since it is in induvidual level
            object_type: event?.group,
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
  
 

  // const parseDateRangeFilter = (fr, to) => {
  //   return (
  //     MomentTz(fr).format('MMM DD, YYYY') +
  //     ' - ' +
  //     MomentTz(to).format('MMM DD, YYYY')
  //   );
  // };

  // const renderFilterContent = () => {
  //   return (
  //     <FaFilterSelectKPI
  //       propOpts={filterDropDownOptions.props}
  //       operatorOpts={filterDropDownOptions.operator}
  //       valueOpts={dropDownValues}
  //       applyFilter={applyFilter}
  //       setValuesByProps={setValuesByProps}
  //       filter={filter}
  //       refValue={refValue}
  //     />
  //   );
  // };

  // const onSelectSearch = (userInput) => {
  //   if (!userInput.currentTarget.value.length) {
  //     if (userInput.keyCode === 8 || userInput.keyCode === 46) {
  //       removeFilter();
  //       return;
  //     }
  //   } else if (
  //     filterTypeState === 'values' &&
  //     userInput.keyCode === 13 &&
  //     newFilterState.props[1] === 'numerical'
  //   ) {
  //     const newFilter = Object.assign({}, newFilterState);
  //     newFilter[filterTypeState].push(userInput.currentTarget.value);
  //     changeFilterTypeState();
  //     insertFilter(newFilter);
  //     closeFilter();
  //   }
  //   setSearchTerm(userInput.currentTarget.value);

  //   if (
  //     (newFilterState.operator === 'contains' ||
  //       newFilterState.operator === 'does not contain') &&
  //     filterTypeState === 'values'
  //   ) {
  //     const newFilter = Object.assign({}, newFilterState);
  //     newFilter[filterTypeState][0]
  //       ? (newFilter[filterTypeState][0] =
  //           newFilter[filterTypeState][0] + userInput.currentTarget.value)
  //       : (newFilter[filterTypeState][0] = userInput.currentTarget.value);
  //     setNewFilterState(newFilter);
  //     setSearchTerm('');
  //   }
  // };

  // const removeFilter = () => {
  //   const filterState = Object.assign({}, newFilterState);
  //   filterTypeState === 'operator'
  //     ? (() => {
  //         filterState.props = [];
  //         changeFilterTypeState(false);
  //       })()
  //     : null;
  //   if (filterTypeState === 'values') {
  //     filterState.values.length
  //       ? filterState.values.pop()
  //       : (() => {
  //           filterState.operator = '';
  //           changeFilterTypeState(false);
  //         })();
  //   }
  //   setNewFilterState(filterState);
  // };

  // const changeFilterTypeState = (next = true) => {
  //   if (next) {
  //     filterTypeState === 'props'
  //       ? setFilterTypeState('operator')
  //       : filterTypeState === 'operator'
  //       ? setFilterTypeState('values')
  //       : (() => {})();
  //   } else {
  //     filterTypeState === 'values'
  //       ? setFilterTypeState('operator')
  //       : filterTypeState === 'operator'
  //       ? setFilterTypeState('props')
  //       : (() => {})();
  //   }
  // };

  // useEffect(() => {

  //   console.log("inside useEffect newFilterState-->>", newFilterState);
  //   console.log("inside useEffect selectedMainCategory-->>", selectedMainCategory);

  //   let filterData = {
  //     "category": selectedMainCategory?.category || "events",
  //     "object_type": newFilterState ? newFilterState.props[3] : "$session",
  //     "property_name": newFilterState ? newFilterState[1] : "$source",
  //     "entity": "event"
  // }

  // if(newFilterState.props.length>0){
  //   fetchKPIFilterValues(activeProject.id,filterData).then(res => {
  //     const ddValues = Object.assign({}, dropDownValues);
  //     ddValues[props[1]] = [...res.data, '$none'];
  //     setDropDownValues(ddValues);
  //   }).catch(err => {
  //     const ddValues = Object.assign({}, dropDownValues);
  //       ddValues[newFilterState.props[0]] = ['$none'];
  //       setDropDownValues(ddValues);
  //   });;
  // }

  //   if(newFilterState.props[1] === 'categorical') {
  //     if(newFilterState.props[2] === 'user') {
  //       if(!dropDownValues[newFilterState.props[0]]) {
  //         fetchUserPropertyValues(activeProject.id, newFilterState.props[0]).then(res => {
  //           const ddValues = Object.assign({}, dropDownValues);
  //           ddValues[newFilterState.props[0]] = [...res.data, '$none'];
  //           setDropDownValues(ddValues);
  //         }).catch(() => {
  //           console.log(err)
  //           const ddValues = Object.assign({}, dropDownValues);
  //           ddValues[newFilterState.props[0]] = ['$none'];
  //           setDropDownValues(ddValues);
  //         });
  //       }
  //   } else if(newFilterState.props[2] === 'event') {
  //     if(!dropDownValues[newFilterState.props[0]]) {
  //       fetchEventPropertyValues(activeProject.id, event.label, newFilterState.props[0]).then(res => {
  //         const ddValues = Object.assign({}, dropDownValues);
  //         ddValues[newFilterState.props[0]] = [...res.data, '$none'];
  //         setDropDownValues(ddValues);
  //       }).catch(() => {
  //         console.log(err)
  //         const ddValues = Object.assign({}, dropDownValues);
  //         ddValues[newFilterState.props[0]] = ['$none'];
  //         setDropDownValues(ddValues);
  //       });
  //     }
  //   } else {
  //     if(filterType === 'channel') {
  //       fetchChannelObjPropertyValues(activeProject.id, typeProps.channel,
  //         newFilterState.props[2].replace(" ", "_"), newFilterState.props[0]).then(res => {
  //           const ddValues = Object.assign({}, dropDownValues);
  //           ddValues[newFilterState.props[0]] = [...res?.data?.result?.filter_values, '$none'];
  //           setDropDownValues(ddValues);
  //       }).catch(() => {
  //         console.log(err)
  //         const ddValues = Object.assign({}, dropDownValues);
  //         ddValues[newFilterState.props[0]] = ['$none'];
  //         setDropDownValues(ddValues);
  //       });
  //     }
  //   }
  // }

  // }, [newFilterState])

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
          object_type: props[3] ? props[3] : event?.group,
          property_name: props[1],
          display_category: selectedMainCategory?.group,
          entity: 'event',
        };
      } else {
        filterData = {
          category: selectedMainCategory?.category,
          object_type: event?.group,
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
        refValue={refValue}
      />
    );
  };

  return (
    <div className={`flex items-center relative`}>
    {!showOr && (
    <Text level={8} type={'title'} extraClass={'m-0 mr-2'} weight={'thin'}>
      {index >= 1 ? 'and' : 'Filter by'}
    </Text>
    )}
    {showOr && (
    <Text level={8} type={'title'} extraClass={'m-0 mr-2 ml-2'} weight={'thin'}>
      or
    </Text>
    )}

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
