import React from 'react';
import Header from '../AppLayout/Header';
import { Button } from 'antd';
import { PoweroffOutlined } from '@ant-design/icons';
import EventsInfo from '../CoreQuery/FunnelsResultPage/EventsInfo';
import ContentTabs from '../../components/ContentTabs';
import TotalEvents from './TotalEvents';

function EventsAnalytics({ queries }) {


    const tabItems = [
        {
            key: '1',
            title: 'Total Events',
            // titleIcon: (<ConversionsOvertimeIcon style={{ fontSize: '24px', color: '#3E516C' }} />),
            content: <TotalEvents />
        },
        {
            key: '2',
            title: 'Total Users',
            // titleIcon: (<TotalConversionsIcon style={{ fontSize: '24px', color: '#3E516C' }} />),
            content: <div>Coming Soon</div>
        },
        {
            key: '3',
            title: 'Active Users',
            // titleIcon: (<TimeToConvertIcon style={{ fontSize: '24px', color: '#3E516C' }} />),
            content: <div>Coming Soon</div>
        },
        {
            key: '4',
            title: 'Frequency',
            // titleIcon: (<ConversionFrequencyIcon style={{ fontSize: '24px', color: '#3E516C' }} />),
            content: <div>Coming Soon</div>
        }
    ]

    return (
        <>
            <Header>
                <div className="flex py-4 justify-end">
                    <Button type="primary" icon={<PoweroffOutlined />} >Save query as</Button>
                </div>
                <div className="py-4">
                    <EventsInfo queries={queries} />
                </div>
            </Header>
            <div className="mt-40 mb-8 fa-container">
                <ContentTabs activeKey="1" tabItems={tabItems} />
            </div>
        </>
    )
}

export default EventsAnalytics;