import React from 'react';
import { formatCount } from '../../../utils/dataFormatter';
import { Number as NumFormat } from '../../../components/factorsComponents';
import NonClickableTableHeader from '../../../components/NonClickableTableHeader';

export const getWebAnalyticsTableData = (tableData, searchText) => {
  const { headers, rows } = tableData;
  const columns = headers.map((header) => {
    return {
      title: <NonClickableTableHeader title={header} />,
      dataIndex: header,
      render: (d) => {
        return isNaN(d) ? d : <NumFormat number={d} />;
      }
    };
  });

  const filteredRows = rows.filter((row) => {
    let isSearchTextAvailable = false;
    row.forEach((elem) => {
      try {
        if (elem.toString().toLowerCase().includes(searchText.toLowerCase())) {
          isSearchTextAvailable = true;
        }
      } catch (err) {
        console.log(err);
      }
    });
    return isSearchTextAvailable;
  });

  const data = filteredRows.map((row, idx) => {
    const rowData = {};
    row.forEach((elem, index) => {
      rowData[headers[index]] = isNaN(elem) ? elem : formatCount(elem, 1);
    });
    return { ...rowData, index: idx };
  });

  return {
    columns,
    data
  };
};

export const getCardsDataInTableFormat = (units, data) => {
  const result = {
    columns: [],
    tableData: [{ index: 0 }]
  };
  units.forEach((unit) => {
    if (data[unit.id]) {
      result.columns.push({
        title: <NonClickableTableHeader title={unit.title} />,
        dataIndex: unit.title,
        render: (d) => {
          return isNaN(d) ? d : <NumFormat number={d} />;
        }
      });
      try {
        result.tableData[0][unit.title] = isNaN(data[unit.id].rows[0][0])
          ? data[unit.id].rows[0][0]
          : formatCount(parseFloat(data[unit.id].rows[0][0]), 1);
      } catch (err) {
        result.tableData[0][unit.title] = data[unit.id].rows[0];
      }
    }
  });
  return result;
};
