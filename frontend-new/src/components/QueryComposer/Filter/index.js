import React, {useState} from 'react';
import styles from '../QueryBlock/index.module.scss';

import { Select} from 'antd';
const { OptGroup, Option } = Select;

function Filter({filter}) {
    const [filterType, setFilterTypeState] = useState("props");

    const [newFilter, setNewFilter] = useState({
            "props": "",
            "operator": "",
            "values": []
        });

    const filterOptions = {
        "props": [
            "City",
            "Country"
        ],
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
        } else {
            filterOptions[filterType].forEach(opt => {
                options.push(<Option value={opt}></Option>)
            });
        }
        
        return options;
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
            <span>Where</span>
            <div className={styles.query_block__filter_query}>
                <Select mode="tags" showSearch style={{ width: 240}} onChange={onFilterEventChange} open={true} showArrow={false}>
                    {renderFilterOptions()}
                </Select>
            </div>
        </>
    )
     

}

export default Filter;