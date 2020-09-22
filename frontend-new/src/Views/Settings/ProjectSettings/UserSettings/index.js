import React, { useState, useEffect } from 'react';
import {
  Row, Col, Button, Table
} from 'antd';
import { Text } from 'factorsComponents';

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
    key: 'name'
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
    key: 'actions'
  }
];

function UserSettings() {
  const [dataLoading, setDataLoading] = useState(true);

  useEffect(() => {
    setTimeout(() => {
      setDataLoading(false);
    }, 500);
  });

  return (
    <>
      <div className={'mb-10 pl-4'}>
        <Row>
          <Col span={12}>
            <Text type={'title'} level={3} weight={'bold'} extraClass={'m-0'}>Users and Roles</Text>
          </Col>
          <Col span={12}>
            <div className={'flex justify-end'}>
              <Button disabled={dataLoading}>Invite Users</Button>
            </div>
          </Col>
        </Row>
        <Row className={'mt-8'}>
          <Col span={24}>
            <Table dataSource={dataSource} columns={columns} pagination={false} />
          </Col>
        </Row>
      </div>

    </>

  );
}

export default UserSettings;
