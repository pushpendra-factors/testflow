import React, { memo, useCallback, useMemo, useState } from 'react';
import { useSelector } from 'react-redux';
import cx from 'classnames';
import { Button, Dropdown, Menu } from 'antd';
import cloneDeep from 'lodash/cloneDeep';
import map from 'lodash/map';
import { SVG, Text } from 'Components/factorsComponents';
import FilterWrapper from 'Components/GlobalFilter/FilterWrapper';
import ControlledComponent from 'Components/ControlledComponent/ControlledComponent';
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
  profileType,
  isActiveSegment,
  selectedFilters,
  setSelectedFilters,
  appliedFilters,
  applyFilters,
  setSaveSegmentModal,
  areFiltersDirty,
  onCancel,
  onClearFilters,
  disableDiscardButton
}) {
  const { newSegmentMode: accountsNewSegmentMode } = useSelector(
    (state) => state.accountProfilesView
  );
  const { newSegmentMode: profilesNewSegmentMode } = useSelector(
    (state) => state.userProfilesView
  );

  const newSegmentMode =
    profileType === 'account' ? accountsNewSegmentMode : profilesNewSegmentMode;
  const activeProject = useSelector((state) => state.global.active_project);
  const [filterDD, setFilterDD] = useState(false);
  const [secondaryFilterDD, setSecondaryFilterDD] = useState(false);
  const userProperties = useSelector(
    (state) => state.coreQuery.userPropertiesV2
  );
  const groupProperties = useSelector(
    (state) => state.coreQuery.groupProperties
  );

  const availableGroups = useSelector((state) => state.coreQuery.groups);

  const handleEventChange = useCallback((eventItem) => {
    setSelectedFilters((curr) => ({
      ...curr,
      eventProp: eventItem.key
    }));
  }, []);

  const handleEventTimelineChange = useCallback((item) => {
    setSelectedFilters((curr) => ({
      ...curr,
      eventTimeline: item.key
    }));
  }, []);

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
        availableGroups: availableGroups?.all_groups,
        profileType
      }),
    [userProperties, groupProperties, availableGroups, profileType]
  );

  const userFilterProps = useMemo(
    () =>
      computeFilterProperties({
        userProperties,
        groupProperties,
        availableGroups: availableGroups?.all_groups,
        profileType: 'user'
      }),
    [userProperties, groupProperties, availableGroups]
  );

  const handleInsertFilter = useCallback((filter, index) => {
    setSelectedFilters((curr) => {
      const newFilters = [...curr.filters];
      if (newFilters.length === index) {
        newFilters.push(filter);
      } else {
        newFilters[index] = filter;
      }
      return {
        ...curr,
        filters: newFilters
      };
    });
  }, []);

  const handleDeleteFilter = useCallback((filterIndex) => {
    setSelectedFilters((curr) => {
      const newFilters = curr.filters.filter(
        (_, index) => index !== filterIndex
      );
      if (filterIndex === curr.filters.length) {
        setFilterDD(false);
      }
      return {
        ...curr,
        filters: newFilters
      };
    });
  }, []);

  const handleInsertSecondaryFilter = useCallback((filter, index) => {
    setSelectedFilters((curr) => {
      const newFilters = [...curr.secondaryFilters];
      if (newFilters.length === index) {
        newFilters.push(filter);
      } else {
        newFilters[index] = filter;
      }
      return {
        ...curr,
        secondaryFilters: newFilters
      };
    });
  }, []);

  const handleDeleteSecondaryFilter = useCallback((filterIndex) => {
    setSelectedFilters((curr) => {
      const newFilters = curr.secondaryFilters.filter(
        (_, index) => index !== filterIndex
      );
      if (filterIndex === curr.secondaryFilters.length - 1) {
        setFilterDD(false);
      }
      return {
        ...curr,
        secondaryFilters: newFilters
      };
    });
  }, []);

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

  const handleQueryChange = useCallback(
    (newEvent, index, changeType = 'add') => {
      setSelectedFilters((curr) => {
        const eventsList = cloneDeep(curr.eventsList);

        if (changeType === 'delete') {
          eventsList.splice(index, 1);
        } else if (eventsList[index]) {
          eventsList[index] = newEvent;
        } else {
          eventsList.push(newEvent);
        }

        return {
          ...curr,
          eventsList: eventsList.map((q) => ({
            ...q,
            key: q.key || generateRandomKey()
          }))
        };
      });
    },
    []
  );

  const { applyButtonDisabled, saveButtonDisabled } = useMemo(
    () =>
      checkFiltersEquality({
        appliedFilters,
        selectedFilters,
        newSegmentMode,
        isActiveSegment,
        areFiltersDirty
      }),
    [
      appliedFilters,
      selectedFilters,
      newSegmentMode,
      isActiveSegment,
      areFiltersDirty
    ]
  );

  const showClearAllButton = useMemo(
    () =>
      appliedFilters.filters.length > 0 || appliedFilters.eventsList.length > 0,
    [appliedFilters.eventsList.length, appliedFilters.filters.length]
  );

  return (
    <div
      className={cx(styles['filters-box-container'], 'flex flex-col gap-y-5')}
    >
      <div className='pt-4 gap-x-5 flex flex-col gap-y-2'>
        <div className='px-6 pb-1 border-b border-neutral-100'>
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
          <ControlledComponent controller={selectedFilters.filters.length > 0}>
            {selectedFilters.filters.map((filter, index) => (
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
                showInList
              />
            ))}
          </ControlledComponent>

          <ControlledComponent controller={filterDD === true}>
            <FilterWrapper
              viewMode={false}
              projectID={activeProject?.id}
              index={selectedFilters.filters.length}
              filterProps={mainFilterProps}
              minEntriesPerGroup={3}
              insertFilter={handleInsertFilter}
              closeFilter={handleCloseFilter}
              deleteFilter={handleDeleteFilter}
              showInList
            />
          </ControlledComponent>

          <Button
            className={cx(
              'flex items-center gap-x-2',
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
      <div className='flex flex-col gap-y-2'>
        <div className='px-6 pb-1 border-b border-neutral-100'>
          <Text
            type='title'
            color='character-secondary'
            extraClass='mb-0'
            weight='medium'
          >
            That matches these events
          </Text>
        </div>
        <div className='px-6 flex flex-col gap-y-2'>
          {selectedFilters.eventsList.map((event, index) => (
            <div key={index}>
              <EventsBlock
                isEngagementConfig={false}
                index={index + 1}
                event={event}
                queries={selectedFilters.eventsList}
                groupAnalysis={selectedFilters.account?.[1]}
                eventChange={handleQueryChange}
                initialDDState={false}
                showInList
              />
            </div>
          ))}
          <ControlledComponent
            controller={selectedFilters.eventsList.length < 10}
          >
            <EventsBlock
              initialDDState={false}
              index={selectedFilters.eventsList.length + 1}
              queries={selectedFilters.eventsList}
              groupAnalysis={selectedFilters.account?.[1]}
              eventChange={handleQueryChange}
              showInList
            />
          </ControlledComponent>
          <ControlledComponent controller={false}>
            <div className='flex gap-x-1 items-center'>
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
                    'flex gap-x-1 cursor-pointer items-center',
                    styles['event-timeline-picker']
                  )}
                >
                  <Text
                    type='title'
                    color='character-primary'
                    extraClass='mb-0'
                    weight='medium'
                  >
                    {eventTimelineMenuList[selectedFilters.eventTimeline].label}
                  </Text>
                </div>
              </Dropdown>
            </div>
          </ControlledComponent>
          <ControlledComponent
            controller={selectedFilters.eventsList.length > 1}
          >
            <div className='flex gap-x-1 items-center'>
              <Text
                type='title'
                extraClass='mb-0'
                color='character-primary'
                weight='medium'
              >
                {profileType === 'account'
                  ? 'Accounts that match conditions for'
                  : 'People who performed'}
              </Text>
              <Dropdown overlay={eventMenuItems}>
                <div className='flex gap-x-1 cursor-pointer items-center'>
                  <Text type='title' color='brand-color-6' extraClass='mb-0'>
                    {eventMenuList[selectedFilters.eventProp].label}
                  </Text>
                  <SVG name='caretDown' color='#1890ff' size={20} />
                </div>
              </Dropdown>
            </div>
          </ControlledComponent>
        </div>
      </div>
      <ControlledComponent controller={profileType === 'account'}>
        <div className='flex flex-col gap-y-2 gap-x-5'>
          <div className='px-6 pb-1 border-b border-neutral-100'>
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
            <ControlledComponent
              controller={selectedFilters.secondaryFilters.length > 0}
            >
              {selectedFilters.secondaryFilters.map((filter, index) => (
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
                  showInList
                />
              ))}
            </ControlledComponent>

            <ControlledComponent controller={secondaryFilterDD === true}>
              <FilterWrapper
                viewMode={false}
                projectID={activeProject?.id}
                index={selectedFilters.secondaryFilters.length}
                filterProps={userFilterProps}
                minEntriesPerGroup={3}
                insertFilter={handleInsertSecondaryFilter}
                closeFilter={handleCloseSecondaryFilter}
                deleteFilter={handleDeleteSecondaryFilter}
                showInList
              />
            </ControlledComponent>

            <Button
              className={cx(
                'flex items-center gap-x-2',
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
          'py-2 px-6 flex items-center justify-between',
          styles['buttons-container']
        )}
      >
        <div className='flex gap-x-2 items-center'>
          <Button
            disabled={applyButtonDisabled}
            onClick={applyFilters}
            type='primary'
          >
            Apply changes
          </Button>
          <Button disabled={disableDiscardButton} onClick={onCancel}>
            Discard changes
          </Button>
        </div>
        <ControlledComponent
          controller={showClearAllButton === true && newSegmentMode === false}
        >
          <Button
            className='flex items-center gap-x-1'
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
            className='flex items-center gap-x-1'
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
