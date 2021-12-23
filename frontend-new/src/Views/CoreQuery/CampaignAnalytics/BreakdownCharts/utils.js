import React from 'react';
import moment from 'moment';
import {
  getClickableTitleSorter,
  SortResults,
} from '../../../../utils/dataFormatter';
import { Number as NumFormat } from '../../../../components/factorsComponents';

export const getDefaultSorterState = (arrayMapper, currentEventIndex) => {
  return [
    {
      key: arrayMapper[currentEventIndex].eventName,
      type: 'numerical',
      subtype: null,
      order: 'descend',
    },
  ];
};

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

export const formatData = (data, arrayMapper, breakdown) => {
  if (
    !data ||
    !Array.isArray(data.result_group) ||
    data.result_group.length < 2 ||
    !Array.isArray(data.result_group[1].headers) ||
    !data.result_group[1].headers.length ||
    !Array.isArray(data.result_group[1].rows) ||
    !data.result_group[1].rows.length
  ) {
    return [];
  }
  console.log('campaigns format data');
  try {
    const { headers, rows } = data.result_group[1];
    const breakdownIndices = getBreakdownIndices(data, breakdown);
    let result = [];
    rows.forEach((elem, index) => {
      const label = [];
      breakdownIndices.forEach((b) => {
        if (b > -1) {
          label.push(elem[b]);
        }
      });
      const measures = {};
      arrayMapper.forEach((mapper) => {
        const currDataIndex = headers.findIndex(
          (elem) => elem === mapper.eventName
        );
        measures[mapper.eventName] = elem[currDataIndex];
      });
      result.push({
        index,
        label: label.join(', '),
        ...measures,
      });
    });
    return result;
  } catch (err) {
    console.log(err);
    return [];
  }
};

export const getTableColumns = (
  arrayMapper,
  breakdown,
  currentSorter,
  handleSorting
) => {
  console.log('campaigns getTableColumns');
  const breakdownCols = breakdown.map((b, index) => {
    return {
      title: getClickableTitleSorter(
        `${b.name}_${b.property}`,
        {
          key: `${b.name}_${b.property}-${index}`,
          type: 'categorical',
          subtype: null,
        },
        currentSorter,
        handleSorting
      ),
      dataIndex: `${b.name}_${b.property}-${index}`,
      fixed: !index ? 'left' : '',
      width: 150,
    };
  });

  const eventCols = arrayMapper.map((elem) => {
    return {
      title: getClickableTitleSorter(
        elem.eventName,
        { key: elem.eventName, type: 'numerical', subtype: null },
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

export const getTableData = (data, breakdown, searchText, currentSorter) => {
  console.log('campaigns getTableData');
  const filteredData = data.filter(
    (d) => d.label.toLowerCase().indexOf(searchText.toLowerCase()) > -1
  );
  const result = filteredData.map(({ label, ...rest }) => {
    const breakdownVals = {};
    const splittedLabel = label.split(', ');
    breakdown.forEach((b, index) => {
      breakdownVals[`${b.name}_${b.property}-${index}`] = splittedLabel[index];
    });
    return {
      ...rest,
      ...breakdownVals,
    };
  });
  return SortResults(result, currentSorter);
};

export const formatDataInHighChartsFormat = (
  data,
  arrayMapper,
  aggregateData
) => {
  if (
    !data ||
    !Array.isArray(data.headers) ||
    !data.headers.length ||
    !Array.isArray(data.rows) ||
    !data.rows.length ||
    !aggregateData.length
  ) {
    return {
      categories: [],
      highchartsData: [],
    };
  }
  console.log('campaigns formatDataInHighChartsFormat');
  const dateIndex = data.headers.findIndex((h) => h === 'datetime');
  let differentDates = new Set();
  data.rows.forEach((row) => {
    differentDates.add(row[dateIndex]);
  });
  differentDates = Array.from(differentDates);
  const initialData = differentDates.map(() => {
    return 0;
  });
  const resultantData = aggregateData.map((property, index) => {
    const measures = {};
    arrayMapper.forEach((mapper) => {
      measures[mapper.eventName] = [...initialData];
    });
    return {
      index: property.index,
      name: property.label,
      ...measures,
      marker: {
        enabled: false,
      },
    };
  });
  data.rows.forEach((row) => {
    const breakdownJoin = row.slice(0, dateIndex).join(', ');
    const bIdx = resultantData.findIndex((d) => d.name === breakdownJoin);
    const idx = differentDates.indexOf(row[dateIndex]);
    if (resultantData[bIdx]) {
      arrayMapper.forEach((mapper) => {
        const currDataIndex = data.headers.findIndex(
          (elem) => elem === mapper.eventName
        );
        resultantData[bIdx][mapper.eventName][idx] = row[currDataIndex];
      });
    }
  });
  return {
    categories: differentDates,
    highchartsData: resultantData,
  };
};

export const getDateBasedColumns = (
  categories,
  breakdown,
  currentSorter,
  handleSorting
) => {
  console.log('campaigns getDateBasedColumns');
  const breakdownCols = breakdown.map((b, index) => {
    return {
      title: getClickableTitleSorter(
        `${b.name}_${b.property}`,
        {
          key: `${b.name}_${b.property}-${index}`,
          type: 'categorical',
          subtype: null,
        },
        currentSorter,
        handleSorting
      ),
      dataIndex: `${b.name}_${b.property}-${index}`,
      fixed: !index ? 'left' : '',
      width: 150,
    };
  });
  const format = 'MMM D';
  const dateColumns = categories.map((cat) => {
    return {
      title: getClickableTitleSorter(
        moment(cat).format(format),
        { key: moment(cat).format(format), type: 'numerical', subtype: null },
        currentSorter,
        handleSorting
      ),
      width: 100,
      dataIndex: moment(cat).format(format),
      render: (d) => {
        return <NumFormat number={d} />;
      },
    };
  });
  return [...breakdownCols, ...dateColumns];
};

export const getDateBasedTableData = (
  seriesData,
  categories,
  breakdown,
  searchText,
  currentSorter,
  arrayMapper,
  currentEventIndex
) => {
  console.log('campaigns getDateBasedTableData');
  const format = 'MMM D';
  const currentEventName = arrayMapper[currentEventIndex].eventName;
  const result = seriesData
    .filter((sd) => sd.name.toLowerCase().includes(searchText.toLowerCase()))
    .map((sd) => {
      const dateWiseData = {};
      categories.forEach((cat, index) => {
        dateWiseData[moment(cat).format(format)] = sd[currentEventName][index];
      });
      const splittedLabel = sd.name.split(',');
      const breakdownData = {};
      breakdown.forEach((b, index) => {
        breakdownData[`${b.name}_${b.property}-${index}`] =
          splittedLabel[index];
      });
      return {
        index: sd.index,
        ...breakdownData,
        ...dateWiseData,
      };
    });
  return SortResults(result, currentSorter);
};
