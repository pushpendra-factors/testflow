import React, {useState} from 'react';
import styles from './index.module.scss';

import {Input} from 'antd';
import {SVG} from 'factorsComponents'

export default function FilterBlock({filter, insertFilter}) {

    const [filterTypeState, setFilterTypeState] = useState("props");
    const [searchTerm, setSearchTerm] = useState("");
    const [newFilterState, setNewFilterState] = useState({
        "props": "",
        "operator": "",
        "values": []
    })

    const placeHolder = {
        "props": "Choose a property",
        "operator": "Choose an operator",
        "values": "Choose values"
    }

    const filterDropDownOptions = {
        "props": [
            {
                "label": "User Properties",
                "icon": "play",
                "values": [
                    "Cart Updated",
                    "Paid"
                ]
            },
            {
                "label": "Event Properties",
                "icon": "mouseevent",
                "values": [
                    "City",
                    "Country"
                ]
            }
        ],
        "operator": [
            "isEqual",
            "lessThan",
            "greaterThan"
        ],
        "values": {
            "Cart Updated": ["cart val1", "cart val2", "cart val3"],
            "Paid": ["paid val1", "paid val2", "paid val3"],
            "City": ["Bangalore", "Delhi", "Mumbai"],
            "Country": ["India", "USA", "France", "UK"]
        }
    }

    const renderFilterContent = () => {
        return (
            <div className={`${styles.filter_block__filter_content} ml-4`}>
                    {filter.props + ' ' + filter.operator + ' ' + filter.values.join(', ')}
            </div> 
        )
    }

    const onSelectSearch = (userInput) => {
        this.setSearchTerm(userInput);
    }

    const changeFilterTypeState = (next = true) => {
        if(next) {
            filterTypeState === 'props' ? setFilterTypeState("operator") :
            filterTypeState === 'operator' ? setFilterTypeState("values") : 
            null;
        } else {
            filterTypeState === 'values' ? setFilterTypeState("operator") :
            filterTypeState === 'operator' ? setFilterTypeState("props") : 
            null;
        }
    }

    const optionClick = (value) => {
        const newFilter = Object.assign({}, newFilterState);
        if(filterTypeState === 'values') {
            newFilter[filterTypeState].push(value);
        } else {
            newFilter[filterTypeState] = value;
        }
        changeFilterTypeState();
        setNewFilterState(newFilter);
    }

    const renderOptions = (options) => {
        let renderOptions = []
        switch (filterTypeState) {
            case "props": 
                renderOptions = options.map(group => {
                    return <>
                        <div className={styles.filter_block__filter_select__option_group}
                            >
                            <SVG name={group["icon"]} extraClass={`self-center`}></SVG>
                            <span className={`ml-1`}>{group["label"]}</span>
                            <SVG name="plus" extraClass={`ml-20 self-center`}></SVG>
                        </div>
                        {
                            group["values"].map(opt => 
                                <span className={styles.filter_block__filter_select__option}
                                    onClick={() => optionClick(opt)}    >
                                    {opt}
                                </span>
                                )
                        }
                    </>
                });
                break;
            case "operator":
                renderOptions = options.map(opt => 
                    <span className={styles.filter_block__filter_select__option}
                        onClick={() => optionClick(opt)}    >
                        {opt}
                    </span>
                );
                break;
            case "values":
                renderOptions = options[newFilterState["props"]].map(opt => 
                    <span className={styles.filter_block__filter_select__option}
                        onClick={() => optionClick(opt)}    >
                        {opt}
                    </span>
                );
                break;
        }

        return renderOptions;
        
    }

    const renderTags = () => {
        const tags = [];
        const tagClass = styles.filter_block__filter_select__tag;
        newFilterState["props"] ? 
            tags.push(<span className={tagClass}>
                    {newFilterState["props"]}
                </span>) : null;
        newFilterState["operator"] ? 
            tags.push(<span className={tagClass}>
                    {newFilterState["operator"]}
                </span>) : null;

        if(newFilterState["values"].length > 0) {
            newFilterState["values"].slice(0, 2).forEach((val, i) => {
                tags.push(<span className={tagClass}>
                    {newFilterState["values"][i]}
                </span>)
            })
            newFilterState["values"].length >= 3 ? tags.push(
                <span>
                    ...+{newFilterState["values"].length - 2}
                </span>
            ) : null;
        }
        if(tags.length < 1) {
            tags.push(<SVG name="plus" />);
        }
        return tags;
    }

    const renderFilterSelect = () => {
        return (
            <div className={`${styles.filter_block__filter_select} ml-4 fa-filter-select`}>
                <Input 
                    className={styles.filter_block__filter_select__input}  
                    placeholder={placeHolder[filterTypeState]} 
                    prefix={renderTags()} 
                    onChange={onSelectSearch}
                />
                <div className={styles.filter_block__filter_select__content}>
                    {renderOptions(filterDropDownOptions[filterTypeState])}
                </div>
            </div> 
        )
    }

    const onClickOutside = () => {
        if(newFilterState["props"].length 
            && newFilterState["operator"].length 
            && newFilterState["values"].length   
        ) {
            insertFilter(newFilterState);
        }
    }
    
    return (
        <div className={styles.filter_block}>
            <span className={`${styles.filter_block__prefix} ml-10`}>where</span>
            {filter? 
                renderFilterContent()
                : 
                renderFilterSelect()
            }
            <div className={styles.filter_block__hd_overlay} onClick={onClickOutside}></div>
        </div>
    )
}