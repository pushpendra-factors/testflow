import tableStyles from './DataTable/index.module.scss'

export const generateGroupedChartsData = (data, groups) => {
    let result = data.map(elem => {
        let values = [];
        for (let key in elem.data) {
            let group = groups.find(g => g.name === key);
            if (group.is_visible) {
                values.push(elem.data[key]);
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
            conversion_rate: data[data.length - 1].data[elem] + "%",
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
        data: {
            'Chennai': 100,
            'Mumbai': 100,
            'New Delhi': 100,
            'Amritsar': 100,
            'Jalandhar': 100,
        }
    },
    {
        index: 2,
        color: '#008BAE',
        data: {
            'Chennai': 40,
            'Mumbai': 40,
            'New Delhi': 60,
            'Amritsar': 50,
            'Jalandhar': 60,
        }
    },
    {
        index: 3,
        color: '#52C07C',
        data: {
            'Chennai': 30,
            'Mumbai': 30,
            'New Delhi': 30,
            'Amritsar': 40,
            'Jalandhar': 40,
        }
    },
    {
        index: 4,
        color: '#F1C859',
        data: {
            'Chennai': 10,
            'Mumbai': 15,
            'New Delhi': 15,
            'Amritsar': 30,
            'Jalandhar': 20,
        }
    },
    {
        index: 5,
        color: '#EEAC4C',
        data: {
            'Chennai': 5,
            'Mumbai': 10,
            'New Delhi': 8,
            'Amritsar': 20,
            'Jalandhar': 5.25,
        }
    },
    {
        index: 6,
        color: '#DE7542',
        data: {
            'Chennai': 3,
            'Mumbai': 8,
            'New Delhi': 6,
            'Amritsar': 18,
            'Jalandhar': 1.5,
        }
    }
];

const ungroupedDummyData = [
    {
        value: 100
    },
    {
        value: 60
    },
    {
        value: 30
    },
    {
        value: 10
    },
    {
        value: 5
    },
    {
        value: 3
    }
]

export const generateDummyData = (labels) => {
    let result = labels.map((elem, index) => {
        return { ...groupedDummyData[index], name: elem };
    });
    return result;
}

export const generateUngroupedChartsData = (data) => {
    let result = data.map((elem, index) => {
        return {
            event: elem.name,
            value: ungroupedDummyData[index].value
        }
    })
    return result;
}