import React from 'react';
import {
    Button, Select
} from 'antd';
import { Text } from '../../components/factorsComponents';
import {
    LockOutlined, ReloadOutlined, UserAddOutlined, MoreOutlined, EditOutlined, UnlockOutlined
} from '@ant-design/icons';

const { Option } = Select;

function DashboardSubMenu({ dashboard, handleEditClick }) {
    let btn = null;

    if (dashboard.type === 'pr') {
        btn = (
            <Button
                style={{ display: 'flex' }}
                size={'large'}
                type={'text'}
                className={'m-0 fa-button-ghost items-center p-0 py-2'}
            >
                <UnlockOutlined /> Public.
            </Button>
        );
    } else {
        btn = (
            <Button
                style={{ display: 'flex' }}
                size={'large'}
                type={'text'}
                className={'m-0 fa-button-ghost items-center p-0 py-2'}
            >
                <LockOutlined /> Private.
            </Button>
        );
    }

    return (
        <div className={'flex justify-between items-center px-4 mb-4'}>
            <div className={'flex justify-between items-center'}>
                <Text type={'title'} level={7} extraClass={'m-0 mr-2'}>Date from</Text>
                <Select className={'fa-select mx-2 mr-4 ml-4'} defaultValue="Last 30 days">
                    <Option value="jack">1 Month</Option>
                    <Option value="lucy2">2 Months</Option>
                    <Option value="lucy3">6 Months</Option>
                    <Option value="lucy4">1 Year</Option>
                    <Option value="lucy5">1+ Year</Option>
                </Select>
                {btn}
                <Button onClick={handleEditClick.bind(this, dashboard)} size={'large'} type={'text'} className={'m-0 fa-button-ghost flex items-center p-0 py-2'}><EditOutlined /> Edit</Button>
            </div>
            <div className={'flex justify-between items-center'}>
                <Button style={{ display: 'flex' }} size={'large'} className={'items-center flex m-0 fa-button-ghost p-0 py-2'}><ReloadOutlined /> Refresh Data.</Button>
                <Button style={{ display: 'flex' }} size={'large'} className={'items-center m-0 fa-button-ghost p-0 py-2'}><UserAddOutlined /></Button>
                <Button style={{ display: 'flex' }} size={'large'} className={'items-center m-0 fa-button-ghost p-0 py-2'}><MoreOutlined /></Button>

            </div>
        </div>
    );
};

export default DashboardSubMenu;