import React, { useCallback, useEffect, useState } from 'react';
import { connect, useDispatch, useSelector } from 'react-redux';
import { useHistory } from 'react-router-dom';
import {
  Avatar,
  Button,
  Col,
  Dropdown,
  Menu,
  Row,
  Table,
  message,
  notification
} from 'antd';
import { MoreOutlined } from '@ant-design/icons';
import MomentTz from 'Components/MomentTz';
import { SVG, Text } from 'Components/factorsComponents';
import {
  QUERY_TYPE_EVENT,
  QUERY_TYPE_KPI,
  QUERY_TYPE_WEB
} from 'Utils/constants';
import { getQueryType } from 'Utils/dataFormatter';
// import { fetchWeeklyIngishts } from 'Reducers/insights';
import ConfirmationModal from 'Components/ConfirmationModal';
import { deleteQuery } from 'Reducers/coreQuery/services';
import {
  createAlert,
  enableSlackIntegration,
  fetchProjectSettingsV1,
  fetchSlackChannels,
  sendAlertNow
} from 'Reducers/global';
import AppModal from 'Components/AppModal/AppModal';
import ShareToSlackModal from 'Components/ShareToSlackModal/ShareToSlackModal';
import ShareToEmailModal from 'Components/ShareToEmailModal/ShareToEmailModal';
import { featureLock } from 'Routes/feature';
import useAgentInfo from 'hooks/useAgentInfo';
import styles from './index.module.scss';

const columns = [
  {
    title: 'Type',
    dataIndex: 'type',
    width: 60,
    key: 'type'
  },
  {
    title: 'Title of the Report',
    dataIndex: 'title',
    key: 'title',
    render: (text) => (
      <Text type='title' level={7} weight='bold' extraClass='m-0'>
        {text}
      </Text>
    )
  },
  {
    title: 'Created By',
    dataIndex: 'author',
    width: 240,
    key: 'author',
    render: (created_by_user) => (
      <div className='flex items-center'>
        <Avatar
          src={
            typeof created_by_user?.email === 'string' &&
            created_by_user?.email?.length !== 0 &&
            created_by_user.email.split('@')[1] === 'factors.ai'
              ? 'https://s3.amazonaws.com/www.factors.ai/assets/img/product/factors-icon.svg'
              : created_by_user?.image
                ? created_by_user?.image
                : 'assets/avatar/avatar.png'
          }
          size={24}
          className='mr-2'
        />
        &nbsp; {created_by_user?.text}
      </div>
    )
  },
  {
    title: 'Date',
    dataIndex: 'date',
    width: 240,
    key: 'date'
  }
];

