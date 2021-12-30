/* eslint-disable */
import React, { useState, useEffect } from 'react';
import { SVG, Text } from '../../factorsComponents';
import styles from './index.module.scss';
import { Button, Tooltip } from 'antd';
import { connect } from 'react-redux';
import { bindActionCreators } from 'redux';

import { setGroupBy, delGroupBy } from '../../../reducers/coreQuery/middleware';

import FilterBlock from '../FilterBlock';
import EventFilterWrapper from '../EventFilterWrapper';

import GroupSelect2 from '../GroupSelect2';
import EventGroupBlock from '../EventGroupBlock';
import { QUERY_TYPE_FUNNEL } from '../../../utils/constants';

import FaSelect from 'Components/FaSelect';
import AliasModal from '../AliasModal';

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
}) {
  const [isDDVisible, setDDVisible] = useState(false);
  const [isFilterDDVisible, setFilterDDVisible] = useState(false);
  const [isGroupByDDVisible, setGroupByDDVisible] = useState(false);
  const [moreOptions, setMoreOptions] = useState(false);
  const [filterProps, setFilterProperties] = useState({
    event: [],
    user: [],
  });

  const [isModalVisible, setIsModalVisible] = useState(false);

  const showModal = () => {
    setIsModalVisible(true);
  };

  const handleCancel = () => {
    setIsModalVisible(false);
  };
  const handleOk = (alias) => {
    const newEvent = Object.assign({}, event);
    newEvent.alias = alias;
    setIsModalVisible(false);
    eventChange(newEvent, index - 1, 'filters_updated');
  };

  const alphabetIndex = 'ABCDEF';

  const onChange = (value) => {
    const newEvent = { alias: '', label: '', filters: [] };
    newEvent.label = value;
    setDDVisible(false);
    eventChange(newEvent, index - 1);
  };

  useEffect(() => {
    if (!event || event === undefined) {
      return undefined;
    } // Akhil please check this line
    const assignFilterProps = Object.assign({}, filterProps);

    if (eventProperties[event.label]) {
      assignFilterProps.event = eventProperties[event.label];
    }
    assignFilterProps.user = userProperties;
    setFilterProperties(assignFilterProps);
  }, [userProperties, eventProperties]);

  const triggerDropDown = () => {
    setDDVisible(true);
  };

  const deleteItem = () => {
    eventChange(event, index - 1, 'delete');
  };

  const selectEvents = () => {
    return (
      <div className={styles.query_block__event_selector}>
        {isDDVisible ? (
          <div className={styles.query_block__event_selector__btn}>
            <GroupSelect2
              groupedProperties={eventOptions}
              placeholder='Select Event'
              optionClick={(group, val) => onChange(val[1]? val[1]: val[0])}
              onClickOutside={() => setDDVisible(false)}
              allowEmpty={true}
            ></GroupSelect2>
          </div>
        ) : null}
      </div>
    );
  };

  const addGroupBy = () => {
    setGroupByDDVisible(true);
  };

  const addFilter = () => {
    setFilterDDVisible(true);
  };

  const insertFilters = (filter, filterIndex) => {
    const newEvent = Object.assign({}, event);
    if(filterIndex >= 0) {
      newEvent.filters = newEvent.filters.map((filt, i) => {
        if(i === filterIndex) {
          return filter;
        } 
        return filt;
      })
    } else {
      newEvent.filters.push(filter);
    }
    
    eventChange(newEvent, index - 1, 'filters_updated');
  };

  const removeFilters = (i) => {
    const newEvent = Object.assign({}, event);
    if (newEvent.filters[i]) {
      newEvent.filters.splice(i, 1);
    }
    eventChange(newEvent, index - 1, 'filters_updated');
  };

  const selectEventFilter = () => {
      return (
        <EventFilterWrapper
          filterProps={filterProps}
          activeProject={activeProject}
          event={event}
          deleteFilter={() => setFilterDDVisible(false)}
          insertFilter={insertFilters}
          closeFilter={() => setFilterDDVisible(false)}
        ></EventFilterWrapper>
      );
  };

  const deleteGroupBy = (groupState, id, type = 'event') => {
    delGroupBy(type, groupState, id);
  };

  const pushGroupBy = (groupState, index) => {
    let ind;
    index >= 0 ? (ind = index) : (ind = groupBy.length);
    setGroupBy('event', groupState, ind);
  };

  const selectGroupByEvent = () => {
    if (isGroupByDDVisible) {
      return (
        <EventGroupBlock
          eventIndex={index}
          event={event}
          setGroupState={pushGroupBy}
          closeDropDown={() => setGroupByDDVisible(false)}
        ></EventGroupBlock>
      );
    }
  };

  const setAdditionalactions = (opt) => {
    if(opt[1] === 'filter') {
      addFilter();
    } else if(opt[1] === 'groupby') {
      addGroupBy();
    } else { 
      showModal();
    }
    setMoreOptions(false);
  }

  const additionalActions = () => {
    return (
      <div className={`fa--query_block--actions-cols flex`}>
        <div className={`relative`}>
          <Button
            type='text'
            onClick={() => setMoreOptions(true)}
            className={`fa-btn--custom ml-1 mr-1`}
          >
            <SVG name='more'></SVG>
          </Button>

          {moreOptions ? <FaSelect
            options={[['Filter By', 'filter'], ['Breakdown', 'groupby'], [ !event?.alias?.length ? 'Create Alias' : 'Edit Alias', 'edit']]}
            optionClick={(val) => setAdditionalactions(val)}
            onClickOutside={() => setMoreOptions(false)}
          ></FaSelect> : false}

          <AliasModal
            visible={isModalVisible}
            event={eventNames[event.label]? eventNames[event.label] : event.label}
            onOk={handleOk}
            onCancel={handleCancel}
            alias={event.alias}
          >
          </AliasModal>
    
        </div>
        <Button type='text' onClick={deleteItem} className={`fa-btn--custom`}>
          <SVG name='trash'></SVG>
        </Button>
      </div>
    );
  };

  const eventFilters = () => {
    const filters = [];
    if (event && event?.filters?.length) {
      event.filters.forEach((filter, index) => {
        filters.push(
          <div key={index} className={'fa--query_block--filters'}>
            <EventFilterWrapper
              index={index}
              filter={filter}
              event={event}
              filterProps={filterProps}
              activeProject={activeProject}
              deleteFilter={removeFilters}
              insertFilter={insertFilters}
              closeFilter={() => setFilterDDVisible(false)}
            ></EventFilterWrapper>
          </div>
        );
      });
    }

    if (isFilterDDVisible) {
      filters.push(
        <div key={'init'} className={'fa--query_block--filters'}>
          {selectEventFilter()}
        </div>
      );
    }

    return filters;
  };

  const groupByItems = () => {
    const groupByEvents = [];
    if (groupBy && groupBy.length && groupBy[0] && groupBy[0].property) {
      groupBy
        .map((gbp, index) => {
          return { ...gbp, groupByIndex: index };
        })
        .filter((gbp) => {
          return gbp.eventName === event.label && gbp.eventIndex === index;
        })
        .forEach((gbp, gbpIndex) => {
          const { groupByIndex, ...orgGbp } = gbp;
          groupByEvents.push(
            <div key={gbpIndex} className={'fa--query_block--filters'}>
              <EventGroupBlock
                index={gbp.groupByIndex}
                grpIndex={gbpIndex}
                eventIndex={index}
                groupByEvent={orgGbp}
                event={event}
                delGroupState={(ev) => deleteGroupBy(ev, gbpIndex)}
                setGroupState={pushGroupBy}
                closeDropDown={() => setGroupByDDVisible(false)}
              ></EventGroupBlock>
            </div>
          );
        });
    }

    if (isGroupByDDVisible) {
      groupByEvents.push(
        <div key={'init'} className={'fa--query_block--filters'}>
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
        className={`${styles.query_block} fa--query_block my-2 ${
          ifQueries ? 'borderless no-padding' : 'borderless no-padding'
        }`}
      >
        <div
          className={`${styles.query_block__event} flex justify-start items-center`}
        >
          { 
              <Button
                type='text'
                onClick={triggerDropDown}
                icon={<SVG name={'plus'} color={'grey'} />}
              >
                {ifQueries ? 'Add another event' : 'Add First Event'}
              </Button>
            }
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
        className={`${!event?.alias?.length ? 'flex justify-start' : ''} ${styles.query_block__event} block_section items-center`}
      >
        <div className={'flex items-center'}>
          <div
            className={
              'fa--query_block--add-event active flex justify-center items-center mr-2'
            }
          >
            <Text
              type={'title'}
              level={7}
              weight={'bold'}
              color={'white'}
              extraClass={'m-0'}
            >
              {queryType === QUERY_TYPE_FUNNEL ? index : alphabetIndex[index - 1]}
            </Text>{' '}
          </div>
          {event?.alias?.length
            ? (
              <Text
                type={'title'}
                level={7}
                weight={'bold'}
                extraClass={'m-0'}
              >
                {event?.alias}
                <Tooltip title={'Edit Alias'}>
                  <Button
                    className={`${styles.custombtn} mx-1`} type="text" onClick={showModal} ><SVG size={20} name="edit" color={'grey'} />
                  </Button>
                </Tooltip>
              </Text>)
            : null}
        </div>
        <div className={`flex ${!event?.alias?.length ? '' : 'ml-8 mt-2'}`}>
          <div className="max-w-7xl">
            <Tooltip title={eventNames[event.label] ? eventNames[event.label] : event.label}>
              <Button
                icon={<SVG name='mouseevent' size={16} color={'purple'} />}
                className={`fa-button--truncate fa-button--truncate-lg`}
                type='link'
                onClick={triggerDropDown}
              >
                {' '}
                {eventNames[event.label] ? eventNames[event.label] : event.label}{' '}
              </Button>
              {selectEvents()}
            </Tooltip>
          </div>
          <div className={styles.query_block__additional_actions}>{additionalActions()}</div>
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
  eventProperties: state.coreQuery.eventProperties,
  groupBy: state.coreQuery.groupBy.event,
  eventNames: state.coreQuery.eventNames
});

const mapDispatchToProps = (dispatch) =>
  bindActionCreators(
    {
      setGroupBy,
      delGroupBy,
    },
    dispatch
  );

export default connect(mapStateToProps, mapDispatchToProps)(QueryBlock);
