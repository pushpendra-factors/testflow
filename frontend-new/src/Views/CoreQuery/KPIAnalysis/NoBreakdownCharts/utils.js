import React from 'react';
import { DATE_FORMATS } from '../../../../utils/constants';
import { Number as NumFormat } from '../../../../components/factorsComponents';
import moment from 'moment';
import {
  getClickableTitleSorter,
  SortResults,
} from '../../../../utils/dataFormatter';

export const getDefaultSortProp = (queries) => {
  if (Array.isArray(queries) && queries.length) {
    return {
      key: `${queries[0]} - 0`,
      type: 'numerical',
      subtype: null,
      order: 'descend',
    };
  }
  return {};
};

export const getDefaultDateSortProp = () => {
  return {
    key: 'Overall',
    type: 'numerical',
    subtype: null,
    order: 'descend',
  };
};

export const formatData = (data, queries) => {
  try {
    const result = queries.map((query, index) => {
      const totalIndex = index * 2;
      const dateSplitIndex = totalIndex + 1;
      const obj = {
        index,
        name: query,
      };
      if (data[totalIndex] && data[dateSplitIndex]) {
        const aggregateIndex = data[totalIndex].headers.findIndex(
          (h) => h === 'aggregate'
        );
        const dateIndex = data[dateSplitIndex].headers.findIndex(
          (h) => h === 'datetime'
        );
        const queryIndex = data[dateSplitIndex].headers.findIndex(
          (h) => h === 'aggregate'
        );
        return {
          ...obj,
          total: data[totalIndex].rows.length
            ? data[totalIndex].rows[0][aggregateIndex]
            : 0,
          dataOverTime: data[dateSplitIndex].rows.map((row) => {
            return {
              date: new Date(row[dateIndex]),
              [query]: row[queryIndex],
            };
          }),
        };
      } else {
        return {
          ...obj,
          total: 0,
        };
      }
    });
    return result;
  } catch (err) {
    console.log('formatData -> err', err);
    return [];
  }
};

export const formatDataInSeriesFormat = (aggData) => {
  try {
    const differentDates = new Set();
    aggData.forEach((d) => {
      d.dataOverTime.forEach((elem) => {
        differentDates.add(new Date(elem.date).getTime());
      });
    });
    const categories = Array.from(differentDates);
    const initializedDatesData = categories.map(() => {
      return 0;
    });
    const data = aggData.map((m) => {
      return {
        index: m.index,
        name: m.name,
        data: [...initializedDatesData],
        marker: {
          enabled: false,
        },
        total: m.total,
      };
    });
    aggData.forEach((m, index) => {
      categories.forEach((cat, catIndex) => {
        const dateIndex = m.dataOverTime.findIndex(
          (elem) => new Date(elem.date).getTime() === cat
        );
        if (dateIndex > -1) {
          data[index].data[catIndex] = m.dataOverTime[dateIndex][m.name];
        }
      });
    });
    return {
      categories,
      data,
    };
  } catch (err) {
    return {
      categories: [],
      data: [],
    };
  }
};

export const getTableColumns = (
  queries,
  currentSorter,
  handleSorting,
  eventNames,
  frequency
) => {
  const format = DATE_FORMATS[frequency] || DATE_FORMATS['date'];
  const result = [
    {
      title: '',
      dataIndex: '',
      width: 37,
    },
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
  ];
  const eventColumns = queries.map((e, idx) => {
    return {
      title: getClickableTitleSorter(
        eventNames[e] || e,
        {
          key: `${e} - ${idx}`,
          type: 'numerical',
          subtype: null,
        },
        currentSorter,
        handleSorting
      ),
      dataIndex: `${e} - ${idx}`,
      render: (d) => {
        return <NumFormat number={d} />;
      },
    };
  });
  return [...result, ...eventColumns];
};

export const getDataInTableFormat = (
  data,
  categories,
  queries,
  currentSorter
) => {
  const result = categories.map((cat, catIndex) => {
    const obj = {
      index: catIndex,
      date: cat,
    };
    queries.forEach((q, qIndex) => {
      obj[`${q} - ${qIndex}`] = data[qIndex].data[catIndex];
    });
    return obj;
  });
  return SortResults(result, currentSorter);
};

export const getDateBasedColumns = (
  categories,
  currentSorter,
  handleSorting,
  eventNames,
  frequency
) => {
  const OverallColumn = {
    title: getClickableTitleSorter(
      'Overall',
      { key: `Overall`, type: 'numerical', subtype: null },
      currentSorter,
      handleSorting
    ),
    dataIndex: `Overall`,
    width: 150,
  };
  const result = [
    {
      title: getClickableTitleSorter(
        'Event',
        {
          key: 'event',
          type: 'categorical',
          subtype: null,
        },
        currentSorter,
        handleSorting
      ),
      dataIndex: 'event',
      fixed: 'left',
      width: 200,
      render: (d) => {
        return eventNames[d] || d;
      },
    },
  ];
  const format = DATE_FORMATS[frequency] || DATE_FORMATS['date'];
  const dateColumns = categories.map((cat) => {
    return {
      title: getClickableTitleSorter(
        moment(cat).format(format),
        {
          key: moment(cat).format(format),
          type: 'numerical',
          subtype: null,
        },
        currentSorter,
        handleSorting
      ),
      width: frequency === 'hour' ? 150 : 100,
      dataIndex: moment(cat).format(format),
      render: (d) => {
        return <NumFormat number={d} />;
      },
    };
  });
  return [...result, ...dateColumns, OverallColumn];
};

export const getDateBasedTableData = (
  seriesData,
  categories,
  searchText,
  currentSorter,
  frequency
) => {
  const format = DATE_FORMATS[frequency] || DATE_FORMATS['date'];
  const result = seriesData.map((sd, index) => {
    const obj = {
      index,
      event: sd.name,
      Overall: sd.total,
    };
    const dateData = {};
    categories.forEach((cat, catIndex) => {
      dateData[moment(cat).format(format)] = sd.data[catIndex];
    });
    return {
      ...obj,
      ...dateData,
    };
  });
  const filteredResult = result.filter(
    (elem) => elem.event.toLowerCase().indexOf(searchText.toLowerCase()) > -1
  );
  return SortResults(filteredResult, currentSorter);
};
