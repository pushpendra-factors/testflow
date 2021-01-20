export const getWebAnalyticsTableData = (tableData) => {
  const { headers, rows } = tableData;
  const columns = headers.map((header) => {
    return {
      title: header,
      dataIndex: header,
    };
  });

  const data = rows.map((row, idx) => {
    const rowData = {};
    row.forEach((elem, index) => {
      rowData[headers[index]] = elem;
    });
    return { ...rowData, index: idx };
  });

  return {
    columns,
    data,
  };
};

export const getCardsDataInTableFormat = (units, data) => {
  const result = {
    columns: [],
    tableData: [{ index: 0 }],
  };
  units.forEach((unit) => {
    if (data[unit.id]) {
      result.columns.push({
        title: unit.title,
        dataIndex: unit.title,
      });
      result.tableData[0][unit.title] = data[unit.id].rows[0];
    }
  });
  return result;
};
