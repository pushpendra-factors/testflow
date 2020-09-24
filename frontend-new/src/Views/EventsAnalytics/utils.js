import moment from 'moment';

import { getTitleWithSorter } from '../CoreQuery/FunnelsResultPage/utils';
import { singleEventResponse, multiEventResponse } from './SampleResponse';

export const getNoGroupingTableData = (data, currentSorter, searchText, reverseEventsMapper) => {
  const clonedData = data.map(elem => {
    const element = { ...elem }
    for (let key in element) {
      if (key !== 'date') {
        element[reverseEventsMapper[key]] = element[key];
        delete element[key];
      }
    }
    return element;
  });

  const result = clonedData.map((elem, index) => {
    return {
      index,
      ...elem,
      date: moment(elem.date).format('MMM D, YYYY')
    };
  });

  result.sort((a, b) => {
    if (currentSorter.order === 'ascend') {
      return parseInt(a[currentSorter.key]) >= parseInt(b[currentSorter.key]) ? 1 : -1;
    }
    if (currentSorter.order === 'descend') {
      return parseInt(a[currentSorter.key]) <= parseInt(b[currentSorter.key]) ? 1 : -1;
    }
    return 0;
  });
  return result;
};

export const getColumns = (events, currentSorter, handleSorting) => {
  const result = [
    {
      title: '',
      dataIndex: '',
      width: 37
    },
    {
      title: 'Date',
      dataIndex: 'date'
    }];

  const eventColumns = events.map(e => {
    return {
      title: getTitleWithSorter(e, e, currentSorter, handleSorting),
      dataIndex: e
    };
  });
  return [...result, ...eventColumns];
};

export const getSingleEventAnalyticsData = (event, eventsMapper) => {
  const response = singleEventResponse;
  const result = response.rows.map(row => {
    return {
      date: new Date(row[0]),
      [eventsMapper[event]]: row[1]
    }
  });
  return result;
};

export const getMultiEventsAnalyticsData = (queries, eventsMapper) => {
  const response = multiEventResponse;
  const dates = [];
  const result = [];
  response.rows.forEach(r => {
    if (dates.indexOf(r[0]) === -1) {
      const currentDateData = response.rows.filter(elem => elem[0] === r[0]);
      const eventsData = {}
      currentDateData.forEach(d => {
        const query = queries.find(q => q === d[1]);
        eventsData[eventsMapper[query]] = d[2];
      });
      result.push({
        date: new Date(r[0]),
        ...eventsData
      })
      dates.push(r[0]);
    }
  });
  return result;
};

export const getDataInLineChartFormat = (data, queries, eventsMapper) => {
  data.sort((a, b) => {
    return moment(a["date"]).utc().unix() > moment(b["date"]).utc().unix() ? 1 : -1;
  });
  const result = [];
  const hashedData = {};
  hashedData["x"] = data.map(elem => {
    return moment(elem["date"]).format("YYYY-MM-DD");
  });
  queries.forEach(q => {
    hashedData[eventsMapper[q]] = data.map(elem => {
      return elem[eventsMapper[q]];
    })
  })
  for (let obj in hashedData) {
    result.push([obj, ...hashedData[obj]]);
  }
  return result;
}