import React from 'react';
import moment from 'moment';
import {
  generateColors,
  SortResults,
  getClickableTitleSorter,
} from '../../../../utils/dataFormatter';
import { Number as NumFormat } from '../../../../components/factorsComponents';
import { DATE_FORMATS } from '../../../../utils/constants';

export const formatData = (data, arrayMapper) => {
  if (
    !data.result_group ||
    !data.result_group.length ||
    !data.result_group[0].headers ||
    !data.result_group[0].headers.length ||
    !data.result_group[0].rows ||
    !data.result_group[0].rows.length
  ) {
    return [];
  }
  console.log('no breakdown campaign format data');
  const result = [];
  arrayMapper.forEach((elem) => {
    const dateTimeIndex = data.result_group[0].headers.indexOf('datetime');
    const dateTimeEventIndex = data.result_group[0].headers.indexOf(
      elem.eventName
    );
    const eventIndex = data.result_group[1].headers.indexOf(elem.eventName);
    if (
      dateTimeEventIndex > -1 &&
      eventIndex > -1 &&
      data.result_group[1].rows.length &&
      data.result_group[1].rows[0][eventIndex]
    ) {
      result.push({
        index: elem.index,
        name: elem.eventName,
        mapper: elem.mapper,
        dataOverTime: data.result_group[0].rows.map((row) => {
          return {
            date: new Date(row[dateTimeIndex]),
            [elem.mapper]: row[dateTimeEventIndex],
          };
        }),
        total: data.result_group[1].rows[0][eventIndex],
      });
    }
  });
  return result;
};

export const getTableColumns = (
  chartsData,
  frequency,
  currentSorter,
  handleSorting
) => {
  console.log('no breakdown campaign getTableColumns');
  const format = DATE_FORMATS[frequency] || DATE_FORMATS['date'];

  const result = chartsData.map((elem) => {
    return {
      title: getClickableTitleSorter(
        elem.name,
        { key: elem.name, type: 'numerical', subtype: null },
        currentSorter,
        handleSorting
      ),
      dataIndex: elem.name,
      render: (d) => {
        return <NumFormat number={d} />;
      },
    };
  });
  return [
    {
      title: getClickableTitleSorter(
        'Date',
        { key: 'date', type: 'datetime', subtype: 'date' },
        currentSorter,
        handleSorting
      ),
      dataIndex: 'date',
      render: (d) => {
        return moment(d).format(format);
      },
    },
    ...result,
  ];
};

export const getTableData = (chartsData, currentSorter) => {
  console.log('no breakdown campaign getTableData');
  const dates = chartsData[0].dataOverTime.map((d) => d.date);
  const columns = chartsData.map((elem) => elem.name);
  const result = dates.map((date, dateIndex) => {
    const colVals = {};
    columns.forEach((col, index) => {
      const mapper = chartsData[index].mapper;
      colVals[col] = chartsData[index].dataOverTime[dateIndex][mapper];
    });
    return {
      index: dateIndex,
      date,
      ...colVals,
    };
  });
  return SortResults(result, currentSorter);
};

export const getDateBaseTableColumns = (
  chartsData,
  frequency,
  currentSorter,
  handleSorting
) => {
  console.log('no breakdown campaign getDateBaseTableColumns');
  const format = DATE_FORMATS[frequency] || DATE_FORMATS['date'];

  const dates = chartsData[0].dataOverTime.map((d) => d.date);
  const dateColumns = dates.map((date) => {
    return {
      title: getClickableTitleSorter(
        moment(date).format(format),
        { key: moment(date).format(format), type: 'numerical', subtype: null },
        currentSorter,
        handleSorting
      ),
      width: 100,
      dataIndex: moment(date).format(format),
      render: (d) => {
        return <NumFormat number={d} />;
      },
    };
  });
  return [
    {
      title: getClickableTitleSorter(
        'Measures',
        {
          key: 'measures',
          type: 'categorical',
          subtype: null,
        },
        currentSorter,
        handleSorting
      ),
      dataIndex: 'measures',
      fixed: 'left',
      width: 150,
    },
    ...dateColumns,
  ];
};

export const getDateBasedTableData = (chartsData, frequency, currentSorter) => {
  console.log('no breakdown campaign getDateBasedTableData');
  const format = DATE_FORMATS[frequency] || DATE_FORMATS['date'];
  const result = chartsData.map((elem) => {
    const dateVals = {};
    elem.dataOverTime.forEach((d) => {
      dateVals[moment(d.date).format(format)] = d[elem.mapper];
    });
    return {
      index: elem.index,
      measures: elem.name,
      ...dateVals,
    };
  });
  return SortResults(result, currentSorter);
};

export const formatDataInHighChartsSeriesFormat = (data, arrayMapper) => {
  if (
    !data.result_group ||
    !data.result_group.length ||
    !data.result_group[0].headers ||
    !data.result_group[0].headers.length ||
    !data.result_group[0].rows ||
    !data.result_group[0].rows.length
  ) {
    return {
      categories: [],
      seriesData: [],
    };
  }
  console.log('no breakdown campaign formatDataInHighChartsSeriesFormat');
  const { headers, rows } = data.result_group[0];
  const dateIndex = headers.findIndex((h) => h === 'datetime');
  let differentDates = new Set();
  rows.forEach((row) => {
    differentDates.add(row[dateIndex]);
  });
  differentDates = Array.from(differentDates);
  const initializedDatesData = differentDates.map(() => {
    return 0;
  });
  const appliedColors = generateColors(arrayMapper.length);
  const eventIndices = [];
  const resultantData = arrayMapper.map((m, index) => {
    eventIndices.push(headers.findIndex((header) => m.eventName === header));
    return {
      name: m.eventName,
      data: [...initializedDatesData],
      index: m.index,
      color: appliedColors[index],
      marker: {
        enabled: false,
      },
    };
  });

  rows.forEach((row) => {
    const idx = differentDates.indexOf(row[dateIndex]);
    eventIndices.forEach((valIndex, index) => {
      if (valIndex > -1) {
        resultantData[index].data[idx] = row[valIndex];
      }
    });
  });
  return {
    categories: differentDates,
    seriesData: resultantData,
  };
};
