import moment from "moment";
import { getTitleWithSorter, SortData } from "../../../../utils/dataFormatter";

export const formatData = (data, arrayMapper) => {
  const result = [];
  arrayMapper.forEach((elem) => {
    const dateTimeIndex = data.result_group[0].headers.indexOf("datetime");
    const dateTimeEventIndex = data.result_group[0].headers.indexOf(
      elem.eventName
    );
    const eventIndex = data.result_group[1].headers.indexOf(elem.eventName);
    if (
      dateTimeEventIndex > -1 &&
      eventIndex > -1 &&
      data.result_group[1].rows[eventIndex]
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
        total: data.result_group[1].rows[eventIndex],
      });
    }
  });
  return result;
};

export const formatDataInLineChartFormat = (chartsData) => {
  const result = [];
  const format = "YYYY-MM-DD HH-mm";
  const dates = chartsData[0].dataOverTime.map((d) =>
    moment(d.date).format(format)
  );
  result.push(["x", ...dates]);
  chartsData.forEach((d) => {
    result.push([d.mapper, ...d.dataOverTime.map((elem) => elem[d.mapper])]);
  });
  return result;
};

export const getTableColumns = (chartsData, currentSorter, handleSorting) => {
  const result = chartsData.map((elem) => {
    return {
      title: getTitleWithSorter(
        elem.name,
        elem.name,
        currentSorter,
        handleSorting
      ),
      dataIndex: elem.name,
    };
  });
  return [
    {
      title: "Date",
      dataIndex: "date",
    },
    ...result,
  ];
};

export const getTableData = (chartsData, frequency, currentSorter) => {
  let format = "MMM D, YYYY";
  if (frequency === "hour") {
    format = "h A, MMM D";
  }
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
      date: moment(date).format(format),
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
  let format = "MMM D";
  if (frequency === "hour") {
    format = "h A, MMM D";
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
    };
  });
  return [
    {
      title: "Measures",
      dataIndex: "measures",
    },
    ...dateColumns,
  ];
};

export const getDateBasedTableData = (chartsData, frequency, currentSorter) => {
	console.log(chartsData);
  let format = "MMM D";
  if (frequency === "hour") {
    format = "h A, MMM D";
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
