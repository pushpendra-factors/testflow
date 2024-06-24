import React, { useState, useEffect, useMemo } from 'react';
import { connect, useDispatch, useSelector } from 'react-redux';
import {
  Row,
  Col,
  Menu,
  Dropdown,
  Button,
  Table,
  notification,
  Tabs,
  Badge,
  Switch,
  Modal,
  Space
} from 'antd';
import { Text, SVG } from 'factorsComponents';
import { MoreOutlined, ExclamationCircleOutlined } from '@ant-design/icons';
import {
  fetchAlerts,
  deleteAlert,
  deleteEventAlert,
  createEventAlert,
  fetchAllAlerts,
  createAlert
} from 'Reducers/global';
import { fetchEventNames, getGroups } from 'Reducers/coreQuery/middleware';
import useQuery from 'hooks/useQuery';
import TableSearchAndRefresh from 'Components/TableSearchAndRefresh';
import {
  NEW_DASHBOARD_TEMPLATES_MODAL_CLOSE,
  NEW_DASHBOARD_TEMPLATES_MODAL_OPEN
} from 'Reducers/types';
import ModalFlow from 'Components/ModalFlow';
import { RESET_GROUPBY } from 'Reducers/coreQuery/actions';
import KPIBasedAlert from './KPIBasedAlert';
import EventBasedAlert from './EventBasedAlert';
import styles from './index.module.scss';
import { getAlertTemplatesTransformation } from './utils';
import RealTimeAlertsIllustration from '../../../../assets/images/illustrations/realtimealerts_illustration.png';

const { TabPane } = Tabs;

