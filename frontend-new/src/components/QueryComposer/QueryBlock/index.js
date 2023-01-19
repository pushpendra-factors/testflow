import React, { useState, useEffect } from 'react';
import cx from 'classnames';
import { Button, Dropdown, Menu, Tooltip } from 'antd';
import { connect } from 'react-redux';
import { bindActionCreators } from 'redux';
import FaSelect from '../../FaSelect';
import { SVG, Text } from '../../factorsComponents';
import styles from './index.module.scss';
import {
  setGroupBy,
  delGroupBy,
  getGroupProperties
} from '../../../reducers/coreQuery/middleware';
import EventFilterWrapper from '../EventFilterWrapper';
import GroupSelect2 from '../GroupSelect2';
import EventGroupBlock from '../EventGroupBlock';
import {
  AvailableGroups,
  QUERY_TYPE_FUNNEL,
  RevAvailableGroups
} from '../../../utils/constants';
import AliasModal from '../AliasModal';
import ORButton from '../../ORButton';
import { compareFilters, groupFilters } from '../../../utils/global';
import { TOOLTIP_CONSTANTS } from '../../../constants/tooltips.constans';

function QueryBlock({
  index,
  event,
  eventChange,
  queries,
  queryType,
  eventOptions,
  eventNames,
  activeProject,
  groupBy,
  setGroupBy,
  delGroupBy,
  userProperties,
  eventProperties,
  groupProperties,
  getGroupProperties,
  groupAnalysis
}) {
  const [isDDVisible, setDDVisible] = useState(false);
  const [isFilterDDVisible, setFilterDDVisible] = useState(false);
  const [isGroupByDDVisible, setGroupByDDVisible] = useState(false);
  const [moreOptions, setMoreOptions] = useState(false);
  const [filterProps, setFilterProperties] = useState({
    event: [],
    user: [],
    group: []
  });
  const [showGroups, setShowGroups] = useState([]);

  useEffect(() => {
    const groupsArray = Object.values(RevAvailableGroups);
    const options = [...eventOptions];
    let showOpts = [];
    if (groupAnalysis === 'users') {
      // showOpts = options.filter((item) => !groupsArray.includes(item?.label));
      showOpts = [...options];
    } else {
      const groupOpts = options.filter(
        (item) => item.label === RevAvailableGroups[groupAnalysis]
      );
      const userOpts = options.filter(
        (item) => !groupsArray.includes(item?.label)
      );
      showOpts = groupOpts.concat(userOpts);
    }
    setShowGroups(showOpts);
  }, [eventOptions, groupAnalysis]);

  const [orFilterIndex, setOrFilterIndex] = useState(-1);

  const [isModalVisible, setIsModalVisible] = useState(false);

  const showModal = () => {
    setIsModalVisible(true);
  };

  const handleCancel = () => {
    setIsModalVisible(false);
  };
  const handleOk = (alias) => {
    const newEvent = { ...event };
    newEvent.alias = alias;
    setIsModalVisible(false);
    eventChange(newEvent, index - 1, 'filters_updated');
  };

  const alphabetIndex = 'ABCDEF';

  const onChange = (group, value) => {
    const newEvent = { alias: '', label: '', filters: [], group: '' };
    newEvent.label = value;
    newEvent.group = group;
    setDDVisible(false);
    eventChange(newEvent, index - 1);
  };

  useEffect(() => {
    if (!event || event === undefined) {
      return;
    }
    if (AvailableGroups[event.group]) {
      getGroupProperties(activeProject.id, AvailableGroups[event.group]);
    }
  }, [event]);

  useEffect(() => {
    if (!event || event === undefined) {
      return;
    }
    const assignFilterProps = { ...filterProps };
    if (AvailableGroups[event.group]) {
      assignFilterProps.group = groupProperties[AvailableGroups[event.group]];
      assignFilterProps.user = [];
    } else {
      assignFilterProps.user = userProperties;
      assignFilterProps.group = [];
    }
    assignFilterProps.event = eventProperties[event.label] || [];
    setFilterProperties(assignFilterProps);
  }, [eventProperties, groupProperties, userProperties]);

  const triggerDropDown = () => {
    setDDVisible(true);
  };

  const deleteItem = () => {
    eventChange(event, index - 1, 'delete');
  };

  const selectEvents = () =>
    isDDVisible ? (
      <div className={styles.query_block__event_selector}>
        <GroupSelect2
          groupedProperties={showGroups}
          placeholder='Select Event'
          optionClick={(group, val) =>
            onChange(group, val[1] ? val[1] : val[0])
          }
          onClickOutside={() => setDDVisible(false)}
          allowEmpty
        />
      </div>
    ) : null;

  const addGroupBy = () => {
    setGroupByDDVisible(true);
  };

  const addFilter = () => {
    setFilterDDVisible(true);
  };

  const insertFilters = (filter, filterIndex) => {
    const newEvent = { ...event };
    const filtersSorted = newEvent.filters;
    filtersSorted.sort(compareFilters);
    if (filterIndex >= 0) {
      newEvent.filters = filtersSorted.map((filt, i) => {
        if (i === filterIndex) {
          return filter;
        }
        return filt;
      });
    } else {
      newEvent.filters.push(filter);
    }

    eventChange(newEvent, index - 1, 'filters_updated');
  };

  const removeFilters = (i) => {
    const newEvent = { ...event };
    const filtersSorted = newEvent.filters;
    filtersSorted.sort(compareFilters);
    if (filtersSorted[i]) {
      filtersSorted.splice(i, 1);
      newEvent.filters = filtersSorted;
    }
    eventChange(newEvent, index - 1, 'filters_updated');
  };
  const closeFilter = () => {
    setFilterDDVisible(false);
    setOrFilterIndex(-1);
  };
  const selectEventFilter = (ind) => (
    <EventFilterWrapper
      filterProps={filterProps}
      activeProject={activeProject}
      event={event}
      deleteFilter={closeFilter}
      insertFilter={insertFilters}
      closeFilter={closeFilter}
      refValue={ind}
    />
  );

  const deleteGroupBy = (groupState, id, type = 'event') => {
    delGroupBy(type, groupState, id);
  };

  const pushGroupBy = (groupState, ind) => {
    const i = ind >= 0 ? ind : groupBy.length;
    setGroupBy('event', groupState, i);
  };

  const selectGroupByEvent = () =>
    isGroupByDDVisible ? (
      <EventGroupBlock
        eventIndex={index}
        event={event}
        setGroupState={pushGroupBy}
        closeDropDown={() => setGroupByDDVisible(false)}
      />
    ) : null;

  const setAdditionalactions = (opt) => {
    if (opt[1] === 'filter') {
      addFilter();
    } else if (opt[1] === 'groupby') {
      addGroupBy();
    } else {
      showModal();
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
      <div className='fa--query_block--actions-cols flex'>
        <div className='relative'>
          <Tooltip
            title={`Filter this ${queryType === 'funnel' ? 'funnel' : 'event'}`}
            color={TOOLTIP_CONSTANTS.DARK}
          >
            <Button
              type='text'
              onClick={addFilter}
              className='fa-btn--custom mr-1 btn-total-round'
            >
              <SVG name='filter' />
            </Button>
          </Tooltip>

          {moreOptions ? (
            <FaSelect
              options={[
                ['Filter By', 'filter'],
                ['Breakdown', 'groupby'],
                [!event?.alias?.length ? 'Create Alias' : 'Edit Alias', 'edit']
              ]}
              optionClick={(val) => setAdditionalactions(val)}
              onClickOutside={() => setMoreOptions(false)}
            />
          ) : (
            false
          )}

          <AliasModal
            visible={isModalVisible}
            event={
              eventNames[event.label] ? eventNames[event.label] : event.label
            }
            onOk={handleOk}
            onCancel={handleCancel}
            alias={event.alias}
          />
        </div>
        <Tooltip
          title={`Delete this ${queryType === 'funnel' ? 'funnel' : 'event'}`}
          color={TOOLTIP_CONSTANTS.DARK}
        >
          <Button
            type='text'
            onClick={deleteItem}
            className='fa-btn--custom btn-total-round'
          >
            <SVG name='trash' />
          </Button>
        </Tooltip>
      </div>
    );
  };

  const eventFilters = () => {
    const filters = [];
    let ind = 0;
    let lastRef = 0;
    if (event && event?.filters?.length) {
      const group = groupFilters(event.filters, 'ref');
      const filtersGroupedByRef = Object.values(group);
      const refValues = Object.keys(group);
      lastRef = parseInt(refValues[refValues.length - 1]);

      filtersGroupedByRef.forEach((filtersGr) => {
        const refValue = filtersGr[0].ref;
        if (filtersGr.length === 1) {
          const filter = filtersGr[0];
          filters.push(
            <div className='fa--query_block--filters flex flex-row'>
              <div key={ind}>
                <EventFilterWrapper
                  index={ind}
                  filter={filter}
                  event={event}
                  filterProps={filterProps}
                  activeProject={activeProject}
                  deleteFilter={removeFilters}
                  insertFilter={insertFilters}
                  closeFilter={closeFilter}
                  refValue={refValue}
                />
              </div>
              {ind !== orFilterIndex && (
                <ORButton index={ind} setOrFilterIndex={setOrFilterIndex} />
              )}
              {ind === orFilterIndex && (
                <div key='init'>
                  <EventFilterWrapper
                    filterProps={filterProps}
                    activeProject={activeProject}
                    event={event}
                    deleteFilter={closeFilter}
                    insertFilter={insertFilters}
                    closeFilter={closeFilter}
                    refValue={refValue}
                    showOr
                  />
                </div>
              )}
            </div>
          );
          ind += 1;
        } else {
          filters.push(
            <div className='fa--query_block--filters flex flex-row'>
              <div key={ind}>
                <EventFilterWrapper
                  index={ind}
                  filter={filtersGr[0]}
                  event={event}
                  filterProps={filterProps}
                  activeProject={activeProject}
                  deleteFilter={removeFilters}
                  insertFilter={insertFilters}
                  closeFilter={closeFilter}
                  refValue={refValue}
                />
              </div>
              <div key={ind + 1}>
                <EventFilterWrapper
                  index={ind + 1}
                  filter={filtersGr[1]}
                  event={event}
                  filterProps={filterProps}
                  activeProject={activeProject}
                  deleteFilter={removeFilters}
                  insertFilter={insertFilters}
                  closeFilter={closeFilter}
                  refValue={refValue}
                  showOr
                />
              </div>
            </div>
          );
          ind += 2;
        }
      });
    }

    if (isFilterDDVisible) {
      filters.push(
        <div key='init' className='fa--query_block--filters'>
          {selectEventFilter(lastRef + 1)}
        </div>
      );
    }

    return filters;
  };

  const groupByItems = () => {
    const groupByEvents = [];
    if (groupBy && groupBy.length && groupBy[0] && groupBy[0].property) {
      groupBy
        .map((gbp, ind) => ({ ...gbp, groupByIndex: ind }))
        .filter(
          (gbp) => gbp.eventName === event.label && gbp.eventIndex === index
        )
        .forEach((gbp, gbpIndex) => {
          const { groupByIndex, ...orgGbp } = gbp;
          groupByEvents.push(
            <div key={gbpIndex} className='fa--query_block--filters'>
              <EventGroupBlock
                index={gbp.groupByIndex}
                grpIndex={gbpIndex}
                eventIndex={index}
                groupByEvent={orgGbp}
                event={event}
                delGroupState={(ev) => deleteGroupBy(ev, gbpIndex)}
                setGroupState={pushGroupBy}
                closeDropDown={() => setGroupByDDVisible(false)}
              />
            </div>
          );
        });
    }

    if (isGroupByDDVisible) {
      groupByEvents.push(
        <div key='init' className='fa--query_block--filters'>
          {selectGroupByEvent()}
        </div>
      );
    }

    return groupByEvents;
  };

  const ifQueries = queries.length > 0;
  let filterOptions = [
    // ['Filter By', 'filter'],
    ['Breakdown', 'groupby'],
    [!event?.alias?.length ? 'Create Alias' : 'Edit Alias', 'edit']
  ];
  if (!event) {
    return (
      <div
        className={`${styles.query_block} fa--query_block my-2 ${
          ifQueries ? 'borderless no-padding' : 'borderless no-padding'
        }`}
      >
        <div
          className={`${styles.query_block__event} flex justify-start items-center`}
        >
          <Button
            type='text'
            onClick={triggerDropDown}
            icon={<SVG name='plus' color='grey' />}
          >
            {ifQueries ? 'Add another event' : 'Add First Event'}
          </Button>
          {selectEvents()}
        </div>
      </div>
    );
  }

  return (
    <div
      className={`${styles.query_block} fa--query_block_section borderless no-padding mt-2`}
    >
      <div
        className={`${!event?.alias?.length ? 'flex justify-start' : ''} ${
          styles.query_block__event
        } block_section items-center`}
      >
        <div className='flex items-center'>
          <div
            className={cx(
              styles.query_block__additional_actions,
              'mr-2',
              styles['drag-icon']
            )}
          >
            <SVG name='drag' />
          </div>
          <div className='fa--query_block--add-event active flex justify-center items-center mr-2'>
            <Text
              type='title'
              level={7}
              weight='bold'
              color='white'
              extraClass='m-0'
            >
              {queryType === QUERY_TYPE_FUNNEL
                ? index
                : alphabetIndex[index - 1]}
            </Text>
          </div>
          {event?.alias?.length ? (
            <Text type='title' level={7} weight='bold' extraClass='m-0'>
              {event?.alias}
              <Tooltip title='Edit Alias'>
                <Button
                  className={`${styles.custombtn} mx-1`}
                  type='text'
                  onClick={showModal}
                >
                  <SVG size={20} name='edit' color='grey' />
                </Button>
              </Tooltip>
            </Text>
          ) : null}
        </div>
        <div className={`flex ${!event?.alias?.length ? '' : 'ml-8 mt-2'}`}>
          <div className='relative ml-2'>
            <Tooltip
              title={
                eventNames[event.label] ? eventNames[event.label] : event.label
              }
            >
              <Button
                icon={<SVG name='mouseevent' size={16} color='purple' />}
                className='fa-button--truncate fa-button--truncate-lg btn-total-round'
                type='link'
                onClick={triggerDropDown}
              >
                {eventNames[event.label]
                  ? eventNames[event.label]
                  : event.label}
              </Button>
              {selectEvents()}
            </Tooltip>
          </div>
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
          <div className={styles.query_block__additional_actions}>
            {additionalActions()}
          </div>
        </div>
      </div>
      {eventFilters()}
      {groupByItems()}
    </div>
  );
}

const mapStateToProps = (state) => ({
  eventOptions: state.coreQuery.eventOptions,
  activeProject: state.global.active_project,
  userProperties: state.coreQuery.userProperties,
  groupProperties: state.coreQuery.groupProperties,
  eventProperties: state.coreQuery.eventProperties,
  groupBy: state.coreQuery.groupBy.event,
  groupByMagic: state.coreQuery.groupBy,
  eventNames: state.coreQuery.eventNames
});

const mapDispatchToProps = (dispatch) =>
  bindActionCreators(
    {
      setGroupBy,
      delGroupBy,
      getGroupProperties
    },
    dispatch
  );

export default connect(mapStateToProps, mapDispatchToProps)(QueryBlock);
