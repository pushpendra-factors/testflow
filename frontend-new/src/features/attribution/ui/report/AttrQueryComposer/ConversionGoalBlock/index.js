import React, { useState, useEffect } from 'react';
import styles from './index.module.scss';
import { connect } from 'react-redux';
import { bindActionCreators } from 'redux';
import FilterWrapper from 'Components/GlobalFilter/FilterWrapper';
import { Button, Tooltip } from 'antd';
import { SVG, Text } from 'factorsComponents';
import { isArray } from 'lodash';
import FaSelect from 'Components/FaSelect';
import ORButton from 'Components/ORButton';
import { getNormalizedKpi } from 'Utils/kpiQueryComposer.helpers';
import { compareFilters, groupFilters } from 'Utils/global';
import { fetchKPIConfigWithoutDerivedKPI } from 'Reducers/kpi';
import { TOOLTIP_CONSTANTS } from 'Constants/tooltips.constans';

import EventFilterWrapper from 'Components/KPIComposer/EventFilterWrapper';
import GroupSelect from 'Components/GenericComponents/GroupSelect';
import getGroupIcon from 'Utils/getGroupIcon';
import { kpiItemsgroupedByCategoryProperty } from './utils';

function ConversionGoalBlock({
  eventGoal,
  eventGoalChange,
  delEvent,
  eventNameOptions,
  eventNames,
  activeProject,
  eventPropertiesV2,
  eventUserPropertiesV2,
  group_analysis = 'users',
  KPI_config,
  KPI_config_without_derived_kpi,
  showDerivedKPI = false,
  fetchKPIConfigWithoutDerivedKPI,
  currentProjectSettings
}) {
  const [selectVisible, setSelectVisible] = useState(false);
  const [filterBlockVisible, setFilterBlockVisible] = useState(false);

  const [moreOptions, setMoreOptions] = useState(false);
  const [orFilterIndex, setOrFilterIndex] = useState(-1);

  const [groupProps, setGroupProps] = useState();

  const [filterProps, setFilterProperties] = useState({
    event: [],
    user: []
  });

  const attrGroupNameMap = {
    hs_kpi: { label: 'Hubspot Deals', value: 'hubspot_deals' },
    sf_kpi: {
      label: 'Salesforce Opportunities',
      value: 'salesforce_opportunities'
    },
    user_kpi: { label: 'Users', value: 'user_kpi' }
  };

  useEffect(() => {
    if (eventGoal) {
      setFilterPropsforKpiGroups();
    }
  }, [eventUserPropertiesV2, eventPropertiesV2, group_analysis, eventGoal]);

  useEffect(() => {
    if (
      currentProjectSettings.attribution_config &&
      currentProjectSettings.attribution_config.kpis_to_attribute
    ) {
      const kpiList =
        currentProjectSettings.attribution_config.kpis_to_attribute;
      const groupedList = [];
      Object.keys(kpiList).forEach((grpName) => {
        const propertyObj = { label: '', iconName: '', value: '', values: [] };
        propertyObj.label = attrGroupNameMap[grpName]?.label;
        propertyObj.value = attrGroupNameMap[grpName]?.value;
        propertyObj.iconName = getGroupIcon(propertyObj.label);
        kpiList[grpName].forEach((item) => {
          propertyObj.values.push({
            label: item.label,
            value: item.value,
            extraProps: {
              valueCategory: item.category,
              kpiQueryType: item.kpi_query_type,
              valueGroup: item.group
            }
          });
        });
        groupedList.push(propertyObj);
      });
      setGroupProps(groupedList);
    }
  }, [activeProject, showDerivedKPI, currentProjectSettings]);

  const setFilterPropsforKpiGroups = () => {
    const assignFilterProps = Object.assign({}, filterProps);
    assignFilterProps.event = getKPIProps(group_analysis);

    setFilterProperties(assignFilterProps);
  };

  const getKPIProps = () => {
    let KPIlist = KPI_config || [];
    let selGroup = KPIlist.find((item) => {
      return item?.display_category == eventGoal.group;
    });
    return kpiItemsgroupedByCategoryProperty(selGroup) || {};
  };

  const toggleEventSelect = () => {
    setSelectVisible(!selectVisible);
  };

  const addFilter = (val) => {
    const updatedEvent = Object.assign({}, eventGoal);

    const filtersSorted = updatedEvent.filters;
    filtersSorted.sort(compareFilters);
    const filt = filtersSorted.filter(
      (fil) => JSON.stringify(fil) === JSON.stringify(val)
    );
    if (filt && filt.length) return;

    updatedEvent.filters.push(val);
    eventGoalChange(updatedEvent);
  };

  const editFiler = (index, val) => {
    let updatedEvent = Object.assign({}, eventGoal);
    const filt = Object.assign({}, val);
    filt.operator = isArray(val.operator) ? val.operator[0] : val.operator;
    const filtersSorted = updatedEvent.filters;
    filtersSorted.sort(compareFilters);
    filtersSorted[index] = filt;
    updatedEvent.filters = filtersSorted;
    eventGoalChange(updatedEvent);
  };

  const delFilter = (val) => {
    const updatedEvent = Object.assign({}, eventGoal);
    const filtersSorted = updatedEvent.filters;
    filtersSorted.sort(compareFilters);
    const filt = filtersSorted.filter((v, i) => i !== val);
    updatedEvent.filters = filt;
    eventGoalChange(updatedEvent);
  };

  const closeFilter = () => {
    setFilterBlockVisible(false);
    setOrFilterIndex(-1);
  };

  const deleteItem = () => {
    delEvent();
    closeFilter();
  };

  const addFilterBlock = () => {
    setFilterBlockVisible(true);
  };

  const selectEventFilter = (index) => {
   return (
      <EventFilterWrapper
        filterProps={filterProps}
        activeProject={activeProject}
        event={eventGoal}
        deleteFilter={closeFilter}
        insertFilter={addFilter}
        closeFilter={closeFilter}
        selectedMainCategory={eventGoal}
        refValue={index}
      />
    );
  };

  const renderFilterWrapper = (
    index,
    refValue,
    filter,
    showOr,
    inFilter,
    deleteFilter
  ) => (
    <EventFilterWrapper
      index={index}
      filter={filter}
      event={eventGoal}
      filterProps={filterProps}
      activeProject={activeProject}
      deleteFilter={deleteFilter}
      insertFilter={inFilter}
      closeFilter={closeFilter}
      selectedMainCategory={eventGoal}
      showOr={showOr}
      refValue={refValue}
    />
  );

  const eventFilters = () => {
    const filters = [];
    let index = 0;
    let lastRef = 0;
    if (eventGoal && eventGoal?.filters?.length) {
      const group = groupFilters(eventGoal.filters, 'ref');
      const filtersGroupedByRef = Object.values(group);
      const refValues = Object.keys(group);
      lastRef = parseInt(refValues[refValues.length - 1]);

      filtersGroupedByRef.forEach((filtersGr) => {
        const refValue = filtersGr[0].ref;
        if (filtersGr.length == 1) {
          const filter = filtersGr[0];
          let filterContent = filter;
          filterContent.values =
            filter.props[1] === 'datetime' && isArray(filter.values)
              ? filter.values[0]
              : filter.values;
          filters.push(
            <div className={'fa--query_block--filters flex flex-row'}>
              <div key={index}>
                {renderFilterWrapper(
                  index,
                  refValue,
                  filter,
                  false,
                  (val, index) => editFiler(index, val),
                  delFilter
                )}
              </div>
              {index !== orFilterIndex && (
                <ORButton index={index} setOrFilterIndex={setOrFilterIndex} />
              )}
              {index === orFilterIndex && (
                <div key={'init'}>
                  {renderFilterWrapper(
                    undefined,
                    refValue,
                    undefined,
                    true,
                    addFilter,
                    closeFilter
                  )}
                </div>
              )}
            </div>
          );
          index += 1;
        } else {
          filters.push(
            <div className={'fa--query_block--filters flex flex-row'}>
              <div key={index}>
                {renderFilterWrapper(
                  index,
                  refValue,
                  filtersGr[0],
                  false,
                  (val, index) => editFiler(index, val),
                  delFilter
                )}
              </div>
              <div key={index + 1}>
                {renderFilterWrapper(
                  index + 1,
                  refValue,
                  filtersGr[1],
                  true,
                  (val, index) => editFiler(index, val),
                  delFilter
                )}
              </div>
            </div>
          );
          index += 2;
        }
      });
    }

    if (filterBlockVisible) {
      filters.push(
        <div key={'init'} className={'fa--query_block--filters'}>
          {selectEventFilter(lastRef + 1)}
        </div>
      );
    }

    return filters;
  };

  const onEventSelect = (option, group) => {
    const currentEventGoal = Object.assign({}, eventGoal);
    currentEventGoal.label = option.value ? option.value : option.label;
    currentEventGoal.filters = [];
    if (group_analysis !== 'users') {
      currentEventGoal.label = option.label;
      currentEventGoal.metric = option.value ? option.value : option.label;
      currentEventGoal.group = group.value;
      const valueCategory = option.extraProps.valueCategory;
      const valueGroup = option.extraProps.valueGroup;
      const kpiQueryType = option.extraProps.kpiQueryType;
      if (valueCategory) {
        currentEventGoal.category = valueCategory;
      }
      if (valueGroup) {
        currentEventGoal.group = valueGroup;
      }
      if (kpiQueryType) {
        currentEventGoal.qt = kpiQueryType;
      }
    }
    eventGoalChange(currentEventGoal);
    setSelectVisible(false);
    closeFilter();
  };

  const additionalActions = () => {
    return (
      <div className={'fa--query_block--actions-cols flex relative ml-2'}>
        <div className={`relative flex`}>
          <Tooltip title='Filter this Attribute' color={TOOLTIP_CONSTANTS.DARK}>
            <Button
              type='text'
              onClick={() => setMoreOptions(true)}
              className={`fa-btn--custom mr-1 btn-total-round`}
            >
              <SVG name='more'></SVG>
            </Button>
          </Tooltip>

          {moreOptions ? (
            <FaSelect
              options={[[`Filter By`, 'filter']]}
              optionClick={(val) => {
                addFilterBlock();
                setMoreOptions(false);
              }}
              onClickOutside={() => setMoreOptions(false)}
              showIcon
            ></FaSelect>
          ) : (
            false
          )}
        </div>
        <Tooltip title='Delete this Attribute'>
          <Button
            type='text'
            onClick={deleteItem}
            className={`fa-btn--custom btn-total-round`}
          >
            <SVG name='trash'></SVG>
          </Button>
        </Tooltip>
      </div>
    );
  };

  const renderCountLabel = () => {
    return (
      <Text
        type={'title'}
        level={7}
        weight={'regular'}
        color={'grey'}
        extraClass={'m-0 ml-2'}
      >
        as count of unique users
      </Text>
    );
  };

  const selectEvents = () => {
    return (
      <div className={styles.block__event_selector}>
        {selectVisible ? (
          <GroupSelect
            options={groupProps}
            onClickOutside={() => setSelectVisible(false)}
            optionClickCallback={onEventSelect}
            placeholder='Select Event'
            allowSearch={true}
            extraClass={styles.block__event_selector__select}
            allowSearchTextSelection={false}
          />
        ) : null}
      </div>
    );
  };

  const renderGoalBlockContent = () => {
    return (
      <div
        className={`${styles.block__content} flex items-center relative mt-4`}
      >
        {
          <Tooltip
            title={
              eventNames[eventGoal?.label]
                ? eventNames[eventGoal?.label]
                : eventGoal?.label
            }
          >
            <Button
              type='link'
              onClick={toggleEventSelect}
              icon={<SVG name={getGroupIcon(eventGoal?.group)} />}
              className={`fa-button--truncate fa-button--truncate-lg btn-total-round`}
            >
              {eventNames[eventGoal?.label]
                ? eventNames[eventGoal?.label]
                : eventGoal?.label}
            </Button>
          </Tooltip>
        }

        {selectEvents()}

        {(!group_analysis || group_analysis === 'users') && renderCountLabel()}

        <div className={styles.block__additional_actions}>
          {additionalActions()}
        </div>
      </div>
    );
  };

  const renderGoalSelect = () => {
    return (
      <div className={'flex justify-start items-center mt-4'}>
        {
          <Button
            type='text'
            onClick={toggleEventSelect}
            icon={<SVG name={'plus'} color={'grey'} />}
          >
            Add KPI
          </Button>
        }
        {selectEvents()}
      </div>
    );
  };

  return (
    <div className={`${styles.block} fa--query_block_section--basic relative`}>
      {eventGoal?.label?.length ? renderGoalBlockContent() : renderGoalSelect()}
      {eventFilters()}
    </div>
  );
}

const mapStateToProps = (state) => ({
  activeProject: state.global.active_project,
  currentProjectSettings: state.global.currentProjectSettings,
  eventPropertiesV2: state.coreQuery.eventPropertiesV2,
  eventUserPropertiesV2: state.coreQuery.eventUserPropertiesV2,
  eventNameOptions: state.coreQuery.eventOptions,
  eventNames: state.coreQuery.eventNames,
  KPI_config: state.kpi?.config,
  KPI_config_without_derived_kpi: state.kpi?.config_without_derived_kpi
});

const mapDispatchToProps = (dispatch) =>
  bindActionCreators({ fetchKPIConfigWithoutDerivedKPI }, dispatch);

export default connect(
  mapStateToProps,
  mapDispatchToProps
)(ConversionGoalBlock);