const Alerts = ({
  activeProject,
  fetchAlerts,
  deleteAlert,
  deleteEventAlert,
  savedAlerts,
  currentAgent,
  createEventAlert,
  fetchAllAlerts,
  createAlert,
  groups,
  getGroups,
  fetchEventNames
}) => {
  const [tableData, setTableData] = useState([]);
  const [tableLoading, setTableLoading] = useState(false);
  const [tableLoaded, setTableLoaded] = useState(false);
  const [viewAlertDetails, setAlertDetails] = useState(false);
  const [tabNo, setTabNo] = useState('2');
  const [alertState, setAlertState] = useState({
    state: 'list',
    index: 0
  });
  const [alertType, setAlertType] = useState('realtime');
  const { confirm } = Modal;
  const [showSearch, setShowSearch] = useState(false);
  const [searchTerm, setSearchTerm] = useState('');
  const [searchTableData, setSearchTableData] = useState([]);

  const routeQuery = useQuery();
  const dashboard_templates_modal_state = useSelector(
    (state) => state.dashboardTemplatesController
  );
  const alertTemplates = useSelector((state) => state.alertTemplates);
  const dispatch = useDispatch();

  useEffect(() => {
    if (!groups || Object.keys(groups).length === 0) {
      getGroups(activeProject?.id);
    }
  }, [activeProject?.id, groups]);
  useEffect(() => {
    fetchEventNames(activeProject?.id, true);
  }, [activeProject]);

  useEffect(() => {
    const type = routeQuery.get('type');
    if (type && ['realtime', 'weekly'].includes(type)) {
      setAlertState({ state: 'list', index: 0 });
      setAlertType(type);

      setSearchTerm('');
      setShowSearch(false);
    }
  }, [routeQuery]);

  const confirmDeleteAlert = (item) => {
    confirm({
      title: 'Do you really want to remove this alert?',
      icon: <ExclamationCircleOutlined />,
      content: 'Please confirm to proceed',
      onOk() {
        if (item?.type == 'kpi_alert') {
          return deleteAlert(activeProject?.id, item?.id).then(
            (res) => {
              fetchAllAlerts(activeProject?.id);
              notification.success({
                message: 'Success',
                description: 'Deleted Alert successfully ',
                duration: 5
              });
              setAlertState({ state: 'list', index: 0 });
            },
            (err) => {
              notification.error({
                message: 'Error',
                description: err.data,
                duration: 5
              });
            }
          );
        }
        return deleteEventAlert(activeProject?.id, item?.id).then(
          (res) => {
            fetchAllAlerts(activeProject?.id);
            notification.success({
              message: 'Success',
              description: 'Deleted Alert successfully ',
              duration: 5
            });
            setAlertState({ state: 'list', index: 0 });
          },
          (err) => {
            notification.error({
              message: 'Error',
              description: err.data,
              duration: 5
            });
          }
        );
      }
    });
  };

  const createDuplicateAlert = (item) => {
    if (item?.type == 'kpi_alert') {
      const payload = {
        ...item?.alert,
        alert_name: `Copy of ${item?.alert?.alert_name}`
      };
      createAlert(activeProject?.id, payload, 0)
        .then((res) => {
          setTableLoading(false);
          fetchAllAlerts(activeProject?.id);
          notification.success({
            message: 'Alert Created',
            description: 'Copy of alert is created and saved successfully.'
          });
        })
        .catch((err) => {
          setTableLoading(false);
          notification.error({
            message: 'Error',
            description: err?.data?.error
          });
        });
    } else {
      const payload = {
        ...item?.alert,
        title: `Copy of ${item?.alert?.title}`
      };

      createEventAlert(activeProject?.id, payload)
        .then((res) => {
          setTableLoading(false);
          fetchAllAlerts(activeProject?.id);
          notification.success({
            message: 'Alert Created',
            description: 'Copy of alert is created and saved successfully.'
          });
        })
        .catch((err) => {
          setTableLoading(false);
          notification.error({
            message: 'Error',
            description: err?.data?.error
          });
        });
    }
  };

  const menu = (item) => (
    <Menu className={`${styles.antdActionMenu}`}>
      <Menu.Item
        key='0'
        onClick={() => {
          setTabNo(item?.type === 'kpi_alert' ? '1' : '2');
          dispatch({ type: RESET_GROUPBY });
          setAlertState({ state: 'edit', index: item });
          setAlertDetails(item);
        }}
      >
        <a>Edit alert</a>
      </Menu.Item>
      <Menu.Item
        key='1'
        onClick={(e) => {
          createDuplicateAlert(item);
        }}
      >
        <a>Create copy</a>
      </Menu.Item>
      <Menu.Divider />
      <Menu.Item
        key='2'
        onClick={() => {
          confirmDeleteAlert(item);
        }}
      >
        <a>
          <span style={{ color: 'red' }}>Remove alert</span>
        </a>
      </Menu.Item>
    </Menu>
  );

  const columns = [
    {
      title: 'Name',
      dataIndex: 'alert_name',
      key: 'alert_name',
      width: '400px',
      render: (item) => (
        <Text
          type='title'
          level={7}
          truncate
          charLimit={50}
          extraClass='cursor-pointer m-0'
          onClick={() => {
            setTabNo(item?.type == 'kpi_alert' ? '1' : '2');
            dispatch({ type: RESET_GROUPBY });
            setAlertState({ state: 'edit', index: item });
            setAlertDetails(item);
          }}
        >
          {item?.alert_name || item?.title}
        </Text>
      )
    },
    {
      title: 'Delivery Options',
      dataIndex: 'dop',
      key: 'dop',
      render: (text) => (
        <Text type='title' level={7} truncate charLimit={25}>
          {text}
        </Text>
      )
      // width: 200,
    },
    {
      title: 'Status',
      dataIndex: 'status',
      key: 'status',
      render: (item) => (
        <div className='flex items-center'>
          {item?.status === 'paused' || item?.status === 'disabled' ? (
            <Badge
              className='fa-custom-badge fa-custom-badge--orange'
              status='default'
              text='Paused'
            />
          ) : (
            <Badge
              className='fa-custom-badge fa-custom-badge--green'
              status='processing'
              text='Active'
            />
          )}
          {item?.error && (
            <SVG name='InfoCircle' extraClass='ml-2' size={18} color='red' />
          )}
        </div>
      )
    },
    {
      title: '',
      dataIndex: 'actions',
      key: 'actions',
      align: 'right',
      width: 75,
      render: (obj) => (
        <Dropdown
          trigger={['click']}
          overlay={menu(obj)}
          placement='bottomRight'
        >
          <Button
            type='text'
            icon={
              <MoreOutlined
                rotate={90}
                style={{ color: 'gray', fontSize: '18px' }}
              />
            }
          />
        </Dropdown>
      )
    }
  ];

  useEffect(() => {
    setTableLoading(true);
    setTableLoaded(false);
    fetchAllAlerts(activeProject?.id).then(() => {
      setTableLoading(false);
      setTableLoaded(true);
    });
  }, [activeProject]);

  const IntegrationIcons = ({ slack, teams, webhook, email }) => {
    const iconSize = 22;
    return (
      <div className='flex items-center'>
        {slack && <SVG name='slack' size={iconSize} extraClass='mr-4' />}
        {teams && <SVG name='MSTeam' size={iconSize + 4} extraClass='mr-4' />}
        {webhook && <SVG name='Webhook' size={iconSize} extraClass='mr-4' />}
        {email && <SVG name='Email' size={iconSize} extraClass='mr-4' />}
      </div>
    );
  };

  useEffect(() => {
    const savedArr = [];
    savedAlerts?.forEach((item, index) => {
      if (alertType === 'weekly') {
        item.type === 'kpi_alert' &&
          savedArr.push({
            key: index,
            alert_name: item,
            type: item?.type == 'kpi_alert' ? 'Weekly alerts' : 'Real-time',
            dop: (
              <IntegrationIcons
                slack={item?.alert?.alert_configuration?.slack_enabled}
                teams={item?.alert?.alert_configuration?.teams_enabled}
                email={item?.alert?.alert_configuration?.email_enabled}
              />
            ),
            status: { status: item?.status, error: item?.last_fail_details },
            actions: item
          });
      } else if (alertType === 'realtime') {
        item.type === 'event_based_alert' &&
          savedArr.push({
            key: index,
            alert_name: item,
            type: item?.type == 'kpi_alert' ? 'Weekly alerts' : 'Real-time',
            dop: (
              <IntegrationIcons
                slack={item?.alert?.slack}
                teams={item?.alert?.teams}
                webhook={item?.alert?.webhook}
              />
            ),
            status: { status: item?.status, error: item?.last_fail_details },
            actions: item
          });
      }
    });
    setTableData(savedArr);
  }, [savedAlerts, tabNo, alertType]);

  function callback(key) {
    setTabNo(key);
  }

  const renderTitle = () => {
    let title = null;
    let titleText;
    switch (alertType) {
      case 'realtime': {
        titleText = 'Real time alerts';
        break;
      }
      case 'weekly': {
        titleText = 'Weekly updates';
        break;
      }
      default: {
        titleText = 'Alerts';
      }
    }
    if (alertState.state === 'list') {
      title = (
        <Text
          type='title'
          level={3}
          weight='bold'
          extraClass='m-0'
          id='fa-at-text--page-title'
        >
          {titleText}
        </Text>
      );
    }
    return title;
  };

  const addMenu = (item) => (
    <Menu className={`${styles.antdActionMenu}`}>
      <Menu.Item
        key='0'
        onClick={() => {
          setTabNo('2');
          setAlertState({ state: 'add', index: 0 });
          setAlertDetails(false);
        }}
      >
        <div className='flex items-center'>
          <SVG name='Event' size={20} color='blue' />
          <div className='pl-2'>
            <Text type='title' level={7} color='grey-2' extraClass='m-0'>
              Real-time alerts
            </Text>
            <Text type='title' level={8} color='grey' extraClass='m-0'>
              Track and chart events
            </Text>
          </div>
        </div>
      </Menu.Item>
      <Menu.Item
        key='0'
        onClick={() => {
          setTabNo('1');
          setAlertState({ state: 'add', index: 0 });
          setAlertDetails(false);
        }}
      >
        <div className='flex items-center'>
          <SVG name='Linechart' size={20} color='blue' />
          <div className='pl-2'>
            <Text type='title' level={7} color='grey-2' extraClass='m-0'>
              Weekly alerts
            </Text>
            <Text type='title' level={8} color='grey' extraClass='m-0'>
              Measure performance over time
            </Text>
          </div>
        </div>
      </Menu.Item>
    </Menu>
  );

  const switchToAddAlertState = (tabNo = '2') => {
    setTabNo(tabNo);
    setAlertState({ state: 'add', index: 0 });
    setAlertDetails(false);
  };

  const newAlertAction = () => {
    switch (alertType) {
      case 'realtime': {
        switchToAddAlertState('2');
        break;
      }
      case 'weekly': {
        switchToAddAlertState('1');
        break;
      }
      default: {
        switchToAddAlertState('2');
      }
    }
  };

  const renderTitleActions = (isFromTitleAction = false) => {
    let titleAction = null;
    if (
      alertType === 'realtime' &&
      !isFromTitleAction &&
      tableData.length === 0
    ) {
      return titleAction;
    }
    if (alertState.state === 'list') {
      titleAction = (
        <div className='p-1' style={{ display: 'flex', gap: '10px' }}>
          {alertType === 'realtime' && (
            <Button
              type='text'
              className='dropdown-btn'
              onClick={() => {
                dispatch({ type: NEW_DASHBOARD_TEMPLATES_MODAL_OPEN });
              }}
            >
              Templates
            </Button>
          )}
          <Button type='primary' onClick={newAlertAction}>
            <Space>
              <SVG name='plus' size={16} color='white' />
              Create New
            </Space>
          </Button>
        </div>
      );
    }

    return titleAction;
  };

  const onSearch = (e) => {
    const term = e.target.value;
    setSearchTerm(term);
    const searchResults = tableData?.filter((item) =>
      item?.alert_name?.title?.toLowerCase().includes(term.toLowerCase())
    );
    setSearchTableData(searchResults);
  };

  const onRefresh = () => {
    setTableLoading(true);
    setTableLoaded(false);
    fetchAllAlerts(activeProject?.id).then(() => {
      setTableLoading(false);
      setTableLoaded(true);
    });
  };
  const RealTimeEmptyIllustration = useMemo(
    () => (
      <div>
        <div className='flex justify-center select-none'>
          <img width={435} src={RealTimeAlertsIllustration} />
        </div>
        <div className='flex justify-center text-center py-3'>
          <Text type='title' level={7} color='grey-2' extraClass='m-0'>
            Setup alerts to get notified about prospect activity on your
            messaging apps <br /> like Slack and Teams to never miss out on a
            high intent lead.
          </Text>
        </div>
        <div className='flex justify-center text-center'>
          {renderTitleActions(true)}
        </div>
      </div>
    ),
    []
  );
  const renderAlertContent = () => {
    let alertContent = null;
    if (alertState.state === 'list') {
      alertContent = (
        <div className='mt-8'>
          {tableLoaded &&
            alertType === 'realtime' &&
            tableData.length === 0 &&
            RealTimeEmptyIllustration}
          {((tableData.length > 0 && alertType === 'realtime') ||
            alertType === 'weekly') && (
            <>
              <TableSearchAndRefresh
                showSearch={showSearch}
                setShowSearch={setShowSearch}
                searchTerm={searchTerm}
                setSearchTerm={setSearchTerm}
                onSearch={onSearch}
                onRefresh={onRefresh}
                tableLoading={tableLoading}
              />
              <Table
                className='fa-table--basic mt-2'
                loading={tableLoading}
                columns={columns}
                dataSource={searchTerm ? searchTableData : tableData}
                pagination
              />
            </>
          )}
        </div>
      );
    }
    return alertContent;
  };

  const renderTable = () => (
    <div className='fa-container'>
      <Row gutter={[24, 24]} justify='center'>
        <Col span={24}>
          <Row>
            <Col span={12}>{renderTitle()}</Col>
            <Col span={12}>
              <div className='flex justify-end'>{renderTitleActions()}</div>
            </Col>
          </Row>
          <Row>
            <Col span={24}>
              <Text type='title' level={7} color='grey-2' extraClass='m-0'>
                Get notified for important actions on your messaging app or send
                it to other platforms via webhook. &nbsp;
                <a
                  href='https://help.factors.ai/en/collections/8479811-real-time-alerts'
                  target='_blank'
                  rel='noreferrer'
                >
                  Learn more
                </a>
              </Text>

              <div className='mt-6'>{renderAlertContent()}</div>
            </Col>
          </Row>
        </Col>
        <ModalFlow
          data={getAlertTemplatesTransformation(alertTemplates?.data)}
          visible={dashboard_templates_modal_state.isNewDashboardTemplateModal}
          onCancel={() => {
            dispatch({ type: NEW_DASHBOARD_TEMPLATES_MODAL_CLOSE });
          }}
          handleLastFinish={(item, currentQuery, message_property) => {
            setTabNo('2');
            setAlertState({ state: 'add', index: 0 });
            setAlertDetails(false);
            setAlertDetails({
              type: 'event_based_alert',
              alert: {
                title: item.alert_name,
                message: item.alert_message,
                currentQuery,
                message_property
              },
              title: item.alert_name,
              extra: item
            });
          }}
        />
      </Row>
    </div>
  );

  return (
    <div>
      {alertState.state == 'list' ? (
        renderTable()
      ) : tabNo === '1' ? (
        <KPIBasedAlert
          alertState={alertState}
          setAlertState={setAlertState}
          viewAlertDetails={viewAlertDetails}
        />
      ) : (
        <EventBasedAlert
          alertState={alertState}
          setAlertState={setAlertState}
          viewAlertDetails={viewAlertDetails}
        />
      )}
    </div>
  );
};

const mapStateToProps = (state) => ({
  activeProject: state.global.active_project,
  savedAlerts: state.global.Alerts,
  kpi: state?.kpi,
  agent_details: state.agent.agent_details,
  slack: state.global.slack,
  projectSettings: state.global.projectSettingsV1,
  currentAgent: state.agent.agent_details,
  groups: state.coreQuery.groups
});

export default connect(mapStateToProps, {
  fetchAlerts,
  deleteAlert,
  deleteEventAlert,
  createEventAlert,
  fetchAllAlerts,
  createAlert,
  getGroups,
  fetchEventNames
})(Alerts);
