import React, { useState } from 'react';
import { formatSingleEventAnalyticsData, formatMultiEventsAnalyticsData, getDataInLineChartFormat } from '../utils';
import { singleEventResponse, multiEventResponse } from '../SampleResponse';
import ChartTypeDropdown from '../../../components/ChartTypeDropdown';
import TotalEventsTable from '../TotalEvents/TotalEventsTable';
import SparkLineChart from '../../../components/SparkLineChart';
import LineChart from '../../../components/LineChart';
import { generateColors } from '../../CoreQuery/FunnelsResultPage/utils';

function NoBreakdownCharts({ queries, eventsMapper, reverseEventsMapper }) {
    const [hiddenEvents, setHiddenEvents] = useState([]);
    const appliedColors = generateColors(queries.length);
    const [chartType, setChartType] = useState('sparklines');

    let chartsData = [];
    if (queries.length === 1) {
        chartsData = formatSingleEventAnalyticsData(singleEventResponse, queries[0], eventsMapper);
    } else {
        chartsData = formatMultiEventsAnalyticsData(multiEventResponse, queries, eventsMapper);
    }

    if (!chartsData.length) {
        return null;
    }

    const menuItems = [
        {
            key: 'sparklines',
            onClick: setChartType,
            name: 'Sparkline'
        },
        {
            key: 'linechart',
            onClick: setChartType,
            name: 'Line Chart'
        }
    ]

    let chartContent = null;

    if (chartType === 'sparklines') {
        chartContent = (
            <SparkLineChart
                queries={queries}
                chartsData={chartsData}
                parentClass="flex justify-center items-center flex-wrap mt-8"
                appliedColors={appliedColors}
                eventsMapper={eventsMapper}
            />
        )
    } else if (chartType === 'linechart') {
        chartContent = (
            <div className="flex mt-8">
                <LineChart
                    chartData={getDataInLineChartFormat(chartsData, queries, eventsMapper, hiddenEvents)}
                    appliedColors={appliedColors}
                    queries={queries}
                    reverseEventsMapper={reverseEventsMapper}
                    eventsMapper={eventsMapper}
                    setHiddenEvents={setHiddenEvents}
                    hiddenEvents={hiddenEvents}
                />
            </div>
        )
    }

    return (
        <div className="total-events">
            <div className="flex items-center justify-between">
                <div className="filters-info">

                </div>
                <div className="user-actions">
                    <ChartTypeDropdown
                        chartType={chartType}
                        menuItems={menuItems}
                        onClick={(item) => {
                            setChartType(item.key);
                        }}
                    />
                </div>
            </div>
            {chartContent}
            <div className="mt-8">
                <TotalEventsTable
                    data={chartsData}
                    events={queries}
                    reverseEventsMapper={reverseEventsMapper}
                    chartType={chartType}
                    setHiddenEvents={setHiddenEvents}
                    hiddenEvents={hiddenEvents}
                />
            </div>
        </div>
    )

}

export default NoBreakdownCharts;