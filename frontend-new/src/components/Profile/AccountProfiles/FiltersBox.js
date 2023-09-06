import React, { memo, useCallback, useEffect, useMemo, useState } from 'react';
import cx from 'classnames';
import styles from './index.module.scss';
import { SVG, Text } from 'Components/factorsComponents';
import { Button, Dropdown, Menu } from 'antd';
import FilterWrapper from 'Components/GlobalFilter/FilterWrapper';
import { useSelector } from 'react-redux';
import map from 'lodash/map';
import {
  checkFiltersEquality,
  computeFilterProperties
} from './accountProfiles.helpers';
import ControlledComponent from 'Components/ControlledComponent/ControlledComponent';
import { eventMenuList } from './accountProfiles.constants';
import EventsBlock from '../MyComponents/EventsBlock';
import { selectGroupsList } from 'Reducers/groups/selectors';
import { generateRandomKey } from 'Utils/global';

const FiltersBox = ({
  filtersList,
  profileType = 'account',
  source,
  appliedFilters,
  setFiltersList,
  applyFilters,
  onCancel,
  setSaveSegmentModal,
  listEvents,
  setListEvents,
  eventProp,
  setEventProp
}) => {
  const { newSegmentMode } = useSelector((state) => state.accountProfilesView);
  const groupsList = useSelector((state) => selectGroupsList(state));
  const activeProject = useSelector((state) => state.global.active_project);
  const [filterProps, setFilterProperties] = useState({});
  const [filterDD, setFilterDD] = useState(false);
  const [isEventsVisible, setEventsVisible] = useState(false);
  const userProperties = useSelector((state) => state.coreQuery.userProperties);
  const groupProperties = useSelector(
    (state) => state.coreQuery.groupProperties
  );

  const availableGroups = useSelector((state) => state.groups.data);

  const handleEventChange = useCallback(
    (eventItem) => {
      setEventProp(eventItem.key);
    },
    [setEventProp]
  );

  const eventMenuItems = (
    <Menu className={styles['dropdown-menu']}>
      {map(eventMenuList, (item) => {
        return (
          <Menu.Item
            className={styles['dropdown-menu-item']}
            onClick={() => handleEventChange(item)}
            key={item.key}
          >
            <Text type='title' extraClass='mb-0'>
              {item.label}
            </Text>
          </Menu.Item>
        );
      })}
    </Menu>
  );

  useEffect(() => {
    const props = computeFilterProperties({
      userProperties,
      groupProperties,
      availableGroups,
      profileType,
      source
    });
    setFilterProperties(props);
  }, [userProperties, groupProperties, availableGroups, profileType, source]);

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

  const showFilterDropdown = useCallback(() => {
    setFilterDD(true);
  }, []);

  const handleCloseFilter = useCallback(() => {
    setFilterDD(false);
  }, []);

  const showEventsDropdown = useCallback(() => {
    setEventsVisible(true);
  }, []);

  const closeEvent = useCallback(() => {
    setEventsVisible(false);
  }, []);

  const handleQueryChange = useCallback(
    (newEvent, index, changeType = 'add') => {
      const updatedQuery = [...listEvents];
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
        updatedQuery.map((q) => {
          return {
            ...q,
            key: q.key || generateRandomKey()
          };
        })
      );
    },
    [listEvents, setListEvents]
  );

  const { saveButtonDisabled, applyButtonDisabled } = useMemo(() => {
    return checkFiltersEquality({
      appliedFilters,
      filtersList,
      newSegmentMode,
      eventsList: listEvents,
      eventProp
    });
  }, [filtersList, appliedFilters, newSegmentMode, listEvents, eventProp]);

  const showEventsSection = filtersList.length > 0 && source !== 'All';

  return (
    <div className={cx(styles['filters-box-container'], 'flex flex-col')}>
      <div className='px-6 pt-4 pb-8 flex flex-col row-gap-3'>
        <Text
          type='title'
          color='character-title'
          extraClass='mb-0'
          weight='medium'
        >
          With Properties
        </Text>
        {filtersList.map((filter, index) => {
          return (
            <FilterWrapper
              key={index}
              groupName={source}
              viewMode={false}
              projectID={activeProject?.id}
              filter={filter}
              index={index}
              filterProps={filterProps}
              minEntriesPerGroup={3}
              insertFilter={handleInsertFilter}
              closeFilter={handleCloseFilter}
              deleteFilter={handleDeleteFilter}
            />
          );
        })}
        {filterDD === true ? (
          <FilterWrapper
            groupName={source}
            viewMode={false}
            projectID={activeProject?.id}
            index={filtersList.length}
            filterProps={filterProps}
            minEntriesPerGroup={3}
            insertFilter={handleInsertFilter}
            closeFilter={handleCloseFilter}
            deleteFilter={handleDeleteFilter}
          />
        ) : null}
        <Button
          className={cx('flex items-center', styles['add-filter-button'])}
          type='text'
          onClick={showFilterDropdown}
        >
          <SVG name='plus' color='#00000073' />
          <Text
            type='title'
            color='character-primary'
            extraClass='mb-0'
            weight='medium'
          >
            Add filter
          </Text>
        </Button>
      </div>
      <ControlledComponent controller={showEventsSection === true}>
        <>
          <div className={styles['and-tag']}>
            <div className={cx(styles['and-tag-box'], 'inline')}>
              <Text
                type='title'
                color='character-primary'
                extraClass='mb-0 inline'
              >
                AND
              </Text>
            </div>
          </div>
          <div className='pt-4 px-6 pb-8 flex flex-col row-gap-3'>
            <Text
              type='title'
              color='character-title'
              extraClass='mb-0'
              weight='medium'
            >
              Who Performed
            </Text>
            {listEvents.map((event, index) => {
              return (
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
                  />
                </div>
              );
            })}
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
                className={cx('flex items-center', styles['add-filter-button'])}
                type='text'
                onClick={showEventsDropdown}
              >
                <SVG name='plus' color='#00000073' />
                <Text
                  type='title'
                  color='character-primary'
                  extraClass='mb-0'
                  weight='medium'
                >
                  Add event
                </Text>
              </Button>
            </ControlledComponent>
            <ControlledComponent controller={listEvents.length > 1}>
              <div className='flex col-gap-1 items-center'>
                <Text
                  type='title'
                  extraClass='mb-0'
                  color='character-primary'
                  weight='medium'
                >
                  Consider users who performed
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
        </>
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
            Apply
          </Button>
          <Button type='secondary' onClick={onCancel}>
            Cancel
          </Button>
        </div>
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
      </div>
    </div>
  );
};

export default memo(FiltersBox);
