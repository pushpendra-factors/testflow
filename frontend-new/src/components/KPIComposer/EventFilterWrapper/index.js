/* eslint-disable */
import React, { useState, useEffect } from 'react';
import { useSelector, connect } from 'react-redux';
import styles from './index.module.scss';
import { DateRangePicker } from 'react-date-range';
import { Input, Button, Result } from 'antd';
import MomentTz from 'Components/MomentTz';
import { SVG, Text } from 'factorsComponents';
import { DEFAULT_DATE_RANGE } from '../../QueryComposer/DateRangeSelector/utils';
import { DEFAULT_OPERATOR_PROPS } from 'Components/FaFilterSelect/utils';
import _ from 'lodash';
import FAFilterSelect from 'Components/KPIComposer/FaFilterSelectKPI';
import { getKPIPropertyValues } from 'Reducers/coreQuery/middleware';

const defaultOpProps = DEFAULT_OPERATOR_PROPS;

function EventFilterWrapper({
  index,
  refValue,
  filterProps,
  activeProject,
  operatorProps = defaultOpProps,
  event,
  filter,
  delIcon = 'remove',
  deleteFilter,
  insertFilter,
  closeFilter,
  KPI_config,
  selectedMainCategory,
  showOr,
  getKPIPropertyValues,
  propertyValuesMap
}) {
  const [filterTypeState, setFilterTypeState] = useState('props');
  const [groupCollapseState, setGroupCollapse] = useState({});
  const [searchTerm, setSearchTerm] = useState('');
  const [newFilterState, setNewFilterState] = useState({
    props: [],
    operator: '',
    values: []
  });

  const [dropDownValues, setDropDownValues] = useState({});
  const [valueOptsLoading, setvalueOptsLoading] = useState(false);
  const [selectedRngState, setSelectedRngState] = useState([
    { ...DEFAULT_DATE_RANGE }
  ]);

  const [filterDropDownOptions, setFiltDD] = useState({
    props: [
      {
        label: ' ',
        icon: 'mouseclick'
      }
    ],
    operator: operatorProps
  });

  const { userPropNames } = useSelector((state) => state.coreQuery);

  useEffect(() => {
    if (filter) {
      setValuesByProps(filter.props);
      setNewFilterState(filter);

      if (filter && filter?.extra) {
        let filterData = {};
        if (
          event?.category == 'channels' ||
          event?.category == 'custom_channels'
        ) {
          filterData = {
            category: event?.category, //use event instead of selectedMainCategory since it is in induvidual level
            object_type: filter?.extra[3],
            property_name: filter?.extra[1],
            display_category: selectedMainCategory?.group,
            entity: 'event'
          };
        } else {
          filterData = {
            category: event?.category, //use event instead of selectedMainCategory since it is in induvidual level
            object_type: event?.pageViewVal ? event?.pageViewVal : event?.group, // depreciated! object_type to display_category key change
            display_category: event?.pageViewVal
              ? event?.pageViewVal
              : event?.group, // object_type to display_category key change
            property_name: filter?.extra[1],
            entity: filter?.extra[3] ? filter?.extra[3] : filter?.extra[2]
          };
        }
        setvalueOptsLoading(true);
        if (propertyValuesMap[filterData?.property_name]) {
          getKPIPropertyValues(activeProject.id, filterData)
            .then((res) => {
              setvalueOptsLoading(false);
            })
            .catch((err) => {
              setvalueOptsLoading(false);
            });
        }
      } else if (!filter?.extra) {
        // filter.extra getiing set null after running query once and after 2nd time it showing loading
        // added here temporary fix for the above
        const ddValues = Object.assign({}, dropDownValues);
        ddValues[filter?.props[0]] = ['$none'];
        setDropDownValues(ddValues);
      }
    }
  }, [filter, event]);

  useEffect(() => {
    const filterDD = Object.assign({}, filterDropDownOptions);
    const propState = [];
    //Needs to Update But not being Used.
    Object.keys(filterProps).forEach((k, i) => {
      propState.push({
        label: k,
        icon: k === 'event' ? 'mouseclick' : k,
        values: filterProps[k]
      });
    });
    //
    let KPIlist = KPI_config || [];
    let selGroup = KPIlist.find((item) => {
      return item.display_category == event?.group;
    });
    let DDvalues = selGroup?.properties?.map((item) => {
      if (item == null) return;
      let ddName = item.display_name ? item.display_name : item.name;
      let ddtype =
        selGroup?.category == 'channels' ||
        selGroup?.category == 'custom_channels'
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
        values: DDvalues
      }
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
      if (
        event?.category == 'channels' ||
        event?.category == 'custom_channels'
      ) {
        //use event instead of selectedMainCategory since it is in induvidual level
        filterData = {
          category: event?.category, //use event instead of selectedMainCategory since it is in induvidual level
          object_type: props[3] ? props[3] : event?.group,
          property_name: props[1],
          display_category: event?.group,
          entity: 'event'
        };
      } else {
        filterData = {
          category: event?.category, //use event instead of selectedMainCategory since it is in induvidual level
          object_type: event?.pageViewVal ? event?.pageViewVal : event?.group, // depreciated! object_type to display_category key change
          display_category: event?.pageViewVal
            ? event?.pageViewVal
            : event?.group, // object_type to display_category key change
          property_name: props[1],
          entity: props[3] ? props[3] : props[2]
        };
      }

      setvalueOptsLoading(true);
      getKPIPropertyValues(activeProject.id, filterData)
        .then((res) => {
          setvalueOptsLoading(false);
        })
        .catch((err) => {
          setvalueOptsLoading(false);
        });
    }
  };

  const renderFilterContent = () => {
    return (
      <FAFilterSelect
        propOpts={filterDropDownOptions.props}
        operatorOpts={filterDropDownOptions.operator}
        valueOpts={propertyValuesMap.data}
        valueOptsLoading={propertyValuesMap.loading}
        applyFilter={applyFilter}
        setValuesByProps={setValuesByProps}
        filter={filter}
        refValue={refValue}
      />
    );
  };

  return (
    <div className={`flex items-center relative ${!showOr ? 'ml-10' : ''}`}>
      {!showOr &&
        (index >= 1 ? (
          <Text
            level={8}
            type={'title'}
            extraClass={'m-0 mr-16'}
            weight={'thin'}
          >
            and
          </Text>
        ) : (
          <Text
            level={8}
            type={'title'}
            extraClass={'m-0 mr-10'}
            weight={'thin'}
          >
            Filter by
          </Text>
        ))}
      {showOr && (
        <Text level={8} type={'title'} extraClass={'m-0 mx-4'} weight={'thin'}>
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
  propertyValuesMap: state.coreQuery.propertyValuesMap
});

export default connect(mapStateToProps, { getKPIPropertyValues })(
  EventFilterWrapper
);
