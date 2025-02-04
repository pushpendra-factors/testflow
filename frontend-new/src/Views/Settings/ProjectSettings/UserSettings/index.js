import React, { useState, useEffect } from 'react';
import {
  Row,
  Col,
  Button,
  Table,
  Avatar,
  Menu,
  Dropdown,
  Modal,
  message
} from 'antd';
import { Text } from 'factorsComponents';
import { MoreOutlined, ExclamationCircleOutlined } from '@ant-design/icons';
import { connect } from 'react-redux';
import {
  fetchProjectAgents,
  projectAgentRemove,
  updateAgentRole
} from 'Reducers/agentActions';
import MomentTz from 'Components/MomentTz';
import CommonSettingsHeader from 'Components/GenericComponents/CommonSettingsHeader';
import InviteUsers from './InviteUsers';

const { confirm } = Modal;

function UserSettings({
  agents,
  currentAgent,
  projectAgentRemove,
  activeProjectID,
  updateAgentRole,
  fetchProjectAgents
}) {
  const [dataLoading, setDataLoading] = useState(true);
  const [dataSource, setdataSource] = useState(null);
  const [inviteModal, setInviteModal] = useState(false);
  const [confirmLoading, setConfirmLoading] = useState(false);

  const confirmRemove = (uuid) => {
    const agent = agents.filter((agent) => agent.email === currentAgent.email);
    confirm({
      title: 'Do you want to remove this user?',
      icon: <ExclamationCircleOutlined />,
      content: 'Please confirm to proceed',
      okText: 'Yes',
      onOk() {
        projectAgentRemove(activeProjectID, uuid)
          .then(() => {
            fetchProjectAgents(activeProjectID);
            const rulesToUpdate = [
              ...dataSource.filter(
                (val) => JSON.stringify(val.uuid) !== JSON.stringify(uuid)
              )
            ];
            setdataSource(rulesToUpdate);
            message.success('User removed successfully!');
          })
          .catch((err) => {
            message.error(err?.data?.error);
          });
      }
    });
  };

  const confirmRoleChange = (uuid) => {
    confirm({
      title: "Do you want to change this user's role?",
      icon: <ExclamationCircleOutlined />,
      content: 'Please confirm to proceed',
      okText: 'Yes',
      onOk() {
        updateAgentRole(activeProjectID, uuid, 2)
          .then(() => {
            fetchProjectAgents(activeProjectID);
            const rulesToUpdate = [
              ...dataSource.filter(
                (val) => JSON.stringify(val.uuid) !== JSON.stringify(uuid)
              )
            ];
            setdataSource(rulesToUpdate);
            message.success('User role updated successfully!');
          })
          .catch((err) => {
            message.error(err.data.error);
          });
      }
    });
  };

  const menu = (values) => (
    <Menu>
      <Menu.Item key='0' onClick={() => confirmRemove(values.uuid)}>
        <a>Remove User</a>
      </Menu.Item>
      {values.role === 1 && (
        <Menu.Item key='1' onClick={() => confirmRoleChange(values.uuid)}>
          <a>Make Project Admin</a>
        </Menu.Item>
      )}
    </Menu>
  );

  const columns = [
    {
      title: 'Name',
      dataIndex: 'name',
      key: 'name',
      render: (text) => (
        <div className='flex items-center'>
          <Avatar src='assets/avatar/avatar.png' className='mr-2' size={24} />
          <Text type='title' level={7} weight='bold' extraClass='m-0 ml-2'>
            {' '}
            {text}
          </Text>{' '}
        </div>
      )
    },
    {
      title: 'Email',
      dataIndex: 'email',
      key: 'email'
    },
    {
      title: 'Role',
      dataIndex: 'role',
      key: 'role'
    },
    {
      title: 'Last activity',
      dataIndex: 'lastActivity',
      key: 'lastActivity'
    },
    {
      title: '',
      dataIndex: 'actions',
      key: 'actions',
      render: (values) => (
        <Dropdown overlay={() => menu(values)} trigger={['click']}>
          <Button size='large' type='text' icon={<MoreOutlined />} />
        </Dropdown>
      )
    }
  ];

  useEffect(() => {
    if (agents) {
      const formattedArray = [];
      agents.map((agent, index) => {
        // console.log(index, 'agent-name-->', agent.first_name);
        const values = {
          uuid: `${agent.uuid}`,
          role: agent.role
        };
        formattedArray.push({
          key: index,
          name: `${
            agent.first_name || agent.last_name
              ? `${agent.first_name} ${agent.last_name}`
              : '---'
          }`,
          email: agent.email,
          role: `${agent.role === 2 ? 'Admin' : 'User'}`,
          lastActivity: `${
            agent.last_logged_in
              ? MomentTz(agent.last_logged_in).fromNow()
              : !agent.is_email_verified
                ? 'Pending Invite'
                : '---'
          }`,
          actions: values
        });
        setdataSource(formattedArray);
      });
    } else {
      setdataSource([]);
    }
    setDataLoading(false);
  }, [agents]);

  const handleOk = () => {
    setConfirmLoading(true);
    setTimeout(() => {
      setInviteModal(false);
      setConfirmLoading(false);
    }, 2000);
  };

  return (
    <div className='fa-container'>
      <Row gutter={[24, 24]} justify='center'>
        <Col span={24}>
          <div className='mb-10'>
            <CommonSettingsHeader
              title='Members'
              description="Manage your project's team by adjusting roles, inviting new members, and overseeing user access."
              actionsNode={
                <div className='flex justify-end'>
                  <Button
                    size='large'
                    disabled={dataLoading}
                    onClick={() => setInviteModal(true)}
                  >
                    Invite Users
                  </Button>
                </div>
              }
            />
            <Row className='mt-2'>
              <Col span={24}>
                <Table
                  className='fa-table--basic'
                  loading={dataLoading}
                  dataSource={dataSource}
                  columns={columns}
                  pagination={false}
                />
              </Col>
            </Row>
          </div>

          <InviteUsers
            visible={inviteModal}
            onCancel={() => setInviteModal(false)}
            onOk={() => handleOk()}
            confirmLoading={confirmLoading}
          />
        </Col>
      </Row>
    </div>
  );
}

const mapStateToProps = (state) => ({
  activeProjectID: state.global.active_project.id,
  agents: state.agent.agents,
  currentAgent: state.agent.agent_details
});

export default connect(mapStateToProps, {
  fetchProjectAgents,
  updateAgentRole,
  projectAgentRemove
})(UserSettings);

// table datasource example
// {
//   key: '1',
//   name: 'Anand Nair',
//   email: 'anand@uxfish.com',
//   role: 'Owner',
//   lastActivity: 'Yesterday',
//   actions: ''
// }
