import React, { useMemo } from 'react';
import { Text } from 'factorsComponents';
import { useSelector } from 'react-redux';


const matchEventName = (item = "", eventPropNames = [], userPropNames = [], stringOnly = false, color = 'grey') => {

    // const {userPropNames, eventPropNames} = useSelector((state) => state.coreQuery)
    // let userPropNames = '';
    // let  eventPropNames = '';
    let findItem = eventPropNames?.[item] || userPropNames?.[item]
    if (item) {
        if (stringOnly) {
            return findItem ? findItem : item
        }
        else {
            return <Text type={"title"} level={8} color={color} extraClass={"m-0"} truncate={true} charLimit={35}>{findItem ? findItem : item}</Text>
        }
    }
    else null
}

export default matchEventName