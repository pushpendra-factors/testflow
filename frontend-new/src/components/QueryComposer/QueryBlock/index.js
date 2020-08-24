import React, {useState} from 'react';
import {SVG} from 'factorsComponents';
import styles from './index.module.scss';
import { Select} from 'antd';

const { OptGroup, Option } = Select;

function QueryBlock({index, event, eventChange}) {

    const [isDDVisible, setDDVisible] = useState(index == 1 && !event ? true : false); 

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

    if(!event) {
        return (
            <div className={styles.query_block}>
                <div className={styles.query_block__add_event}><SVG name={'plus'}></SVG> </div>

                <div className={styles.query_block__placehoder} onClick={triggerDropDown}> Add Event</div>
                {selectEvents()}
                
            </div>
        )
    }

    return(
        <div className={styles.query_block}>
            <span className={styles.query_block__index}>{index}</span>
            <span className={styles.query_block__event_tag} onClick={triggerDropDown}> <SVG name="mouseevent"></SVG> {event.label} </span>
            {selectEvents()}
        </div>
    )
}

export default QueryBlock;