import moment from 'moment';
import { getTitleWithSorter } from '../../CoreQuery/FunnelsResultPage/utils';

export const formatData = (data, queries, colors) => {
  const splittedData = {};
  queries.forEach(query => {
    splittedData[query] = [];
  });
  let gIdx = 0;

  data.rows.forEach(d => {
    const date = d[0];
    const str = d.slice(2, d.length - 1).join(',');
    const idx = splittedData[d[1]].findIndex(r => r.label === str);
    if (idx === -1) {
      const queryIndex = queries.findIndex(q => q === d[1]);
      splittedData[d[1]].push({
        label: str,
        value: d[d.length - 1],
        index: gIdx,
        event: d[1],
        color: colors[queryIndex],
        dateWise: [{
          date,
          value: d[d.length - 1]
        }]
      });
      gIdx++;
    } else {
      splittedData[d[1]][idx].dateWise.push({
        date,
        value: d[d.length - 1]
      });
      splittedData[d[1]][idx].value += d[d.length - 1];
    }
  });

  let allData = [];

  for (const key in splittedData) {
    splittedData[key].sort((a, b) => {
      return parseInt(a.value) <= parseInt(b.value) ? 1 : -1;
    });
  }

  const result = [];

  for (const key in splittedData) {
    if (splittedData[key].length) {
      allData = [...allData, ...splittedData[key]];
      result.push(splittedData[key][0]);
    }
  }

  allData.sort((a, b) => {
    return parseInt(a.value) <= parseInt(b.value) ? 1 : -1;
  });

  for (let j = 0; j < allData.length; j++) {
    const obj = result.find(elem => elem.index === allData[j].index);
    if (!obj) {
      result.push(allData[j]);
    }
  }

  return result;
};

export const formatVisibleProperties = (data, queries) => {
  const vp = data.map(d => {
    return { ...d, label: `${d.label}; [${d.event}]` };
  });
  vp.sort((a, b) => {
    return parseInt(a.value) <= parseInt(b.value) ? 1 : -1;
  });
  vp.sort((a, b) => {
    const idx1 = queries.findIndex(q => q === a.event);
    const idx2 = queries.findIndex(q => q === b.event);
    return idx1 >= idx2 ? 1 : -1;
  });
  return vp;
};

export const getTableColumns = (breakdown, currentSorter, handleSorting) => {
  const result = [];
  result.push({
    title: 'Event',
    dataIndex: 'event'
  });
  breakdown.forEach(b => {
    result.push({
      title: b.property,
      dataIndex: b.property
    });
  });
  result.push({
    title: getTitleWithSorter('Event Count', 'Event Count', currentSorter, handleSorting),
    dataIndex: 'Event Count'
  });
  return result;
};

export const getTableData = (data, breakdown, searchText, currentSorter) => {
  const filteredData = data.filter(elem => elem.label.toLowerCase().includes(searchText.toLowerCase()) || elem.event.toLowerCase().includes(searchText.toLowerCase()));
  const result = [];
  filteredData.forEach(d => {
    const breakdownValues = {};
    breakdown.forEach((b, index) => {
      breakdownValues[b.property] = d.label.split(',')[index];
    });
    result.push({
      ...d, 'Event Count': d.value, ...breakdownValues
    });
  });
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
};

export const formatDataInLineChartFormat = (visibleProperties, mapper, hiddenProperties) => {
  const result = [];
  const dates = visibleProperties[0].dateWise.map(elem => moment(elem.date).format('YYYY-MM-DD'));
  result.push(['x', ...dates]);
  visibleProperties.forEach(v => {
    const label = `${v.event},${v.label}`;
    if (hiddenProperties.indexOf(label) === -1) {
      const values = v.dateWise.map(elem => elem.value);
      result.push([mapper[label], ...values]);
    }
  });
  return result;
};

export const getDateBasedColumns = (data, breakdown, currentSorter, handleSorting) => {
  const breakdownColumns = breakdown.map(elem => {
    return {
      title: elem.property,
      dataIndex: elem.property,
      fixed: 'left',
      width: 200
    };
  });
  const dateColumns = data[0].slice(1).map(elem => {
    return {
      title: getTitleWithSorter(moment(elem).format('MMM D'), moment(elem).format('MMM D'), currentSorter, handleSorting),
      width: 100,
      dataIndex: moment(elem).format('MMM D')
    };
  });
  const eventCol = {
    title: 'Event',
    dataIndex: 'Event',
    fixed: 'left',
    width: 200
  };
  return [eventCol, ...breakdownColumns, ...dateColumns];
};

export const getDateBasedTableData = (data, breakdown, currentSorter) => {
  const result = data.map(d => {
    const breakdownValues = {};
    breakdown.forEach((b, index) => {
      breakdownValues[b.property] = d.label.split(',')[index];
    });

    const dateWiseValues = {};
    d.dateWise.forEach(w => {
      const key = moment(w.date).format('MMM D');
      dateWiseValues[key] = w.value;
    });
    return {
      index: d.index,
      Event: d.event,
      ...breakdownValues,
      ...dateWiseValues
    };
  });
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
};
