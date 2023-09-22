import React, { useState, useEffect } from 'react';
import { Button, Tooltip } from 'antd';
import { connect } from 'react-redux';
import { bindActionCreators } from 'redux';
import { SVG } from 'Components/factorsComponents';
import styles from './index.module.scss';
import {
  getEventPropertiesV2,
  getUserPropertiesV2
} from 'Reducers/coreQuery/middleware';
import FilterWrapper from 'Components/GlobalFilter/FilterWrapper';
import ORButton from 'Components/ORButton';
import { compareFilters, groupFilters } from 'Utils/global';
import { DEFAULT_OPERATOR_PROPS } from 'Components/FaFilterSelect/utils';
import { OPERATORS } from 'Utils/constants';
import GroupSelect from 'Components/GenericComponents/GroupSelect';
import getGroupIcon from 'Utils/getGroupIcon';
import { processProperties } from 'Utils/dataFormatter';

const ENGAGEMENT_SUPPORTED_OPERATORS = [
  OPERATORS['equalTo'],
  OPERATORS['notEqualTo'],
  OPERATORS['contain'],
  OPERATORS['doesNotContain']
];

function EventsBlock({
  isEngagementConfig = false,
  availableGroups,
  index,
  event,
  disableEventEdit = false,
  closeEvent,
  eventChange,
  eventOptions,
  eventNames,
  activeProject,
  eventPropertiesV2,
  getEventPropertiesV2,
  eventUserPropertiesV2,
  getUserPropertiesV2,
  groupAnalysis,
  viewMode,
  dropdownPlacement = 'top',
  propertiesScope = ['event'],
  initialDDState = true
}) {
  const [isDDVisible, setDDVisible] = useState(initialDDState);
  useEffect(() => {
    if (viewMode) {
      setDDVisible(false);
    }
  }, [viewMode]);
  const [isFilterDDVisible, setFilterDDVisible] = useState(false);
  const [filterProps, setFilterProperties] = useState();
  const [showGroups, setShowGroups] = useState([]);
  const [orFilterIndex, setOrFilterIndex] = useState(-1);

  const operatorProps = isEngagementConfig
    ? {
        categorical: DEFAULT_OPERATOR_PROPS.categorical.filter((item) =>
          ENGAGEMENT_SUPPORTED_OPERATORS.includes(item)
        ),
        numerical: DEFAULT_OPERATOR_PROPS.numerical.filter((item) =>
          ENGAGEMENT_SUPPORTED_OPERATORS.includes(item)
        )
      }
    : DEFAULT_OPERATOR_PROPS;

  useEffect(() => {
    let showOpts = [];

    if (groupAnalysis === 'users') {
      showOpts = [
        ...eventOptions.filter(
          (group) =>
            !['Linkedin Company Engagements', 'G2 Engagements'].includes(
              group?.label
            )
        )
      ];
    } else if (groupAnalysis === 'events') {
      showOpts = [...eventOptions];
    } else {
      const [label] =
        availableGroups?.find((group) => group[1] === groupAnalysis) || [];
      const groupOpts = eventOptions?.filter((item) => item?.label === label);
      const userOpts = eventOptions?.filter(
        (item) =>
          !availableGroups?.map((group) => group[0]).includes(item?.label)
      );
      showOpts = groupOpts.concat(userOpts);
    }
    showOpts = showOpts?.map((opt) => {
      return {
        iconName: getGroupIcon(opt?.icon),
        label: opt?.label,
        values: processProperties(opt?.values)
      };
    });
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
  }, [eventOptions, groupAnalysis]);

  const onChange = (option, group) => {
    const newEvent = { label: '', filters: [], group: '' };
    newEvent.label = option?.value;
    newEvent.group = group?.label;
    setDDVisible(false);
    eventChange(newEvent, index - 1);
    closeEvent();
  };

  useEffect(() => {
    if (!event || event === undefined) {
      return;
    }
    if (!eventPropertiesV2[event.label] && !viewMode) {
      getEventPropertiesV2(activeProject?.id, event.label);
    }
    getUserPropertiesV2(activeProject.id);
  }, [activeProject?.id, event, eventPropertiesV2, viewMode]);

  useEffect(() => {
    if (!event || event === undefined) {
      return;
    }
    const eventPropertiesFiltered = {};
    if (eventPropertiesV2?.[event?.label]) {
      for (const key in eventPropertiesV2[event?.label]) {
        if (eventPropertiesV2[event?.label].hasOwnProperty(key)) {
          eventPropertiesFiltered[key] = eventPropertiesV2[event?.label][
            key
          ].filter((item) => item?.[2] === 'categorical');
        }
      }
    }
    const eventUserPropertiesFiltered = {};
    if (eventUserPropertiesV2) {
      for (const key in eventUserPropertiesV2) {
        if (eventUserPropertiesV2.hasOwnProperty(key)) {
          eventUserPropertiesFiltered[key] = eventUserPropertiesV2[key].filter(
            (item) => item?.[2] === 'categorical'
          );
        }
      }
    }
    const assignFilterProps = {};
    propertiesScope.forEach((scope) => {
      if (scope === 'event') {
        assignFilterProps.event = isEngagementConfig
          ? eventPropertiesFiltered
          : eventPropertiesV2[event?.label] || {};
      }
      if (scope === 'user') {
        assignFilterProps.user = isEngagementConfig
          ? eventUserPropertiesFiltered
          : eventUserPropertiesV2 || {};
      }
    });
    setFilterProperties(assignFilterProps);
  }, [eventPropertiesV2, eventUserPropertiesV2, event]);

  const deleteItem = () => {
    eventChange(event, index - 1, 'delete');
  };

  const selectEvents = () =>
    isDDVisible && !disableEventEdit ? (
      <div className={styles.query_block__event_selector}>
        <GroupSelect
          options={showGroups}
          searchPlaceHolder='Select Event'
          optionClickCallback={onChange}
          allowSearch={true}
          placement={dropdownPlacement}
          onClickOutside={() => {
            setDDVisible(false);
            closeEvent();
          }}
          extraClass={`${styles.query_block__event_selector__select}`}
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
    <FilterWrapper
      operatorsMap={operatorProps}
      viewMode={viewMode}
      filterProps={filterProps}
      projectID={activeProject?.id}
      event={event}
      deleteFilter={closeFilter}
      insertFilter={insertFilters}
      closeFilter={closeFilter}
      refValue={ind}
      caller='profiles'
      dropdownPlacement={dropdownPlacement}
      dropdownMaxHeight={344}
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
        {!disableEventEdit && (
          <Tooltip title={`Delete this event`} color='#0B1E39'>
            <Button
              type='text'
              onClick={deleteItem}
              className='fa-btn--custom btn-total-round'
            >
              <SVG name='trash' />
            </Button>
          </Tooltip>
        )}
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
                  <FilterWrapper
                    operatorsMap={operatorProps}
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
                  />
                </div>
                {ind !== orFilterIndex && !viewMode && (
                  <ORButton index={ind} setOrFilterIndex={setOrFilterIndex} />
                )}
              </div>
              {ind === orFilterIndex && (
                <div key='init'>
                  <FilterWrapper
                    operatorsMap={operatorProps}
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
                  operatorsMap={operatorProps}
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
                />
              </div>
              <div key={ind + 1}>
                <FilterWrapper
                  operatorsMap={operatorProps}
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
            <div className='relative'>
              <Tooltip
                title={
                  eventNames[event?.label]
                    ? eventNames[event?.label]
                    : event?.label
                }
              >
                {!event ? (
                  <Button
                    className='btn-total-round'
                    type='link'
                    onClick={() => setDDVisible(true)}
                  >
                    Select Event
                  </Button>
                ) : (
                  <Button
                    icon={
                      <SVG
                        name='mouseevent'
                        size={16}
                        color={viewMode ? 'grey' : 'purple'}
                      />
                    }
                    className={`fa-button--truncate fa-button--truncate-lg ${
                      viewMode ? 'static-button' : ''
                    } btn-total-round ${
                      disableEventEdit ? 'pointer-events-none' : ''
                    }`}
                    type={viewMode ? 'default' : 'link'}
                    onClick={() =>
                      viewMode || disableEventEdit ? null : setDDVisible(true)
                    }
                  >
                    {eventNames[event.label]
                      ? eventNames[event.label]
                      : event.label}
                  </Button>
                )}
                {selectEvents()}
              </Tooltip>
            </div>
            {event && !viewMode ? additionalActions() : null}
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
  eventUserPropertiesV2: state.coreQuery.eventUserPropertiesV2,
  groupProperties: state.coreQuery.groupProperties,
  eventPropertiesV2: state.coreQuery.eventPropertiesV2,
  eventNames: state.coreQuery.eventNames
});

const mapDispatchToProps = (dispatch) =>
  bindActionCreators({ getEventPropertiesV2, getUserPropertiesV2 }, dispatch);

export default connect(mapStateToProps, mapDispatchToProps)(EventsBlock);
