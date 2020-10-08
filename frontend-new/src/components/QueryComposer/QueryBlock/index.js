import React, { useState } from 'react';
import { SVG, Text } from 'factorsComponents';
import styles from './index.module.scss';
import { Select, Button } from 'antd';
import { connect } from 'react-redux';

// import Filter from '../Filter';
import FilterBlock from '../FilterBlock';

const { OptGroup, Option } = Select;

function QueryBlock({
  index, event, eventChange, queries, queryType, eventOptions
}) {
  const [isDDVisible, setDDVisible] = useState(!!(index === 1 && !event));
  const [isFilterDDVisible, setFilterDDVisible] = useState(false);

  const alphabetIndex = 'ABCDEF';

  const onChange = (value) => {
    const newEvent = event || { label: '', filters: [] };
    newEvent.label = value;
    setDDVisible(false);
    eventChange(newEvent, index - 1);
  };

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
                   {isDDVisible ? <Select showSearch
                        style={{ width: 240 }}
                        onChange={onChange} defaultOpen={true}
                        showArrow={false}
                        onDropdownVisibleChange={() => setDDVisible(false)}
                        dropdownRender={menu => (
                            <div className={styles.query_block__selector_body}>
                              {menu}
                            </div>
                        )}
                    >
                            {eventOptions && eventOptions.map((group, index) => (
                                <OptGroup key={index} label={(
                                        <div className={styles.query_block__selector_group}>
                                            <SVG name={group.icon}></SVG>
                                            <span >{group.label}</span>
                                        </div>
                                    )}>
                                        {group.values.map((option, index) => (
                                            <Option key={index} value={option}></Option>
                                        ))}
                                </OptGroup>
                            ))}
                    </Select> : null }
                </div>
    );
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
      return <FilterBlock insertFilter={insertFilters} closeFilter={() => setFilterDDVisible(false)}></FilterBlock>;
    }
  };

  const additionalActions = () => {
    return (
            <div className={'fa--query_block--actions'}>
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
        </div>
  );
}

const mapStateToProps = (state) => ({
  eventOptions: state.coreQuery.eventOptions,
  activeProject: state.global.activeProject
});

// const mapDispatchToProps = dispatch => bindActionCreators({
//   fetchEvents,
// }, dispatch)

export default connect(mapStateToProps)(QueryBlock);
