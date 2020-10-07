import { getTitleWithSorter } from "../../CoreQuery/FunnelsResultPage/utils";

export const formatData = (data, breakdown) => {
    const result = [];
    data.rows.forEach(d => {
        const str = d.slice(2, d.length - 1).join(",");
        const idx = result.findIndex(r => r.label === str);
        if (idx === -1) {
            result.push({
                label: str,
                value: d[d.length - 1]
            });
        } else {
            result[idx].value += d[d.length - 1]
        }
    });
    result.sort((a, b) => {
        return parseInt(a.value) <= parseInt(b.value) ? 1 : -1;
    });
    return result;
}

export const getTableColumns = (breakdown, currentSorter, handleSorting) => {
    const eventBreakdowns = breakdown.filter(elem => elem.prop_category === 'event').map(elem => elem.property).join(",");
    const userBreakdowns = breakdown.filter(elem => elem.prop_category === 'user').map(elem => elem.property).join(",");
    const result = [];
    if (eventBreakdowns) {
        result.push({
            title: eventBreakdowns,
            dataIndex: eventBreakdowns
        })
    }
    if (userBreakdowns) {
        result.push({
            title: userBreakdowns,
            dataIndex: userBreakdowns
        })
    }
    result.push({
        title: getTitleWithSorter(`Event Count`, `Event Count`, currentSorter, handleSorting),
        dataIndex: 'Event Count'
    })
    return result;
}

export const getDataInTableFormat = (data, columns, breakdown, searchText, currentSorter) => {
    const filteredData = data.filter(elem => elem.label.toLowerCase().includes(searchText.toLowerCase()))
    const result = filteredData.map((d, index) => {
        const obj = {}
        columns.slice(0, columns.length - 1).forEach(c => {
            const keys = c.title.split(",");
            const val = keys.map(k => {
                const idx = breakdown.findIndex(b => b.property === k);
                return d.label.split(",")[idx];
            });
            obj[c.title] = val.join(",");
        })
        return { ...obj, [`Event Count`]: d.value, index }
    })
    result.sort((a, b) => {
        if (currentSorter.order === 'ascend') {
            return parseInt(a[currentSorter.key]) >= parseInt(b[currentSorter.key]) ? 1 : -1;
        }
        if (currentSorter.order === 'descend') {
            return parseInt(a[currentSorter.key]) <= parseInt(b[currentSorter.key]) ? 1 : -1;
        }
        return 0;
    });
    return result;
}