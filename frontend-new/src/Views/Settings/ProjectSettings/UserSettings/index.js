import React, { useState, useEffect } from 'react';
import {
  Row, Col, Button, Table, Avatar, Menu, Dropdown, Modal, message
} from 'antd';
import { Text } from 'factorsComponents';
import { MoreOutlined, ExclamationCircleOutlined } from '@ant-design/icons';
import InviteUsers from './InviteUsers';
import { connect } from 'react-redux';
import { fetchProjectAgents, projectAgentRemove, updateAgentRole } from 'Reducers/agentActions';
import { fetchDemoProject } from 'Reducers/global';
import MomentTz from 'Components/MomentTz';

const { confirm } = Modal;

function UserSettings({
  agents,
  currentAgent,
  projectAgentRemove,
  activeProjectID,
  updateAgentRole,
  fetchProjectAgents,
  fetchDemoProject
}) {
  const [dataLoading, setDataLoading] = useState(true);
  const [dataSource, setdataSource] = useState(null);
  const [inviteModal, setInviteModal] = useState(false);
  const [confirmLoading, setConfirmLoading] = useState(false);
  const [demoProjectID, setdemoProjectID] = useState(null);

  const confirmRemove = (uuid) => {
    let agent = agents.filter(agent => agent.email === currentAgent.email);
    confirm({
      title: 'Do you want to remove this user?',
      icon: <ExclamationCircleOutlined />,
      content: 'Please confirm to proceed',
      okText: 'Yes',
      onOk() {
        if(agent[0].role === 2) {
          projectAgentRemove(activeProjectID, uuid).then(() => {
            fetchProjectAgents(activeProjectID);
            const rulesToUpdate = [...dataSource.filter((val) => JSON.stringify(val.uuid) !== JSON.stringify(uuid))];
            setdataSource(rulesToUpdate);
            message.success('User removed!');
          }).catch((err) => {
            if(!err) {
              console.log('rm err', err)
              message.error(err?.data?.error);
            } else {
              // temporary fix for now will fix it later
              fetchProjectAgents(activeProjectID);
              const rulesToUpdate = [...dataSource.filter((val) => JSON.stringify(val.uuid) !== JSON.stringify(uuid))];
              setdataSource(rulesToUpdate);
              message.success('User removed!');
            }
          });
        } else {
          message.error('Agent user can not remove other users');
        }
      }
    });
  };

  const confirmRoleChange = (uuid) => {
    confirm({
      title: 'Do you want to change this user\'s role?',
      icon: <ExclamationCircleOutlined />,
      content: 'Please confirm to proceed',
      okText: 'Yes',
      onOk() {
        updateAgentRole(activeProjectID, uuid, 2).then(() => {
          fetchProjectAgents(activeProjectID);
          const rulesToUpdate = [...dataSource.filter((val) => JSON.stringify(val.uuid) !== JSON.stringify(uuid))];
          setdataSource(rulesToUpdate);
          message.success('User role updated!');
        }).catch((err) => {
          message.error(err.data.error); 
        });
      }
    });
  };

  const menu = (values) => {
    return (
    <Menu>
      <Menu.Item key="0" onClick={() => confirmRemove(values.uuid)}>
        <a>Remove User</a>
      </Menu.Item>
      {values.role === 1 &&
        <Menu.Item key="1" onClick={() => confirmRoleChange(values.uuid)}>
          <a>Make Project Admin</a>
        </Menu.Item>
      }
    </Menu>
    );
  };

  const columns = [
    {
      title: 'Name',
      dataIndex: 'name',
      key: 'name',
      render: (text) => <div className="flex items-center">
        <Avatar src="assets/avatar/avatar.png" className={'mr-2'} size={24} /><Text type={'title'} level={7} weight={'bold'} extraClass={'m-0 ml-2'}> {text}</Text> </div>
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
          <Button size={'large'} type="text" icon={<MoreOutlined />} />
        </Dropdown>
      )
    }
  ];

  useEffect(() => {
    fetchDemoProject().then((res) => {
      let id = res.data[0];
      setdemoProjectID(id);
    })
    if(activeProjectID === demoProjectID) {
      setdataSource([]);
    } else
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
          name: `${agent.first_name || agent.last_name ? (agent.first_name + ' ' + agent.last_name) : '---'}`,
          email: agent.email,
          role: `${agent.role === 2 ? 'Admin' : 'User'}`,
          lastActivity: `${agent.last_logged_in ? MomentTz(agent.last_logged_in).fromNow() : !agent.is_email_verified ? 'Pending Invite' : '---'}`,
          actions: values
        });
        setdataSource(formattedArray);
      });
    }
    else{
      setdataSource([]);
    }
    setDataLoading(false);
  }, [agents, demoProjectID]);

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
  agents: state.agent.agents,
  currentAgent: state.agent.agent_details
});

export default connect(mapStateToProps, { fetchProjectAgents, updateAgentRole, projectAgentRemove, fetchDemoProject })(UserSettings);

// table datasource example
// {
//   key: '1',
//   name: 'Anand Nair',
//   email: 'anand@uxfish.com',
//   role: 'Owner',
//   lastActivity: 'Yesterday',
//   actions: ''
// }
