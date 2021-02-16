import React from 'react';
import {
  Button, Tooltip
} from 'antd';
import { Text, SVG } from '../../components/factorsComponents';
import {
  // LockOutlined, ReloadOutlined, UserAddOutlined, MoreOutlined, EditOutlined, UnlockOutlined
  LockOutlined, ReloadOutlined, EditOutlined, UnlockOutlined
} from '@ant-design/icons';
import FaDatepicker from '../../components/FaDatepicker';
import moment from 'moment';

function DashboardSubMenu({
  dashboard, handleEditClick, durationObj, handleDurationChange, refreshClicked, setRefreshClicked
}) {
  let btn = null;

  if (dashboard.type === 'pr') {
    btn = (
      <Button
        style={{ display: 'flex' }}
         
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
         
        type={'text'}
        className={'m-0 fa-button-ghost items-center p-0 py-2'}
      >
        <UnlockOutlined /> Public.
      </Button>
    );
  }

  return (
    <div className={'flex justify-between items-center px-0 mb-5'}>
      <div className={'flex justify-between items-center'}>
        <Text type={'title'} level={7} extraClass={'m-0 mr-2'}>Data from</Text>
        <FaDatepicker
            customPicker
            presetRange
            monthPicker
            quarterPicker
            range={{
              startDate: durationObj.from,
              endDate: durationObj.to,
            }}
            placement="topRight"
            onSelect={handleDurationChange}
            buttonSize={'default'}
          />
        {btn}
        <Button onClick={handleEditClick.bind(this, dashboard)} type={'text'} className={'m-0 fa-button-ghost'} icon={<SVG name={'edit'}/>}>Edit</Button>
      </div>
      <div className={'flex justify-between items-center'}>
      
      <Tooltip placement="bottom" title={"Refresh data now"} mouseEnterDelay={0.2}>
        <Button type={"text"} onClick={setRefreshClicked.bind(this, true)} icon={refreshClicked ? null : <SVG name={'syncAlt'}/> }  loading={refreshClicked} style={{minWidth:'142px'}} className={'fa-button-ghost p-0 py-2'}>
          {dashboard?.updated_at ? moment(dashboard.updated_at).fromNow() : 'Refresh Data'} 
        </Button> 
      </Tooltip>
        {/* <Button style={{ display: 'flex' }} size={'large'} className={'items-center m-0 fa-button-ghost p-0 py-2'}><UserAddOutlined /></Button>
        <Button style={{ display: 'flex' }} size={'large'} className={'items-center m-0 fa-button-ghost p-0 py-2'}><MoreOutlined /></Button> */}

      </div>
    </div>
  );
};

export default DashboardSubMenu;
