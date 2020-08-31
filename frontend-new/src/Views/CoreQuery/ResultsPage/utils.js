import tableStyles from './DataTable/index.module.scss'

const windowSize = {
    w: window.outerWidth,
    h: window.outerHeight,
    iw: window.innerWidth,
    ih: window.innerHeight
};

export const generateGroupedChartsData = (data, groups) => {
    const displayedData = data.filter(elem => elem.display);
    let result = displayedData.map(elem => {
        let values = [];
        for (let key in elem.data) {
            let group = groups.find(g => g.name === key);
            if (group.is_visible) {
                values.push(((elem.data[key] / (data[0].data[key])) * 100).toFixed(2));
            }
        }
        return [elem.name, ...values];
    });
    return result;
}

export const generateColors = (data) => {
    let result = {};
    data.forEach(elem => {
        result[elem.name] = elem.color;
    });
    return result;
}

export const generateGroups = (data) => {
    let cat_names = Object.keys(data[0].data)
    let result = cat_names.map(elem => {
        return {
            name: elem,
            conversion_rate: ((((data[data.length - 1].data[elem]) / (data[0].data[elem])) * 100).toFixed(2)) + "%",
            is_visible: true,
        }
    });
    return result;
}

export const generateTableColumns = (data) => {
    let result = [
        {
            title: 'Grouping',
            dataIndex: 'name',
            className: tableStyles.groupColumn,
        },
        {
            title: 'Conversion',
            dataIndex: 'conversion',
            className: tableStyles.conversionColumn,
            sorter: (a, b) => {
                return parseFloat(a.conversion.split("%")[0]) - parseFloat(b.conversion.split("%")[0])
            },
        }
    ]
    let eventColumns = data.map((elem, index) => {
        return {
            title: elem.name,
            dataIndex: elem.name,
            className: index === data.length - 1 ? tableStyles.lastColumn : '',
            sorter: (a, b) => {
                return a[elem.name] - b[elem.name]
            },
        };
    });
    return [...result, ...eventColumns];
}

export const generateTableData = (data, groups) => {
    let appliedGroups = groups.map(elem => elem.name);
    const result = appliedGroups.map((group, index) => {
        let eventsData = {};
        data.forEach(d => {
            eventsData[d.name] = d.data[group];
        });
        return {
            index: index,
            name: group,
            conversion: data[data.length - 1].data[group] + "%",
            ...eventsData
        }
    })
    return result;
}

const groupedDummyData = [
    {
        index: 1,
        color: '#014694',
        display: true,
        data: {
            'Chennai': 20000,
            'Mumbai': 20000,
            'New Delhi': 20000,
            'Amritsar': 20000,
            'Jalandhar': 20000,
            'Kolkatta': 20000
        }
    },
    {
        index: 2,
        color: '#008BAE',
        display: true,
        data: {
            'Chennai': 8000,
            'Mumbai': 8000,
            'New Delhi': 12000,
            'Amritsar': 10000,
            'Jalandhar': 12000,
            'Kolkatta': 6000
        }
    },
    {
        index: 3,
        color: '#52C07C',
        display: true,
        data: {
            'Chennai': 6000,
            'Mumbai': 6000,
            'New Delhi': 6000,
            'Amritsar': 8000,
            'Jalandhar': 8000,
            'Kolkatta': 5000
        }
    },
    {
        index: 4,
        color: '#F1C859',
        display: true,
        data: {
            'Chennai': 2000,
            'Mumbai': 3000,
            'New Delhi': 3000,
            'Amritsar': 6000,
            'Jalandhar': 4000,
            'Kolkatta': 4000
        }
    },
    {
        index: 5,
        color: '#EEAC4C',
        display: true,
        data: {
            'Chennai': 1000,
            'Mumbai': 2000,
            'New Delhi': 1600,
            'Amritsar': 4000,
            'Jalandhar': 1050,
            'Kolkatta': 3000
        }
    },
    {
        index: 6,
        color: '#DE7542',
        display: true,
        data: {
            'Chennai': 600,
            'Mumbai': 1600,
            'New Delhi': 1200,
            'Amritsar': 3600,
            'Jalandhar': 300,
            'Kolkatta': 2000
        }
    }
];

export const generateDummyData = (labels) => {
    let result = labels.map((elem, index) => {
        return { ...groupedDummyData[index], name: elem };
    });
    return result;
}

export const generateUngroupedChartsData = (data) => {

    const displayedData = data.filter(elem => elem.display);
    let totalData = 0;

    if (displayedData.length) {
        Object.keys(displayedData[0].data).forEach(elem => {
            totalData += displayedData[0].data[elem];
        });
    }

    let result = displayedData.map((elem, index) => {

        let obj = data.find(d => d.name === elem.name);

        let netCount = 0;

        Object.keys(obj.data).forEach(d => {
            netCount += obj.data[d]
        });

        return {
            event: elem.name,
            color: elem.color,
            netCount,
            value: ((netCount / totalData) * 100).toFixed(2),
        }
    })
    return result;
}

export const checkForWindowSizeChange = (callback) => {
    if (window.outerWidth !== windowSize.w || window.outerHeight !== windowSize.h) {
        setTimeout(() => {
            windowSize.w = window.outerWidth; // update object with current window properties
            windowSize.h = window.outerHeight;
            windowSize.iw = window.innerWidth;
            windowSize.ih = window.innerHeight;
        }, 0)
        callback();
    }

    //if the window doesn't resize but the content inside does by + or - 5%
    else if (window.innerWidth + window.innerWidth * .05 < windowSize.iw ||
        window.innerWidth - window.innerWidth * .05 > windowSize.iw) {
        setTimeout(() => {
            windowSize.iw = window.innerWidth;
        }, 0);
        callback();
    }
}