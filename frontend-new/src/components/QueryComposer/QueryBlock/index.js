/* eslint-disable */
import React, { useState, useEffect } from 'react';
import { SVG, Text } from '../../factorsComponents';
import styles from './index.module.scss';
import { Button } from 'antd';
import { connect } from 'react-redux';
import { bindActionCreators } from 'redux';

import { setGroupBy, delGroupBy } from '../../../reducers/coreQuery/middleware';

import FilterBlock from '../FilterBlock';

import GroupSelect from '../GroupSelect';
import EventGroupBlock from '../EventGroupBlock';
import { QUERY_TYPE_FUNNEL } from '../../../utils/constants';

function QueryBlock({
  index, event, eventChange, queries, queryType, eventOptions,
  activeProject, groupBy, setGroupBy,
  delGroupBy, userProperties, eventProperties
}) {
  const [isDDVisible, setDDVisible] = useState(!!(index === 1 && !event));
  const [isFilterDDVisible, setFilterDDVisible] = useState(false);
  const [isGroupByDDVisible, setGroupByDDVisible] = useState(false);
  const [filterProps, setFilterProperties] = useState({
    user: [],
    event: []
  });

  const alphabetIndex = 'ABCDEF';

  const onChange = (value) => {
    const newEvent = event || { label: '', filters: [] };
    newEvent.label = value;
    setDDVisible(false);
    eventChange(newEvent, index - 1);
  };

  useEffect(() => {
    if (!event || event === undefined) { return undefined; }; // Akhil please check this line
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
    // const selectDisplay = isDDVisible ? 'block' : 'none';

    return (
            <div className={styles.query_block__event_selector}>
                   {isDDVisible
                     ? <GroupSelect
                  groupedProperties={eventOptions}
                  placeholder="Select Event"
                  optionClick={(group, val) => onChange(val[0])}
                  onClickOutside={() => setDDVisible(false)}

                  ></GroupSelect>

                     : null }
                </div>
    );
  };

  const addGroupBy = () => {
    setGroupByDDVisible(true);
  };

  const addFilter = () => {
    setFilterDDVisible(true);
  };

  const insertFilters = (filter) => {
    const newEvent = Object.assign({}, event);
    const filt = newEvent.filters.filter(fil => JSON.stringify(fil) === JSON.stringify(filter));
    if (filt && filt.length) return;
    newEvent.filters.push(filter);
    eventChange(newEvent, index - 1);
  };

  const removeFilters = (i) => {
    const newEvent = Object.assign({}, event);
    if (newEvent.filters[i]) {
      newEvent.filters.splice(i, 1);
    }
    eventChange(newEvent, index-1);
  };

  const selectEventFilter = () => {
    if (isFilterDDVisible) {
      return <FilterBlock
      filterProps={filterProps}
      activeProject={activeProject}
      event={event}
      insertFilter={insertFilters}
      closeFilter={() => setFilterDDVisible(false)}
      >

      </FilterBlock>;
    }
  };

  const deleteGroupBy = (groupState, id, type = 'event') => {
    delGroupBy(type, groupState, id);
  };

  const pushGroupBy = (groupState, index) => {
    const ind = index || groupBy.length;
    setGroupBy('event', groupState, ind);
  };

  const selectGroupByEvent = () => {
    if (isGroupByDDVisible) {
      return <EventGroupBlock
        eventIndex={index}
        event={event}
        setGroupState={pushGroupBy}
        closeDropDown={() => setGroupByDDVisible(false)}
      ></EventGroupBlock>;
    }
  };

  const additionalActions = () => {
    return (
            <div className={'fa--query_block--actions'}>
              <Button size={'large'} type="text" onClick={addGroupBy} className={'mr-1'}><SVG name="groupby"></SVG></Button>
               <Button size={'large'} type="text" onClick={addFilter} className={'mr-1'}><SVG name="filter"></SVG></Button>
               <Button size={'large'} type="text" onClick={deleteItem}><SVG name="trash"></SVG></Button>
            </div>
    );
  };

  const eventFilters = () => {
    const filters = [];
    if (event && event.filters.length) {
      event.filters.forEach((filter, index) => {
        filters.push(
                    <div key={index} className={'fa--query_block--filters'}>
                        <FilterBlock index={index} filter={filter} deleteFilter={removeFilters} insertFilter={insertFilters} closeFilter={() => setFilterDDVisible(false)}></FilterBlock>
                    </div>
        );
      });
    }

    if (isFilterDDVisible) {
      filters.push(<div key={'init'} className={'fa--query_block--filters'}>
            {selectEventFilter()}
        </div>);
    }

    return filters;
  };

  const groupByItems = () => {
    const groupByEvents = [];
    if (groupBy && groupBy.length && groupBy[0] && groupBy[0].property) {
      groupBy.filter(gbp => gbp.eventName === event.label && gbp.eventIndex === index).forEach((gbp, gbpIndex) => {
        groupByEvents.push(<div key={gbpIndex} className={'fa--query_block--filters'}>
          <EventGroupBlock
            index={gbpIndex}
            eventIndex={index}
            groupByEvent={gbp} event={event}
            delGroupState={(ev) => deleteGroupBy(ev, gbpIndex)}
            setGroupState={pushGroupBy}
            closeDropDown={() => setGroupByDDVisible(false)}
            ></EventGroupBlock>
        </div>);
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
            <div className={`${styles.query_block} fa--query_block ${ifQueries ? 'bordered' : ''}`}>
                <div className={`${styles.query_block__event} flex justify-start items-center`}>
                    <div className={'fa--query_block--add-event flex justify-center items-center mr-2'}><SVG name={'plus'} color={'purple'}></SVG></div>
                        {!isDDVisible && <Button size={'large'} type="link" onClick={triggerDropDown}>{ifQueries ? 'Add another event' : 'Add First Event'}</Button> }
                    {selectEvents()}
                </div>
            </div>
    );
  }

  return (
        <div className={`${styles.query_block} fa--query_block bordered `}>
            <div className={`${styles.query_block__event} flex justify-start items-center`}>
                <div className={'fa--query_block--add-event active flex justify-center items-center mr-2'}><Text type={'title'} level={7} weight={'bold'} color={'white'} extraClass={'m-0'}>{queryType === QUERY_TYPE_FUNNEL ? index : alphabetIndex[index - 1]}</Text> </div>
                {!isDDVisible && <Button size={'large'} type="link" onClick={triggerDropDown}><SVG name="mouseevent" extraClass={'mr-1'}></SVG> {event.label} </Button> }
                {additionalActions()}
                {selectEvents()}
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
  groupBy: state.coreQuery.groupBy.event
});

const mapDispatchToProps = dispatch => bindActionCreators({
  setGroupBy,
  delGroupBy
}, dispatch);

export default connect(mapStateToProps, mapDispatchToProps)(QueryBlock);
