import React from "react";
import moment from "moment";
import { SortData, getTitleWithSorter } from "../../../../utils/dataFormatter";
import { Number as NumFormat } from "../../../../components/factorsComponents";

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
  if (currentSorter.key) {
    const sortMapper = arrayMapper.find(
      (elem) => elem.eventName === currentSorter.key
    );
    if (sortMapper) {
      return SortData(result, sortMapper.mapper, currentSorter.order);
    }
  }
  return SortData(result, currentSorter.key, currentSorter.order);
};

export const getColumns = (
  events,
  arrayMapper,
  frequency,
  currentSorter,
  handleSorting
) => {
  let format = "MMM D, YYYY";
  if (frequency === "hour") {
    format = "h A, MMM D";
  }
  const result = [
    {
      title: "",
      dataIndex: "",
      width: 37,
    },
    {
      title: getTitleWithSorter("Date", "date", currentSorter, handleSorting),
      dataIndex: "date",
      render: (d) => {
        return moment(d).format(format);
      },
    },
  ];

  const eventColumns = events.map((e, idx) => {
    return {
      title: getTitleWithSorter(e, e, currentSorter, handleSorting),
      dataIndex: arrayMapper.find((elem) => elem.index === idx).mapper,
      render: (d) => {
        return <NumFormat number={d} />;
      },
    };
  });
  return [...result, ...eventColumns];
};

export const formatSingleEventAnalyticsData = (response, arrayMapper) => {
  const result = response.rows.map((row) => {
    const key = arrayMapper[0].mapper;
    return {
      date: new Date(row[0]),
      [key]: row[1],
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

export const getDataInLineChartFormat = (
  data,
  queries,
  hiddenEvents = [],
  arrayMapper
) => {
  data.sort((a, b) => {
    return moment(a.date).utc().unix() > moment(b.date).utc().unix() ? 1 : -1;
  });
  const result = [];
  const hashedData = {};
  const format = "YYYY-MM-DD HH-mm";
  hashedData.x = data.map((elem) => {
    return moment(elem.date).format(format);
  });
  queries.forEach((q, index) => {
    if (hiddenEvents.indexOf(q) === -1) {
      const key = arrayMapper.find((m) => m.index === index).mapper;
      hashedData[key] = data.map((elem) => {
        return elem[key];
      });
    }
  });
  for (const obj in hashedData) {
    result.push([obj, ...hashedData[obj]]);
  }
  return result;
};

export const getDateBasedColumns = (
  data,
  currentSorter,
  handleSorting,
  frequency
) => {
  const result = [
    {
      title: "Events",
      dataIndex: "event",
      fixed: "left",
      width: 200,
    },
  ];
  let format = "MMM D";
  if (frequency === "hour") {
    format = "h A, MMM D";
  }

  const dateColumns = data.map((elem) => {
    return {
      title: getTitleWithSorter(
        moment(elem.date).format(format),
        moment(elem.date).format(format),
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
  let format = "MMM D";
  if (frequency === "hour") {
    format = "h A, MMM D";
  }
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

  return SortData(result, currentSorter.key, currentSorter.order);
};
