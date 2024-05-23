import React, { useState, useEffect, useMemo } from 'react';
import { Button, Dropdown, Input, Select, Tooltip, Menu } from 'antd';
import { connect, useSelector } from 'react-redux';
import { bindActionCreators } from 'redux';
import { SVG, Text } from 'Components/factorsComponents';
import {
  getEventPropertiesV2,
  getGroupProperties,
  getUserPropertiesV2
} from 'Reducers/coreQuery/middleware';
import FilterWrapper from 'Components/GlobalFilter/FilterWrapper';
import ORButton from 'Components/ORButton';
import { compareFilters, groupFilters } from 'Utils/global';
import GroupSelect from 'Components/GenericComponents/GroupSelect';
import getGroupIcon from 'Utils/getGroupIcon';
import { processProperties } from 'Utils/dataFormatter';
import { GROUP_NAME_DOMAINS } from 'Components/GlobalFilter/FilterWrapper/utils';
import { DownOutlined } from '@ant-design/icons';
import {
  EVENT_FREQ_OPERATORS,
  INITIAL_EVENT_WITH_PROPERTIES_STATE
} from 'Views/CoreQuery/constants';
import styles from './index.module.scss';

function EventsBlock({
  isEngagementConfig = false,
  index,
  event,
  closeEventDD,
  eventChange,
  getEventPropertiesV2,
  getUserPropertiesV2,
  getGroupProperties,
  groupAnalysis,
  viewMode,
  dropdownPlacement = 'top',
  propertiesScope = ['event'],
  initialDDState = true,
  showInList = false,
  isSpecialEvent = false
}) {
  const [isDDVisible, setDDVisible] = useState(initialDDState);
  const [isFilterDDVisible, setFilterDDVisible] = useState(false);
  const [filterProps, setFilterProperties] = useState();
  const [showGroups, setShowGroups] = useState([]);
  const [orFilterIndex, setOrFilterIndex] = useState(-1);

  const activeProject = useSelector((state) => state.global.active_project);

  const {
    eventNames,
    eventNamesSpecial,
    eventOptions,
    eventOptionsSpecial,
    eventPropertiesV2,
    eventUserPropertiesV2,
    groupProperties,
    groups
  } = useSelector((state) => state.coreQuery);

  const eventGroup = useMemo(() => {
    if (!event || !groups) {
      return null;
    }

    const group =
      Object.entries(groups).find((grp) => grp[1] === event.group) || [];
    return group[0];
  }, [event, groups]);

  const eventNamesApplicable = useMemo(
    () => (isSpecialEvent ? eventNamesSpecial : eventNames),
    [eventNamesSpecial, eventNames]
  );

  const eventTitle = useMemo(() => {
    if (!event) {
      return '';
    }

    return eventNamesApplicable[event.label] || event.label;
  }, [event, eventNamesApplicable]);

  const eventOptionsApplicable = useMemo(
    () => (isSpecialEvent ? eventOptionsSpecial : eventOptions),
    [eventOptions, eventOptionsSpecial]
  );

  const showEngagementGroups = useMemo(() => {
    if (!isEngagementConfig) {
      return showGroups;
    }

    const customEvent = {
      label: 'All Events',
      value: 'all_events',
      extraProps: {
        groupName: undefined,
        propertyType: undefined,
        queryType: undefined,
        valueType: undefined
      }
    };

    const listGroups = [...showGroups];
    const othersIndex = listGroups.findIndex(
      (group) => group.label === 'Others'
    );

    if (othersIndex === -1) {
      listGroups.push({
        label: 'Others',
        iconName: 'Others',
        values: [customEvent]
      });
    } else {
      const allEventsIndex = listGroups[othersIndex].values.findIndex(
        (ev) => ev.value === customEvent.value
      );

      if (allEventsIndex === -1) {
        listGroups[othersIndex].values?.push(customEvent);
      }
    }

    return listGroups;
  }, [showGroups]);

  useEffect(() => {
    if (viewMode) {
      setDDVisible(false);
    }
  }, [viewMode]);

  useEffect(() => {
    if (!event || viewMode) {
      return;
    }

    if (eventGroup?.length && !groupProperties[eventGroup]) {
      getGroupProperties(activeProject?.id, eventGroup);
    }

    if (!eventPropertiesV2[event.label]) {
      getEventPropertiesV2(activeProject?.id, event.label);
    }

    getUserPropertiesV2(activeProject?.id);
  }, [
    activeProject?.id,
    viewMode,
    event,
    eventGroup,
    eventPropertiesV2,
    groupProperties
  ]);

  useEffect(() => {
    let showOpts = [];
    if (groupAnalysis === 'users') {
      showOpts = [
        ...eventOptionsApplicable.filter(
          (group) =>
            !['Linkedin Company Engagements', 'G2 Engagements'].includes(
              group?.label
            )
        )
      ];
    } else if (
      groupAnalysis === 'events' ||
      groupAnalysis === GROUP_NAME_DOMAINS
    ) {
      showOpts = [...eventOptionsApplicable];
    } else {
      const [label] =
        Object.entries(groups || {})?.find(
          (group) => group[0] === groupAnalysis
        ) || [];
      const groupOpts = eventOptionsApplicable?.filter(
        (item) => item?.label === label
      );
      const userOpts = eventOptionsApplicable?.filter(
        (item) =>
          !Object.entries(groups || {})
            ?.map((group) => group[0])
            .includes(item?.label)
      );
      showOpts = groupOpts.concat(userOpts);
    }
    showOpts = showOpts?.map((opt) => ({
      iconName: getGroupIcon(opt?.icon),
      label: opt?.label,
      values: processProperties(opt?.values)
    }));
    // Moving MostRecent as first Option.
    const mostRecentGroupindex = showOpts
      ?.map((opt) => opt.label)
      ?.indexOf('Most Recent');
    if (mostRecentGroupindex > 0) {
      showOpts = [
        showOpts[mostRecentGroupindex],
        ...showOpts.slice(0, mostRecentGroupindex),
        ...showOpts.slice(mostRecentGroupindex + 1)
      ];
    }
    setShowGroups(showOpts);
  }, [eventOptionsApplicable, groupAnalysis]);

  useEffect(() => {
    if (!event || event === undefined) {
      return;
    }

    const assignFilterProps = {};
    propertiesScope.forEach((scope) => {
      if (scope === 'event') {
        assignFilterProps.event = eventPropertiesV2[event?.label] || {};
      }
      if (scope === 'user') {
        if (!eventGroup) {
          assignFilterProps.user = eventUserPropertiesV2 || {};
        }
      }
      if (scope === 'group' && eventGroup && groupProperties[eventGroup]) {
        assignFilterProps[eventGroup] = groupProperties[eventGroup];
        assignFilterProps.user = {};
      }
    });

    setFilterProperties(assignFilterProps);
  }, [
    eventPropertiesV2,
    eventUserPropertiesV2,
    event?.label,
    eventGroup,
    groupProperties
  ]);

  const createNewEventObj = (grpa) => {
    if (grpa === GROUP_NAME_DOMAINS) {
      return INITIAL_EVENT_WITH_PROPERTIES_STATE;
    }
    return { label: '', filters: [], group: '' };
  };

  const onEventChange = (option, group) => {
    const newEvent = createNewEventObj(groupAnalysis);

    if (option?.value) {
      newEvent.label = option.value;
    }

    if (group?.label) {
      newEvent.group = group.label;
    }

    setDDVisible(false);
    eventChange(newEvent, index - 1);
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

  const deleteEvent = () => {
    eventChange(event, index - 1, 'delete');
  };

  const handleEventPerformedChange = (value) => {
    const newEvent = { ...event, isEventPerformed: value };
    eventChange(newEvent, index - 1);
  };

  const handleOperatorChange = (value) => {
    const newEvent = { ...event };
    newEvent.frequencyOperator = value;
    eventChange(newEvent, index - 1);
  };

  const handleFrequencyChange = (e) => {
    const newEvent = { ...event };
    const { value: inputValue } = e.target;
    const reg = /^-?\d*(\.\d*)?$/;
    if (reg.test(inputValue) || inputValue === '' || inputValue === '-') {
      newEvent.frequency = inputValue;
      eventChange(newEvent, index - 1);
    }
  };

  const handleDurationChange = (value) => {
    const newEvent = { ...event };
    newEvent.range = value;
    eventChange(newEvent, index - 1);
  };

  const selectEvents = () => {
    if (!isDDVisible) {
      return null;
    }

    return (
      <div className={styles.query_block__event_selector}>
        <GroupSelect
          options={isEngagementConfig ? showEngagementGroups : showGroups}
          searchPlaceHolder='Select Event'
          optionClickCallback={onEventChange}
          allowSearch
          placement={dropdownPlacement}
          onClickOutside={() => {
            setDDVisible(false);
            closeEventDD();
          }}
          extraClass={`${styles.query_block__event_selector__select}`}
        />
      </div>
    );
  };

  const closeFilter = () => {
    setFilterDDVisible(false);
    setOrFilterIndex(-1);
  };

  const renderAdditionalActions = () => {
    if (!event || viewMode) {
      return null;
    }

    return (
      <div
        className='fa--query_block--actions-cols flex'
        id='additional_actions_events_block'
      >
        <Tooltip
          overlayInnerStyle={{ width: 'max-content' }}
          getPopupContainer={() =>
            document.getElementById('additional_actions_events_block')
          }
          title='Filter this event'
          color='#0B1E39'
        >
          <Button
            type='text'
            onClick={() => setFilterDDVisible(true)}
            className='fa-btn--custom mr-1 btn-total-round'
          >
            <SVG name='filter' />
          </Button>
        </Tooltip>
        <Tooltip
          overlayInnerStyle={{ width: 'max-content' }}
          getPopupContainer={() =>
            document.getElementById('additional_actions_events_block')
          }
          title='Delete this event'
          color='#0B1E39'
        >
          <Button
            type='text'
            onClick={deleteEvent}
            className='fa-btn--custom btn-total-round'
          >
            <SVG name='trash' />
          </Button>
        </Tooltip>
      </div>
    );
  };

  // needs a cleanup
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
                  <FilterWrapper
                    viewMode={viewMode}
                    index={ind}
                    filter={filter}
                    event={event}
                    filterProps={filterProps}
                    projectID={activeProject?.id}
                    deleteFilter={removeFilters}
                    insertFilter={insertFilters}
                    closeFilter={closeFilter}
                    refValue={refValue}
                    caller='profiles'
                    dropdownPlacement={dropdownPlacement}
                    dropdownMaxHeight={344}
                    showInList={showInList}
                  />
                </div>
                {ind !== orFilterIndex && !viewMode && (
                  <ORButton index={ind} setOrFilterIndex={setOrFilterIndex} />
                )}
              </div>
              {ind === orFilterIndex && (
                <div key='init'>
                  <FilterWrapper
                    viewMode={viewMode}
                    filterProps={filterProps}
                    projectID={activeProject?.id}
                    event={event}
                    deleteFilter={closeFilter}
                    insertFilter={insertFilters}
                    closeFilter={closeFilter}
                    refValue={refValue}
                    showOr
                    caller='profiles'
                    dropdownPlacement={dropdownPlacement}
                    dropdownMaxHeight={344}
                    showInList={showInList}
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
                <FilterWrapper
                  viewMode={viewMode}
                  index={ind}
                  filter={filtersGr[0]}
                  event={event}
                  filterProps={filterProps}
                  projectID={activeProject?.id}
                  deleteFilter={removeFilters}
                  insertFilter={insertFilters}
                  closeFilter={closeFilter}
                  refValue={refValue}
                  caller='profiles'
                  dropdownPlacement={dropdownPlacement}
                  dropdownMaxHeight={344}
                  showInList={showInList}
                />
              </div>
              <div key={ind + 1}>
                <FilterWrapper
                  viewMode={viewMode}
                  index={ind + 1}
                  filter={filtersGr[1]}
                  event={event}
                  filterProps={filterProps}
                  projectID={activeProject?.id}
                  deleteFilter={removeFilters}
                  insertFilter={insertFilters}
                  closeFilter={closeFilter}
                  refValue={refValue}
                  showOr
                  caller='profiles'
                  dropdownPlacement={dropdownPlacement}
                  dropdownMaxHeight={344}
                  showInList={showInList}
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
          <FilterWrapper
            viewMode={viewMode}
            filterProps={filterProps}
            projectID={activeProject?.id}
            event={event}
            deleteFilter={closeFilter}
            insertFilter={insertFilters}
            closeFilter={closeFilter}
            refValue={lastRef + 1}
            caller='profiles'
            dropdownPlacement={dropdownPlacement}
            dropdownMaxHeight={344}
            showInList={showInList}
          />
        </div>
      );
    }

    return filters;
  };

  const isEventPerformedOptions = [
    {
      value: true,
      label: 'did'
    },
    {
      value: false,
      label: 'did not do'
    }
  ];

  const renderIsEventPerformedSelect = () => {
    if (!event || isEngagementConfig || groupAnalysis !== GROUP_NAME_DOMAINS) {
      return null;
    }

    return (
      <Select
        className='h-8'
        dropdownMatchSelectWidth={false}
        bordered={false}
        value={event.isEventPerformed}
        options={isEventPerformedOptions}
        onChange={handleEventPerformedChange}
      />
    );
  };

  const renderSelectEventButton = () => (
    <Button
      className='btn-total-round'
      type='link'
      onClick={() => setDDVisible(true)}
    >
      Select Event
    </Button>
  );

  const renderAddEventButton = () => (
    <Button
      className='flex items-center gap-x-2'
      type='text'
      onClick={() => setDDVisible(true)}
    >
      <SVG name='plus' color='#00000073' />
      <Text
        type='title'
        color='character-title'
        extraClass='mb-0'
        weight='medium'
      >
        Add event
      </Text>
    </Button>
  );

  const renderActiveEventButton = () => {
    const currGroup = showGroups.find((group) => group.label === event.group);
    const iconName = currGroup?.iconName || 'mouseclick';

    return (
      <Button
        icon={<SVG name={iconName} size={20} />}
        className={`fa-button--truncate fa-button--truncate-lg ${
          viewMode ? 'static-button' : ''
        } btn-total-round`}
        type={viewMode ? 'default' : 'link'}
        onClick={() => (viewMode ? null : setDDVisible(true))}
      >
        {eventTitle}
      </Button>
    );
  };

  const renderEventButton = () => {
    if (!event) {
      return isEngagementConfig
        ? renderSelectEventButton()
        : renderAddEventButton();
    }
    return renderActiveEventButton();
  };

  const renderEventSection = () => (
    <div className='relative'>
      <Tooltip zIndex={99999} title={eventTitle}>
        {renderEventButton()}
        {selectEvents()}
      </Tooltip>
    </div>
  );

  const renderFrequencyControls = () => {
    if (!event.isEventPerformed) {
      return null;
    }

    const frequencyOptions = Object.entries(EVENT_FREQ_OPERATORS).map(
      ([label, value]) => ({ value, label })
    );

    return (
      <>
        <Select
          className='h-8'
          dropdownMatchSelectWidth={false}
          bordered={false}
          value={event.frequencyOperator}
          options={frequencyOptions}
          onChange={handleOperatorChange}
        />
        <Input
          style={{ width: '56px' }}
          onChange={handleFrequencyChange}
          maxLength={3}
          value={event.frequency}
        />
      </>
    );
  };

  const renderDaysMenu = () => {
    const daysArray = [7, 14, 30, 60, 90];
    return (
      <Menu
        onClick={(info) => {
          const selectedDays = Number(info.key);
          handleDurationChange(selectedDays);
        }}
        style={{ overflowY: 'scroll', maxHeight: '185px' }}
      >
        {daysArray.map((days) => (
          <Menu.Item style={{ padding: '10px' }} key={days}>
            Last {days} Days
          </Menu.Item>
        ))}
      </Menu>
    );
  };

  const renderDaysDropdown = () => {
    const prefix = event.isEventPerformed ? 'times in' : 'in';
    const timePeriod = event.range ? `Last ${event.range} Days` : 'Select';

    return (
      <>
        <div className='mx-2'>{prefix}</div>
        <Dropdown overlay={renderDaysMenu()}>
          <Button className='dropdown-btn gap-x-2 justify-between' type='text'>
            <div className='flex gap-x-1 items-center'>
              <SVG name='calendar' />
              {timePeriod}
            </div>
            <DownOutlined />
          </Button>
        </Dropdown>
      </>
    );
  };

  const renderAdditionalFilters = () => {
    if (!event || isEngagementConfig || groupAnalysis !== GROUP_NAME_DOMAINS) {
      return null;
    }

    return (
      <>
        {renderFrequencyControls()}
        {renderDaysDropdown()}
      </>
    );
  };

  return (
    <div
      className={`${styles.query_block} fa--query_block_section borderless no-padding`}
    >
      <div
        className={`${styles.query_block__event} block_section items-center`}
      >
        <div className='flex items-center'>
          {renderIsEventPerformedSelect()}
          {renderEventSection()}
          {renderAdditionalFilters()}
          {renderAdditionalActions()}
        </div>
      </div>
      {eventFilters()}
    </div>
  );
}

const mapDispatchToProps = (dispatch) =>
  bindActionCreators(
    { getEventPropertiesV2, getUserPropertiesV2, getGroupProperties },
    dispatch
  );

export default connect(null, mapDispatchToProps)(EventsBlock);
