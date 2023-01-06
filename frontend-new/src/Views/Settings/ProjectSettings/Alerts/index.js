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
  Tabs
} from 'antd';
import { Text, SVG } from 'factorsComponents';
import { MoreOutlined } from '@ant-design/icons';
import _ from 'lodash';
import {
  fetchAlerts,
  deleteAlert,
  fetchEventAlerts,
  deleteEventAlert
} from 'Reducers/global';
import ConfirmationModal from '../../../../components/ConfirmationModal';
import KPIBasedAlert from './KPIBasedAlert';
import EventBasedAlert from './EventBasedAlert';

const { TabPane } = Tabs;

const Alerts = ({
  activeProject,
  fetchAlerts,
  deleteAlert,
  fetchEventAlerts,
  deleteEventAlert,
  savedAlerts,
  savedEventAlerts,
  currentAgent
}) => {
  const [tableData, setTableData] = useState([]);
  const [tableLoading, setTableLoading] = useState(false);
  const [viewAlertDetails, setAlertDetails] = useState(false);
  const [deleteWidgetModal, showDeleteWidgetModal] = useState(false);
  const [deleteApiCalled, setDeleteApiCalled] = useState(false);
  const [tabNo, setTabNo] = useState('1');
  const [alertState, setAlertState] = useState({
    state: 'list',
    index: 0
  });

  const confirmRemove = (id) => {
    if (tabNo === '1') {
      return deleteAlert(activeProject.id, id).then(
        (res) => {
          fetchAlerts(activeProject.id);
          notification.success({
            message: 'Success',
            description: 'Deleted Alert successfully ',
            duration: 5
          });
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
      return deleteEventAlert(activeProject.id, id).then(
        (res) => {
          fetchEventAlerts(activeProject.id);
          notification.success({
            message: 'Success',
            description: 'Deleted Alert successfully ',
            duration: 5
          });
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
  };

  const confirmDelete = useCallback(async () => {
    try {
      setDeleteApiCalled(true);
      await confirmRemove(deleteWidgetModal);
      setDeleteApiCalled(false);
      showDeleteWidgetModal(false);
      setAlertState({ state: 'list', index: 0 });
    } catch (err) {
      console.log(err);
      console.log(err.response);
    }
  }, [deleteWidgetModal]);

  const menu = (item) => {
    return (
      <Menu>
        <Menu.Item
          key='0'
          onClick={() => {
            setAlertState({ state: 'view', index: item });
            setAlertDetails(item);
          }}
        >
          <a>View</a>
        </Menu.Item>
        <Menu.Item
          key='1'
          onClick={() => {
            setAlertState({ state: 'edit', index: item });
            setAlertDetails(item);
          }}
        >
          <a>Edit</a>
        </Menu.Item>
        <Menu.Item
          key='2'
          onClick={() => {
            showDeleteWidgetModal(item.id);
          }}
        >
          <a>Remove</a>
        </Menu.Item>
      </Menu>
    );
  };

  const columns = [
    {
      title: 'Name',
      dataIndex: 'alert_name',
      key: 'alert_name',
      render: (text) => (
        <Text type={'title'} level={7} truncate={true} charLimit={50}>
          {text}
        </Text>
      )
      // width: 100,
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
      title: '',
      dataIndex: 'actions',
      key: 'actions',
      align: 'right',
      width: 75,
      render: (obj) => (
        <Dropdown overlay={menu(obj)} trigger={['hover']}>
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
    if (tabNo === '1') {
      setTableLoading(true);
      fetchAlerts(activeProject.id).then(() => {
        setTableLoading(false);
      });
    } else {
      setTableLoading(true);
      fetchEventAlerts(activeProject.id).then(() => {
        setTableLoading(false);
      });
    }
  }, [activeProject, tabNo]);

  useEffect(() => {
    if (tabNo === '1') {
      if (savedAlerts) {
        let savedArr = [];
        savedAlerts?.map((item, index) => {
          savedArr.push({
            key: index,
            alert_name: item.alert_name,
            dop:
              (item.alert_configuration.email_enabled ? 'Email' : '') +
              ' ' +
              (item.alert_configuration.slack_enabled ? 'Slack' : ''),
            actions: item
          });
        });
        setTableData(savedArr);
      } else {
        setTableData([]);
      }
    } else {
      if (savedEventAlerts) {
        let savedArr = [];
        savedEventAlerts?.map((item, index) => {
          savedArr.push({
            key: index,
            alert_name: item.title,
            dop: item?.slack && 'Slack',
            actions: item
          });
        });
        setTableData(savedArr);
      } else {
        setTableData([]);
      }
    }
  }, [savedAlerts, savedEventAlerts, tabNo]);

  function callback(key) {
    setTabNo(key);
  }

  const whiteListedAccounts = [
    'junaid@factors.ai',
    'solutions@factors.ai',
  ];

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

  const renderTitleActions = () => {
    let titleAction = null;
    if (alertState.state === 'list') {
      titleAction = (
        <Button
          size={'large'}
          onClick={() => {
            setAlertState({ state: 'add', index: 0 });
          }}
        >
          <SVG name={'plus'} extraClass={'mr-2'} size={16} />
          Add New
        </Button>
      );
    }

    return titleAction;
  };

  const renderAlertContent = () => {
    let alertContent = null;
    if (alertState.state === 'list') {
      alertContent = (
        <Tabs activeKey={`${tabNo}`} onChange={callback}>
          <TabPane tab='Track KPIs' key='1'>
            <Table
              loading={tableLoading}
              className='fa-table--basic mt-8'
              columns={columns}
              dataSource={tableData}
              pagination={false}
            />
          </TabPane>
          {whiteListedAccounts.includes(currentAgent?.email) &&
          <TabPane tab='Event based' key='2'>
            <Table
              className='fa-table--basic mt-8'
              loading={tableLoading}
              columns={columns}
              dataSource={tableData}
              pagination={false}
            />
          </TabPane>}
        </Tabs>
      );
    }
    return alertContent;
  };

  const renderTable = () => {
    return (
      <div className={'fa-container mt-32 mb-12 min-h-screen'}>
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
        >
          {' '}
        </KPIBasedAlert>
      ) : (
        <EventBasedAlert
          alertState={alertState}
          setAlertState={setAlertState}
          viewAlertDetails={viewAlertDetails}
        >
          {' '}
        </EventBasedAlert>
      )}
      <ConfirmationModal
        visible={deleteWidgetModal ? true : false}
        confirmationText='Do you really want to remove this alert?'
        onOk={confirmDelete}
        onCancel={showDeleteWidgetModal.bind(this, false)}
        title='Remove Alert'
        okText='Confirm'
        cancelText='Cancel'
        confirmLoading={deleteApiCalled}
      />
    </div>
  );
};

const mapStateToProps = (state) => ({
  activeProject: state.global.active_project,
  savedAlerts: state.global.Alerts,
  savedEventAlerts: state.global.eventAlerts,
  kpi: state?.kpi,
  agent_details: state.agent.agent_details,
  slack: state.global.slack,
  projectSettings: state.global.projectSettingsV1,
  currentAgent: state.agent.agent_details,
});

export default connect(mapStateToProps, {
  fetchAlerts,
  deleteAlert,
  fetchEventAlerts,
  deleteEventAlert
})(Alerts);