const SavedQueriesTable = ({
  sendAlertNow,
  enableSlackIntegration,
  createAlert,
  fetchProjectSettingsV1,
  fetchSlackChannels
}) => {
  const dispatch = useDispatch();
  // const { metadata } = useSelector((state) => state.insights);

  const [selectedRow, setSelectedRow] = useState(null);
  const [loading, setLoading] = useState(false);
  const [showShareToSlackModal, setShowShareToSlackModal] = useState(false);
  const [showShareToEmailModal, setShowShareToEmailModal] = useState(false);

  const [channelOpts, setChannelOpts] = useState([]);
  const [allChannels, setAllChannels] = useState([]);
  const { email } = useAgentInfo();
  const { agent_details } = useSelector((state) => state.agent);
  const [deleteModal, showDeleteModal] = useState(false);
  const history = useHistory();
  const activeProjectProfilePicture = useSelector(
    (state) => state.global.active_project.profile_picture
  );
  const { slack } = useSelector((state) => state.global);
  const { projectSettingsV1 } = useSelector((state) => state.global);
  const activeProject = useSelector((state) => state.global.active_project);
  const queriesState = useSelector((state) => state.queries);

  const showSlackModal = useCallback((row, event) => {
    event.stopPropagation();
    event.preventDefault();
    setShowShareToSlackModal(true);
    setSelectedRow(row);
  }, []);

  const showEmailModal = useCallback((row, event) => {
    event.stopPropagation();
    event.preventDefault();
    setShowShareToEmailModal(true);
    setSelectedRow(row);
  }, []);

  const handleRowClick = (record) => {
    const query = queriesState.data.find((query) => query.id === record.key);
    if (query != null) {
      let analyseQueryParamsPath = '/analyse';
      if (query?.query?.query_group?.[0]?.cl === 'events') {
        analyseQueryParamsPath = `${analyseQueryParamsPath}/events/${query.id_text}`;
      } else if (query?.query?.cl === 'funnel') {
        analyseQueryParamsPath = `${analyseQueryParamsPath}/funnel/${query.id_text}`;
      }

      history.push({
        pathname: analyseQueryParamsPath,
        state: {
          query,
          global_search: true,
          navigatedFromDashboard: query
        }
      });
    }
  };

  const onConnectSlack = () => {
    enableSlackIntegration(activeProject.id)
      .then((r) => {
        if (r.status === 200) {
          window.open(r.data.redirectURL, '_blank');
          setShowShareToSlackModal(false);
        }
        if (r.status >= 400) {
          message.error('Error fetching Slack redirect url');
        }
      })
      .catch((err) => {
        console.log('Slack error-->', err);
      });
  };

  const handleSlackClick = ({ data, frequency, onSuccess }) => {
    setLoading(true);

    let slackChannels = {};
    const selected = allChannels.filter((c) => c.id === data.channel);
    const map = new Map();
    map.set(agent_details.uuid, selected);
    for (const [key, value] of map) {
      slackChannels = { ...slackChannels, [key]: value };
    }

    const payload = {
      alert_name: selectedRow?.title || data?.subject,
      alert_type: 3,
      // "query_id": selectedRow?.key || selectedRow?.id,
      alert_description: {
        message: data?.message,
        date_range: frequency === 'send_now' ? '' : frequency,
        subject: data?.subject
      },
      alert_configuration: {
        email_enabled: false,
        slack_enabled: true,
        emails: [],
        slack_channels_and_user_groups: slackChannels
      }
    };

    if (frequency === 'send_now') {
      sendAlertNow(
        activeProject.id,
        payload,
        selectedRow?.key || selectedRow?.id,
        { from: '', to: '' },
        false
      )
        .then((r) => {
          notification.success({
            message: 'Report Sent Successfully',
            description: 'Report has been sent to the selected Slack channel',
            duration: 5
          });
        })
        .catch((err) => {
          message.error(err?.data?.error);
        });
    } else {
      createAlert(
        activeProject.id,
        payload,
        selectedRow?.key || selectedRow?.id
      )
        .then((r) => {
          notification.success({
            message: 'Report Saved Successfully',
            description: 'Report will be sent on the specified date.',
            duration: 5
          });
        })
        .catch((err) => {
          message.error(err?.data?.error);
        });
    }
    setLoading(false);
    onSuccess();
  };

  const handleDelete = useCallback((row, event) => {
    event.stopPropagation();
    event.preventDefault();
    setSelectedRow(row);
    showDeleteModal(true);
  }, []);

  const handleEmailClick = ({ data, frequency, onSuccess }) => {
    setLoading(true);

    let emails = [];
    if (data?.emails) {
      emails = data.emails.map((item) => item.email);
    }
    if (data.email) {
      emails.push(data.email);
    }

    const payload = {
      alert_name: selectedRow?.title || data?.subject,
      alert_type: 3,
      // "query_id": selectedRow?.key || selectedRow?.id,
      alert_description: {
        message: data?.message,
        date_range: frequency === 'send_now' ? '' : frequency,
        subject: data?.subject
      },
      alert_configuration: {
        email_enabled: true,
        slack_enabled: false,
        emails,
        slack_channels_and_user_groups: {}
      }
    };

    if (frequency === 'send_now') {
      sendAlertNow(
        activeProject.id,
        payload,
        selectedRow?.key || selectedRow?.id,
        { from: '', to: '' },
        false
      )
        .then((r) => {
          notification.success({
            message: 'Report Sent Successfully',
            description: 'Report has been sent to the selected emails',
            duration: 5
          });
        })
        .catch((err) => {
          message.error(err?.data?.error);
        });
    } else {
      createAlert(
        activeProject.id,
        payload,
        selectedRow?.key || selectedRow?.id
      )
        .then((r) => {
          notification.success({
            message: 'Report Saved Successfully',
            description: 'Report will be sent on the specified date.',
            duration: 5
          });
        })
        .catch((err) => {
          message.error(err?.data?.error);
        });
    }
    setLoading(false);
    onSuccess();
  };

  const confirmDelete = useCallback(() => {
    const queryDetails = {
      ...selectedRow,
      project_id: activeProject?.id
    };
    dispatch(deleteQuery(queryDetails));
    setSelectedRow(null);
    showDeleteModal(false);
  }, [activeProject?.id, selectedRow, dispatch]);

  const getMenu = ({ row }) => (
    <Menu className={`${styles.antdActionMenu}`}>
      <Menu.Item key='0'>
        <div onClick={handleRowClick.bind(this, row)}>
          <SVG name='eye' size={18} color='grey' extraClass='inline mr-2' />
          View Report
        </div>
      </Menu.Item>
      {getQueryType(row.query) === QUERY_TYPE_KPI ||
      getQueryType(row.query) === QUERY_TYPE_EVENT ? (
        <Menu.Item key='1'>
          <a onClick={showEmailModal.bind(this, row)} href='#!'>
            <SVG
              name='envelope'
              size={18}
              color='grey'
              extraClass='inline mr-2'
            />
            Email this report
          </a>
        </Menu.Item>
      ) : null}
      {getQueryType(row.query) === QUERY_TYPE_KPI ||
      getQueryType(row.query) === QUERY_TYPE_EVENT ? (
        <Menu.Item key='2'>
          <a onClick={showSlackModal.bind(this, row)} href='#!'>
            <SVG
              name='SlackStroke'
              size={18}
              color='grey'
              extraClass='inline mr-2'
            />
            Share to Slack
          </a>
        </Menu.Item>
      ) : null}
      <Menu.Item key='3'>
        <a onClick={handleDelete.bind(this, row)} href='#!'>
          <SVG name='trash' size={18} color='grey' extraClass='inline mr-2' />
          Delete Report
        </a>
      </Menu.Item>
    </Menu>
  );

  const getFormattedRow = (q) => {
    const requestQuery = q.query;
    const queryType = getQueryType(q.query);
    const queryTypeName = {
      events: 'events_cq',
      funnel: 'funnels_cq',
      channel_v1: 'campaigns_cq',
      attribution: 'attributions_cq',
      profiles: 'profiles_cq',
      kpi: 'KPI_cq'
    };
    let svgName = '';
    Object.entries(queryTypeName).forEach(([k, v]) => {
      if (queryType === k) {
        svgName = v;
      }
    });

    return {
      key: q.id,
      id_text: q.id_text,
      type: <SVG name={svgName} size={24} color='blue' />,
      title: q.title,
      author: {
        image: activeProjectProfilePicture,
        text: q.created_by_name,
        email: q.created_by_email
      },
      settings: q.settings,
      date: (
        <div className='flex justify-between items-center'>
          <div>{MomentTz(q.created_at).format('MMM DD, YYYY')}</div>
          <div>
            <Dropdown overlay={getMenu({ row: q })} placement='bottomRight'>
              <Button type='text' icon={<MoreOutlined />} />
            </Dropdown>
          </div>
        </div>
      ),
      query: requestQuery,
      actions: ''
    };
  };

  const onRow = (record) => ({
    onClick: () => handleRowClick(record)
  });

  const data = queriesState.data
    .filter((q) => !(q.query && q.query.cl === QUERY_TYPE_WEB))
    .map((q) => getFormattedRow(q));

  useEffect(() => {
    fetchProjectSettingsV1(activeProject.id);
    if (projectSettingsV1?.int_slack) {
      fetchSlackChannels(activeProject.id);
    }
  }, [
    activeProject,
    fetchProjectSettingsV1,
    fetchSlackChannels,
    projectSettingsV1?.int_slack,
    showShareToSlackModal
  ]);

  useEffect(() => {
    if (slack?.length > 0) {
      const tempArr = [];
      const allArr = [];
      for (let i = 0; i < slack.length; i++) {
        tempArr.push({ label: `#${slack[i].name}`, value: slack[i].id });
        allArr.push({
          name: slack[i].name,
          id: slack[i].id,
          is_private: slack[i].is_private
        });
      }
      setChannelOpts(tempArr);
      setAllChannels(allArr);
    }
  }, [activeProject, agent_details, slack]);

  return (
    <>
      <Table
        onRow={onRow}
        // loading={queriesState.loading}
        className='fa-table--basic'
        columns={columns}
        dataSource={data}
        pagination
        rowClassName='cursor-pointer'
      />
      <ConfirmationModal
        visible={deleteModal}
        confirmationText='Are you sure you want to delete this report?'
        onOk={confirmDelete}
        onCancel={showDeleteModal.bind(this, false)}
        title='Delete Report'
        okText='Confirm'
        cancelText='Cancel'
      />
      <ShareToEmailModal
        visible={showShareToEmailModal}
        onSubmit={handleEmailClick}
        isLoading={loading}
        setShowShareToEmailModal={setShowShareToEmailModal}
        queryTitle={selectedRow?.title}
      />
      {projectSettingsV1?.int_slack ? (
        <ShareToSlackModal
          visible={showShareToSlackModal}
          onSubmit={handleSlackClick}
          channelOpts={channelOpts}
          isLoading={loading}
          setShowShareToSlackModal={setShowShareToSlackModal}
          queryTitle={selectedRow?.title}
        />
      ) : (
        <AppModal
          title={null}
          visible={showShareToSlackModal}
          footer={null}
          centered
          mask
          maskClosable={false}
          maskStyle={{ backgroundColor: 'rgb(0 0 0 / 70%)' }}
          closable
          isLoading={loading}
          onCancel={() => setShowShareToSlackModal(false)}
          className='fa-modal--regular'
          width='470px'
        >
          <div className='m-0 mb-2'>
            <Row className='m-0'>
              <Col>
                <SVG name='Slack' size={25} extraClass='inline mr-2 -mt-2' />
                <Text
                  type='title'
                  level={5}
                  weight='bold'
                  extraClass='inline m-0'
                >
                  Slack Integration
                </Text>
              </Col>
            </Row>
            <Row className='m-0 mt-4'>
              <Col>
                <Text
                  type='title'
                  level={6}
                  color='grey-2'
                  weight='regular'
                  extraClass='m-0'
                >
                  Slack is not integrated, Do you want to integrate with your
                  Slack account now?
                </Text>
              </Col>
            </Row>
            <Col>
              <Row justify='end' className='w-full mb-1 mt-4'>
                <Col className='mr-2'>
                  <Button
                    type='default'
                    onClick={() => setShowShareToSlackModal(false)}
                  >
                    Cancel
                  </Button>
                </Col>
                <Col className='mr-2'>
                  <Button type='primary' onClick={onConnectSlack}>
                    Connect to Slack
                  </Button>
                </Col>
              </Row>
            </Col>
          </div>
        </AppModal>
      )}
    </>
  );
};

export default connect(undefined, {
  createAlert,
  sendAlertNow,
  fetchSlackChannels,
  fetchProjectSettingsV1,
  enableSlackIntegration
})(SavedQueriesTable);
