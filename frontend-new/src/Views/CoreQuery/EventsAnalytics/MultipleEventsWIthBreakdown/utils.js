import React from "react";
import moment from "moment";
import { labelsObj } from "../../utils";
import { SortData, getTitleWithSorter, generateColors } from "../../../../utils/dataFormatter";
import { Number as NumFormat } from "../../../../components/factorsComponents";

export const getBreakdownTitle = (breakdown) => {
  const charArr = ["A", "B", "C", "D", "E", "F", "G", "H"];
  if (!breakdown.eventIndex) {
    return breakdown.property;
  }
  return (
    <div className="flex items-center">
      <div className="mr-1">{breakdown.property} of </div>
      <div
        style={{ backgroundColor: "#3E516C" }}
        className="text-white w-4 h-4 flex justify-center items-center rounded-full font-semibold leading-5 text-xs"
      >
        {charArr[breakdown.eventIndex - 1]}
      </div>
    </div>
  );
};

export const formatData = (data, queries, colors) => {
  const splittedData = {};
  queries.forEach((query) => {
    splittedData[query] = [];
  });

  const result = data.metrics.rows.map((d, index) => {
    const str = d.slice(2, d.length - 1).join(",");
    const queryIndex = queries.findIndex((_, index) => index === d[0]);
    const dateRows = data.rows
      .filter((row) => {
        const rowStr = row.slice(3, row.length - 1).join(",");
        return row[0] === d[0] && rowStr === str;
      })
      .map((row) => {
        return {
          date: row[1],
          value: row[row.length - 1],
        };
      });
    return {
      label: str,
      value: d[d.length - 1],
      index,
      event: d[1],
      eventIndex: d[0],
      color: colors[queryIndex],
      dateWise: dateRows,
    };
  });

  const sortedData = SortData(result, "value", "descend");
  const maxIndices = [];
  queries.forEach((_, qIdx) => {
    const idx = sortedData.findIndex((elem) => elem.eventIndex === qIdx);
    if (idx > -1) {
      maxIndices.push(idx);
    }
  });
  const finalResult = maxIndices.map((m) => {
    return sortedData[m];
  });
  sortedData.forEach((sd, idx) => {
    if (maxIndices.indexOf(idx) === -1) {
      finalResult.push(sd);
    }
  });
  return finalResult;
};

export const formatVisibleProperties = (data, queries) => {
  const vp = data.map((d) => {
    return { ...d, label: `${d.label}; [${d.eventIndex}]` };
  });
  vp.sort((a, b) => {
    return parseInt(a.value) <= parseInt(b.value) ? 1 : -1;
  });
  vp.sort((a, b) => {
    const idx1 = queries.findIndex((_, index) => index === a.eventIndex);
    const idx2 = queries.findIndex((_, index) => index === b.eventIndex);
    return idx1 >= idx2 ? 1 : -1;
  });
  return vp;
};

export const getTableColumns = (
  breakdown,
  currentSorter,
  handleSorting,
  page
) => {
  const result = [];
  result.push({
    title: "Event",
    dataIndex: "event",
  });
  breakdown.forEach((b, index) => {
    result.push({
      title: getBreakdownTitle(b),
      dataIndex: b.property + ";" + index,
    });
  });
  result.push({
    title: getTitleWithSorter(
      labelsObj[page],
      "Event Count",
      currentSorter,
      handleSorting
    ),
    dataIndex: "Event Count",
    render: (d) => {
      return <NumFormat number={d} />;
    },
  });
  return result;
};

export const getTableData = (data, breakdown, searchText, currentSorter) => {
  const filteredData = data.filter(
    (elem) =>
      elem.label.toLowerCase().includes(searchText.toLowerCase()) ||
      elem.event.toLowerCase().includes(searchText.toLowerCase())
  );
  const result = [];
  filteredData.forEach((d) => {
    const breakdownValues = {};
    breakdown.forEach((b, index) => {
      breakdownValues[b.property + ";" + index] = d.label.split(",")[index];
    });
    result.push({
      ...d,
      "Event Count": d.value,
      ...breakdownValues,
    });
  });
  return SortData(result, currentSorter.key, currentSorter.order);
};

