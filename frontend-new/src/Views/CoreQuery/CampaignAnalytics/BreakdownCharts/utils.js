import moment from "moment";
import {
  generateColors,
  SortData,
  getTitleWithSorter,
} from "../../../../utils/dataFormatter";

export const getBreakdownIndices = (data, breakdown) => {
  const result = breakdown.map((elem) => {
    const str = elem.name + "_" + elem.property;
    return data.result_group[1].headers.findIndex((elem) => elem === str);
  });
  return result;
};

export const getDateBreakdownIndices = (data, breakdown) => {
  const result = breakdown.map((elem) => {
    const str = elem.name + "_" + elem.property;
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
          label: label.join(", "),
          value: elem[currDataIndex],
          color: colors[currentEventIndex],
        });
        // if (elem[currDataIndex]) {
        //   const dateLabel = [];
        //   const dataOverTime = [];
        //   // dateRows.forEach((row) => {
        //   //   if (!row.is_done) {
        //   //     dateBreakdownIndices.forEach((b) => {
        //   //       if (b > -1) {
        //   //         dateLabel.push(row[b]);
        //   //       }
        //   //     });
        //   //     if (label.join(", ") === dateLabel.join(", ")) {
        //   //       dataOverTime.push(row);
        //   //       row.is_done = true;
        //   //     }
        //   //   }
        //   // });
        //   console.log(elem[currDataIndex])
        //   console.log(dataOverTime)

        // } else {
        //   result.push({
        //     index,
        //     label: label.join(", "),
        //     value: elem[currDataIndex],
        //     color: colors[currentEventIndex],
        //     dataOverTime: [],
        //   });
        // }
      });
    }
    return SortData(result, "value", "descend");
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
  const breakdownCols = breakdownIndices.map((b) => {
    return {
      title: data.result_group[1].headers[b],
      dataIndex: data.result_group[1].headers[b],
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
    };
  });
  return [...breakdownCols, ...eventCols];
};

export const getTableData = (
  data,
  breakdown,
  currentEventIndex,
  arrayMapper,
  currentSorter
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
  if (!currentSorter.key) {
    return SortData(result, currEventName, "descend");
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
  const format = "YYYY-MM-DD HH-mm";
  let dates = new Set();
  const dateTimeIndex = data.result_group[0].headers.indexOf("datetime");
  data.result_group[0].rows.forEach((row) => {
    dates.add(moment(row[dateTimeIndex]).format(format));
  });
  dates = Array.from(dates);
  const xDates = ["x", ...dates];
  const result = visibleProperties.map((v) => {
    const dateBreakdownIndices = getDateBreakdownIndices(data, breakdown);
    const breakdownRows = data.result_group[0].rows.filter((row) => {
      const dateLabel = [];
      dateBreakdownIndices.forEach((b) => {
        if (b > -1) {
          dateLabel.push(row[b]);
        }
      });
      return dateLabel.join(", ") === v.label;
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
