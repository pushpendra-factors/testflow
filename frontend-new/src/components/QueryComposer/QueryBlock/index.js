import React, { useState, useEffect } from 'react';
import { SVG, Text } from 'factorsComponents';
import styles from './index.module.scss';
import { Button } from 'antd';
import { connect } from 'react-redux';
import { bindActionCreators } from 'redux';

import { setGroupBy } from '../../../reducers/coreQuery/middleware';

import FilterBlock from '../FilterBlock';

import GroupSelect from '../GroupSelect';
import EventGroupBlock from '../EventGroupBlock';

function QueryBlock({
  index, event, eventChange, queries, queryType, eventOptions,
  activeProject, groupBy, setGroupBy, userProperties, eventProperties
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
                  placeholder="Select Property"
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
    newEvent.filters.push(filter);
    eventChange(newEvent, index - 1);
  };

  const selectEventFilter = () => {
    if (isFilterDDVisible) {
      return <FilterBlock filterProps={filterProps} activeProject={activeProject} event={event} insertFilter={insertFilters} closeFilter={() => setFilterDDVisible(false)}></FilterBlock>;
    }
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
              <Button type="link" onClick={addGroupBy} className={'mr-1'}><SVG name="groupby"></SVG></Button>
               <Button type="link" onClick={addFilter} className={'mr-1'}><SVG name="filter"></SVG></Button>
               <Button type="link" onClick={deleteItem}><SVG name="trash"></SVG></Button>
            </div>
    );
  };

  const eventFilters = () => {
    const filters = [];
    if (event && event.filters.length) {
      event.filters.forEach((filter, index) => {
        filters.push(
                    <div key={index} className={'fa--query_block--filters'}>
                        <FilterBlock filter={filter} insertFilter={insertFilters} closeFilter={() => setFilterDDVisible(false)}></FilterBlock>
                    </div>
        );
      });
    }

    filters.push(<div key={'init'} className={'fa--query_block--filters'}>
            {additionalActions()}
            {selectEventFilter()}
        </div>);

    return filters;
  };

  const groupByItems = () => {
    const groupByEvents = [];
    if (groupBy && groupBy.length && groupBy[0].property) {
      groupBy.forEach((gbp, gbpIndex) => {
        groupByEvents.push(<div key={gbpIndex} className={'fa--query_block--filters'}>
          <EventGroupBlock
            index={gbpIndex}
            eventIndex={index}
            groupByEvent={gbp} event={event}
            setGroupState={pushGroupBy}
            closeDropDown={() => setGroupByDDVisible(false)}
            ></EventGroupBlock>
        </div>);
      });
    }
    groupByEvents.push(
      <div key={'init'} className={'fa--query_block--filters'}>
        {selectGroupByEvent()}
      </div>
    );
    return groupByEvents;
  };

  const ifQueries = queries.length > 0;
  if (!event) {
    return (
            <div className={`${styles.query_block} fa--query_block ${ifQueries ? 'bordered' : ''}`}>
                <div className={`${styles.query_block__event} flex justify-start items-center`}>
                    <div className={'fa--query_block--add-event flex justify-center items-center mr-2'}><SVG name={'plus'} color={'purple'}></SVG></div>
                        {!isDDVisible && <Button type="link" onClick={triggerDropDown}>{ifQueries ? 'Add another event' : 'Add First Event'}</Button> }
                    {selectEvents()}
                </div>
            </div>
    );
  }

  return (
        <div className={`${styles.query_block} fa--query_block bordered `}>
            <div className={`${styles.query_block__event} flex justify-start items-center`}>
                <div className={'fa--query_block--add-event active flex justify-center items-center mr-2'}><Text type={'title'} level={7} weight={'bold'} color={'white'} extraClass={'m-0'}>{queryType === 'funnel' ? index : alphabetIndex[index - 1]}</Text> </div>
                {!isDDVisible && <Button type="link" onClick={triggerDropDown}><SVG name="mouseevent" extraClass={'mr-1'}></SVG> {event.label} </Button> }
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
  setGroupBy
}, dispatch);

export default connect(mapStateToProps, mapDispatchToProps)(QueryBlock);
