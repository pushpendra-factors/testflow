import React, { useState, useCallback } from 'react';
import { getSingleEventNoGroupingTableData, getSingleEventNoGroupingTableColumns } from '../utils';
import DataTable from '../../CoreQuery/FunnelsResultPage/DataTable';

function TotalEventsTable({ data, event }) {

    const [sorter, setSorter] = useState({});
    const [searchText, setSearchText] = useState('');

    const handleSorting = useCallback((sorter) => {
        setSorter(sorter);
    }, []);

    const columns = getSingleEventNoGroupingTableColumns(data, event, sorter, handleSorting);

    const tableData = getSingleEventNoGroupingTableData(data, event, sorter, searchText);

    return (
        <DataTable
            tableData={tableData}
            searchText={searchText}
            setSearchText={setSearchText}
            columns={columns}
            rowSelection={null}
        />
    )
}

export default TotalEventsTable;