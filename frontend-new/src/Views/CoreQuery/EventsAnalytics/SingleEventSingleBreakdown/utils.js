import moment from 'moment';
import { labelsObj } from '../../utils';
import { SortData, getTitleWithSorter } from '../../../../utils/dataFormatter';

export const getTableColumns = (events, breakdown, currentSorter, handleSorting, page) => {
  const breakdownColumns = breakdown.map(e => {
    return {
      title: e,
      dataIndex: e
    };
  });

  const eventColumns = events.map(e => {
    return {
      title: getTitleWithSorter(`${e}: ${labelsObj[page]}`, e, currentSorter, handleSorting),
      dataIndex: e
    };
  });
  return [...breakdownColumns, ...eventColumns];
};

export const getDataInTableFormat = (data, events, breakdown, searchText, currentSorter) => {
  if (breakdown.length === 1 && events.length === 1) {
    const filteredData = data.filter(d => d.label.toLowerCase().indexOf(searchText.toLowerCase()) > -1);
    const result = filteredData.map((d, index) => {
      return {
        index,
        [breakdown[0]]: d.label,
        [events[0]]: d.value
      };
    });
    return SortData(result, currentSorter.key, currentSorter.order);
  }
  return [];
};

export const formatData = (data) => {
  const result = data.metrics.rows.map(elem => {
    return {
      label: elem[1],
      value: elem[2]
    };
  });
  return SortData(result, 'value', 'descend');
};

export const formatDataInLineChartFormat = (data, visibleProperties, mapper, hiddenProperties, frequency) => {
  const visibleLabels = visibleProperties.map(v => v.label).filter(l => hiddenProperties.indexOf(l) === -1);
  const resultInObjFormat = {};
  const result = [];
  data.rows.forEach(elem => {
    if (visibleLabels.indexOf(elem[2]) > -1) {
      if (resultInObjFormat[elem[0]]) {
        resultInObjFormat[elem[0]][elem[2]] = elem[3];
      } else {
        resultInObjFormat[elem[0]] = {
          [elem[2]]: elem[3]
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
  const format = 'YYYY-MM-DD HH-mm';
  for (const key in resultInObjFormat) {
    result[0].push(moment(key).format(format));
    for (const b in resultInObjFormat[key]) {
      result[keysMapper[b]].push(resultInObjFormat[key][b]);
    }
  }
  return result;
};

export const getDateBasedColumns = (data, breakdown, currentSorter, handleSorting, frequency) => {
  const result = [
    {
      title: breakdown[0],
      dataIndex: breakdown[0],
      fixed: 'left',
      width: 200
    }];

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
  return [...result, ...dateColumns];
};

export const getDateBasedTableData = (labels, data, breakdown, searchText, currentSorter, frequency) => {
  const filteredLabels = labels.filter(d => d.toLowerCase().indexOf(searchText.toLowerCase()) > -1);
  let format = 'MMM D';
  if (frequency === 'hour') {
    format = 'h A, MMM D'
  }
  const result = filteredLabels.map((elem, index) => {
    const entries = data.rows.filter(d => d[2] === elem);
    const obj = {
      index,
      [breakdown[0]]: elem
    };
    entries.forEach(entry => {
      obj[moment(entry[0]).format(format)] = entry[3];
    });
    return obj;
  });
  return SortData(result, currentSorter.key, currentSorter.order);;
};
