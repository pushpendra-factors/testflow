import React from 'react';
import moment from 'moment';
import {
  generateColors,
  SortData,
  getTitleWithSorter,
} from '../../../../utils/dataFormatter';
import { Number as NumFormat } from '../../../../components/factorsComponents';

export const getBreakdownIndices = (data, breakdown) => {
  const result = breakdown.map((elem) => {
    const str = elem.name + '_' + elem.property;
    return data.result_group[1].headers.findIndex((elem) => elem === str);
  });
  return result;
};

export const getDateBreakdownIndices = (data, breakdown) => {
  const result = breakdown.map((elem) => {
    const str = elem.name + '_' + elem.property;
    return data.result_group[0].headers.findIndex((elem) => elem === str);
  });
  return result;
};

export const formatData = (data, arrayMapper, breakdown, currentEventIndex) => {
  try {
    const breakdownIndices = getBreakdownIndices(data, breakdown);
    // const dateBreakdownIndices = getDateBreakdownIndices(data, breakdown);
    const colors = generateColors(arrayMapper.length);
    const currEventName = arrayMapper.find(
      (elem) => elem.index === currentEventIndex
    ).eventName;
    const currDataIndex = data.result_group[1].headers.findIndex(
      (elem) => elem === currEventName
    );
    let result = [];
    // const dateRows = [...data.result_group[0].rows];
    if (currDataIndex > -1) {
      data.result_group[1].rows.forEach((elem, index) => {
        const label = [];
        breakdownIndices.forEach((b) => {
          if (b > -1) {
            label.push(elem[b]);
          }
        });
        result.push({
          index,
          label: label.join(', '),
          value: elem[currDataIndex],
          color: colors[currentEventIndex],
        });
      });
    }
    return SortData(result, 'value', 'descend');
  } catch (err) {
    console.log(err);
    return [];
  }
};

export const getTableColumns = (
  data,
  breakdown,
  arrayMapper,
  currentSorter,
  handleSorting
) => {
  const breakdownIndices = getBreakdownIndices(data, breakdown);
  const breakdownCols = breakdownIndices.map((b, index) => {
    return {
      title: data.result_group[1].headers[b],
      dataIndex: data.result_group[1].headers[b],
      fixed: index < 2 ? 'left' : '',
      width: 150,
    };
  });
  const eventCols = arrayMapper.map((elem) => {
    return {
      title: getTitleWithSorter(
        elem.eventName,
        elem.eventName,
        currentSorter,
        handleSorting
      ),
      dataIndex: elem.eventName,
      width: 150,
      render: (d) => {
        return <NumFormat number={d} />;
      },
    };
  });
  return [...breakdownCols, ...eventCols];
};

export const getTableData = (
  data,
  breakdown,
  currentEventIndex,
  arrayMapper,
  currentSorter,
  searchText
) => {
  const breakdownIndices = getBreakdownIndices(data, breakdown);
  const currEventName = arrayMapper.find(
    (elem) => elem.index === currentEventIndex
  ).eventName;
  const filteredRows = data.result_group[1].rows.filter((row) =>
    row[0].toString().toLowerCase().includes(searchText.toLowerCase())
  );
  const result = filteredRows.map((d, index) => {
    const breakdownVals = {};
    breakdownIndices.forEach((b) => {
      const dataIndex = data.result_group[1].headers[b];
      breakdownVals[dataIndex] = d[b];
    });
    const eventVals = {};
    arrayMapper.forEach((elem) => {
      const currDataIndex = data.result_group[1].headers.findIndex(
        (header) => header === elem.eventName
      );
      eventVals[elem.eventName] = d[currDataIndex];
    });
    return {
      ...breakdownVals,
      index,
      ...eventVals,
    };
  });
  if (!currentSorter.key) {
    return SortData(result, currEventName, 'descend');
  }
  return SortData(result, currentSorter.key, currentSorter.order);
};

export const formatDataInLineChartFormat = (
  visibleProperties,
  data,
  breakdown,
  currentEventIndex,
  arrayMapper,
  breakdownMapper
) => {
  const currEventName = arrayMapper.find(
    (elem) => elem.index === currentEventIndex
  ).eventName;
  const currDataIndex = data.result_group[0].headers.findIndex(
    (elem) => elem === currEventName
  );
  const format = 'YYYY-MM-DD HH-mm';
  let dates = new Set();
  const dateTimeIndex = data.result_group[0].headers.indexOf('datetime');
  data.result_group[0].rows.forEach((row) => {
    dates.add(moment(row[dateTimeIndex]).format(format));
  });
  dates = Array.from(dates);
  const xDates = ['x', ...dates];
  const result = visibleProperties.map((v) => {
    const dateBreakdownIndices = getDateBreakdownIndices(data, breakdown);
    const breakdownRows = data.result_group[0].rows.filter((row) => {
      const dateLabel = [];
      dateBreakdownIndices.forEach((b) => {
        if (b > -1) {
          dateLabel.push(row[b]);
        }
      });
      return dateLabel.join(', ') === v.label;
    });
    const breakdownLabel = breakdownMapper.find(
      (elem) => elem.eventName === v.label
    ).mapper;
    const values = [breakdownLabel];
    dates.forEach((d) => {
      const idx = breakdownRows.findIndex(
        (bRow) => moment(bRow[dateTimeIndex]).format(format) === d
      );
      if (idx > -1) {
        values.push(breakdownRows[idx][currDataIndex]);
      } else {
        values.push(0);
      }
    });
    return values;
  });
  return [xDates, ...result];
};

export const formatDataInHighChartsFormat = (
  data,
  arrayMapper,
  currentEventIndex,
  visibleProperties
) => {
  if (
    !data.headers ||
    !data.headers.length ||
    !data.rows ||
    !data.rows.length
  ) {
    return {
      categories: [],
      highchartsData: [],
    };
  }
  const colors = generateColors(visibleProperties.length);
  const event = arrayMapper.find((elem) => elem.index === currentEventIndex)
    .eventName;
  const eventIndex = data.headers.findIndex((h) => h === event);
  const dateIndex = data.headers.findIndex((h) => h === 'datetime');
  let differentDates = new Set();
  data.rows.forEach((row) => {
    differentDates.add(row[dateIndex]);
  });
  differentDates = Array.from(differentDates);
  const resultantData = visibleProperties.map((property, index) => {
    const initialData = differentDates.map(() => {
      return 0;
    });
    return {
      name: property.label,
      data: initialData,
      color: colors[index],
      marker: {
        enabled: false,
      },
    };
  });
  data.rows.forEach((row) => {
    const breakdownJoin = row.slice(0, dateIndex).join(', ');
    const bIdx = resultantData.findIndex((d) => d.name === breakdownJoin);
    if (bIdx > -1) {
      const idx = differentDates.indexOf(row[dateIndex]);
      resultantData[bIdx].data[idx] = row[eventIndex];
    }
  });
  return {
    categories: differentDates,
    highchartsData: resultantData,
  };
};
