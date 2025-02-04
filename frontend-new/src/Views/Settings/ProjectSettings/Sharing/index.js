import React, { useState, useEffect, useCallback } from 'react';
import { connect } from 'react-redux';
import {
  Row,
  Col,
  Select,
  Menu,
  Dropdown,
  Button,
  Table,
  Input,
  notification,
  Checkbox
} from 'antd';
import { Text, SVG } from 'factorsComponents';
import { MoreOutlined, PlusOutlined } from '@ant-design/icons';
import _ from 'lodash';
import { fetchSharedAlerts, deleteAlert } from 'Reducers/global';
import CommonSettingsHeader from 'Components/GenericComponents/CommonSettingsHeader';
import ConfirmationModal from '../../../../components/ConfirmationModal';

const { Option } = Select;

const Sharing = ({
  activeProject,
  fetchSharedAlerts,
  deleteAlert,
  sharedAlerts
}) => {
  const [tableData, setTableData] = useState([]);
  const [tableLoading, setTableLoading] = useState(false);
  const [errorInfo, seterrorInfo] = useState(null);
  const [loading, setLoading] = useState(false);
  const [viewMode, SetViewMode] = useState(false);
  const [viewAlertDetails, setAlertDetails] = useState(false);
  const [viewSelectedChannels, setViewSelectedChannels] = useState([]);

  const [deleteWidgetModal, showDeleteWidgetModal] = useState(false);
  const [deleteApiCalled, setDeleteApiCalled] = useState(false);

  const confirmRemove = (id) =>
    deleteAlert(activeProject.id, id).then(
      (res) => {
        fetchSharedAlerts(activeProject.id);
        notification.success({
          message: 'Success',
          description: 'Unsubscribed Alert successfully ',
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

  const confirmDelete = useCallback(async () => {
    try {
      setDeleteApiCalled(true);
      await confirmRemove(deleteWidgetModal);
      setDeleteApiCalled(false);
      showDeleteWidgetModal(false);
      SetViewMode(false);
    } catch (err) {
      console.log(err);
      console.log(err.response);
    }
  }, [deleteWidgetModal]);

  const menu = (item) => (
    <Menu>
      <Menu.Item
        key='0'
        onClick={() => {
          SetViewMode(true);
          setAlertDetails(item);
        }}
      >
        <a>View</a>
      </Menu.Item>
      <Menu.Item
        key='1'
        onClick={() => {
          showDeleteWidgetModal(item.id);
        }}
      >
        <a>Unsubscribe</a>
      </Menu.Item>
    </Menu>
  );

  const columns = [
    {
      title: 'Name',
      dataIndex: 'alert_name',
      key: 'alert_name',
      render: (text) => (
        <Text type='title' level={7} truncate charLimit={50}>
          {text}
        </Text>
      ),
      width: 400
    },
    {
      title: 'Type',
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
      title: 'Frequncy',
      dataIndex: 'date_range',
      key: 'date_range',
      render: (text) => (
        <Text type='title' level={7} truncate charLimit={25}>
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
        <div>
          <Dropdown
            overlay={() => menu(obj)}
            trigger={['hover']}
            placement='bottomRight'
          >
            <Button
              type='text'
              icon={
                <MoreOutlined style={{ color: 'gray', fontSize: '18px' }} />
              }
            />
          </Dropdown>
        </div>
      )
    }
  ];

  const emailView = () => {
    if (viewAlertDetails.alert_configuration.emails) {
      return viewAlertDetails.alert_configuration.emails.map((item, index) => (
        <div className='mb-3'>
          <Input
            disabled
            key={index}
            value={item}
            className='fa-input'
            placeholder='yourmail@gmail.com'
          />
        </div>
      ));
    }
  };

  useEffect(() => {
    if (viewAlertDetails?.alert_configuration?.slack_channels_and_user_groups) {
      const obj =
        viewAlertDetails?.alert_configuration?.slack_channels_and_user_groups;
      for (const key in obj) {
        if (obj[key].length > 0) {
          setViewSelectedChannels(obj[key]);
        }
      }
    }
  }, [viewAlertDetails]);

  useEffect(() => {
    setTableLoading(true);
    fetchSharedAlerts(activeProject.id).then(() => {
      setTableLoading(false);
    });
  }, [activeProject]);

  useEffect(() => {
    if (sharedAlerts) {
      const savedArr = [];
      sharedAlerts?.map((item, index) => {
        savedArr.push({
          key: index,
          alert_name: item.alert_name,
          dop: `${item.alert_configuration.email_enabled ? 'Email' : ''} ${
            item.alert_configuration.slack_enabled ? 'Slack' : ''
          }`,
          date_range:
            item?.alert_description?.date_range === 'last_week'
              ? 'Weekly'
              : item?.alert_description?.date_range,
          actions: item
        });
      });
      setTableData(savedArr);
    } else {
      setTableData([]);
    }
  }, [sharedAlerts]);

  return (
    <div className='fa-container'>
      <Row gutter={[24, 24]} justify='center'>
        <Col span={24}>
          <div className='mb-10'>
            {!viewMode && (
              <>
                <CommonSettingsHeader
                  title='Sharing'
                  description='Effortlessly manage automated report sharing to keep your team informed and aligned.'
                />

                <Row className='m-0'>
                  <Col span={24}>
                    <div className='m-0 '>
                      <Table
                        className='fa-table--basic'
                        onRow={(record, rowIndex) => ({
                          onClick: (event) => {
                            SetViewMode(true);
                            setAlertDetails(record.actions);
                          } // click row
                        })}
                        columns={columns}
                        dataSource={tableData}
                        pagination={false}
                        loading={tableLoading}
                        tableLayout='fixed'
                        rowClassName='cursor-pointer'
                      />
                    </div>
                  </Col>
                </Row>
              </>
            )}

            {viewMode && (
              <>
                <CommonSettingsHeader
                  title='View Shared Report'
                  actionsNode={
                    <div className='flex justify-end'>
                      <Button
                        size='large'
                        disabled={loading}
                        onClick={() => {
                          SetViewMode(false);
                        }}
                      >
                        Back
                      </Button>
                    </div>
                  }
                />

                <Row className='mt-2'>
                  <Col span={18}>
                    <Text
                      type='title'
                      level={7}
                      weight='bold'
                      color='grey-2'
                      extraClass='m-0'
                    >
                      Report name
                    </Text>
                  </Col>
                </Row>
                <Row className='mt-2'>
                  <Col span={8} className='m-0'>
                    <Input
                      disabled
                      className='fa-input'
                      value={viewAlertDetails?.alert_name}
                    />
                  </Col>
                </Row>
                <Row className='mt-2'>
                  <Col span={18}>
                    <Text
                      type='title'
                      level={7}
                      weight='bold'
                      color='grey-2'
                      extraClass='m-0'
                    >
                      Subject
                    </Text>
                  </Col>
                </Row>
                <Row className='mt-2'>
                  <Col span={8} className='m-0'>
                    <Input
                      disabled
                      className='fa-input'
                      value={viewAlertDetails?.alert_description?.subject}
                    />
                  </Col>
                </Row>

                <Row className='mt-2'>
                  <Col span={24}>
                    <div className='border-top--thin-2 pt-2 mt-2' />
                    <Text
                      type='title'
                      level={7}
                      weight='bold'
                      color='grey-2'
                      extraClass='m-0'
                    >
                      Delivery options
                    </Text>
                  </Col>
                </Row>

                <Row className='mt-2 ml-2'>
                  <Col span={4}>
                    <Checkbox
                      disabled
                      checked={
                        viewAlertDetails?.alert_configuration?.email_enabled
                      }
                    >
                      Email
                    </Checkbox>
                  </Col>
                </Row>
                <Row className='mt-4'>
                  <Col span={8}>{emailView()}</Col>
                </Row>
                <Row className='mt-2 ml-2'>
                  <Col span={4}>
                    <Checkbox
                      disabled
                      checked={
                        viewAlertDetails?.alert_configuration?.slack_enabled
                      }
                    >
                      Slack
                    </Checkbox>
                  </Col>
                </Row>
                {viewAlertDetails?.alert_configuration?.slack_enabled &&
                  viewAlertDetails?.alert_configuration
                    ?.slack_channels_and_user_groups && (
                    <Row className='mt-4'>
                      <Col span={8}>
                        {viewSelectedChannels.map((channel, index) => (
                          <div className='mb-3'>
                            <Input
                              disabled
                              key={index}
                              value={`#${channel.name}`}
                              className='fa-input'
                            />
                          </div>
                        ))}
                      </Col>
                    </Row>
                  )}
                <Row className='mt-2'>
                  <Col span={24}>
                    <div className='border-top--thin-2 mt-2 mb-2' />
                    <Text
                      type='title'
                      level={7}
                      weight='bold'
                      color='grey-2'
                      extraClass='m-0'
                    >
                      Message
                    </Text>
                  </Col>
                </Row>
                <Row className='mt-2'>
                  <Col span={8} className='m-0'>
                    <Input
                      disabled
                      className='fa-input'
                      value={viewAlertDetails?.alert_description?.message}
                    />
                  </Col>
                </Row>
                <Row className='mt-2'>
                  <Col span={18}>
                    <Text
                      type='title'
                      level={7}
                      weight='bold'
                      color='grey-2'
                      extraClass='m-0'
                    >
                      Frequency
                    </Text>
                  </Col>
                </Row>
                <Row className='mt-2'>
                  <Col span={8} className='m-0'>
                    <Input
                      disabled
                      className='fa-input'
                      value={
                        viewAlertDetails?.alert_description?.date_range ===
                        'last_week'
                          ? 'Weekly'
                          : viewAlertDetails?.alert_description?.date_range
                      }
                    />
                  </Col>
                </Row>
                <Row className='mt-2'>
                  <Col span={24}>
                    <div className='border-top--thin-2 mt-2 mb-4' />
                    <Button
                      type='text'
                      size='large'
                      style={{ color: '#EE3C3C' }}
                      className='m-0'
                      onClick={() =>
                        showDeleteWidgetModal(viewAlertDetails?.id)
                      }
                    >
                      <SVG
                        name='Delete1'
                        extraClass='-mt-1 -mr-1'
                        size={18}
                        color='#EE3C3C'
                      />
                      Unsubscribe
                    </Button>
                  </Col>
                </Row>
              </>
            )}

            <ConfirmationModal
              visible={!!deleteWidgetModal}
              confirmationText='Do you really want to unsubscribe this alert?'
              onOk={confirmDelete}
              onCancel={showDeleteWidgetModal.bind(this, false)}
              title='Unsubscribe Alert'
              okText='Confirm'
              cancelText='Cancel'
              confirmLoading={deleteApiCalled}
            />
          </div>
        </Col>
      </Row>
    </div>
  );
};

const mapStateToProps = (state) => ({
  activeProject: state.global.active_project,
  sharedAlerts: state.global.sharedAlerts
});

export default connect(mapStateToProps, { fetchSharedAlerts, deleteAlert })(
  Sharing
);
