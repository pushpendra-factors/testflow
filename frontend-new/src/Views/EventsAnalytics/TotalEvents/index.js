import React, { useEffect, useCallback, useState } from 'react';
import styles from './index.module.scss';
import SpikeChart from './SpikeChart';

function TotalEvents() {

    const [chartData, setChartData] = useState([]);

    const randomDate = (start, end) => {
        return new Date(start.getTime() + Math.random() * (end.getTime() - start.getTime()));
    }

    const getData = useCallback(() => {
        const result = [];
        const dates = [];
        for (let i = 0; i < 30; i++) {
            let date = randomDate(new Date(2020, 0, 1), new Date());
            let convertedDate = date.getFullYear() + date.getMonth() + date.getDate();
            while (dates.indexOf(convertedDate) > -1) {
                date = randomDate(new Date(2020, 0, 1), new Date());
                convertedDate = date.getFullYear() + date.getMonth() + date.getDate();
            }
            dates.push(convertedDate);
            result.push({
                date: date,
                value: Math.floor(Math.random() * 11)
            });
        }
        return result
    }, [])

    useEffect(() => {
        setChartData(getData());
    }, [getData]);

    const total = 2340;

    if (!chartData.length) {
        return null;
    }

    return (
        <div className="total-events">
            <div className="flex justify-center items-center mt-4">
                <div className="w-1/4 flex flex-col items-center justify-center">
                    <div className="flex items-center mb-4">
                        <div className={`mr-1 ${styles.eventCircle}`}></div>
                        <div className={styles.eventText}>Add to Wishlist</div>
                    </div>
                    <div className={styles.totalText}>{total}</div>
                </div>
                <div className="w-3/4">
                    <SpikeChart chartData={chartData} chartColor="#4D7DB4" />
                </div>
            </div>
            {/* <div className="mt-8">
                <DataTable
                    eventsData={eventsData}
                    groups={groups}
                    setGroups={setGroups}
                />
            </div> */}
        </div>
    )
}

export default TotalEvents;