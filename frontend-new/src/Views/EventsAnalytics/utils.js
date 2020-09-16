import moment from 'moment';

import { getTitleWithSorter } from '../CoreQuery/FunnelsResultPage/utils';

export const getSingleEventNoGroupingTableData = (data, currentSorter) => {
    const result = data.map((elem, index) => {
        return {
            index,
            ...elem,
            date: moment(elem.date).format('MMM D, YYYY'),
        }
    })

    result.sort((a, b) => {
        if (currentSorter.order === 'ascend') {
            return parseInt(a[currentSorter.key]) >= parseInt(b[currentSorter.key]) ? 1 : -1;
        }
        if (currentSorter.order === 'descend') {
            return parseInt(a[currentSorter.key]) <= parseInt(b[currentSorter.key]) ? 1 : -1;
        }
        return 0;
    })
    return result;
}

export const getColumns = (events, currentSorter, handleSorting) => {
    let result = [
        {
            title: '',
            dataIndex: '',
            width: 37
        },
        {
            title: 'Date',
            dataIndex: 'date',
        }]

    const eventColumns = events.map(e => {
        return {
            title: getTitleWithSorter(e, e, currentSorter, handleSorting),
            dataIndex: e
        }
    });
    return [...result, ...eventColumns];
}

const randomDate = (start, end) => {
    return new Date(start.getTime() + Math.random() * (end.getTime() - start.getTime()));
}

export const getSpikeChartData = (events) => {
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
        let eventsData = {};
        events.forEach(event=>{
            eventsData[event] = Math.floor(Math.random() * 11)
        });
        result.push({
            date: date,
            ...eventsData    
        });
    }
    return result
}