import React, {useState} from 'react';
import {SVG, Text} from 'factorsComponents';
import styles from './index.module.scss';
import { Select, Button} from 'antd';

import Filter from '../Filter';

const { OptGroup, Option } = Select;

function QueryBlock({index, event, eventChange}) {

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
                            {eventOptions.map(group => (
                                <OptGroup label={(
                                        <div className={styles.query_block__selector_group}>
                                            <SVG name={group.icon}></SVG>
                                            <span >{group.label}</span>
                                        </div>
                                    )}>
                                        {group.values.map((option) => (
                                            <Option value={option}></Option>
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

    const selectEventFilter = () => {
        if(isFilterDDVisible) {
            return <Filter></Filter>
        }
    }

    const additionalActions = () => {
        return(
            <div className={styles.query_block__actions}>
               <Button type="text" onClick={addFilter}><SVG name="filter"></SVG></Button>
               <Button type="text" onClick={deleteItem}><SVG name="trash"></SVG></Button> 
            </div>
        )
    }

    const eventFilters = () => {
        const filters = [];
        if(event && event.filters.length) {
            event.filters.forEach((filter, index) => {
                filters.push(
                <div className={styles.query_block__filters}>
                    <span className={styles.query_block__filters__label}>Where</span>
                    <div className={styles.query_block__filter_query}>
                        <span>{filter.prop}</span>
                        <span>{filter.operator}</span>
                        <span>{filter.values.join(',')}</span>
                    </div>
                    {index === event.filters.length -1? additionalActions() : null}
                </div>)
            });
        }

        filters.push(<div className={styles.query_block__filters}>
            {additionalActions()}
            {selectEventFilter()}
        </div>)

        return filters;
    }

    if(!event) {
        return (
            <div className={`${styles.query_block} fa--query_block `}>
                <div className={`${styles.query_block__event} flex justify-start items-center`}> 
                        <div className={`fa--query_block--add-event flex justify-center items-center mr-2`}><SVG name={'plus'} color={`purple`}></SVG></div>
                        {!isDDVisible && <Button type="link" onClick={triggerDropDown}>Add First Event</Button> }
                        {selectEvents()} 
                </div>
                
            </div>
        )
    }

    return(
        <div className={`${styles.query_block} fa--query_block `}>
            <div className={`${styles.query_block__event} flex justify-start items-center`}> 
    <div className={`fa--query_block--add-event active flex justify-center items-center mr-2`}><Text type={'title'} level={7} weight={'bold'} color={`white`} extraClass={`m-0`}>{index}</Text> </div>
                {!isDDVisible && <Button type="link" onClick={triggerDropDown}><SVG name="mouseevent"></SVG> {event.label} </Button> } 
                {selectEvents()}
            </div>
            {eventFilters()}
        </div>
    )
}

export default QueryBlock;