import React, { useState, useEffect, useCallback } from 'react';
import { getTableColumns, getDataInTableFormat } from './utils';
import DataTable from '../../CoreQuery/FunnelsResultPage/DataTable';

function SingleEventMultipleBreakdownTable({ chartType, breakdown, data }) {
    const [sorter, setSorter] = useState({});
    const [searchText, setSearchText] = useState('');

    useEffect(() => {
        // reset sorter on change of chart type
        setSorter({});
    }, [chartType]);

    const handleSorting = useCallback((sorter) => {
        setSorter(sorter);
    }, []);

    let columns;
    let tableData;

    if (chartType === 'linechart') {
        // columns = getDateBasedColumns(lineChartData, breakdown, sorter, handleSorting);
        // tableData = getDateBasedTableData(data.map(elem => elem.label), originalData, breakdown, searchText, sorter);
    } else {
        columns = getTableColumns(breakdown, sorter, handleSorting);
        tableData = getDataInTableFormat(data, columns, breakdown, searchText, sorter);
    }

    return (
        <DataTable
            tableData={tableData}
            searchText={searchText}
            setSearchText={setSearchText}
            columns={columns}
        // rowSelection={rowSelection}
        />
    );
}

export default SingleEventMultipleBreakdownTable;