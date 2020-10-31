import React, { useState, useEffect } from 'react';
import Header from '../AppLayout/Header';
import SearchBar from '../../components/SearchBar';
import {
  Row, Col, Table, Avatar, Button
} from 'antd';
import { Text, SVG } from 'factorsComponents';
import { PlusOutlined, SlackOutlined } from '@ant-design/icons';
import ConfigureDP from './ConfigureDP';
import CreateGoalDrawer from './CreateGoalDrawer';

const columns = [
  {
    title: 'Saved Goals',
    dataIndex: 'title',
    key: 'title'
  },
  {
    title: 'Created By',
    dataIndex: 'author',
    key: 'author',
    render: (text) => <div className="flex items-center">
        <Avatar src="assets/avatar/avatar.png" className={'mr-2'} />&nbsp; {text} </div>
  }
];

const data = [
  {
    key: '1',
    type: <SVG name={'events_cq'} size={24} />,
    title: 'Monthly User signups from Google Campaigns',
    author: 'Vishnu Baliga',
    date: 'Jan 10, 2020'
  },
  {
    key: '2',
    type: <SVG name={'attributions_cq'} size={24} />,
    title: 'Quarterly Lead Acquisition Rate by Region',
    author: 'Praveen Das',
    date: 'Feb 21, 2020'
  },
  {
    key: '3',
    type: <SVG name={'funnels_cq'} size={24} />,
    title: 'Onboarding Funnel Over month',
    author: 'Anand Nair',
    date: 'Jan 04, 2020'
  }
];

const suggestionList = [
  {
    name: 'User_SignUp',
    img: ''
  },
  {
    name: 'chargebee.com...ebinar/Thank-you',
    img: ''
  },
  {
    name: 'chargebee.com/Plans',
    img: ''
  },
  {
    name: 'chargebee.com/Features',
    img: ''
  },
  {
    name: 'chargebee.com...ebinar/Thank-you',
    img: ''
  }
];

const Factors = () => {
  const [loadingTable, SetLoadingTable] = useState(true);
  const [showConfigureDPModal, setConfigureDPModal] = useState(false);
  const [showGoalDrawer, setGoalDrawer] = useState(false);

  useEffect(() => {
    setInterval(() => {
      SetLoadingTable(false);
    }, 2000);
  });
  const handleCancel = () => {
    setConfigureDPModal(false);
  };
  return (
    <>
         <Header>
                <div className="w-full h-full py-4 flex flex-col justify-center items-center">
                    <SearchBar />
                </div>
            </Header>

            <div className={'fa-container mt-24'}>
                <Row gutter={[24, 24]}>
                    <Col span={8} className={'border-right--thin-2'}>
                        <Row gutter={[24, 24]} className={'p-4'}>
                            <Col span={24}>
                                <Text type={'title'} level={5} weight={'bold'} extraClass={'m-0'} >What’s being tracked?</Text>
                                <Text type={'title'} level={7} extraClass={'m-0'} >Factors periodically track a pre-configured set of data points for faster and efficient retrieval of insights. </Text>
                                <Button className={'m-0 mt-4'} size={'large'} onClick={() => setConfigureDPModal(true)}>Configure Data Points</Button>
                            </Col>
                            <Col span={24}>
                                <Text type={'title'} level={7} weight={'bold'} extraClass={'mt-8'} >Suggestions based on your activity</Text>
                                {suggestionList.map((item, index) => {
                                  return (
                                    <div key={index} className={'flex justify-between items-center mt-2'}>
                                        <Text type={'title'} level={7} weight={'thin'} extraClass={'m-0'} ><SlackOutlined className={'mr-1'} />{item.name}</Text>
                                        <Button size={'small'} icon={<PlusOutlined />}></Button>
                                    </div>
                                  );
                                })}
                            </Col>
                        </Row>
                    </Col>
                <Col span={16}>
                    <Row gutter={[24, 24]} justify="center">
                        <Col span={20}>
                            <Text type={'title'} level={3} weight={'bold'} extraClass={'m-0'} >Factors</Text>
                            <Text type={'title'} level={5} extraClass={'m-0 mt-2'} >Periodically track website events, pages, user properties that are important to you and get insights that influence your goals.</Text>
                        </Col>
                        <Col span={20}>
                            <Button size={'large'} type={'primary'} onClick={() => setGoalDrawer(true)}>Create a New Goal</Button>
                            <a className={'ml-4'}>learn more</a>
                        </Col>
                    </Row>
                    <Row gutter={[24, 24]} justify="center">
                        <Col span={20}>
                        <Table loading={loadingTable} className="ant-table--custom mt-8" columns={columns} dataSource={data} pagination={false} />
                        </Col>
                    </Row>
                </Col>
                </Row>
            </div>

            <ConfigureDP
            visible={showConfigureDPModal}
            handleCancel={handleCancel}
            />

            <CreateGoalDrawer
                visible={showGoalDrawer}
                onClose={() => setGoalDrawer(false)}
            />

    </>
  );
};

export default Factors;
