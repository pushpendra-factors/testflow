import React, { useState, useEffect } from 'react';
import {
  Row, Col, Button, Table, Avatar, Menu, Dropdown, Modal, message
} from 'antd';
import { Text } from 'factorsComponents';
import { MoreOutlined, ExclamationCircleOutlined } from '@ant-design/icons';
import InviteUsers from './InviteUsers';
import { connect } from 'react-redux';
import { fetchProjectAgents, projectAgentRemove } from 'Reducers/agentActions';
import moment from 'moment';

const { confirm } = Modal;

function UserSettings({
  agents,
  projectAgentRemove,
  activeProjectID
}) {
  const [dataLoading, setDataLoading] = useState(true);
  const [dataSource, setdataSource] = useState(null);
  const [inviteModal, setInviteModal] = useState(false);
  const [confirmLoading, setConfirmLoading] = useState(false);

  const confirmRemove = (uuid) => {
    confirm({
      title: 'Do you Want to remove this user?',
      icon: <ExclamationCircleOutlined />,
      content: 'Please confirm to proceed',
      okText: 'Yes',
      onOk() {
        projectAgentRemove(activeProjectID, uuid).then(() => {
          const getData = async () => {
            await fetchProjectAgents(activeProjectID);
          };
          message.success('User removed!');
          getData();
        }).catch((err) => {
          message.error(err);
        });
      }
    });
  };

  const menu = (uuid) => {
    return (
    <Menu>
      <Menu.Item key="0" onClick={() => confirmRemove(uuid)}>
        <a>Remove User</a>
      </Menu.Item>
      <Menu.Item key="1">
        <a href="#!">Make Project Admin</a>
      </Menu.Item>
    </Menu>
    );
  };

  const columns = [
    {
      title: 'Name',
      dataIndex: 'name',
      key: 'name',
      render: (text) => <div className="flex items-center">
        <Avatar src="assets/avatar/avatar.png" className={'mr-2'} size={32} />&nbsp; {text} </div>
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
      render: (uuid) => (
        <Dropdown overlay={() => menu(uuid)} trigger={['click']}>
          <Button size={'large'} type="text" icon={<MoreOutlined />} />
        </Dropdown>
      )
    }
  ];

  useEffect(() => {
    if (agents) {
      const formattedArray = [];
      agents.map((agent, index) => {
        // console.log(index, 'agent-name-->', agent.first_name);
        formattedArray.push({
          key: index,
          name: `${agent.first_name || agent.last_name ? (agent.first_name + ' ' + agent.last_name) : '---'}`,
          email: agent.email,
          role: `${agent.role === 2 ? 'Admin' : 'User'}`,
          lastActivity: `${agent.last_logged_in ? moment(agent.last_logged_in).fromNow() : !agent.is_email_verified ? 'Pending Invite' : '---'}`,
          actions: `${agent.uuid}`
        });
        setdataSource(formattedArray);
      });
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
    <>
      <div className={'mb-10 pl-4'}>
        <Row>
          <Col span={12}>
            <Text type={'title'} level={3} weight={'bold'} extraClass={'m-0'}>Users and Roles</Text>
          </Col>
          <Col span={12}>
            <div className={'flex justify-end'}>
              <Button size={'large'} disabled={dataLoading} onClick={() => setInviteModal(true)}>Invite Users</Button>
            </div>
          </Col>
        </Row>
        <Row className={'mt-8'}>
          <Col span={24}>
            <Table className={'fa-table--basic'} loading={dataLoading} dataSource={dataSource} columns={columns} pagination={false} />
          </Col>
        </Row>
      </div>

      <InviteUsers
       visible={inviteModal}
       onCancel={() => setInviteModal(false)}
       onOk={() => handleOk()}
       confirmLoading={confirmLoading}
      />

    </>

  );
}

const mapStateToProps = (state) => ({
  activeProjectID: state.global.active_project.id,
  agents: state.agent.agents
});

export default connect(mapStateToProps, { fetchProjectAgents, projectAgentRemove })(UserSettings);

// table datasource example
// {
//   key: '1',
//   name: 'Anand Nair',
//   email: 'anand@uxfish.com',
//   role: 'Owner',
//   lastActivity: 'Yesterday',
//   actions: ''
// }
