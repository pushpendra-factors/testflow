import React, { memo, useCallback, useMemo, useState } from 'react';
import { useSelector } from 'react-redux';
import cx from 'classnames';
import { Button, Dropdown, Menu } from 'antd';
import cloneDeep from 'lodash/cloneDeep';
import map from 'lodash/map';
import { SVG, Text } from 'Components/factorsComponents';
import FilterWrapper from 'Components/GlobalFilter/FilterWrapper';
import ControlledComponent from 'Components/ControlledComponent/ControlledComponent';
import { selectGroupsList } from 'Reducers/groups/selectors';
import { generateRandomKey } from 'Utils/global';
import {
  eventMenuList,
  eventTimelineMenuList
} from './accountProfiles.constants';
import {
  checkFiltersEquality,
  computeFilterProperties
} from './accountProfiles.helpers';
import EventsBlock from '../MyComponents/EventsBlock';
import styles from './index.module.scss';

function FiltersBox({
  filtersList,
  secondaryFiltersList,
  setSecondaryFiltersList,
  profileType = 'account',
  source,
  appliedFilters,
  setFiltersList,
  applyFilters,
  onCancel,
  setSaveSegmentModal,
  listEvents,
  areFiltersDirty,
  setListEvents,
  eventProp,
  setEventProp,
  onClearFilters,
  disableDiscardButton,
  isActiveSegment,
  eventTimeline,
  setEventTimeline
}) {
  const { newSegmentMode: accountsNewSegmentMode } = useSelector(
    (state) => state.accountProfilesView
  );
  const { newSegmentMode: profilesNewSegmentMode } = useSelector(
    (state) => state.userProfilesView
  );

  const newSegmentMode = accountsNewSegmentMode || profilesNewSegmentMode;
  const groupsList = useSelector((state) => selectGroupsList(state));
  const activeProject = useSelector((state) => state.global.active_project);
  const [filterDD, setFilterDD] = useState(false);
  const [secondaryFilterDD, setSecondaryFilterDD] = useState(false);
  const [isEventsVisible, setEventsVisible] = useState(false);
  const userProperties = useSelector(
    (state) => state.coreQuery.userPropertiesV2
  );
  const groupProperties = useSelector(
    (state) => state.coreQuery.groupProperties
  );

  const availableGroups = useSelector((state) => state.coreQuery.groups);

  const handleEventChange = useCallback(
    (eventItem) => {
      setEventProp(eventItem.key);
    },
    [setEventProp]
  );

  const handleEventTimelineChange = useCallback(
    (item) => {
      setEventTimeline(item.key);
    },
    [setEventTimeline]
  );

  const eventMenuItems = (
    <Menu className={styles['dropdown-menu']}>
      {map(eventMenuList, (item) => (
        <Menu.Item
          className={styles['dropdown-menu-item']}
          onClick={() => handleEventChange(item)}
          key={item.key}
        >
          <Text type='title' extraClass='mb-0'>
            {item.label}
          </Text>
        </Menu.Item>
      ))}
    </Menu>
  );

  const eventTimelineMenuItems = (
    <Menu className={styles['dropdown-menu']}>
      {map(eventTimelineMenuList, (item) => (
        <Menu.Item
          className={styles['dropdown-menu-item']}
          onClick={() => handleEventTimelineChange(item)}
          key={item.key}
        >
          <Text type='title' extraClass='mb-0'>
            {item.label}
          </Text>
        </Menu.Item>
      ))}
    </Menu>
  );

  const mainFilterProps = useMemo(
    () =>
      computeFilterProperties({
        userProperties,
        groupProperties,
        availableGroups: availableGroups?.account_groups,
        profileType
      }),
    [userProperties, groupProperties, availableGroups, profileType]
  );

  const userFilterProps = useMemo(
    () =>
      computeFilterProperties({
        userProperties,
        groupProperties,
        availableGroups: availableGroups?.account_groups,
        profileType: 'user',
        source: 'users'
      }),
    [userProperties, groupProperties, availableGroups]
  );

  const handleInsertFilter = useCallback(
    (filter, index) => {
      if (filtersList.length === index) {
        setFiltersList([...filtersList, filter]);
      } else {
        setFiltersList([
          ...filtersList.slice(0, index),
          filter,
          ...filtersList.slice(index + 1)
        ]);
      }
    },
    [filtersList, setFiltersList]
  );

  const handleDeleteFilter = useCallback(
    (filterIndex) => {
      if (filterIndex === filtersList.length) {
        setFilterDD(false);
        return;
      }
      setFiltersList(filtersList.filter((_, index) => index !== filterIndex));
    },
    [setFiltersList, filtersList]
  );

  const handleInsertSecondaryFilter = useCallback(
    (filter, index) => {
      if (secondaryFiltersList.length === index) {
        setSecondaryFiltersList([...secondaryFiltersList, filter]);
      } else {
        setSecondaryFiltersList([
          ...secondaryFiltersList.slice(0, index),
          filter,
          ...secondaryFiltersList.slice(index + 1)
        ]);
      }
    },
    [secondaryFiltersList, setSecondaryFiltersList]
  );

  const handleDeleteSecondaryFilter = useCallback(
    (filterIndex) => {
      if (filterIndex === secondaryFiltersList.length) {
        setSecondaryFilterDD(false);
        return;
      }
      setSecondaryFiltersList(
        secondaryFiltersList.filter((_, index) => index !== filterIndex)
      );
    },
    [setSecondaryFiltersList, secondaryFiltersList]
  );

  const showFilterDropdown = useCallback(() => {
    setFilterDD(true);
  }, []);

  const showSecondaryFilterDropdown = useCallback(() => {
    setSecondaryFilterDD(true);
  }, []);

  const handleCloseFilter = useCallback(() => {
    setFilterDD(false);
  }, []);

  const handleCloseSecondaryFilter = useCallback(() => {
    setSecondaryFilterDD(false);
  }, []);

  const showEventsDropdown = useCallback(() => {
    setEventsVisible(true);
  }, []);

  const closeEvent = useCallback(() => {
    setEventsVisible(false);
  }, []);

  const handleQueryChange = useCallback(
    (newEvent, index, changeType = 'add') => {
      const updatedQuery = cloneDeep(listEvents);
      if (updatedQuery[index]) {
        if (changeType === 'add' || changeType === 'filters_updated') {
          updatedQuery[index] = newEvent;
        } else if (changeType === 'delete') {
          updatedQuery.splice(index, 1);
        }
      } else {
        updatedQuery.push(newEvent);
      }
      setListEvents(
        updatedQuery.map((q) => ({
          ...q,
          key: q.key || generateRandomKey()
        }))
      );
    },
    [listEvents, setListEvents]
  );

  const { applyButtonDisabled, saveButtonDisabled } = useMemo(
    () =>
      checkFiltersEquality({
        appliedFilters,
        filtersList,
        newSegmentMode,
        eventsList: listEvents,
        eventProp,
        isActiveSegment,
        areFiltersDirty,
        secondaryFiltersList
      }),
    [
      appliedFilters,
      filtersList,
      newSegmentMode,
      listEvents,
      eventProp,
      isActiveSegment,
      areFiltersDirty,
      secondaryFiltersList
    ]
  );

  const showClearAllButton = useMemo(
    () =>
      appliedFilters.filters.length > 0 || appliedFilters.eventsList.length > 0,
    [appliedFilters.eventsList.length, appliedFilters.filters.length]
  );

  return (
    <div
      className={cx(styles['filters-box-container'], 'flex flex-col row-gap-5')}
    >
      <div className='pt-4 col-gap-5 flex flex-col row-gap-2'>
        <div className={cx('px-6 pb-1', styles['section-title-container'])}>
          <Text
            type='title'
            color='character-secondary'
            extraClass='mb-0'
            weight='medium'
          >
            {profileType === 'account'
              ? 'With account properties'
              : 'With properties'}
          </Text>
        </div>
        <div className='px-6'>
          <ControlledComponent controller={filtersList.length > 0}>
            {filtersList.map((filter, index) => (
              <FilterWrapper
                key={index}
                viewMode={false}
                projectID={activeProject?.id}
                filter={filter}
                index={index}
                filterProps={mainFilterProps}
                minEntriesPerGroup={3}
                insertFilter={handleInsertFilter}
                closeFilter={handleCloseFilter}
                deleteFilter={handleDeleteFilter}
              />
            ))}
          </ControlledComponent>

          <ControlledComponent controller={filterDD === true}>
            <FilterWrapper
              viewMode={false}
              projectID={activeProject?.id}
              index={filtersList.length}
              filterProps={mainFilterProps}
              minEntriesPerGroup={3}
              insertFilter={handleInsertFilter}
              closeFilter={handleCloseFilter}
              deleteFilter={handleDeleteFilter}
            />
          </ControlledComponent>

          <Button
            className={cx(
              'flex items-center col-gap-2',
              styles['add-filter-button']
            )}
            type='text'
            onClick={showFilterDropdown}
          >
            <SVG name='plus' color='#00000073' />
            <Text
              type='title'
              color='character-title'
              extraClass='mb-0'
              weight='medium'
            >
              Add filter
            </Text>
          </Button>
        </div>
      </div>
      <div className='flex flex-col row-gap-2'>
        <div className={cx('px-6 pb-1', styles['section-title-container'])}>
          <Text
            type='title'
            color='character-secondary'
            extraClass='mb-0'
            weight='medium'
          >
            Who Performed
          </Text>
        </div>
        <div className='px-6 flex flex-col row-gap-2'>
          {listEvents.map((event, index) => (
            <div key={index}>
              <EventsBlock
                isEngagementConfig={false}
                availableGroups={groupsList}
                index={index + 1}
                event={event}
                queries={listEvents}
                groupAnalysis={source}
                eventChange={handleQueryChange}
                closeEvent={closeEvent}
                initialDDState={false}
              />
            </div>
          ))}
          <ControlledComponent
            controller={isEventsVisible === true && listEvents.length < 3}
          >
            <div key={listEvents.length}>
              <EventsBlock
                isEngagementConfig={false}
                availableGroups={groupsList}
                index={listEvents.length + 1}
                queries={listEvents}
                groupAnalysis={source}
                eventChange={handleQueryChange}
                closeEvent={closeEvent}
              />
            </div>
          </ControlledComponent>
          <ControlledComponent controller={listEvents.length < 3}>
            <Button
              className={cx('flex items-center col-gap-2', styles['add-filter-button'])}
              type='text'
              onClick={showEventsDropdown}
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
          </ControlledComponent>
          <ControlledComponent controller={false}>
            <div className='flex col-gap-1 items-center'>
              <Text
                type='title'
                extraClass='mb-0'
                color='character-primary'
                weight='medium'
              >
                Events performed in
              </Text>
              <Dropdown overlay={eventTimelineMenuItems}>
                <div
                  className={cx(
                    'flex col-gap-1 cursor-pointer items-center',
                    styles['event-timeline-picker']
                  )}
                >
                  <Text
                    type='title'
                    color='character-primary'
                    extraClass='mb-0'
                    weight='medium'
                  >
                    {eventTimelineMenuList[eventTimeline].label}
                  </Text>
                </div>
              </Dropdown>
            </div>
          </ControlledComponent>
          <ControlledComponent controller={listEvents.length > 1}>
            <div className='flex col-gap-1 items-center'>
              <Text
                type='title'
                extraClass='mb-0'
                color='character-primary'
                weight='medium'
              >
                {profileType === 'account'
                  ? 'Accounts that performed'
                  : 'People who performed'}
              </Text>
              <Dropdown overlay={eventMenuItems}>
                <div className='flex col-gap-1 cursor-pointer items-center'>
                  <Text type='title' color='brand-color-6' extraClass='mb-0'>
                    {eventMenuList[eventProp].label}
                  </Text>
                  <SVG name='caretDown' color='#1890ff' size={20} />
                </div>
              </Dropdown>
            </div>
          </ControlledComponent>
        </div>
      </div>
      <ControlledComponent controller={profileType === 'account'}>
        <div className='flex flex-col row-gap-2 col-gap-5'>
          <div className={cx('px-6 pb-1', styles['section-title-container'])}>
            <Text
              type='title'
              color='character-secondary'
              extraClass='mb-0'
              weight='medium'
            >
              With at least 1 person that matches
            </Text>
          </div>
          <div className='px-6'>
            <ControlledComponent controller={secondaryFiltersList.length > 0}>
              {secondaryFiltersList.map((filter, index) => (
                <FilterWrapper
                  key={index}
                  viewMode={false}
                  projectID={activeProject?.id}
                  filter={filter}
                  index={index}
                  filterProps={userFilterProps}
                  minEntriesPerGroup={3}
                  insertFilter={handleInsertSecondaryFilter}
                  closeFilter={handleCloseSecondaryFilter}
                  deleteFilter={handleDeleteSecondaryFilter}
                />
              ))}
            </ControlledComponent>

            <ControlledComponent controller={secondaryFilterDD === true}>
              <FilterWrapper
                viewMode={false}
                projectID={activeProject?.id}
                index={secondaryFiltersList.length}
                filterProps={userFilterProps}
                minEntriesPerGroup={3}
                insertFilter={handleInsertSecondaryFilter}
                closeFilter={handleCloseSecondaryFilter}
                deleteFilter={handleDeleteSecondaryFilter}
              />
            </ControlledComponent>

            <Button
              className={cx(
                'flex items-center col-gap-2',
                styles['add-filter-button']
              )}
              type='text'
              onClick={showSecondaryFilterDropdown}
            >
              <SVG name='plus' color='#00000073' />
              <Text
                type='title'
                color='character-title'
                extraClass='mb-0'
                weight='medium'
              >
                Add filter
              </Text>
            </Button>
          </div>
        </div>
      </ControlledComponent>
      <div
        className={cx(
          'py-4 px-6 flex items-center justify-between',
          styles['buttons-container']
        )}
      >
        <div className='flex col-gap-2 items-center'>
          <Button
            disabled={applyButtonDisabled}
            onClick={applyFilters}
            type='primary'
          >
            Apply changes
          </Button>
          <Button
            disabled={disableDiscardButton}
            type='secondary'
            onClick={onCancel}
          >
            Discard changes
          </Button>
        </div>
        <ControlledComponent
          controller={showClearAllButton === true && newSegmentMode === false}
        >
          <Button
            type='text'
            className='flex items-center col-gap-1'
            onClick={onClearFilters}
          >
            <Text type='title' extraClass='mb-0' color='character-title'>
              Clear all filters
            </Text>
          </Button>
        </ControlledComponent>
        <ControlledComponent controller={newSegmentMode === true}>
          <Button
            type='default'
            className='flex items-center col-gap-1'
            disabled={saveButtonDisabled}
            onClick={() => setSaveSegmentModal(true)}
          >
            <SVG
              color={saveButtonDisabled ? '#BFBFBF' : '#1890ff'}
              size={16}
              name='pieChart'
            />
            <Text
              type='title'
              extraClass='mb-0'
              color={saveButtonDisabled ? 'disabled' : 'brand-color-6'}
            >
              Save segment
            </Text>
          </Button>
        </ControlledComponent>
      </div>
    </div>
  );
}

export default memo(FiltersBox);
