import React, { useState, useEffect } from 'react';
import { Button, Tooltip } from 'antd';
import { connect } from 'react-redux';
import { bindActionCreators } from 'redux';
import { SVG } from 'Components/factorsComponents';
import styles from './index.module.scss';
import { getGroupProperties } from 'Reducers/coreQuery/middleware';
import EventFilterWrapper from 'Components/QueryComposer/EventFilterWrapper';
import GroupSelect2 from 'Components/QueryComposer/GroupSelect2';
import { AvailableGroups, RevAvailableGroups } from 'Utils/constants';
import ORButton from 'Components/ORButton';
import { compareFilters, groupFilters } from 'Utils/global';

function EventsBlock({
  index,
  event,
  closeEvent,
  eventChange,
  queries,
  eventOptions,
  eventNames,
  activeProject,
  userProperties,
  eventProperties,
  groupProperties,
  getGroupProperties,
  groupAnalysis,
  displayMode
}) {
  const [isDDVisible, setDDVisible] = useState(true);
  useEffect(() => {
    if (displayMode) {
      setDDVisible(false);
    }
  }, [displayMode]);
  const [isFilterDDVisible, setFilterDDVisible] = useState(false);
  const [filterProps, setFilterProperties] = useState({
    event: [],
    user: [],
    group: []
  });
  const [showGroups, setShowGroups] = useState([]);
  const [orFilterIndex, setOrFilterIndex] = useState(-1);

  useEffect(() => {
    const groupsArray = Object.values(RevAvailableGroups);
    const options = [...eventOptions];
    let showOpts = [];
    if (groupAnalysis === 'users') {
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

  const onChange = (group, value) => {
    const newEvent = { alias: '', label: '', filters: [], group: '' };
    newEvent.label = value;
    newEvent.group = group;
    setDDVisible(false);
    eventChange(newEvent, index - 1);
    closeEvent();
  };

  useEffect(() => {
    if (!event || event === undefined) {
      return;
    }
    if (AvailableGroups[event.group] && !displayMode) {
      getGroupProperties(activeProject.id, AvailableGroups[event.group]);
    }
  }, [event, displayMode]);

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
  }, [eventProperties, groupProperties, userProperties, event]);

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
          onClickOutside={() => {
            setDDVisible(false);
            closeEvent();
          }}
          allowEmpty
        />
      </div>
    ) : null;

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
      displayMode={displayMode}
      filterProps={filterProps}
      activeProject={activeProject}
      event={event}
      deleteFilter={closeFilter}
      insertFilter={insertFilters}
      closeFilter={closeFilter}
      refValue={ind}
      caller={'profiles'}
    />
  );

  const additionalActions = () => {
    return (
      <div className='fa--query_block--actions-cols flex'>
        <Tooltip title={`Filter this event`} color='#0B1E39'>
          <Button
            type='text'
            onClick={addFilter}
            className='fa-btn--custom mr-1 btn-total-round'
          >
            <SVG name='filter' />
          </Button>
        </Tooltip>
        <Tooltip title={`Delete this event`} color='#0B1E39'>
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
            <div className='fa--query_block--filters flex flex-col'>
              <div className='flex flex-row'>
                <div key={ind}>
                  <EventFilterWrapper
                    displayMode={displayMode}
                    index={ind}
                    filter={filter}
                    event={event}
                    filterProps={filterProps}
                    activeProject={activeProject}
                    deleteFilter={removeFilters}
                    insertFilter={insertFilters}
                    closeFilter={closeFilter}
                    refValue={refValue}
                    caller={'profiles'}
                  />
                </div>
                {ind !== orFilterIndex && !displayMode && (
                  <ORButton index={ind} setOrFilterIndex={setOrFilterIndex} />
                )}
              </div>
              {ind === orFilterIndex && (
                <div key='init'>
                  <EventFilterWrapper
                    displayMode={displayMode}
                    filterProps={filterProps}
                    activeProject={activeProject}
                    event={event}
                    deleteFilter={closeFilter}
                    insertFilter={insertFilters}
                    closeFilter={closeFilter}
                    refValue={refValue}
                    showOr
                    caller={'profiles'}
                  />
                </div>
              )}
            </div>
          );
          ind += 1;
        } else {
          filters.push(
            <div className='fa--query_block--filters flex flex-col'>
              <div key={ind}>
                <EventFilterWrapper
                  displayMode={displayMode}
                  index={ind}
                  filter={filtersGr[0]}
                  event={event}
                  filterProps={filterProps}
                  activeProject={activeProject}
                  deleteFilter={removeFilters}
                  insertFilter={insertFilters}
                  closeFilter={closeFilter}
                  refValue={refValue}
                  caller={'profiles'}
                />
              </div>
              <div key={ind + 1}>
                <EventFilterWrapper
                  displayMode={displayMode}
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
                  caller={'profiles'}
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

  return (
    <div
      className={`${styles.query_block} fa--query_block_section borderless no-padding`}
    >
      <div
        className={`${styles.query_block__event} block_section items-center`}
      >
        <div className='flex items-center'>
          <div className={`flex items-center`}>
            <div className='relative ml-2'>
              <Tooltip
                title={
                  eventNames[event?.label]
                    ? eventNames[event?.label]
                    : event?.label
                }
              >
                {!event ? (
                  <Button
                    type='text'
                    onClick={() => setDDVisible(true)}
                    icon={<SVG name='plus' color='grey' />}
                  >
                    {queries.length ? 'Add another event' : 'Add First Event'}
                  </Button>
                ) : (
                  <Button
                    icon={
                      <SVG
                        name='mouseevent'
                        size={16}
                        color={displayMode ? 'grey' : 'purple'}
                      />
                    }
                    className={`fa-button--truncate fa-button--truncate-lg ${
                      displayMode ? 'static-button' : ''
                    } btn-total-round`}
                    type={displayMode ? 'default' : 'link'}
                    onClick={() => (displayMode ? null : setDDVisible(true))}
                  >
                    {eventNames[event.label]
                      ? eventNames[event.label]
                      : event.label}
                  </Button>
                )}
                {selectEvents()}
              </Tooltip>
            </div>
            {event && !displayMode ? (
              <div className={styles.query_block__additional_actions}>
                {additionalActions()}
              </div>
            ) : null}
          </div>
        </div>
      </div>
      {eventFilters()}
    </div>
  );
}

const mapStateToProps = (state) => ({
  eventOptions: state.coreQuery.eventOptions,
  activeProject: state.global.active_project,
  userProperties: state.coreQuery.userProperties,
  groupProperties: state.coreQuery.groupProperties,
  eventProperties: state.coreQuery.eventProperties,
  eventNames: state.coreQuery.eventNames
});

const mapDispatchToProps = (dispatch) =>
  bindActionCreators(
    {
      getGroupProperties
    },
    dispatch
  );

export default connect(mapStateToProps, mapDispatchToProps)(EventsBlock);
