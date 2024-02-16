import React, { useEffect, useState } from 'react';
import { Text, SVG } from 'factorsComponents';
import {
    Button,
    Table,
    Avatar,
    Menu,
    Dropdown,
    Modal,
    message,
    Badge,
    Input
} from 'antd';

const TableSearchAndRefresh = ({
    showSearch,
    setShowSearch,
    searchTerm,
    setSearchTerm,
    onSearch,
    onRefresh,
    tableLoading
}) => {

    return (
        <div className='flex justify-end'>
            {/* refresh table */}
            <div className='flex items-center mr-1'>
                <Button
                    loading={tableLoading}
                    icon={<SVG name={'syncAlt'} size={18} color={'grey'} />}
                    type='text'
                    ghost={true}
                    shape='square'
                    onClick={() => onRefresh()}
                />
            </div>

            {/* search table */}
            <div className={'flex items-center justify-between'}>
                {showSearch ? (
                    <Input
                        autoFocus
                        onChange={(e) => onSearch(e)}
                        placeholder={'Search reports'}
                        style={{ width: '220px', 'border-radius': '5px' }}
                        prefix={<SVG name='search' size={16} color={'grey'} />}
                    />
                ) : null}
                <Button
                    type='text'
                    ghost={true}
                    shape='circle'
                    className={'p-2 bg-white'}
                    onClick={() => {
                        setShowSearch(!showSearch);
                        if (showSearch) {
                            setSearchTerm('');
                        }
                    }}
                >
                    <SVG
                        name={!showSearch ? 'search' : 'close'}
                        size={20}
                        color={'grey'}
                    />
                </Button>
            </div>
        </div>
    )

}

export default TableSearchAndRefresh