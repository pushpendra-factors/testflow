import React from 'react';
import moment from 'moment';
import {
  getClickableTitleSorter,
  SortResults,
} from '../../../../utils/dataFormatter';
import { Number as NumFormat } from '../../../../components/factorsComponents';
import { DATE_FORMATS } from '../../../../utils/constants';

export const getNoGroupingTableData = (data, arrayMapper, currentSorter) => {
  const clonedData = data.map((elem) => {
    const element = { ...elem };
    return element;
  });

  const result = clonedData.map((elem, index) => {
    return {
      index,
      ...elem,
      date: elem.date,
    };
  });

  return SortResults(result, currentSorter);
};

export const getColumns = (
  events,
  arrayMapper,
  frequency,
  currentSorter,
  handleSorting,
  eventNames
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

  const eventColumns = events.map((e, idx) => {
    return {
      title: getClickableTitleSorter(
        eventNames[e] || e,
        {
          key: arrayMapper.find((elem) => elem.index === idx).mapper,
          type: 'numerical',
          subtype: null,
        },
        currentSorter,
        handleSorting
      ),
      dataIndex: arrayMapper.find((elem) => elem.index === idx).mapper,
      render: (d) => {
        return <NumFormat number={d} />;
      },
    };
  });
  return [...result, ...eventColumns];
};

export const formatData = (response, arrayMapper, noOfQueries) => {
  if (noOfQueries > 1) {
    return formatMultiEventsAnalyticsData(response, arrayMapper);
  } else {
    return formatSingleEventAnalyticsData(response, arrayMapper);
  }
};

export const formatSingleEventAnalyticsData = (response, arrayMapper) => {
  if (
    !response.headers ||
    !response.headers.length ||
    !response.rows ||
    !response.rows.length
  ) {
    return [];
  }
  const { headers } = response;
  const dateIndex = headers.findIndex((h) => h === 'datetime');
  const result = response.rows.map((row) => {
    const key = arrayMapper[0].mapper;
    return {
      date: new Date(row[dateIndex]),
      [key]: row[dateIndex + 1],
    };
  });
  return result;
};

export const formatMultiEventsAnalyticsData = (response, arrayMapper) => {
  const result = [];
  response.rows.forEach((r) => {
    const eventsData = {};
    response.headers.slice(1).forEach((_, index) => {
      const key = arrayMapper.find((m) => m.index === index).mapper;
      eventsData[key] = r[index + 1];
    });
    result.push({
      date: new Date(r[0]),
      ...eventsData,
    });
  });
  return result;
};

export const getDataInLineChartFormat = (data, arrayMapper, eventNames) => {
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
  const { headers } = data;
  const dateIndex = headers.findIndex((h) => h === 'datetime');
  let differentDates = new Set();
  data.rows.forEach((row) => {
    differentDates.add(row[dateIndex]);
  });
  differentDates = Array.from(differentDates);
  const initializedDatesData = differentDates.map(() => {
    return 0;
  });
  const eventIndices = [];
  const resultantData = arrayMapper.map((m) => {
    eventIndices.push(headers.findIndex((header) => m.eventName === header));
    return {
      name: eventNames[m.eventName] || m.eventName,
      data: [...initializedDatesData],
      index: m.index,
      marker: {
        enabled: false,
      },
    };
  });

  data.rows.forEach((row) => {
    const idx = differentDates.indexOf(row[dateIndex]);
    eventIndices.forEach((valIndex, index) => {
      if (valIndex > -1) {
        resultantData[index].data[idx] = row[valIndex];
      }
    });
  });
  return {
    categories: differentDates,
    data: resultantData,
  };
};

export const getDateBasedColumns = (
  data,
  currentSorter,
  handleSorting,
  frequency,
  eventNames
) => {
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

  const dateColumns = data.map((elem) => {
    return {
      title: getClickableTitleSorter(
        moment(elem.date).format(format),
        {
          key: moment(elem.date).format(format),
          type: 'numerical',
          subtype: null,
        },
        currentSorter,
        handleSorting
      ),
      width: 100,
      dataIndex: moment(elem.date).format(format),
      render: (d) => {
        return <NumFormat number={d} />;
      },
    };
  });
  return [...result, ...dateColumns];
};

export const getNoGroupingTablularDatesBasedData = (
  data,
  currentSorter,
  searchText,
  arrayMapper,
  frequency
) => {
  const filteredEvents = arrayMapper
    .filter((elem) =>
      elem.eventName.toLowerCase().includes(searchText.toLowerCase())
    )
    .map((elem) => elem.mapper);
  const format = DATE_FORMATS[frequency] || DATE_FORMATS['date'];
  const dates = data.map((elem) => moment(elem.date).format(format));
  const result = filteredEvents.map((elem, index) => {
    const eventsData = {};
    dates.forEach((date) => {
      eventsData[date] = data.find(
        (elem) => moment(elem.date).format(format) === date
      )[elem];
    });
    return {
      index,
      event: arrayMapper.find((m) => m.mapper === elem).eventName,
      ...eventsData,
    };
  });

  return SortResults(result, currentSorter, currentSorter.order);
};
