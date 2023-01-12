import React, { useState, useEffect } from 'react';
import styles from './index.module.scss';
import { connect } from 'react-redux';
import { bindActionCreators } from 'redux';

import GroupSelect2 from '../../QueryComposer/GroupSelect2';
import EventFilterWrapper from '../EventFilterWrapper';
import { default as EventFilter } from 'Components/QueryComposer/EventFilterWrapper';

import { Button, Dropdown, Menu, Tooltip } from 'antd';
import { SVG, Text } from 'factorsComponents';
import { before, isArray } from 'lodash';
import FaSelect from 'Components/FaSelect';
import ORButton from '../../ORButton';
import { getNormalizedKpi } from '../../../utils/kpiQueryComposer.helpers';
import { compareFilters, groupFilters } from '../../../utils/global';
import { fetchKPIConfigWithoutDerivedKPI } from 'Reducers/kpi';

import { TOOLTIP_CONSTANTS } from '../../../constants/tooltips.constans';

const ConversionGoalBlock = ({
  eventGoal,
  eventGoalChange,
  delEvent,
  eventNameOptions,
  eventNames,
  activeProject,
  eventProperties,
  userProperties,
  group_analysis = 'users',
  KPI_config,
  KPI_config_without_derived_kpi,
  showDerivedKPI = false,
  fetchKPIConfigWithoutDerivedKPI
}) => {
  const [selectVisible, setSelectVisible] = useState(false);
  const [filterBlockVisible, setFilterBlockVisible] = useState(false);

  const [moreOptions, setMoreOptions] = useState(false);
  const [orFilterIndex, setOrFilterIndex] = useState(-1);

  const [filterProps, setFilterProperties] = useState({
    event: [],
    user: []
  });

  useEffect(() => {
    if (!group_analysis || group_analysis === 'users') {
      setEventPropsForUserGroup();
    } else {
      setFilterPropsforKpiGroups();
    }
  }, [userProperties, eventProperties, group_analysis]);

  useEffect(() => {
    if (!showDerivedKPI && !KPI_config_without_derived_kpi)
      fetchKPIConfigWithoutDerivedKPI(activeProject.id);
  }, [activeProject, showDerivedKPI, KPI_config_without_derived_kpi]);

  const setEventPropsForUserGroup = () => {
    if (!eventGoal || !eventGoal?.label?.length) {
      return;
    }
    const assignFilterProps = Object.assign({}, filterProps);

    if (eventProperties[eventGoal.label]) {
      assignFilterProps.event = eventProperties[eventGoal.label];
    }
    assignFilterProps.user = userProperties;
    setFilterProperties(assignFilterProps);
  };

  const setFilterPropsforKpiGroups = () => {
    const assignFilterProps = Object.assign({}, filterProps);
    assignFilterProps.event = getKPIProps(group_analysis);
    setFilterProperties(assignFilterProps);
  };

  const getKPIProps = (groupName) => {
    let KPIlist = KPI_config || [];
    let selGroup = KPIlist.find((item) => {
      return item?.display_category == groupName;
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
    return DDvalues;
  };

  const getKpiGroupList = (groupName) => {
    let KPIlist = showDerivedKPI
      ? KPI_config
      : KPI_config_without_derived_kpi || [];
    let selGroup = KPIlist.find((item) => {
      return item?.display_category == groupName;
    });

    const group = ((selGroup) => {
      return getNormalizedKpi({ kpi: selGroup });
    })(selGroup);
    return [group];
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
    if (group_analysis !== 'users') {
      return (
        <EventFilterWrapper
          filterProps={filterProps}
          activeProject={activeProject}
          event={eventGoal}
          deleteFilter={() => closeFilter()}
          insertFilter={addFilter}
          closeFilter={closeFilter}
          refValue={index}
        ></EventFilterWrapper>
      );
    } else {
      return (
        <EventFilter
          filterProps={filterProps}
          activeProject={activeProject}
          event={eventGoal}
          deleteFilter={() => closeFilter()}
          insertFilter={addFilter}
          closeFilter={closeFilter}
          refValue={index}
        ></EventFilter>
      );
    }
  };

  const renderFilterWrapper = (index, refValue, filter, showOr = false) =>
    group_analysis !== 'users' ? (
      <EventFilterWrapper
        index={index}
        filter={filter}
        event={eventGoal}
        filterProps={filterProps}
        activeProject={activeProject}
        deleteFilter={delFilter}
        insertFilter={addFilter}
        closeFilter={closeFilter}
        selectedMainCategory={eventGoal}
        showOr={showOr}
        refValue={refValue}
      ></EventFilterWrapper>
    ) : (
      <EventFilter
        index={index}
        filter={filter}
        event={eventGoal}
        filterProps={filterProps}
        activeProject={activeProject}
        deleteFilter={delFilter}
        insertFilter={addFilter}
        closeFilter={closeFilter}
        selectedMainCategory={eventGoal}
        showOr={showOr}
        refValue={refValue}
      ></EventFilter>
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
          filters.push(
            <div className={'fa--query_block--filters flex flex-row'}>
              <div key={index}>
                {renderFilterWrapper(index, refValue, filter)}
              </div>
              {index !== orFilterIndex && (
                <ORButton index={index} setOrFilterIndex={setOrFilterIndex} />
              )}
              {index === orFilterIndex && (
                <div key={'init'}>
                  {renderFilterWrapper(index, refValue, filter, true)}
                </div>
              )}
            </div>
          );
          index += 1;
        } else {
          filters.push(
            <div className={'fa--query_block--filters flex flex-row'}>
              <div key={index}>
                {renderFilterWrapper(index, refValue, filtersGr[0])}
              </div>
              <div key={index + 1}>
                {renderFilterWrapper(index + 1, refValue, filtersGr[1], true)}
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

  const onEventSelect = (val, group, category) => {
    const currentEventGoal = Object.assign({}, eventGoal);
    currentEventGoal.label = val[1] ? val[1] : val[0];
    currentEventGoal.group = group;
    currentEventGoal.filters = [];
    if (group_analysis !== 'users') {
      currentEventGoal.label = val[0];
      currentEventGoal.metric = val[1] ? val[1] : val[0];
      currentEventGoal.group = group;
      if (category) {
        currentEventGoal.category = category;
      }
    }
    eventGoalChange(currentEventGoal);
    setSelectVisible(false);
    closeFilter();
  };

  const setAdditionalactions = (opt) => {
    if (opt[1] === 'filter') {
      addFilterBlock();
    }
    setMoreOptions(false);
  };

  const getMenu = (filterOptions) => (
    <Menu style={{ minWidth: '200px', padding: '10px' }}>
      {filterOptions.map((eachFilter, eachIndex) => {
        return (
          <Menu.Item
            icon={
              <SVG
                name={eachFilter[1]}
                extraClass={'self-center'}
                style={{ marginRight: '10px' }}
              ></SVG>
            }
            style={{ display: 'flex', padding: '10px', margin: '5px' }}
            key={eachIndex}
            onClick={() => setAdditionalactions(eachFilter)}
          >
            <span style={{ paddingLeft: '5px' }}>{eachFilter[0]}</span>
          </Menu.Item>
        );
      })}
    </Menu>
  );
  const additionalActions = () => {
    return (
      <div className={'fa--query_block--actions-cols flex relative ml-2'}>
        <div className={`relative flex`}>
          <Tooltip title='Filter this Attribute' color={TOOLTIP_CONSTANTS.DARK}>
            <Button
              type='text'
              onClick={addFilterBlock}
              className={`fa-btn--custom btn-total-round`}
            >
              <SVG name='filter'></SVG>
            </Button>
          </Tooltip>
          {/* {moreOptions ? (
            <FaSelect
              options={[[`Filter By`, 'filter']]}
              optionClick={(val) => {
                addFilterBlock();
                setMoreOptions(false);
              }}
              onClickOutside={() => setMoreOptions(false)}
            ></FaSelect>
          ) : (
            false
          )} */}
        </div>
        <Tooltip title='Delete this Attribute' color={TOOLTIP_CONSTANTS.DARK}>
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
    const groupedProps =
      !group_analysis || group_analysis === 'users'
        ? eventNameOptions
        : getKpiGroupList(group_analysis);
    return (
      <div className={styles.block__event_selector}>
        {selectVisible ? (
          <GroupSelect2
            groupedProperties={groupedProps}
            placeholder='Select Event'
            optionClick={(group, val, category) =>
              onEventSelect(val, group, category)
            }
            onClickOutside={() => setSelectVisible(false)}
          ></GroupSelect2>
        ) : null}
      </div>
    );
  };

  const renderGoalBlockContent = () => {
    let filterOptions = [
      // ['Filter By', 'filter']
    ];
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
              icon={<SVG name='mouseevent' />}
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
        {filterOptions.length != 0 ? (
          <Dropdown
            placement='bottomLeft'
            overlay={getMenu(filterOptions)}
            trigger={['hover']}
          >
            <Button
              type='text'
              size={'large'}
              className={`fa-btn--custom mr-1 btn-total-round`}
            >
              <SVG name='more' />
            </Button>
          </Dropdown>
        ) : (
          ''
        )}
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
            Add a goal event
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
};

const mapStateToProps = (state) => ({
  activeProject: state.global.active_project,
  eventProperties: state.coreQuery.eventProperties,
  userProperties: state.coreQuery.userProperties,
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
