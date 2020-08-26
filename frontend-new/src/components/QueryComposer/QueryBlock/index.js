import React, {useState} from 'react';
import {SVG} from 'factorsComponents';
import styles from './index.module.scss';
import { Select, Button} from 'antd';

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
                    <Select showSearch style={{ width: 240, display: selectDisplay}} onChange={onChange} open={isDDVisible} showArrow={false}
                    

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
                    </Select>
                </div>
        )
    }

    const addFilter = () => {
        setFilterDDVisible(true);
    }

    const selectFilters = () => {
        // return(
        //     <div>
        //         <Select>

        //         </Select>
        //     </div>
        // )
    }

    // const additionalActions = () => {
    //     if(!event) {return null};

    //     return(
    //         <div className={styles.query_block__actions}>
    //            <Button type="text" onClick={addFilter}>filter</Button>
    //            <Button type="text" onClick={deleteItem}>delete</Button> 
    //         </div>
    //     )
    // }

    const eventFilters = () => {
        const filters = [];
        if(event.filters.length) {
            event.filters.forEach(filter => {
                filters.push(<div className={styles.query_block__filters}>

                </div>)
            });
        }
        return filters;
    }

    if(!event) {
        return (
            <div className={styles.query_block}>
                <div className={styles.query_block__event}>
                    
                        <div className={styles.query_block__add_event}><SVG name={'plus'}></SVG> </div>

                        <div className={styles.query_block__placehoder} onClick={triggerDropDown}> Add Event</div>
                        {selectEvents()}
                    
                </div>
                
            </div>
        )
    }

    return(
        <div className={styles.query_block}>
            <div className={styles.query_block__event}>
                <span className={styles.query_block__index}>{index}</span>
                <span className={styles.query_block__event_tag} onClick={triggerDropDown}> <SVG name="mouseevent"></SVG> {event.label} </span>
                {/* {additionalActions()} */}
                {selectEvents()}
            </div>
            {eventFilters()}
        </div>
    )
}

export default QueryBlock;