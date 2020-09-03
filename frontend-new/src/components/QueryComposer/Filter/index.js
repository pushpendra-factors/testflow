import React, {useState} from 'react';
import styles from '../QueryBlock/index.module.scss';
import {SVG, Text} from 'factorsComponents';

import { Select} from 'antd';
const { OptGroup, Option } = Select;

function Filter({filter, insertFilters}) {
    const [filterType, setFilterTypeState] = useState("props");

    const [newFilter, setNewFilter] = useState({
            "props": "",
            "operator": "",
            "values": []
        });


    const eventOptions = [
        {
                label: "Event Properties",
                icon: "fav",
                values: [
                    "Country",
                    
                ]
        }, {
                label: "User Properties",
                icon: "virtual",
                values: [
                    "City",
                ]
        }
    ]

    const filterOptions = {
        "props": eventOptions,
        "operator": [
            "is",
            "less than",
            "greater than",
        ],
        "values": {
            "City": [
                "Delhi",
                "Hyderabad",
                "Mumbai",
                "Bangalore",
                "Chennai"
            ],
            "Country": [
                "India",
                "USA",
                "UK",
                "Egypt",
                "Japan",
                "China"
            ]
        }
    }

    const setFilterType = () => {
        return !newFilter["props"] ? "props" : !newFilter["operator"] ? "operator" : "values";
    }

    const renderFilterOptions = () => {
        const options = [];
        if(filterType === 'values') {
            filterOptions[filterType][newFilter['props']].forEach(opt => {
                options.push(<Option value={opt}></Option>)
            });
        }
        else if(filterType === 'props') {
            filterOptions[filterType].forEach((group, index) => {
                options.push(<OptGroup key={index} label={(
                    <div className={styles.query_block__selector_group}>
                        <SVG name={group.icon}></SVG>
                        <span >{group.label}</span>
                    </div>
                        )}>
                            {group.values.map((option,index) => (
                                <Option key={index} value={option}></Option>
                            ))}
                    </OptGroup>)
                })
        }
        else {
            filterOptions[filterType].forEach(opt => {
                options.push(<Option value={opt}></Option>)
            });
        }
        
        return options;
    }

    const sendFilters = () => {
        if(newFilter["props"].length < 1) {return null};
        if(newFilter["operator"].length < 1) {return null};
        if(newFilter["values"].length < 1) {return null};
        insertFilters(newFilter);
    }

    
    if(filter) {return null};

    const onFilterEventChange = (opt) => {
            if(filterType === 'values') {
                newFilter[filterType].push(opt[opt.length-1]);
            } else {
                newFilter[filterType] = opt[opt.length-1];
            }
            setFilterTypeState(setFilterType(filter));
        }

    return(
        <>
        <div className={`fa--query_block--filters flex justify-start items-center`}>
            <span  className={`ml-10`}><Text type={'title'} level={7} weight={'thin'} extraClass={`m-0`}>Where</Text> </span>
             <div className={`fa--query_block--filters-values flex justify-start items-center ml-4`}>
                <Select mode="tags" maxTagCount={6} onDropdownVisibleChange={(open) => !open? sendFilters(): null} showSearch style={{ width: 240}} onChange={onFilterEventChange} defaultOpen={true} showArrow={false}>
                    {renderFilterOptions()}
                </Select>
            </div>
        </div>
        </>
    )
     

}

export default Filter;