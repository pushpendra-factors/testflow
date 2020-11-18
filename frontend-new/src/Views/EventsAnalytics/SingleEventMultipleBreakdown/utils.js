import { getTitleWithSorter } from '../../CoreQuery/FunnelsResultPage/utils';
import moment from 'moment';
import { labelsObj, SortData } from '../../CoreQuery/utils';

export const formatData = (data) => {
  const result = [];
  data.rows.forEach(d => {
    const str = d.slice(2, d.length - 1).join(',');
    const idx = result.findIndex(r => r.label === str);
    if (idx === -1) {
      result.push({
        label: str,
        value: d[d.length - 1]
      });
    } else {
      result[idx].value += d[d.length - 1];
    }
  });
  result.sort((a, b) => {
    return parseInt(a.value) <= parseInt(b.value) ? 1 : -1;
  });
  return result;
};

export const getTableColumns = (events, breakdown, currentSorter, handleSorting, page) => {
  const eventBreakdowns = breakdown
    .filter(elem => elem.prop_category === 'event')
    .map(elem => {
      return {
        title: elem.property,
        dataIndex: elem.property
      };
    });
  const userBreakdowns = breakdown
    .filter(elem => elem.prop_category === 'user')
    .map(elem => {
      return {
        title: elem.property,
        dataIndex: elem.property
      };
    });
  const valCol = {
    title: getTitleWithSorter(`${events[0]}: ${labelsObj[page]}`, 'Event Count', currentSorter, handleSorting),
    dataIndex: 'Event Count'
  };
  return [...eventBreakdowns, ...userBreakdowns, valCol];
};

export const getDataInTableFormat = (data, columns, searchText, currentSorter) => {
  const filteredData = data.filter(elem => elem.label.toLowerCase().includes(searchText.toLowerCase()));
  const result = filteredData.map((d, index) => {
    const obj = {};
    columns.slice(0, columns.length - 1).forEach((c, idx) => {
      const keys = c.title.split(',');
      const val = keys.map(() => {
        return d.label.split(',')[idx];
      });
      obj[c.title] = val.join(',');
    });
    return { ...obj, 'Event Count': d.value, index };
  });

  return SortData(result, currentSorter.key, currentSorter.order);
};

export const formatDataInLineChartFormat = (data, visibleProperties, mapper, hiddenProperties) => {
  const visibleLabels = visibleProperties.map(v => v.label).filter(l => hiddenProperties.indexOf(l) === -1);
  const resultInObjFormat = {};
  const result = [];
  data.rows.forEach(elem => {
    const str = elem.slice(2, elem.length - 1).join(',');
    const val = elem[elem.length - 1];
    if (visibleLabels.indexOf(str) > -1) {
      if (resultInObjFormat[elem[0]]) {
        resultInObjFormat[elem[0]][str] = val;
      } else {
        resultInObjFormat[elem[0]] = {
          [str]: val
        };
      }
    }
  });
  result.push(['x']);
  const keysMapper = {};
  visibleLabels.forEach(v => {
    result.push([mapper[v]]);
    keysMapper[v] = result.length - 1;
  });
  for (const key in resultInObjFormat) {
    result[0].push(moment(key).format('YYYY-MM-DD'));
    for (const b in resultInObjFormat[key]) {
      result[keysMapper[b]].push(resultInObjFormat[key][b]);
    }
  }
  return result;
};

export const getDateBasedColumns = (data, breakdown, currentSorter, handleSorting) => {
  const eventBreakdowns = breakdown
    .filter(elem => elem.prop_category === 'event')
    .map(elem => {
      return {
        title: elem.property,
        dataIndex: elem.property,
        fixed: 'left',
        width: 200
      };
    });
  const userBreakdowns = breakdown
    .filter(elem => elem.prop_category === 'user')
    .map(elem => {
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
  return [...eventBreakdowns, ...userBreakdowns, ...dateColumns];
};

export const getDateBasedTableData = (labels, data, columns, searchText, currentSorter) => {
  const filteredLabels = labels.filter(d => d.toLowerCase().indexOf(searchText.toLowerCase()) > -1);
  const result = filteredLabels.map((elem, index) => {
    const entries = data.rows.filter(d => d.slice(2, d.length - 1).join(',') === elem);
    const obj = {
      index
    };
    let idx = -1;
    columns.slice(0, columns.length - 1).forEach(c => {
      const keys = c.title.split(',');
      const val = keys.map(_ => {
        idx++;
        return elem.split(',')[idx];
      });
      obj[c.title] = val.join(',');
    });
    entries.forEach(entry => {
      obj[moment(entry[0]).format('MMM D')] = entry[entry.length - 1];
    });
    return obj;
  });
  return SortData(result, currentSorter.key, currentSorter.order);
};
