import React, { useState, useEffect } from 'react';
import GroupedChart from './GroupedChart';
import DataTable from './DataTable';
import { generateGroupedChartsData, generateDummyData, generateGroups, generateColors } from './utils';
import EventsInfo from './EventsInfo';
import FiltersInfo from './FiltersInfo';


function PageContent({ queries, setDrawerVisible }) {
    
    const [eventsData, setEventsData] = useState([]);
    const [groups, setGroups] = useState([]);

    useEffect(() => {
        const dummyData = generateDummyData(queries);
        setEventsData(dummyData);
        setGroups(generateGroups(dummyData));
    }, [queries]);

    const chartData = generateGroupedChartsData(eventsData, groups);
    const chartColors = generateColors(eventsData);

    if (!eventsData.length) {
        return null;
    }

    return (
        <div>
            <EventsInfo queries={queries} />
            <FiltersInfo setDrawerVisible={setDrawerVisible} />
            <GroupedChart
                chartData={chartData}
                chartColors={chartColors}
                groups={groups.filter(elem => elem.is_visible)}
                eventsData={eventsData}
            />
            <div className="mt-8 pl-4">
                <DataTable
                    eventsData={eventsData}
                    groups={groups}
                    setGroups={setGroups}
                />
            </div>
        </div>

    )
}

export default PageContent;