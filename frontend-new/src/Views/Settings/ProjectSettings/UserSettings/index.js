import React, { useState, useEffect } from 'react';
import {
  Row, Col, Button, Table, Avatar, Menu, Dropdown
} from 'antd';
import { Text } from 'factorsComponents';
import { MoreOutlined } from '@ant-design/icons';
import InviteUsers from './InviteUsers';

const menu = (
  <Menu>
    <Menu.Item key="0">
      <a href="#!">Remove User</a>
    </Menu.Item>
    <Menu.Item key="1">
      <a href="#!">Make Project Admin</a>
    </Menu.Item>
  </Menu>
);

const dataSource = [
  {
    key: '1',
    name: 'Anand Nair',
    email: 'anand@uxfish.com',
    role: 'Owner',
    lastActivity: 'Yesterday',
    actions: ''
  },
  {
    key: '2',
    name: 'Vishnu Baliga',
    email: 'baliga@factors.ai',
    role: 'Owner',
    lastActivity: 'Today',
    actions: ''
  },
  {
    key: '3',
    name: 'Praveen Das',
    email: 'praveen@factors.ai',
    role: 'Admin',
    lastActivity: 'A long time ago',
    actions: ''
  },
  {
    key: '4',
    name: 'Aravind Murthy',
    email: 'aravind@factors.ai',
    role: 'Owner',
    lastActivity: 'One hour ago',
    actions: ''
  }
];

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
    render: () => (
      <Dropdown overlay={menu} trigger={['click']}>
        <Button type="text" icon={<MoreOutlined />} />
      </Dropdown>
    )
  }
];

function UserSettings() {
  const [dataLoading, setDataLoading] = useState(true);
  const [inviteModal, setInviteModal] = useState(false);
  const [confirmLoading, setConfirmLoading] = useState(false);

  useEffect(() => {
    setTimeout(() => {
      setDataLoading(false);
    }, 500);
  });

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
              <Button disabled={dataLoading} onClick={() => setInviteModal(true)}>Invite Users</Button>
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

export default UserSettings;
