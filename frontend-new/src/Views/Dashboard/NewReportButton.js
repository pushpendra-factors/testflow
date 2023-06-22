import React, { useMemo } from 'react';
import { SVG } from 'Components/factorsComponents';
import { Button, Divider, Dropdown, Menu, Space } from 'antd';
import {
  QUERY_TYPE_ATTRIBUTION,
  QUERY_TYPE_EVENT,
  QUERY_TYPE_FUNNEL,
  QUERY_TYPE_KPI,
  QUERY_TYPE_PROFILE
} from 'Utils/constants';
import { useHistory } from 'react-router-dom';

const getMenuItems = ({ showSavedReport }) => {
  const items = [
    {
      label: 'KPI Report',
      key: 1,
      icon: (
        <div style={{ padding: '0 10px 0 0px' }}>
          <SVG name={`KPI_cq`} size={24} color={'blue'} />
        </div>
      ),
      description: 'Measure performance over time'
    },
    {
      label: 'Funnel Report',
      key: 2,
      icon: (
        <div style={{ padding: '0 10px 0 0px' }}>
          <SVG name={`funnels_cq`} size={24} color={'blue'} />
        </div>
      ),
      description: 'Track how users navigate'
    },
    {
      label: 'Attribution Report',
      key: 3,
      icon: (
        <div style={{ padding: '0 10px 0 0px' }}>
          <SVG name={`attributions_cq`} size={24} color={'blue'} />
        </div>
      ),
      description: 'Identify the channels that contribute'
    },
    {
      label: 'Event Report',
      key: 4,
      icon: (
        <div style={{ padding: '0 10px 0 0px' }}>
          <SVG name={`events_cq`} size={24} color={'blue'} />
        </div>
      ),
      description: 'Track and Chart Events'
    },
    {
      label: 'Profile Report',
      key: 5,
      icon: (
        <div style={{ padding: '0 10px 0 0px' }}>
          <SVG name={`profiles_cq`} size={24} color={'blue'} />
        </div>
      ),
      description: 'Slice and dice your visitors'
    }
  ];

  if (showSavedReport === true) {
    items.push({
      label: 'Saved Report',
      key: 6,
      icon: (
        <div style={{ padding: '0 10px 0 0px' }}>
          {' '}
          <SVG name={'FileSignature'} size={24} color={'blue'} />
        </div>
      ),
      description: 'Select from saved Reports'
    });
  }
  return items;
};

const NewReportButton = ({ setIsReportsModalOpen, showSavedReport }) => {
  const history = useHistory();

  const HandleMenuItemClick = ({ item, key, keyPath, domEvent }) => {
    if (key === '0') {
      history.push({
        pathname: '/analyse/' + QUERY_TYPE_KPI,
        state: {
          navigatedFromDashboardExistingReports: true
        }
      });
    } else if (key === '1') {
      history.push({
        pathname: '/analyse/' + QUERY_TYPE_FUNNEL,
        state: {
          navigatedFromDashboardExistingReports: true
        }
      });
    } else if (key === '2') {
      history.push({
        pathname: '/analyse/' + QUERY_TYPE_ATTRIBUTION,
        state: {
          navigatedFromDashboardExistingReports: true
        }
      });
    } else if (key === '3') {
      history.push({
        pathname: '/analyse/' + QUERY_TYPE_EVENT,
        state: {
          navigatedFromDashboardExistingReports: true
        }
      });
    } else if (key === '4') {
      history.push({
        pathname: '/analyse/' + QUERY_TYPE_PROFILE,
        state: {
          navigatedFromDashboardExistingReports: true
        }
      });
    } else if (key === '5') {
      setIsReportsModalOpen((prev) => !prev);
    }
  };

  const items = useMemo(() => {
    return getMenuItems({ showSavedReport });
  }, [showSavedReport]);

  const menu = (
    <Menu
      onClick={HandleMenuItemClick}
      style={{ borderRadius: '5px', paddingTop: '8px' }}
    >
      {items.map((eachItem, eachKey) => {
        return (
          <React.Fragment key={eachItem.key}>
            {eachKey === items.length - 1 && showSavedReport ? (
              <Divider style={{ margin: 0 }} />
            ) : (
              ''
            )}
            <Menu.Item
              icon={eachItem.icon}
              key={eachKey}
              style={{
                margin: '2px 6px 2px 6px',
                display: 'flex',
                flexWrap: 'nowrap',
                borderRadius: '5px'
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
        );
      })}
    </Menu>
  );

  return (
    <Dropdown overlay={menu} placement='bottomRight' trigger={'click'}>
      <Button type='primary'>
        <Space>
          <SVG name={'plus'} size={16} color='white' />
          New Report
        </Space>
      </Button>
    </Dropdown>
  );
};

export default NewReportButton;
