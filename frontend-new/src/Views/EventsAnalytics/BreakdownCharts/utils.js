import { getTitleWithSorter } from '../../CoreQuery/FunnelsResultPage/utils';
import moment from 'moment';

export const getTableColumns = (events, breakdown, currentSorter, handleSorting) => {
  const breakdownColumns = breakdown.map(e => {
    return {
      title: e,
      dataIndex: e
    };
  });

  const eventColumns = events.map(e => {
    return {
      title: getTitleWithSorter(`${e}: Event Count`, e, currentSorter, handleSorting),
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
  return [];
};

export const formatSingleEventSinglePropertyData = (data) => {
  const properties = {};
  const result = [];
  data.rows.forEach(elem => {
    if (elem[1] !== '$none') {
      if (Object.prototype.hasOwnProperty.call(properties, elem[1])) {
        result[properties[elem[1]]].value += elem[2];
      } else {
        properties[elem[1]] = result.length;
        result.push({
          label: elem[1],
          value: elem[2]
        });
      }
    }
  });
  result.sort((a, b) => {
    return parseInt(a.value) <= parseInt(b.value) ? 1 : -1;
  });
  return result;
};

export const formatDataInLineChartFormat = (data, visibleProperties, mapper, hiddenProperties) => {
  const visibleLabels = visibleProperties.map(v => v.label).filter(l => hiddenProperties.indexOf(l) === -1);
  const resultInObjFormat = {};
  const result = [];
  data.rows.forEach(elem => {
    if (visibleLabels.indexOf(elem[1]) > -1) {
      if (resultInObjFormat[elem[0]]) {
        resultInObjFormat[elem[0]][elem[1]] = elem[2];
      } else {
        resultInObjFormat[elem[0]] = {
          [elem[1]]: elem[2]
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
  const result = [
    {
      title: breakdown[0],
      dataIndex: breakdown[0],
      fixed: 'left',
      width: 200
    }];

  const dateColumns = data[0].slice(1).map(elem => {
    return {
      title: getTitleWithSorter(moment(elem).format('MMM D'), moment(elem).format('MMM D'), currentSorter, handleSorting),
      width: 100,
      dataIndex: moment(elem).format('MMM D')
    };
  });
  return [...result, ...dateColumns];
};

export const getDateBasedTableData = (labels, data, breakdown, searchText, currentSorter) => {
  const filteredLabels = labels.filter(d => d.toLowerCase().indexOf(searchText.toLowerCase()) > -1);
  const result = filteredLabels.map((elem, index) => {
    const entries = data.rows.filter(d => d[1] === elem);
    const obj = {
      index,
      [breakdown[0]]: elem
    };
    entries.forEach(entry => {
      obj[moment(entry[0]).format('MMM D')] = entry[2];
    });
    return obj;
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
