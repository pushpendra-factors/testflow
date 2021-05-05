import React from 'react';
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

  const result = data.result_group[1].rows.map((d, index) => {
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

  const filteredResult = result.filter((r) => {
    for (let key in r) {
      try {
        return r[key]
          .toString()
          .toLowerCase()
          .includes(searchText.toLowerCase());
      } catch (err) {
        console.log(err);
        return false;
      }
    }
  });

  if (!currentSorter.key) {
    return SortData(filteredResult, currEventName, 'descend');
  }
  return SortData(filteredResult, currentSorter.key, currentSorter.order);
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
