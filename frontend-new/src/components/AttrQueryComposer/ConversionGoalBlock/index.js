import React, {useState, useEffect} from 'react';
import styles from './index.module.scss';
import { connect } from 'react-redux';
import { bindActionCreators } from 'redux';

import GroupSelect2 from '../../QueryComposer/GroupSelect2';
import EventFilterWrapper from '../../QueryComposer/EventFilterWrapper';

import { Button, Tooltip } from 'antd';
import { SVG, Text } from 'factorsComponents';
import { isArray } from 'lodash';
import FaSelect from 'Components/FaSelect';

const ConversionGoalBlock = ({
    eventGoal, 
    eventGoalChange, 
    delEvent, 
    eventNameOptions, 
    eventNames,
    activeProject, 
    eventProperties,
    userProperties
}) => {

    const [selectVisible, setSelectVisible] = useState(false);
    const [filterBlockVisible, setFilterBlockVisible] = useState(false);

    const [moreOptions, setMoreOptions] = useState(false);

    const [filterProps, setFilterProperties] = useState({
        event: [],
        user: []
    });
    
    useEffect(() => {
        if(!eventGoal || !eventGoal?.label?.length) {return};
        const assignFilterProps = Object.assign({}, filterProps);
    
        if (eventProperties[eventGoal.label]) {
          assignFilterProps.event = eventProperties[eventGoal.label];
        }
        assignFilterProps.user = userProperties;
        setFilterProperties(assignFilterProps);
    }, [userProperties, eventProperties]);

    const toggleEventSelect = () => {
        setSelectVisible(!selectVisible);
    }

    const addFilter = (val) => {
        const updatedEvent = Object.assign({}, eventGoal);
        const filt = updatedEvent.filters.filter(fil => JSON.stringify(fil) === JSON.stringify(val));
        if (filt && filt.length) return;
        updatedEvent.filters.push(val);
        eventGoalChange(updatedEvent);
    };

    const editFiler = (index, val) => {
        const updatedEvent = Object.assign({}, eventGoal);
        const filt = Object.assign({}, val);
        filt.operator = isArray(val.operator) ? val.operator[0] : val.operator;
        updatedEvent.filters[index] = filt;
        eventGoalChange(updatedEvent);
    }

    const delFilter = (val) => {
        const updatedEvent = Object.assign({}, eventGoal);
        const filt = updatedEvent.filters.filter((v, i) => i !== val);
        updatedEvent.filters = filt;
        eventGoalChange(updatedEvent);
    };

    const closeFilter = () => {
        setFilterBlockVisible(false);
    };

    const deleteItem = () => {
        delEvent();
        closeFilter();
    };

    const addFilterBlock = () => {
        setFilterBlockVisible(true);
    }

    const selectEventFilter = () => {
          return <EventFilterWrapper
          filterProps={filterProps}
          activeProject={activeProject}
          event={eventGoal}
          deleteFilter={() => closeFilter()}
          insertFilter={addFilter}
          closeFilter={closeFilter}
          >
          </EventFilterWrapper>;
    };
    

    const eventFilters = () => {
        const filters = [];
        if (eventGoal && eventGoal?.filters?.length) {
            eventGoal.filters.forEach((filter, index) => {
                let filterContent = filter;
                filterContent.values = filter.props[1] === 'datetime' && isArray(filter.values)? filter.values[0] : filter.values;
                filters.push(
                            <div key={index} className={'fa--query_block--filters'}>
                                <EventFilterWrapper index={index} 
                                    filter={filter} 
                                    filterProps={filterProps} 
                                    activeProject={activeProject} 
                                    deleteFilter={delFilter} 
                                    insertFilter={(val) => editFiler(index, val)} 
                                    closeFilter={closeFilter}
                                    event={eventGoal}
                                ></EventFilterWrapper>
                            </div>
                );
          });
        }
    
        if (filterBlockVisible) {
          filters.push(<div key={'init'} className={'fa--query_block--filters'}>
                {selectEventFilter()}
            </div>);
        }
    
        return filters;
      };

    const onEventSelect = (val) => {
        const currentEventGoal = Object.assign({}, eventGoal);
        currentEventGoal.label = val;
        currentEventGoal.filters = [];
        eventGoalChange(currentEventGoal);
        setSelectVisible(false);
        closeFilter();
    };

    const additionalActions = () => {
        return (
                <div className={'fa--query_block--actions-cols flex relative ml-2'}>
                <div className={`relative flex`}>
                    <Button
                        type='text'
                        onClick={() => setMoreOptions(true)}
                        className={'fa-btn--custom mr-1'}
                    >
                        <SVG name='more'></SVG>
                    </Button>

                    {moreOptions ? <FaSelect
                        options={[[`Filter By`, 'filter']]}
                        optionClick={(val) => {addFilterBlock(); setMoreOptions(false)}}
                        onClickOutside={() => setMoreOptions(false)}
                    ></FaSelect> : false}
                </div>
                <Button className={'fa-btn--custom'} type="text" onClick={deleteItem}><SVG name="trash"></SVG></Button>
                </div>
        );
    };

    const selectEvents = () => {
    
        return (
            <div className={styles.block__event_selector}>
                   {selectVisible
                     ? <GroupSelect2
                            groupedProperties={eventNameOptions}
                            placeholder="Select Event"
                            optionClick={(group, val) => onEventSelect(val[1]? val[1]: val[0])}
                            onClickOutside={() => setSelectVisible(false)}
                        ></GroupSelect2>
                     : null }
            </div>
        );
    };

    const renderGoalBlockContent = () => {
        return (
            <div className={`${styles.block__content} flex items-center relative mt-1`}>
                {<Tooltip title={eventNames[eventGoal?.label]? eventNames[eventGoal?.label] : eventGoal?.label}>
                <Button 
                    type="link" 
                    onClick={toggleEventSelect}
                    icon={<SVG name="mouseevent" />}
                    className={`fa-button--truncate fa-button--truncate-lg`}
                    >
                        {eventNames[eventGoal?.label]? eventNames[eventGoal?.label] : eventGoal?.label}
                </Button> 
                </Tooltip> }

                {selectEvents()}

                <Text type={'title'} level={7} weight={'regular'} color={'grey'} extraClass={'m-0 ml-2'}>as count of unique users</Text>

                <div className={styles.block__additional_actions}>{additionalActions()}</div>
            </div>
        )
    };

    const renderGoalSelect = () => {
        return (
            <div className={'flex justify-start items-center pt-3 mt-1'}>
                {<Button type="text" onClick={toggleEventSelect} icon={<SVG name={'plus'} color={'grey'} />}>Add a goal event</Button>}
                {selectEvents()}
            </div>
        );
    };

    return (
        <div className={`${styles.block} fa--query_block_section--basic relative`}>
            {eventGoal?.label?.length ? renderGoalBlockContent() : renderGoalSelect()}
            {eventFilters()}
        </div>
    )
}

const mapStateToProps = (state) => ({
    activeProject: state.global.active_project,
    eventProperties: state.coreQuery.eventProperties,
    userProperties: state.coreQuery.userProperties,
    eventNameOptions: state.coreQuery.eventOptions,
    eventNames: state.coreQuery.eventNames
});
  
const mapDispatchToProps = dispatch => bindActionCreators({}, dispatch);


export default connect(mapStateToProps, mapDispatchToProps)(ConversionGoalBlock);