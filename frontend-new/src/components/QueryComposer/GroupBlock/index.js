import React, {useState} from 'react';
import styles from './index.module.scss';
import {SVG} from 'factorsComponents';

import {Select} from 'antd';
import { group } from 'd3';
import { queries } from '@testing-library/react';

const {OptGroup, Option} = Select;

export default function GroupBlock({events, groupBy}){

    const [isDDVisible, setDDVisible] = useState(false);
    const [isValueDDVisible, setValueDDVisible] = useState(false);
    const [groupByState, setGroupByState] = useState(groupBy);

    const filterOptions = [
        {
            label: "User Properties",
            icon: "fav",
            values: [
                "Country",
                "City",
                "Age",
            ]
        }, {
            label: "Event Properties",
            icon: "virtual",
            values: [
                "Add to WishList",
                "Applied Coupon",
                "Cart Updated"
            ]
        }
    ]


    const onChange = (value) => {
        const newGroupByState = Object.assign({}, groupByState);
        newGroupByState.property = value;
        setGroupByState(newGroupByState);
        setDDVisible(false);
    }

    const onEventValueChange = (value) => {
        const newGroupByState = Object.assign({}, groupByState);
        newGroupByState.eventValue = value;
        setGroupByState(newGroupByState);
        setValueDDVisible(false);
    }

    const triggerDropDown = () => {
        setDDVisible(true);
    }

    const triggerValueDropDown = () => {
        setValueDDVisible(true);
    }

    const renderGroupedEventOptions = () => {
        return filterOptions.map(group => (
            <OptGroup label={(
                    <div className={styles.group_block__selector_group}>
                        <SVG name={group.icon}></SVG>
                        <span >{group.label}</span>
                    </div>
                )}>
                    {group.values.map((option) => (
                        <Option value={option}></Option>
                    ))}
            </OptGroup>
        ))
    }

    return (
        <div className={styles.group_block}>
            <span className={styles.group_block__group_icon}>
                <SVG name="play"></SVG>
            </span>

            <div className={styles.group_block__property} onClick={triggerDropDown}>
                <span>Group By</span>
                <div className={styles.group_block__property__selection} >
                    {
                    groupByState.property? 
                        <span className={styles.group_block__event_tag} 
                        onClick={triggerDropDown}> <SVG name="plus"></SVG> {groupByState.property} </span>
                        : null
                    }
                    {isDDVisible ?
                    (<><Select 
                        onClick={triggerDropDown}
                        onChange={onChange} open={isDDVisible} 
                        dropdownRender={menu => (
                            <div className={styles.group_block__selector_body}>
                              {menu}
                            </div>
                          )} style={{width: 200}} showArrow={false} showSearch>
                        {renderGroupedEventOptions()}
                    </Select>
                      

                    <span> with values</span>

                    {
                    groupByState.eventValue? 
                        <span className={styles.group_block__event_tag} 
                        onClick={triggerValueDropDown}> {groupByState.eventValue} </span>
                        : <span className={styles.group_block__event_tag} 
                        onClick={triggerValueDropDown}> {events[0].label} </span>
                    }
                    <Select style={{width: 200}} showArrow={false} 
                    onClick={triggerValueDropDown}
                        onChange={onEventValueChange} open={isValueDDVisible} >
                        {events.map(event => (
                            <Option value={event.label}></Option>
                        ))}
                    </Select> </>)

                    : null
                } 
                </div>

            </div>

            
        </div>
    );
};