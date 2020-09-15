import React, { useCallback, useState } from 'react';
import styles from './index.module.scss';
import { generateTableColumns, generateTableData } from '../utils';
import DataTable from '../DataTable';

function FunnelsResultTable({ eventsData, groups, setGroups }) {

    const [sorter, setSorter] = useState({});
    const [searchText, setSearchText] = useState('');

    const handleSorting = useCallback((sorter) => {
        setSorter(sorter);
    }, []);

    const columns = generateTableColumns(eventsData, sorter, handleSorting);
    const tableData = generateTableData(eventsData, groups, sorter, searchText);

    const onSelectionChange = (selectedRowKeys, selectedRows) => {
        const selectedGroups = selectedRows.map(elem => elem.name);
        setGroups(currentData => {
            return currentData.map(elem => {
                return { ...elem, is_visible: selectedGroups.indexOf(elem.name) >= 0 };
            });
        })
    };

    const selectedRowKeys = groups
        .filter(elem => elem.is_visible)
        .map(elem => elem.name)
        .map(elem => {
            return tableData.findIndex(d => d.name === elem)
        })

    const rowSelection = {
        selectedRowKeys,
        onChange: onSelectionChange,
    };

    console.log(columns)
    console.log(tableData)

    return (
        <DataTable
            tableData={tableData}
            searchText={searchText}
            setSearchText={setSearchText}
            columns={columns}
            rowSelection={rowSelection}
            className={styles.funnelResultsTable}
        />
    )
}

export default FunnelsResultTable;