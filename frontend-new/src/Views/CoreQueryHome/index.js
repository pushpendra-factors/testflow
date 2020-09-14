
import React, { useState, useEffect } from 'react';
import { Text, SVG } from '../../components/factorsComponents';
import { Row, Col, Table, Avatar, Button } from 'antd';
import Header from '../AppLayout/Header';
import SearchBar from '../../components/SearchBar';


const coreQueryoptions = [
  {
    title: 'Events',
    icon: 'events_cq',
    desc: 'Create charts from events and related properties'
  },
  {
    title: 'Funnels',
    icon: 'funnels_cq',
    desc: 'Find how users are navigating a defined path'
  },
  {
    title: 'Campaigns',
    icon: 'campaigns_cq',
    desc: 'Find the effect of your marketing campaigns'
  },
  {
    title: 'Attributions',
    icon: 'attributions_cq',
    desc: 'Analyse Multi Touch Attributions'
  },
  {
    title: 'Templates',
    icon: 'templates_cq',
    desc: 'A list of advanced queries crafter by experts'
  },
];


const columns = [
  {
    title: 'Type',
    dataIndex: 'type',
    width: 60,
    key: 'type',
  },
  {
    title: 'Title of the Query',
    dataIndex: 'title',
    key: 'title',
  },
  {
    title: 'Created By',
    dataIndex: 'author',
    key: 'author',
    render: text => <div className="flex items-center">
      <Avatar src="assets/avatar/avatar.png" className={'mr-2'} />&nbsp; {text} </div>,
  },
  {
    title: 'Date',
    dataIndex: 'date',
    key: 'date',
  },
];

const data = [
  {
    key: '1',
    type: <SVG name={`events_cq`} size={24} />,
    title: `Monthly User signups from Google Campaigns`,
    author: `Vishnu Baliga`,
    date: `Jan 10, 2020`,
  },
  {
    key: '2',
    type: <SVG name={`attributions_cq`} size={24} />,
    title: `Quarterly Lead Acquisition Rate by Region`,
    author: `Praveen Das`,
    date: `Feb 21, 2020`,
  },
  {
    key: '3',
    type: <SVG name={`funnels_cq`} size={24} />,
    title: `Onboarding Funnel Over month`,
    author: `Anand Nair`,
    date: `Jan 04, 2020`,
  },
  {
    key: '4',
    type: <SVG name={`events_cq`} size={24} />,
    title: `Check out by category`,
    author: `Akhil Nair`,
    date: `Jan 06, 2020`,
  },
  {
    key: '5',
    type: <SVG name={`events_cq`} size={24} />,
    title: `Quarterly Lead Acquisition Rate by Region`,
    author: `Jitesh Kriplani`,
    date: `Feb 11, 2020`,
  },
  {
    key: '6',
    type: <SVG name={`campaigns_cq`} size={24} />,
    title: `Monthly User signups from Google Campaigns`,
    author: `Anand Nair`,
    date: `Mar 14, 2020`,
  },
];

function CoreQuery({ setDrawerVisible }) {
  const [loadingTable, SetLoadingTable] = useState(true);

  useEffect(()=>{
    setInterval(() => {
      SetLoadingTable(false)
    }, 2000);
  })
  return (
    <>
      <Header>
        <div className="w-full h-full py-4 flex flex-col justify-center items-center">
          <SearchBar />
        </div>
      </Header>
      <div className={"fa-container mt-24"}>
        <Row gutter={[24, 24]} justify="center">
            <Col span={20}>
              <Text type={'title'} level={2} weight={'bold'} extraClass={`m-0`} >Core Query</Text>
              <Text type={'title'} level={5} weight={'regular'} color={`grey`} extraClass={`m-0`} >Use these tools to Analyse and get to the bottom of User Behaviors and Marketing Funnels</Text>
            </Col>
          </Row>
          <Row gutter={[24, 24]} justify="center" className={'mt-10'}>
            {coreQueryoptions.map((item, index) => {
              return (
                <Col span={4} key={index}>
                  <div onClick={() => setDrawerVisible(item.title == `Funnels`)} className="fai--custom-card flex justify-center items-center flex-col ">
                      <div className={`fai--custom-card--icon`}><SVG name={item.icon} size={48} /> </div> 
                    <div className="flex justify-start items-center flex-col before-hover">
                        <Text type={'title'} level={3} weight={'bold'} extraClass={`fai--custom-card--title`} >{item.title}</Text>  
                    </div>
                    <div className="flex justify-start items-center flex-col after-hover"> 
                      <div className={`fai--custom-card--content flex-col flex justify-start items-center`}> 
                        <Text type={'title'} level={7} weight={'bold'} extraClass={`fai--custom-card--desc`} >{item.desc}</Text>
                        <a className={`fai--custom-card--cta`}>New Query <SVG name={`next`} size={20} /> </a>
                      </div>
                    </div>
                  </div>
                </Col>
              )
            })}
          </Row>

          <Row justify="center" className={'mt-12'}>
            <Col span={20}>
              <Row className={`flex justify-between items-center`}>
                <Col span={10}>
                  <Text type={'title'} level={4} weight={'bold'} extraClass={`m-0`} >Saved Queries</Text>
                </Col>
                <Col span={5} >
                  <div className={`flex flex-row justify-end items-end `}>
                    <Button icon={<SVG name={`help`} size={12} color={'grey'} />} type="text">Learn More</Button>
                  </div>
                </Col>
              </Row>
            </Col>
          </Row>
          <Row justify="center" className={'mt-2 mb-20'}>
            <Col span={20}>
              <Table loading={loadingTable} className="ant-table--custom" columns={columns} dataSource={data} pagination={false} />
            </Col>
          </Row>
      </div>
    </>
  )
}

export default CoreQuery;