import React, { useState, useEffect, useCallback } from 'react';
import { connect } from 'react-redux';
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
import { MoreOutlined } from '@ant-design/icons';
import _ from 'lodash';
import {
  fetchAlerts,
  deleteAlert,
  deleteEventAlert,
  createEventAlert,
  fetchAllAlerts,
  createAlert
} from 'Reducers/global';
import KPIBasedAlert from './KPIBasedAlert';
import EventBasedAlert from './EventBasedAlert';
import styles from './index.module.scss';
import { ExclamationCircleOutlined } from '@ant-design/icons';
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
  createAlert
}) => {
  const [tableData, setTableData] = useState([]);
  const [tableLoading, setTableLoading] = useState(false);
  const [viewAlertDetails, setAlertDetails] = useState(false);
  const [tabNo, setTabNo] = useState('2');
  const [alertState, setAlertState] = useState({
    state: 'list',
    index: 0
  });
  const { confirm } = Modal;

  const confirmDeleteAlert = (item) => {
    confirm({
      title: 'Do you really want to remove this alert?',
      icon: <ExclamationCircleOutlined />,
      content: 'Please confirm to proceed',
      onOk() {
        if (item?.type == "kpi_alert") {
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
        } else {
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
      }
    });
  };


  const createDuplicateAlert = (item) => {
    if(item?.type == "kpi_alert"){
      let payload = {
        ...item?.alert,
        alert_name: `Copy of ${item?.alert?.alert_name}`
      }
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
    }
    else{
      let payload = {
        ...item?.alert,
        title: `Copy of ${item?.alert?.title}`
      }

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
  }

  const menu = (item) => {
    return (
      <Menu className={`${styles.antdActionMenu}`}>
        <Menu.Item
          key='0'
          onClick={() => {
            setTabNo(item?.type == "kpi_alert" ? "1" : "2")
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
  };

  const columns = [
    {
      title: 'Name',
      dataIndex: 'alert_name',
      key: 'alert_name',
      width: '300px',
      render: (item) => (
        <Text
          type={'title'}
          level={7}
          truncate={true}
          extraClass={`cursor-pointer m-0`}
          onClick={() => { 
            setTabNo(item?.type == "kpi_alert" ? "1" : "2")
            setAlertState({ state: 'edit', index: item });
            setAlertDetails(item);
          }}
        >
          {item?.alert_name || item?.title}
        </Text>
      )
      // width: 100,
    },
    {
      title: 'Type',
      dataIndex: 'type',
      key: 'type',
      render: (text) => (
        <Text type={'title'} level={7} truncate={true} charLimit={25}>
          {text}
        </Text>
      )
      // width: 200,
    },
    {
      title: 'Delivery Options',
      dataIndex: 'dop',
      key: 'dop',
      render: (text) => (
        <Text type={'title'} level={7} truncate={true} charLimit={25}>
          {text}
        </Text>
      )
      // width: 200,
    },
    {
      title: 'Status',
      dataIndex: 'status',
      key: 'status',
      render: (status) => (
        <div className='flex items-center'>
          {' '}
          {status === 'paused' || status === 'disabled' ? (
            <Badge
              className={'fa-custom-badge fa-custom-badge--orange'}
              status='processing'
              text={'Paused'}
            />
          ) : (
            <Badge
              className={'fa-custom-badge fa-custom-badge--green'}
              status='success'
              text={'Active'}
            />
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
        <Dropdown trigger={["click"]} overlay={menu(obj)} placement='bottomRight'>
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
    fetchAllAlerts(activeProject?.id).then(() => {
      setTableLoading(false);
    }); 
  }, [activeProject]);

  useEffect(() => { 
    let savedArr = [];
      if (savedAlerts && savedAlerts?.type == "kpi_alert") {
        savedAlerts?.map((item, index) => {
          savedArr.push({
            key: index,
            alert_name: item,
            type: item?.type == "kpi_alert" ? "Weekly alerts" : "Real-time",
            dop:
              (item.alert_configuration.email_enabled ? 'Email' : '') +
              ' ' +
              (item.alert_configuration.slack_enabled ? 'Slack' : '') +
              ' ' +
              (item.alert_configuration.teams_enabled ? 'Teams' : ''),
            status: item?.status,
            actions: item
          });
        }); 
      }  else {  
        savedAlerts?.map((item, index) => {
          savedArr.push({
            key: index,
            alert_name: item,
            type: item?.type == "kpi_alert" ? "Weekly alerts" : "Real-time",
            dop: item?.delivery_options,
            status: item?.status,
            actions: item
          });
        });
      }
      setTableData(savedArr); 
  }, [savedAlerts, tabNo]);

  function callback(key) {
    setTabNo(key);
  }

  const renderTitle = () => {
    let title = null;
    if (alertState.state === 'list') {
      title = (
        <Text type={'title'} level={3} weight={'bold'} extraClass={'m-0'}>
          Alerts
        </Text>
      );
    }
    return title;
  };


  const addMenu = (item) => {
    return (
      <Menu className={`${styles.antdActionMenu}`}>
        <Menu.Item
          key='0'
          onClick={() => {
            setTabNo("2")
            setAlertState({ state: 'add', index: 0 });
            setAlertDetails(false)
          }}
        >
            <div className='flex items-center'>
              <SVG name={'Event'} size={20} color='blue' />
              <div className='pl-2'> 
                <Text type={'title'} level={7} color={'grey-2'} extraClass={'m-0'} >Real-time alerts</Text>
                <Text type={'title'} level={8} color={'grey'} extraClass={'m-0'} >Track and chart events</Text>
              </div>
          </div>
        </Menu.Item> 
        <Menu.Item
          key='0'
          onClick={() => {
            setTabNo("1")
            setAlertState({ state: 'add', index: 0 });
            setAlertDetails(false)
          }}
        >
            <div className='flex items-center'>
              <SVG name={'Linechart'} size={20} color='blue' />
              <div className='pl-2'> 
            <Text type={'title'} level={7} color={'grey-2'} extraClass={'m-0'} >Weekly alerts</Text>
            <Text type={'title'} level={8} color={'grey'} extraClass={'m-0'} >Measure performance over time</Text>
            </div>  
          </div>
        </Menu.Item> 
        </Menu>
    )
  }

  const renderTitleActions = () => {
    let titleAction = null;
    if (alertState.state === 'list') {
      titleAction = ( 

        <Dropdown overlay={addMenu} placement='bottomRight' trigger={'click'}>
        <Button type='primary'>
        <Space>
          <SVG name={'plus'} size={16} color='white' />
          New Alert
        </Space>
        </Button>
        </Dropdown>


      );
    }

    return titleAction;
  };

  const renderAlertContent = () => {
    let alertContent = null;
    if (alertState.state === 'list') {
      alertContent = ( 
          <Table
            className='fa-table--basic mt-8'
            loading={tableLoading}
            columns={columns}
            dataSource={tableData}
            pagination={false}
            // onRow={(data,id)=>{
            //   return {
            //     onClick: (e) => { 
            //       setTabNo(data.actions?.type == "kpi_alert" ? "1" : "2")
            //       setAlertState({ state: 'edit', index: data.actions});
            //       setAlertDetails(data.actions);
            //       e.stopPropagation();
            //     }
            //   }
            // }}
          />
      );
    }
    return alertContent;
  };

  const renderTable = () => {
    return (
      <div className={'fa-container'}>
        <Row gutter={[24, 24]} justify='center'>
          <Col span={18}>
            <Row>
              <Col span={12}>{renderTitle()}</Col>
              <Col span={12}>
                <div className={'flex justify-end'}>{renderTitleActions()}</div>
              </Col>
            </Row>
            <Row className={'mt-4'}>
              <Col span={24}> 
                <Text
                  type={'title'}
                  level={7}
                  color={'grey-2'}
                  extraClass={'m-0 mt-2'}
                >
                  Set up alerts to never miss out on any prospect activity or changes in metrics you care about.
                  &nbsp;<a href='https://help.factors.ai/en/articles/7284705-alerts' target='_blank'>
                    Learn more
                  </a>
                </Text>
                <div className={'mt-6'}>{renderAlertContent()}</div>
              </Col>
            </Row>
          </Col>
        </Row>
      </div>
    );
  };

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
  currentAgent: state.agent.agent_details
});

export default connect(mapStateToProps, {
  fetchAlerts,
  deleteAlert,
  deleteEventAlert,
  createEventAlert,
  fetchAllAlerts,
  createAlert
})(Alerts);