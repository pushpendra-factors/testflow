import React, { useState, useEffect } from 'react';
import { Table } from 'antd';
import { generateTableColumns, generateTableData } from '../utils';
import styles from './index.module.scss';

function DataTable({ eventsData, groups, setGroups }) {

    const [tableData, setTableData] = useState([]);

    useEffect(() => {
        setTableData(generateTableData(eventsData, groups));
    }, [eventsData, groups]);

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

    const columns = generateTableColumns(eventsData);

    return (
        <Table
            pagination={false}
            bordered={true}
            rowKey='index'
            rowSelection={rowSelection}
            columns={columns}
            dataSource={tableData}
            className={styles.table}
        />
    )
}

export default DataTable;