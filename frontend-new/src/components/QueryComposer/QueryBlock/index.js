import React, {useState} from 'react';
import {SVG, Text} from 'factorsComponents';
import styles from './index.module.scss';
import { Select, Button} from 'antd';

import Filter from '../Filter';

const { OptGroup, Option } = Select;

function QueryBlock({index, event, eventChange,queries}) {

    const [isDDVisible, setDDVisible] = useState(index == 1 && !event ? true : false); 
    const [isFilterDDVisible, setFilterDDVisible] = useState(false);


    const eventOptions = [
        {
            label: "Frequently Asked",
            icon: "fav",
            values: [
                "Cart Updated",
                "Paid",
                "Add to WishList",
                "Checkout"
                
            ]
        }, {
            label: "Virtual Events",
            icon: "virtual",
            values: [
                "Download Song or Video",
                "Applied Coupon",
                "Cart Updated"
            ]
        }
    ]

    const onChange = (value) => {
        const newEvent = event? event: {label:'', filters: []};
        newEvent.label = value;
        setDDVisible(false);
        eventChange(newEvent, index-1);
    }

    const triggerDropDown = () => {
        setDDVisible(true);
    }

    const deleteItem = () => {
        eventChange(event, index-1, 'delete');
    }

    const selectEvents = () => {
        const selectDisplay = isDDVisible? 'block': 'none';

        return (
            <div className={styles.query_block__event_selector}>
                   {isDDVisible ? <Select showSearch 
                        style={{ width: 240}} 
                        onChange={onChange} defaultOpen={true}
                        showArrow={false}
                        onDropdownVisibleChange={() => setDDVisible(false)}
                        dropdownRender={menu => (
                            <div className={styles.query_block__selector_body}>
                              {menu}
                            </div>
                          )}
                    >
                            {eventOptions.map((group,index) => (
                                <OptGroup key={index} label={(
                                        <div className={styles.query_block__selector_group}>
                                            <SVG name={group.icon}></SVG>
                                            <span >{group.label}</span>
                                        </div>
                                    )}>
                                        {group.values.map((option,index) => (
                                            <Option key={index} value={option}></Option>
                                        ))}
                                </OptGroup>
                            ))}
                    </Select> : null }
                </div>
        )
    }

    const addFilter = () => {
        setFilterDDVisible(true);
    }

    const insertFilters = (filter) => {
        const newEvent = Object.assign({} ,event);
        newEvent.filters.push(filter);
        eventChange(newEvent, index-1);
    }

    const selectEventFilter = () => {
        if(isFilterDDVisible) {
            return <Filter insertFilters={insertFilters}></Filter>
        }
    }

    const additionalActions = () => {
        return(
            <div className={`fa--query_block--actions`}>
               <Button type="link" onClick={addFilter} className={`mr-1`}><SVG name="filter"></SVG></Button>
               <Button type="link" onClick={deleteItem}><SVG name="trash"></SVG></Button> 
            </div>
        )
    }

    const eventFilters = () => {
        const filters = [];
        if(event && event.filters.length) {
            event.filters.forEach((filter, index) => {
                filters.push(
                <div className={`fa--query_block--filters`}>
                    <span ><Text type={'title'} level={7} weight={'thin'} extraClass={`m-0`}>Where</Text> </span>
                    <div className={`fa--query_block--filters-values`}>
                        <span>{filter.prop}</span>
                        <span>{filter.operator}</span>
                        <span>{filter.values.join(',')}</span>
                    </div>
                    {index === event.filters.length -1? additionalActions() : null}
                </div>)
            });
        }

        filters.push(<div className={``}>
            {additionalActions()}
            {selectEventFilter()}
        </div>)

        return filters;
    } 
    const ifQueries = queries.length>0;
    if(!event) {
        return (
            <div className={`${styles.query_block} fa--query_block ${ifQueries?`bordered`:``}`}>
                <div className={`${styles.query_block__event} flex justify-start items-center`}> 
                    <div className={`fa--query_block--add-event flex justify-center items-center mr-2`}><SVG name={'plus'} color={`purple`}></SVG></div>
                        {!isDDVisible && <Button type="link" onClick={triggerDropDown}>{ifQueries ? `Add another event` : `Add First Event`}</Button> }
                    {selectEvents()} 
                </div> 
            </div>
        )
    }

    return(
        <div className={`${styles.query_block} fa--query_block bordered `}>
            <div className={`${styles.query_block__event} flex justify-start items-center`}> 
                <div className={`fa--query_block--add-event active flex justify-center items-center mr-2`}><Text type={'title'} level={7} weight={'bold'} color={`white`} extraClass={`m-0`}>{index}</Text> </div>
                {!isDDVisible && <Button type="link" onClick={triggerDropDown}><SVG name="mouseevent" extraClass={`mr-1`}></SVG> {event.label} </Button> } 
                {selectEvents()}
            </div>
            {eventFilters()}
        </div>
    )
}

export default QueryBlock;