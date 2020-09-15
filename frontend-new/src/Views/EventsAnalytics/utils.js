import moment from 'moment';

import { getTitleWithSorter } from '../CoreQuery/FunnelsResultPage/utils';

export const getSingleEventNoGroupingTableData = (data, event, currentSorter) => {
    const result = data.map((elem, index) => {
        return {
            index,
            date: moment(elem.date).format('MMM D, YYYY'),
            [event]: elem.value
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

export const getSingleEventNoGroupingTableColumns = (data, event, currentSorter, handleSorting) => {
    let result = [
        {
            title: 'Date',
            dataIndex: 'date',
        },
        {
            title: getTitleWithSorter(event, event, currentSorter, handleSorting),
            dataIndex: event
        }
    ]
    return result;
}