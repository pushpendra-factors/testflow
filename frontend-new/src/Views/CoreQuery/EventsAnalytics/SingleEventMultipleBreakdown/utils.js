import React from 'react';
import moment from 'moment';
import { labelsObj } from '../../utils';
import {
  SortData,
  getTitleWithSorter,
  generateColors,
} from '../../../../utils/dataFormatter';
import { Number as NumFormat } from '../../../../components/factorsComponents';

export const formatData = (data) => {
  const result = data.metrics.rows.map((d) => {
    const str = d.slice(2, d.length - 1).join(',');
    return {
      label: str,
      value: d[d.length - 1],
    };
  });
  return SortData(result, 'value', 'descend');
};

export const getTableColumns = (
  events,
  breakdown,
  currentSorter,
  handleSorting,
  page
) => {
  const dataIndices = [];
  const eventBreakdowns = breakdown
    .filter((elem) => elem.prop_category === 'event')
    .map((elem) => {
      let dataIndex = elem.property;
      if (dataIndices.indexOf(dataIndex) > -1) {
        const count = dataIndices.filter((i) => i === dataIndex);
        dataIndex = elem.property + '-' + count;
      }
      dataIndices.push(elem.property);
      return {
        title: elem.property,
        dataIndex,
      };
    });
  const userBreakdowns = breakdown
    .filter((elem) => elem.prop_category === 'user')
    .map((elem) => {
      let dataIndex = elem.property;
      if (dataIndices.indexOf(dataIndex) > -1) {
        const count = dataIndices.filter((i) => i === dataIndex).length;
        dataIndex = elem.property + '-' + count;
      }
      dataIndices.push(elem.property);
      return {
        title: elem.property,
        dataIndex,
      };
    });
  const valCol = {
    title: getTitleWithSorter(
      `${events[0]}: ${labelsObj[page]}`,
      'Event Count',
      currentSorter,
      handleSorting
    ),
    dataIndex: 'Event Count',
    render: (d) => {
      return <NumFormat number={d} />;
    },
  };
  return [...eventBreakdowns, ...userBreakdowns, valCol];
};

export const getDataInTableFormat = (
  data,
  columns,
  searchText,
  currentSorter
) => {
  const filteredData = data.filter((elem) =>
    elem.label.toLowerCase().includes(searchText.toLowerCase())
  );
  const result = filteredData.map((d, index) => {
    const obj = {};
    columns.slice(0, columns.length - 1).forEach((c, idx) => {
      obj[c.dataIndex] = d.label.split(',')[idx];
    });
    return { ...obj, 'Event Count': d.value, index };
  });

  return SortData(result, currentSorter.key, currentSorter.order);
};

export const formatDataInLineChartFormat = (
  data,
  visibleProperties,
  mapper,
  hiddenProperties
) => {
  const visibleLabels = visibleProperties
    .map((v) => v.label)
    .filter((l) => hiddenProperties.indexOf(l) === -1);
  const resultInObjFormat = {};
  const result = [];
  data.rows.forEach((elem) => {
    const str = elem.slice(3, elem.length - 1).join(',');
    const val = elem[elem.length - 1];
    if (visibleLabels.indexOf(str) > -1) {
      if (resultInObjFormat[elem[1]]) {
        resultInObjFormat[elem[1]][str] = val;
      } else {
        resultInObjFormat[elem[1]] = {
          [str]: val,
        };
      }
    }
  });
  result.push(['x']);
  const keysMapper = {};
  visibleLabels.forEach((v) => {
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

export const getDateBasedColumns = (
  data,
  breakdown,
  currentSorter,
  handleSorting,
  frequency
) => {
  const dataIndices = [];
  const eventBreakdowns = breakdown
    .filter((elem) => elem.prop_category === 'event')
    .map((elem, bIndex) => {
      let dataIndex = elem.property;
      if (dataIndices.indexOf(dataIndex) > -1) {
        const count = dataIndices.filter((i) => i === dataIndex);
        dataIndex = elem.property + '-' + count;
      }
      dataIndices.push(elem.property);
      return {
        title: elem.property,
        dataIndex,
        fixed: !bIndex ? 'left' : '', //fixed to left if this is the first column
      };
    });
  const userBreakdowns = breakdown
    .filter((elem) => elem.prop_category === 'user')
    .map((elem, bIndex) => {
      let dataIndex = elem.property;
      if (dataIndices.indexOf(dataIndex) > -1) {
        const count = dataIndices.filter((i) => i === dataIndex).length;
        dataIndex = elem.property + '-' + count;
      }
      dataIndices.push(elem.property);
      return {
        title: elem.property,
        dataIndex,
        width: 150,
        fixed: !eventBreakdowns.length && !bIndex ? 'left' : '', //fixed to left if this is the first column
      };
    });
  let format = 'MMM D';
  if (frequency === 'hour') {
    format = 'h A, MMM D';
  }
  const dateColumns = data[0].slice(1).map((elem) => {
    return {
      title: getTitleWithSorter(
        moment(elem).utc().format(format),
        moment(elem).utc().format(format),
        currentSorter,
        handleSorting
      ),
      width: 100,
      dataIndex: moment(elem).utc().format(format),
      render: (d) => {
        return <NumFormat number={d} />;
      },
    };
  });
  return [...eventBreakdowns, ...userBreakdowns, ...dateColumns];
};

export const getDateBasedTableData = (
  labels,
  data,
  columns,
  searchText,
  currentSorter,
  frequency
) => {
  const filteredLabels = labels.filter(
    (d) => d.toLowerCase().indexOf(searchText.toLowerCase()) > -1
  );
  let format = 'MMM D';
  if (frequency === 'hour') {
    format = 'h A, MMM D';
  }
  const result = filteredLabels.map((elem, index) => {
    const entries = data.rows.filter(
      (d) => d.slice(3, d.length - 1).join(',') === elem
    );
    const obj = {
      index,
    };
    columns.slice(0, columns.length - 1).forEach((c, idx) => {
      obj[c.dataIndex] = elem.split(',')[idx];
    });
    entries.forEach((entry) => {
      obj[moment(entry[1]).format(format)] = entry[entry.length - 1];
    });
    return obj;
  });
  return SortData(result, currentSorter.key, currentSorter.order);
};

export const formatDataInStackedAreaFormat = (
  data,
  visibleLabels,
  arrayMapper
) => {
  if (
    !data.headers ||
    !data.headers.length ||
    !data.rows ||
    !data.rows.length
  ) {
    return {
      categories: [],
      data: [],
    };
  }
  const colors = generateColors(visibleLabels.length);
  const dateIndex = data.headers.findIndex((h) => h === 'datetime');
  const countIndex = data.headers.findIndex((h) => h === 'count');
  const eventIndex = data.headers.findIndex((h) => h === 'event_name');
  const breakdownIndex = eventIndex + 1;
  let differentDates = new Set();
  data.rows.forEach((row) => {
    differentDates.add(row[dateIndex]);
  });
  differentDates = Array.from(differentDates);
  const resultantData = visibleLabels.map((name, index) => {
    const data = differentDates.map(() => {
      return 0;
    });
    return {
      name,
      data,
      color: colors[index],
      marker: {
        enabled: false,
      },
    };
  });
  data.rows.forEach((row) => {
    const breakdownJoin = row.slice(breakdownIndex, countIndex).join(',');
    const bIdx = visibleLabels.indexOf(breakdownJoin);
    if (bIdx > -1) {
      const idx = differentDates.indexOf(row[dateIndex]);
      resultantData[bIdx].data[idx] = row[countIndex];
    }
  });
  return {
    categories: differentDates,
    data: resultantData,
  };
};
