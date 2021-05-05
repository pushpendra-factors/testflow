import React from 'react';
import moment from 'moment';
import {
  getTitleWithSorter,
  SortData,
  generateColors,
} from '../../../../utils/dataFormatter';
import { Number as NumFormat } from '../../../../components/factorsComponents';

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
  let format = 'MMM D, YYYY';
  if (frequency === 'hour') {
    format = 'h A, MMM D';
  }
  const result = chartsData.map((elem) => {
    return {
      title: getTitleWithSorter(
        elem.name,
        elem.name,
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
      title: getTitleWithSorter('Date', 'date', currentSorter, handleSorting),
      dataIndex: 'date',
      render: (d) => {
        return moment(d).format(format);
      },
    },
    ...result,
  ];
};

export const getTableData = (chartsData, currentSorter) => {
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
  return SortData(result, currentSorter.key, currentSorter.order);
};

export const getDateBaseTableColumns = (
  chartsData,
  frequency,
  currentSorter,
  handleSorting
) => {
  let format = 'MMM D';
  if (frequency === 'hour') {
    format = 'h A, MMM D';
  }
  const dates = chartsData[0].dataOverTime.map((d) => d.date);
  const dateColumns = dates.map((date) => {
    return {
      title: getTitleWithSorter(
        moment(date).format(format),
        moment(date).format(format),
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
      title: 'Measures',
      dataIndex: 'measures',
      fixed: 'left',
      width: 150,
    },
    ...dateColumns,
  ];
};

export const getDateBasedTableData = (chartsData, frequency, currentSorter) => {
  let format = 'MMM D';
  if (frequency === 'hour') {
    format = 'h A, MMM D';
  }
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
  return SortData(result, currentSorter.key, currentSorter.order);
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
