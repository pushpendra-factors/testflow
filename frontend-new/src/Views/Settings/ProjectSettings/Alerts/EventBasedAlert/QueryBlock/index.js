import React, { useState, useEffect, useMemo } from 'react';
import { Button, Tooltip } from 'antd';
import { connect } from 'react-redux';
import { bindActionCreators } from 'redux';
import { SVG, Text } from 'factorsComponents';
import {
  setGroupBy,
  delGroupBy,
  getGroupProperties,
  getEventPropertiesV2
} from 'Reducers/coreQuery/middleware';
import FilterWrapper from 'Components/GlobalFilter/FilterWrapper';
import EventGroupBlock from 'Components/QueryComposer/EventGroupBlock';
import AliasModal from 'Components/QueryComposer/AliasModal';
import ORButton from 'Components/ORButton';
import { compareFilters, groupFilters } from 'Utils/global';
import GroupSelect from 'Components/GenericComponents/GroupSelect';
import getGroupIcon from 'Utils/getGroupIcon';
import { processProperties } from 'Utils/dataFormatter';
import styles from './index.module.scss';
import { removeDuplicateAndEmptyKeys, blackListedCategories } from './util';

function QueryBlock({
  availableGroups,
  index,
  event,
  eventChange,
  queries,
  queryType,
  eventOptionsSpecial,
  eventNamesSpecial,
  activeProject,
  groupBy,
  setGroupBy,
  delGroupBy,
  eventUserPropertiesV2,
  eventPropertiesV2,
  groupProperties,
  getGroupProperties,
  groupAnalysis,
  getEventPropertiesV2
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
  const [orFilterIndex, setOrFilterIndex] = useState(-1);
  const [isModalVisible, setIsModalVisible] = useState(false);

  const eventGroup = useMemo(() => {
    const group =
      availableGroups?.find((group) => group?.[0] === event?.group) || [];
    return group[1];
  }, [availableGroups, event]);

  useEffect(() => {
    let showOpts = [];

    const groupNamesList = availableGroups?.map((item) => item[0]);
    if (groupAnalysis === 'users') {
      const userOpts = eventOptionsSpecial?.filter(
        (item) => !groupNamesList?.includes(item?.label)
      );
      // remove blacklisted events
      const userOptsNew = userOpts?.filter(
        (item) => !blackListedCategories?.includes(item?.label)
      );
      showOpts = userOptsNew;
    } else {
      // remove blacklisted events
      const groupOpts = eventOptionsSpecial?.filter(
        (item) => !blackListedCategories?.includes(item?.label)
      );

      const filteredData = groupOpts?.map((group) => {
        if (group?.icon === 'hubspot_contacts') {
          const values = group?.values?.filter(
            (vl) => vl?.[1] === '$hubspot_form_submission'
          );
          group.values = values;
        }
        return group;
      });
      showOpts = filteredData;
    }

    showOpts = showOpts?.map((opt) => ({
      iconName: getGroupIcon(opt?.icon),
      label: opt?.label,
      values: processProperties(opt?.values)
    }));

    setShowGroups(showOpts);
  }, [eventOptionsSpecial, groupAnalysis, availableGroups]);

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

  const onChange = (option, group) => {
    const newEvent = { alias: '', label: '', filters: [], group: '' };
    newEvent.label = option.value;
    newEvent.group = group.label;
    setDDVisible(false);
    eventChange(newEvent, index - 1);
  };

  useEffect(() => {
    if (!event || event === undefined) {
      return;
    }
    if (eventGroup && !groupProperties[eventGroup]) {
      getGroupProperties(activeProject.id, eventGroup);
    }
  }, [event]);

  useEffect(() => {
    queries.forEach((ev) => {
      if (!eventPropertiesV2[ev.label]) {
        getEventPropertiesV2(activeProject.id, ev.label);
      }
    });
  }, [queries]);

  const filterProperties = useMemo(() => {
    const Company_identification = 'Company Identification';
    if (!event) return {};
    const props = {
      event: eventPropertiesV2[event.label] || []
    };
    if (eventGroup) {
      if (groupProperties) {
        Object?.keys(groupProperties)?.map((key) => {
          props[key] = groupProperties[key];
        });
      }
    } else {
      props.user = eventUserPropertiesV2;
      if (groupAnalysis === 'events') {
        if (groupProperties) {
          Object?.keys(groupProperties)?.map((key) => {
            props[key] = groupProperties[key];
          });
        }
        // only for accounts - user events
        if (!eventGroup) {
          // remove 'Company identification' from user properties since the same duplicate properties is available in $6_signal
          if (props?.hasOwnProperty('user')) {
            if (props?.user?.hasOwnProperty(Company_identification)) {
              delete props?.user[Company_identification];
            }
          }
        }
      }
    }
    const finalProps = removeDuplicateAndEmptyKeys(props);
    return finalProps;
  }, [
    event,
    eventGroup,
    eventPropertiesV2,
    groupProperties,
    eventUserPropertiesV2
  ]);

  const triggerDropDown = () => {
    setDDVisible(true);
  };

  const deleteItem = () => {
    eventChange(event, index - 1, 'delete');
  };

  const selectEvents = () =>
    isDDVisible ? (
      <div className={styles.query_block__event_selector}>
        <GroupSelect
          options={showGroups}
          optionClickCallback={onChange}
          allowSearch
          onClickOutside={() => setDDVisible(false)}
          extraClass={`${styles.query_block__event_selector__select}`}
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
    <FilterWrapper
      filterProps={filterProperties}
      projectID={activeProject?.id}
      event={event}
      deleteFilter={closeFilter}
      insertFilter={insertFilters}
      closeFilter={closeFilter}
      refValue={ind}
      showInList
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

  const additionalActions = () => (
    <div className='fa--query_block--actions-cols flex'>
      <div className='relative'>
        <Button
          type='text'
          style={{ color: '#8692A3' }}
          onClick={() => setAdditionalactions(['Filter By', 'filter'])}
          className='-ml-2'
          icon={<SVG name='plus' color='#8692A3' />}
        >
          Add a filter
        </Button>

        <AliasModal
          visible={isModalVisible}
          event={
            eventNamesSpecial[event.label]
              ? eventNamesSpecial[event.label]
              : event.label
          }
          onOk={handleOk}
          onCancel={handleCancel}
          alias={event.alias}
        />
      </div>
    </div>
  );

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
            <div className='fa--query_block--filters flex flex-wrap'>
              <div key={ind}>
                <FilterWrapper
                  index={ind}
                  filter={filter}
                  event={event}
                  filterProps={filterProperties}
                  projectID={activeProject?.id}
                  deleteFilter={removeFilters}
                  insertFilter={insertFilters}
                  closeFilter={closeFilter}
                  refValue={refValue}
                  showInList
                />
              </div>
              {ind !== orFilterIndex && (
                <ORButton index={ind} setOrFilterIndex={setOrFilterIndex} />
              )}
              {ind === orFilterIndex && (
                <div key='init'>
                  <FilterWrapper
                    filterProps={filterProperties}
                    projectID={activeProject?.id}
                    event={event}
                    deleteFilter={closeFilter}
                    insertFilter={insertFilters}
                    closeFilter={closeFilter}
                    refValue={refValue}
                    showOr
                    showInList
                  />
                </div>
              )}
            </div>
          );
          ind += 1;
        } else {
          filters.push(
            <div className='fa--query_block--filters flex flex-wrap'>
              <div key={ind}>
                <FilterWrapper
                  index={ind}
                  filter={filtersGr[0]}
                  event={event}
                  filterProps={filterProperties}
                  projectID={activeProject?.id}
                  deleteFilter={removeFilters}
                  insertFilter={insertFilters}
                  closeFilter={closeFilter}
                  refValue={refValue}
                  showInList
                />
              </div>
              <div key={ind + 1}>
                <FilterWrapper
                  index={ind + 1}
                  filter={filtersGr[1]}
                  event={event}
                  filterProps={filterProperties}
                  projectID={activeProject?.id}
                  deleteFilter={removeFilters}
                  insertFilter={insertFilters}
                  closeFilter={closeFilter}
                  refValue={refValue}
                  showOr
                  showInList
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
  if (!event) {
    return (
      <div
        className={`${styles.query_block} fa--query_block my-1 ${
          ifQueries ? 'borderless no-padding' : 'borderless no-padding'
        }`}
      >
        <div
          className={`${styles.query_block__event} flex justify-start items-center`}
        >
          <Button
            type='link'
            onClick={triggerDropDown}
            // icon={<SVG name='plus' color='grey' />}
          >
            {ifQueries ? 'Select another event' : 'Select Event'}
          </Button>
          {selectEvents()}
        </div>
      </div>
    );
  }

  return (
    <div
      className={`${styles.query_block} fa--query_block_section borderless no-padding mt-0`}
    >
      <div
        className={`${!event?.alias?.length ? 'flex justify-start' : ''} ${
          styles.query_block__event
        } block_section items-center`}
      >
        <div className='flex items-center'>
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
        <div className={`flex ${!event?.alias?.length ? '' : 'ml-8 mt-0'}`}>
          <div className='max-w-7xl'>
            <Tooltip
              title={
                eventNamesSpecial[event.label]
                  ? eventNamesSpecial[event.label]
                  : event.label
              }
            >
              <Button
                icon={
                  <SVG
                    name={
                      showGroups.find((group) => group.label === event.group)
                        ?.iconName
                    }
                    size={20}
                  />
                }
                className='fa-button--truncate fa-button--truncate-lg'
                type='link'
                onClick={triggerDropDown}
              >
                {eventNamesSpecial[event.label]
                  ? eventNamesSpecial[event.label]
                  : event.label}
              </Button>
              {selectEvents()}
            </Tooltip>
          </div>
        </div>
      </div>
      {queries?.[0]?.filters?.length > 0 && (
        <Text type='title' level={7} extraClass='m-0 ml-2 mt-2'>
          Filter By
        </Text>
      )}
      {eventFilters()}
      <div className='mt-2'>{additionalActions()}</div>
      {/* {groupByItems()} */}
    </div>
  );
}

const mapStateToProps = (state) => ({
  eventOptionsSpecial: state.coreQuery.eventOptionsSpecial,
  activeProject: state.global.active_project,
  eventUserPropertiesV2: state.coreQuery.eventUserPropertiesV2,
  groupProperties: state.coreQuery.groupProperties,
  eventPropertiesV2: state.coreQuery.eventPropertiesV2,
  groupBy: state.coreQuery.groupBy.event,
  groupByMagic: state.coreQuery.groupBy,
  eventNamesSpecial: state.coreQuery.eventNamesSpecial
});

const mapDispatchToProps = (dispatch) =>
  bindActionCreators(
    {
      setGroupBy,
      delGroupBy,
      getGroupProperties,
      getEventPropertiesV2
    },
    dispatch
  );

export default connect(mapStateToProps, mapDispatchToProps)(QueryBlock);
