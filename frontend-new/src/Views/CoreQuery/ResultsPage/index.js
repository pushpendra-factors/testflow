import React, { useState, useEffect } from 'react';
import GroupedChart from './GroupedChart';
import DataTable from './DataTable';
import { generateGroupedChartsData, generateDummyData, generateGroups, generateColors, generateUngroupedChartsData } from './utils';
import EventsInfo from './EventsInfo';
import FiltersInfo from './FiltersInfo';
import UngroupedChart from './UngroupedChart';


function PageContent({ queries, setDrawerVisible }) {

    const [eventsData, setEventsData] = useState([]);
    const [groups, setGroups] = useState([]);

    useEffect(() => {
        const dummyData = generateDummyData(queries);
        setEventsData(dummyData);
        setGroups(generateGroups(dummyData));
    }, [queries]);

    const groupedChartData = generateGroupedChartsData(eventsData, groups);
    const ungroupedChartsData = generateUngroupedChartsData(eventsData);
    const chartColors = generateColors(eventsData);

    if (!eventsData.length) {
        return null;
    }

    const handleEventsVisibilityChange = (value, index) => {
        setEventsData(currEventData => {
            return currEventData.map(event => {
                if (event.index !== index) {
                    return event;
                }
                event.display = value;
                return event;
            })
        })
    }

    const handleGroupDataChange = (value, index, group) => {
        setEventsData(currEventData => {
            return currEventData.map(event => {
                if (event.index !== index) {
                    return event;
                }
                event.data[group] = parseInt(value);
                return event;
            })
        })
    }

    return (
        <div>
            <EventsInfo queries={queries} />
            <FiltersInfo
                handleGroupDataChange={handleGroupDataChange}
                handleEventsVisibilityChange={handleEventsVisibilityChange}
                eventsData={eventsData} groups={groups}
                setDrawerVisible={setDrawerVisible}
            />

            <UngroupedChart
                chartData={ungroupedChartsData}
            />
            
                <div className="mt-16">
                <GroupedChart
                    chartData={groupedChartData}
                    chartColors={chartColors}
                    groups={groups.filter(elem => elem.is_visible)}
                    eventsData={eventsData}
                />
            </div>


            <div className="mt-4 pl-4">
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