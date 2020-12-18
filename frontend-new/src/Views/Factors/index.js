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
import { fetchFactorsGoals, fetchFactorsModels } from 'Reducers/factors';
import { connect } from 'react-redux';
import { fetchProjectAgents } from 'Reducers/agentActions';
import { fetchEventNames } from 'Reducers/coreQuery/middleware'; 

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

// const data = [
//   {
//     key: '1',
//     type: <SVG name={'events_cq'} size={24} />,
//     title: 'Monthly User signups from Google Campaigns',
//     author: 'Vishnu Baliga',
//     date: 'Jan 10, 2020'
//   },
// ];

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

const Factors = ({
  fetchFactorsGoals
  , activeProject
  , goals
  , agents
  , fetchProjectAgents
  ,fetchEventNames
  ,fetchFactorsModels
}) => {
  const [loadingTable, SetLoadingTable] = useState(true);
  const [showConfigureDPModal, setConfigureDPModal] = useState(false);
  const [showGoalDrawer, setGoalDrawer] = useState(false);
  const [dataSource, setdataSource] = useState(null);

  useEffect(() => { 
    if (!goals || !agents) {
      const getData = async () => {
        await fetchProjectAgents(activeProject.id);
        await fetchFactorsGoals(activeProject.id); 
      };
      getData();  
    }
    const getData1 = async () => { 
      await fetchEventNames(activeProject.id);
      await fetchFactorsModels(activeProject.id); 
    };
    getData1();
    setdataSource(null);
    if (goals) {
      const formattedArray = [];
      goals.map((goal, index) => {
        let createdUser = '';
        agents.map((agent) => {
          if (agent.uuid === goal.created_by) {
            createdUser = `${agent.first_name} ${agent.last_name}`;
          }
        });
        formattedArray.push({
          key: index,
          title: goal.name,
          author: createdUser
        });
        setdataSource(formattedArray);
      });
      SetLoadingTable(false);
    }
  }, [activeProject, goals, agents]);
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
                                <Text type={'title'} level={7} extraClass={'m-0'} >Explain periodically tracks a pre-configured set of data points for faster and efficient retrieval of insights. </Text>
                                <Button className={'m-0 mt-4'} size={'large'} onClick={() => setConfigureDPModal(true)}>Configure Data Points</Button>
                            </Col>
                            <Col span={24}>
                                <Text type={'title'} level={7} weight={'bold'} extraClass={'mt-8'} >Suggestions based on your activity</Text>
                                {suggestionList.map((item, index) => {
                                  return (
                                    <div key={index} className={'flex justify-between items-center mt-2'}>
                                        <Text type={'title'} level={7} weight={'thin'} extraClass={'m-0'} ><SlackOutlined className={'mr-1'} />{item.name}</Text>
                                        <Button size={'small'} type="text"><SVG name="plus" color={'grey'} /></Button>
                                    </div>
                                  );
                                })}
                            </Col>
                        </Row>
                    </Col>
                <Col span={16}>
                    <Row gutter={[24, 24]} justify="center">
                        <Col span={20}>
                            <Text type={'title'} level={3} weight={'bold'} extraClass={'m-0'} >Explain</Text>
                            <Text type={'title'} level={5} extraClass={'m-0 mt-2'} >Periodically tracks website events, pages, user properties that are important to you and get insights that influence your goals.</Text>
                        </Col>
                        <Col span={20}>
                            <Button size={'large'} type={'primary'} onClick={() => setGoalDrawer(true)}>Create a New Goal</Button>
                            <a className={'ml-4'}>learn more</a>
                        </Col>
                    </Row>
                    <Row gutter={[24, 24]} justify="center">
                        <Col span={20}>
                        <Table loading={loadingTable} className="ant-table--custom mt-8" columns={columns} dataSource={dataSource} pagination={false} />
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
const mapStateToProps = (state) => {
  return {
    activeProject: state.global.active_project,
    goals: state.factors.goals,
    agents: state.agent.agents
  };
};
export default connect(mapStateToProps, { fetchFactorsGoals, fetchProjectAgents, fetchEventNames, fetchFactorsModels })(Factors);
