import React, { useEffect, useState } from 'react';
import { Button, Tooltip } from 'antd';
import { Text, SVG } from '../../components/factorsComponents';
import { LockOutlined, UnlockOutlined } from '@ant-design/icons';
import FaDatepicker from '../../components/FaDatepicker';
import moment from 'moment';
import { connect } from 'react-redux';
import { DASHBOARD_TYPES } from '../../utils/constants';

function DashboardSubMenu({
  dashboard,
  handleEditClick,
  durationObj,
  handleDurationChange,
  refreshClicked,
  setRefreshClicked,
  activeDashboard,
}) {
  let btn = null;
  const [showRefreshBtn, setShowRefreshBtn] = useState(false);

  if (dashboard?.type === 'pr') {
    btn = (
      <Tooltip
        overlayStyle={{ maxWidth: '160px' }}
        placement='bottom'
        title={'This dashboard is visible only to you.'}
        mouseEnterDelay={0.2}
      >
        <Button
          style={{ cursor: 'default' }}
          type={'text'}
          className={'m-0 fa-button-ghost items-center p-0 py-2'}
        >
          <LockOutlined /> Private.
        </Button>
      </Tooltip>
    );
  } else {
    btn = (
      <Tooltip
        overlayStyle={{ maxWidth: '160px' }}
        placement='bottom'
        title={'This dashboard is visible to everyone.'}
        mouseEnterDelay={0.2}
      >
        <Button
          style={{ cursor: 'default' }}
          type={'text'}
          className={'m-0 fa-button-ghost items-center p-0 py-2'}
        >
          <UnlockOutlined /> Public.
        </Button>
      </Tooltip>
    );
  }
  useEffect(() => {
    let isRefresh =
      durationObj?.dateType === 'today' || durationObj?.dateType === 'now';
    setShowRefreshBtn(isRefresh);
  }, [durationObj, dashboard, activeDashboard]);

  return (
    <div className={'flex justify-between items-center px-0 mb-5'}>
      <div className={'flex justify-between items-center'}>
        <Text type={'title'} level={7} extraClass={'m-0 mr-2'}>
          Data from
        </Text>
        <FaDatepicker
          customPicker
          nowPicker={dashboard?.name === 'Website Analytics' ? true : false}
          presetRange
          range={{
            startDate: durationObj.from,
            endDate: durationObj.to,
          }}
          placement='bottomLeft'
          onSelect={handleDurationChange}
          buttonSize={'default'}
          className={'datepicker-minWidth'}
        />
        {btn}
        {/* {dashboard?.class === DASHBOARD_TYPES.USER_CREATED ? (
          <Button
            onClick={handleEditClick.bind(this, dashboard)}
            type={'text'}
            className={'m-0 fa-button-ghost'}
            icon={<SVG name={'edit'} />}
          >
            Edit
          </Button>
        ) : null} */}
      </div>
      <div className={'flex justify-between items-center'}>
        {showRefreshBtn && (
          <Tooltip
            placement='bottom'
            title={'Refresh data now'}
            mouseEnterDelay={0.2}
          >
            <Button
              type={'text'}
              onClick={setRefreshClicked.bind(this, true)}
              icon={refreshClicked ? null : <SVG name={'syncAlt'} />}
              loading={refreshClicked}
              style={{ minWidth: '142px' }}
              className={'fa-button-ghost p-0 py-2'}
            >
              {activeDashboard?.refreshed_at
                ? moment.unix(activeDashboard.refreshed_at).fromNow()
                : 'Refresh Data'}
            </Button>
          </Tooltip>
        )}

        {/* <Button style={{ display: 'flex' }} size={'large'} className={'items-center m-0 fa-button-ghost p-0 py-2'}><UserAddOutlined /></Button>
        <Button style={{ display: 'flex' }} size={'large'} className={'items-center m-0 fa-button-ghost p-0 py-2'}><MoreOutlined /></Button> */}
      </div>
    </div>
  );
}

const mapStateToProps = (state) => {
  return {
    activeDashboard: state.dashboard.activeDashboard,
  };
};

export default connect(mapStateToProps)(DashboardSubMenu);
