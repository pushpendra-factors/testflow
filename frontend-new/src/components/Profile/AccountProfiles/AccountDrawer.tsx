import React from 'react';
import { Drawer, Button } from 'antd';
import { SVG, Text } from 'Components/factorsComponents';
import { AccountDrawerProps } from 'Components/Profile/types';
import AccountTimelineTableView from './AccountTimelineTableView';

function AccountDrawer({
  domain,
  events = [],
  visible,
  onClose,
  onClickMore,
  onClickOpenNewtab
}: AccountDrawerProps): JSX.Element {
  return (
    <Drawer
      title={
        <div className='flex justify-between items-center'>
          <Text type='title' level={4} weight='bold' extraClass='m-0'>
            {domain}
          </Text>
          <div className='inline-flex gap--8'>
            <Button onClick={onClickMore}>
              <div className='inline-flex gap--4'>
                <SVG name='expand' />
                Open
              </div>
            </Button>
            <Button onClick={onClickOpenNewtab} className='flex items-center'>
              <SVG name='ArrowUpRightSquare' size={16} />
            </Button>
            <Button onClick={onClose} className='flex items-center'>
              <SVG name='times' size={16} />
            </Button>
          </div>
        </div>
      }
      placement='right'
      closable={false}
      mask={false}
      visible={visible}
      className='fa-account-drawer--right'
      onClose={onClose}
      width='large'
    >
      <div>
        <AccountTimelineTableView
          timelineEvents={events}
          eventPropsType={{}}
          loading={false}
        />
      </div>
    </Drawer>
  );
}

export default AccountDrawer;
