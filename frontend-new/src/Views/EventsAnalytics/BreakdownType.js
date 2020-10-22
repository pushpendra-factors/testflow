import React from 'react';
import { Radio } from 'antd';

function BreakdownType({ breakdownType, handleBreakdownTypeChange, breakdown }) {

    const isAnyDisabled = breakdown.filter(elem => !elem.eventIndex).length ? false : true

    return (
        <Radio.Group value={breakdownType} onChange={handleBreakdownTypeChange}>
            <Radio.Button value="each">Each Event</Radio.Button>
            <Radio.Button disabled={isAnyDisabled} value="any">Any Event</Radio.Button>
            <Radio.Button value="all">All Events</Radio.Button>
        </Radio.Group>
    )
}

export default BreakdownType;