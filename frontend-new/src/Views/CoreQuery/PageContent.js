import React, { useState } from 'react';
import ContentTabs from '../../components/ContentTabs';
import TotalConversionsIcon from '../../components/Icons/TotalConversions';
import TotalConversions from './TotalConversions';
import ConversionsOvertimeIcon from '../../components/Icons/ConversionsOvertime';
import TimeToConvertIcon from '../../components/Icons/TimeToConvert';
import ConversionFrequencyIcon from '../../components/Icons/ConversionFrequency';
import ConversionsOverTime from './ConversionsOverTime';
import TimeToConvert from './TimeToConvert';
import ConversionsFrequency from './ConversionsFrequency';


function Content() {
    
    const [activeKey, setActiveKey] = useState('1');
    
    const handleTabChange = (key) => {
        setActiveKey(key);
    }

    const tabItems = [
        {
            key: '1',
            title: 'Total Conversions',
            titleIcon: (<TotalConversionsIcon style={{ fontSize: '24px', color: '#3E516C' }} />),
            content: <TotalConversions activeKey={activeKey} />
        },
        {
            key: '2',
            title: 'Conversions Over Time',
            titleIcon: (<ConversionsOvertimeIcon style={{ fontSize: '24px', color: '#3E516C' }} />),
            content: <ConversionsOverTime activeKey={activeKey} />
        },
        {
            key: '3',
            title: 'Time to Convert',
            titleIcon: (<TimeToConvertIcon style={{ fontSize: '24px', color: '#3E516C' }} />),
            content: <TimeToConvert activeKey={activeKey} />
        },
        {
            key: '4',
            title: 'Conversion Frequency',
            titleIcon: (<ConversionFrequencyIcon style={{ fontSize: '24px', color: '#3E516C' }} />),
            content: <ConversionsFrequency activeKey={activeKey} />
        }
    ]

    return (
        <div className="mt-4">
            <ContentTabs tabItems={tabItems} onChange={handleTabChange} activeKey={activeKey} />
        </div>

    )
}

export default Content;