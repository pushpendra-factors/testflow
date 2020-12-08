import moment from 'moment';
import { labelsObj } from '../../utils';
import { SortData, getTitleWithSorter } from '../../../../utils/dataFormatter';

export const formatData = (data) => {
  const result = data.metrics.rows.map(d => {
    const str = d.slice(2, d.length - 1).join(',');
    return {
      label: str,
      value: d[d.length - 1]
    };
  });
  return SortData(result, 'value', 'descend');
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
    const str = elem.slice(3, elem.length - 1).join(',');
    const val = elem[elem.length - 1];
    if (visibleLabels.indexOf(str) > -1) {
      if (resultInObjFormat[elem[1]]) {
        resultInObjFormat[elem[1]][str] = val;
      } else {
        resultInObjFormat[elem[1]] = {
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
    const format = 'YYYY-MM-DD HH-mm';
    result[0].push(moment(key).format(format));
    for (const b in resultInObjFormat[key]) {
      result[keysMapper[b]].push(resultInObjFormat[key][b]);
    }
  }
  return result;
};

export const getDateBasedColumns = (data, breakdown, currentSorter, handleSorting, frequency) => {
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
  let format = 'MMM D';
  if (frequency === 'hour') {
    format = 'h A, MMM D'
  }
  const dateColumns = data[0].slice(1).map(elem => {
    return {
      title: getTitleWithSorter(moment(elem).utc().format(format), moment(elem).utc().format(format), currentSorter, handleSorting),
      width: 100,
      dataIndex: moment(elem).utc().format(format)
    };
  });
  return [...eventBreakdowns, ...userBreakdowns, ...dateColumns];
};

export const getDateBasedTableData = (labels, data, columns, searchText, currentSorter, frequency) => {
  const filteredLabels = labels.filter(d => d.toLowerCase().indexOf(searchText.toLowerCase()) > -1);
  let format = 'MMM D';
  if (frequency === 'hour') {
    format = 'h A, MMM D'
  }
  const result = filteredLabels.map((elem, index) => {
    const entries = data.rows.filter(d => d.slice(3, d.length - 1).join(',') === elem);
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
      obj[moment(entry[1]).format(format)] = entry[entry.length - 1];
    });
    return obj;
  });
  return SortData(result, currentSorter.key, currentSorter.order);
};
