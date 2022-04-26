import React from 'react';
import MomentTz from 'Components/MomentTz';
import {
  addQforQuarter,
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

export const getDefaultSortProp = (arrayMapper) => {
  if (Array.isArray(arrayMapper) && arrayMapper.length) {
    return [
      {
        key: arrayMapper[0]?.mapper,
        type: 'numerical',
        subtype: null,
        order: 'descend',
      },
    ];
  }
  return [];
};

export const getDefaultDateSortProp = () => {
  return [
    {
      key: 'Overall',
      type: 'numerical',
      subtype: null,
      order: 'descend',
    },
  ];
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
      title: getClickableTitleSorter(
        'Date',
        { key: 'date', type: 'datetime', subtype: 'date' },
        currentSorter,
        handleSorting
      ),
      dataIndex: 'date',
      render: (d) => {
        return addQforQuarter(frequency) + MomentTz(d).format(format);
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
        handleSorting,
        'right'
      ),
      className: 'text-right',
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
  const resultantData = arrayMapper.map((m) => {
    return {
      name: m.displayName
        ? m.displayName
        : eventNames[m.eventName] || m.eventName,
      data: [...initializedDatesData],
      index: m.index,
      marker: {
        enabled: false,
      },
    };
  });

  data.rows.forEach((row) => {
    const idx = differentDates.indexOf(row[dateIndex]);
    arrayMapper.forEach((_, index) => {
      resultantData[index].data[idx] = row[dateIndex + index + 1];
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
  const OverallColumn = {
    title: getClickableTitleSorter(
      'Overall',
      { key: `Overall`, type: 'numerical', subtype: null },
      currentSorter,
      handleSorting,
      'right'
    ),
    dataIndex: `Overall`,
    className: 'text-right',
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

  const dateColumns = data.map((elem) => {
    return {
      title: getClickableTitleSorter(
        addQforQuarter(frequency) + MomentTz(elem.date).format(format),
        {
          key: addQforQuarter(frequency) + MomentTz(elem.date).format(format),
          type: 'numerical',
          subtype: null,
        },
        currentSorter,
        handleSorting,
        'right'
      ),
      width: frequency === 'hour' ? 200 : 150,
      dataIndex: addQforQuarter(frequency) + MomentTz(elem.date).format(format),
      className: 'text-right',
      render: (d) => {
        return <NumFormat number={d} />;
      },
    };
  });
  return [...result, ...dateColumns, OverallColumn];
};

export const getNoGroupingTablularDatesBasedData = (
  data,
  currentSorter,
  searchText,
  arrayMapper,
  frequency,
  metrics
) => {
  const filteredMapper = arrayMapper.filter((elem) =>
    elem.eventName.toLowerCase().includes(searchText.toLowerCase())
  );
  const format = DATE_FORMATS[frequency] || DATE_FORMATS['date'];
  const dates = data.map(
    (elem) => addQforQuarter(frequency) + MomentTz(elem.date).format(format)
  );
  const result = filteredMapper.map((elem, index) => {
    let total = 0;
    if (
      metrics &&
      metrics.headers &&
      Array.isArray(metrics.headers) &&
      metrics.headers.length &&
      metrics.rows &&
      Array.isArray(metrics.rows) &&
      metrics.rows.length
    ) {
      const countIdx = metrics.headers.findIndex(
        (h) => h === 'count' || h === 'aggregate'
      );
      const event_indexIdx = metrics.headers.findIndex(
        (h) => h === 'event_index'
      );
      const metricRow = metrics.rows.find(
        (mr) => mr[event_indexIdx] === elem.index
      );
      total = metricRow ? metricRow[countIdx] : 0;
    }
    const eventsData = {};
    dates.forEach((date) => {
      eventsData[date] = data.find(
        (d) =>
          addQforQuarter(frequency) + MomentTz(d.date).format(format) === date
      )[elem.mapper];
    });
    return {
      index,
      event: elem.eventName,
      Overall: total,
      ...eventsData,
    };
  });

  return SortResults(result, currentSorter, currentSorter.order);
};