export const formatDataInLineChartFormat = (
  visibleProperties,
  arrayMapper,
  hiddenProperties
) => {
  const result = [];
  const format = "YYYY-MM-DD HH-mm";
  const dates = visibleProperties[0].dateWise.map((elem) =>
    moment(elem.date).format(format)
  );
  result.push(["x", ...dates]);
  visibleProperties.forEach((v, index) => {
    const label = `${v.event},${v.label}`;
    if (hiddenProperties.indexOf(label) === -1) {
      const values = v.dateWise.map((elem) => elem.value);
      const key = arrayMapper.find((m) => m.index === index).mapper;
      result.push([key, ...values]);
    }
  });
  return result;
};

export const getDateBasedColumns = (
  data,
  breakdown,
  currentSorter,
  handleSorting,
  frequency
) => {
  const breakdownColumns = breakdown.map((elem, index) => {
    return {
      title: getBreakdownTitle(elem),
      dataIndex: elem.property + ";" + index,
      fixed: "left",
      width: 200,
    };
  });
  let format = "MMM D";
  if (frequency === "hour") {
    format = "h A, MMM D";
  }
  const dateColumns = data[0].slice(1).map((elem) => {
    return {
      title: getTitleWithSorter(
        moment(elem).utc().format(format),
        moment(elem).utc().format(format),
        currentSorter,
        handleSorting
      ),
      width: 100,
      dataIndex: moment(elem).utc().format(format),
      render: (d) => {
        return <NumFormat number={d} />;
      },
    };
  });
  const eventCol = {
    title: "Event",
    dataIndex: "Event",
    fixed: "left",
    width: 200,
  };
  return [eventCol, ...breakdownColumns, ...dateColumns];
};

export const getDateBasedTableData = (
  data,
  breakdown,
  currentSorter,
  searchText,
  frequency
) => {
  const filteredData = data.filter(
    (elem) =>
      elem.label.toLowerCase().includes(searchText.toLowerCase()) ||
      elem.event.toLowerCase().includes(searchText.toLowerCase())
  );
  let format = "MMM D";
  if (frequency === "hour") {
    format = "h A, MMM D";
  }
  const result = filteredData.map((d) => {
    const breakdownValues = {};
    breakdown.forEach((b, index) => {
      breakdownValues[b.property + ";" + index] = d.label.split(",")[index];
    });

    const dateWiseValues = {};
    d.dateWise.forEach((w) => {
      const key = moment(w.date).format(format);
      dateWiseValues[key] = w.value;
    });
    return {
      index: d.index,
      Event: d.event,
      ...breakdownValues,
      ...dateWiseValues,
    };
  });
  return SortData(result, currentSorter.key, currentSorter.order);
};

export const formatDataInStackedAreaFormat = (
  data,
  visibleLabels,
  arrayMapper
) => {
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
  const colors = generateColors(visibleLabels.length);
  const dateIndex = data.headers.findIndex((h) => h === "datetime");
  const countIndex = data.headers.findIndex((h) => h === "count");
  const eventIndex = data.headers.findIndex((h) => h === "event_name");
  let differentDates = new Set();
  data.rows.forEach((row) => {
    differentDates.add(row[dateIndex]);
  });
  differentDates = Array.from(differentDates);
  const resultantData = visibleLabels.map((name, index) => {
    const data = differentDates.map(() => {
      return 0;
    });
    return {
      name,
      data,
      color: colors[index],
      marker: {
        enabled: false,
      },
    };
  });
  data.rows.forEach((row) => {
    const breakdownJoin = row.slice(eventIndex, countIndex).join(",");
    const bIdx = visibleLabels.indexOf(breakdownJoin);
    if (bIdx > -1) {
      const idx = differentDates.indexOf(row[dateIndex]);
      resultantData[bIdx].data[idx] = row[countIndex];
    }
  });
  return {
    categories: differentDates,
    data: resultantData,
  };
};
