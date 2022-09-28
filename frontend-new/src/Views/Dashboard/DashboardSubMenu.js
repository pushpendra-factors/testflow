import React from 'react';
import { Button, Tooltip } from 'antd';
import { Text, SVG } from '../../components/factorsComponents';
import { LockOutlined, UnlockOutlined } from '@ant-design/icons';
import FaDatepicker from '../../components/FaDatepicker';
import moment from 'moment';
import { connect } from 'react-redux';

function DashboardSubMenu({
  dashboard,
  durationObj,
  handleDurationChange,
  refreshInProgress,
  oldestRefreshTime,
  handleRefreshClick
}) {
  let btn = null;

  if (dashboard?.type === 'pr') {
    btn = (
      <Tooltip
        overlayStyle={{ maxWidth: '160px' }}
        placement="bottom"
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
        placement="bottom"
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
          quarterPicker
          range={{
            startDate: durationObj.from,
            endDate: durationObj.to
          }}
          placement="bottomLeft"
          onSelect={handleDurationChange}
          buttonSize={'default'}
          className={'datepicker-minWidth'}
        />
        {btn}
      </div>
      <div className={'flex justify-between items-center'}>
        <div className="border-right--thin-3 px-3">
          {!!oldestRefreshTime && (
            <Text type={'title'} level={7} extraClass={'m-0'}>
              {moment.unix(oldestRefreshTime).fromNow()}
            </Text>
          )}
        </div>

        <Tooltip
          placement="bottom"
          title={'Refresh data now'}
          mouseEnterDelay={0.2}
        >
          <Button
            type={'text'}
            onClick={handleRefreshClick}
            icon={refreshInProgress ? null : <SVG name={'syncAlt'} />}
            loading={refreshInProgress}
            style={{ minWidth: '125px' }}
            className={'fa-button-ghost p-0 py-2'}
          >
            {' '}
            {'Refresh now'}
          </Button>
        </Tooltip>
      </div>
    </div>
  );
}

const mapStateToProps = (state) => {
  return {
    activeDashboard: state.dashboard.activeDashboard
  };
};

export default connect(mapStateToProps)(DashboardSubMenu);
