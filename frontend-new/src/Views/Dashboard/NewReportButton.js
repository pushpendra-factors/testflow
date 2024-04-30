import React, { useEffect, useMemo, useRef, useState } from 'react';
import { SVG, Text } from 'Components/factorsComponents';
import { Button, Divider, Dropdown, Menu, Space } from 'antd';
import {
  QUERY_TYPE_EVENT,
  QUERY_TYPE_FUNNEL,
  QUERY_TYPE_KPI,
  SAVED_QUERY,
  QUERY_TYPE_PROFILE
} from 'Utils/constants';
import { useHistory } from 'react-router-dom';
import { PlusOutlined } from '@ant-design/icons';
import ExistingReportsModal from './ExistingReportsModal';

const getMenuItems = ({ showSavedReport }) => {
  const items = [
    {
      label: 'New KPI Report',
      key: QUERY_TYPE_KPI,
      // '#FFF2E8'

      icon: <SVG name='KPI_cq' size={24} />,
      description: 'Measure performance over time',
      bgColor: '#FFF2E8'
    },
    {
      label: 'New Funnel Report',
      key: QUERY_TYPE_FUNNEL,
      icon: <SVG name='funnels_cq' size={24} />,
      description: 'Track how users navigate',
      bgColor: '#FFF0F6'
    },
    {
      label: 'New Event Report',
      key: QUERY_TYPE_EVENT,
      icon: <SVG name='events_cq' size={24} />,
      description: 'Track and Chart Events',
      bgColor: '#F0F5FF'
    },
    {
      label: 'New Profile Report',
      key: QUERY_TYPE_PROFILE,
      icon: <SVG name='profiles_cq' size={24} />,
      description: 'Slice and dice your visitors',
      bgColor: '#FFF7E6'
    }
  ];

  if (showSavedReport === true) {
    items.push({
      label: 'Add from draft',
      key: SAVED_QUERY,
      icon: <SVG name='FileSignature' size={24} color='#8C8C8C' />,
      description: 'Select from saved Reports',
      bgColor: '#F5F5F5'
    });
  }
  return items;
};

const AddReportWidgetWrapper = ({ children, isWidget, ...otherprops }) => {
  if (!isWidget) return children;
  return (
    <div className='w-full h-full flex justify-center ' {...otherprops}>
      <div className='flex items-center gap-2'>{children}</div>
    </div>
  );
};
const NewReportButton = ({
  showSavedReport,
  placement = 'bottomRight',
  isWidget = false
}) => {
  const [isReportsModalOpen, setIsReportsModalOpen] = useState(false);
  const history = useHistory();
  const WrapperRef = useRef();
  const HandleMenuItemClick = ({ key }) => {
    if (key !== SAVED_QUERY) {
      history.push({
        pathname: `/analyse/${key}`,
        state: {
          navigatedFromDashboardExistingReports: true
        }
      });
    } else {
      setIsReportsModalOpen((prev) => !prev);
    }
  };

  const items = useMemo(
    () => getMenuItems({ showSavedReport }),
    [showSavedReport]
  );

  const menu = (
    <Menu
      onClick={HandleMenuItemClick}
      style={{ borderRadius: '5px', paddingTop: '8px' }}
    >
      {items.map((eachItem, eachKey) => (
        <React.Fragment key={eachItem.key}>
          {eachKey === items.length - 1 && showSavedReport ? (
            <Divider style={{ margin: 0 }} />
          ) : (
            ''
          )}
          <Menu.Item
            icon={
              <div
                className='flex items-center rounded justify-center'
                style={{
                  padding: '5px',
                  background: eachItem.bgColor,
                  height: 40,
                  width: 40
                }}
              >
                {eachItem.icon}
              </div>
            }
            key={eachItem.key}
            style={{
              margin: '2px 6px 2px 6px',
              display: 'flex',
              flexWrap: 'nowrap',
              borderRadius: '5px',
              gap: '10px'
            }}
          >
            <div style={{ display: 'block' }}>
              {' '}
              <div>{eachItem.label}</div>
              <div style={{ fontSize: '12px', color: '#8692A3' }}>
                {eachItem.description}
              </div>
            </div>
          </Menu.Item>
        </React.Fragment>
      ))}
    </Menu>
  );
  const reportsModal = (
    <ExistingReportsModal
      isReportsModalOpen={isReportsModalOpen}
      setIsReportsModalOpen={setIsReportsModalOpen}
    />
  );
  if (isWidget) {
    return (
      <>
        <Dropdown
          overlay={menu}
          overlayClassName='fa-at-overlay--new-report'
          trigger={['contextMenu']}
          placement={placement}
          align={{ overflow: { adjustX: true, adjustY: true } }}
        >
          <div
            ref={WrapperRef}
            className='w-full h-full flex items-center justify-center'
            onClick={(e) => {
              var rightClickEvent = new MouseEvent('contextmenu', {
                bubbles: true,
                cancelable: true,
                view: window,
                button: 2,
                buttons: 2,
                clientX: e.clientX,
                clientY: e.clientY
              });

              WrapperRef.current.dispatchEvent(rightClickEvent);
            }}
          >
            <div className='flex items-center gap-2'>
              <PlusOutlined />{' '}
              <Text level={5} type='title' extraClass='mb-0' weight='bold'>
                Add Report
              </Text>
            </div>
          </div>
        </Dropdown>{' '}
        {reportsModal}
      </>
    );
  }
  return (
    <>
      <Dropdown
        overlay={menu}
        overlayClassName='fa-at-overlay--new-report'
        placement={placement}
        trigger='click'
      >
        <Button type='primary' id='fa-at-btn--new-report'>
          <Space>
            <SVG name='plus' size={16} color='white' />
            Add Report
          </Space>
        </Button>
      </Dropdown>
      {reportsModal}
    </>
  );
};

export default NewReportButton;
