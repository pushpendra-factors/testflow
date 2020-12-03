import React from 'react';
import {
  Button
} from 'antd';
import { Text } from '../../components/factorsComponents';
import {
  // LockOutlined, ReloadOutlined, UserAddOutlined, MoreOutlined, EditOutlined, UnlockOutlined
  LockOutlined, ReloadOutlined, EditOutlined, UnlockOutlined
} from '@ant-design/icons';
import DurationInfo from '../CoreQuery/DurationInfo';

function DashboardSubMenu({
  dashboard, handleEditClick, durationObj, handleDurationChange, refreshClicked, setRefreshClicked
}) {
  let btn = null;

  if (dashboard.type === 'pr') {
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
  } else {
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
  }

  return (
    <div className={'flex justify-between items-center px-4 mb-4'}>
      <div className={'flex justify-between items-center'}>
        <Text type={'title'} level={7} extraClass={'m-0 mr-2'}>Data from</Text>
        <DurationInfo
          durationObj={durationObj}
          handleDurationChange={handleDurationChange}
        />
        {btn}
        <Button onClick={handleEditClick.bind(this, dashboard)} size={'large'} type={'text'} className={'m-0 fa-button-ghost flex items-center p-0 py-2'}><EditOutlined /> Edit</Button>
      </div>
      <div className={'flex justify-between items-center'}>
        <Button onClick={setRefreshClicked.bind(this, true)} disabled={refreshClicked} style={{ display: 'flex' }} size={'large'} className={'items-center flex m-0 fa-button-ghost p-0 py-2'}><ReloadOutlined /> Refresh Data</Button>
        {/* <Button style={{ display: 'flex' }} size={'large'} className={'items-center m-0 fa-button-ghost p-0 py-2'}><UserAddOutlined /></Button>
        <Button style={{ display: 'flex' }} size={'large'} className={'items-center m-0 fa-button-ghost p-0 py-2'}><MoreOutlined /></Button> */}

      </div>
    </div>
  );
};

export default DashboardSubMenu;
