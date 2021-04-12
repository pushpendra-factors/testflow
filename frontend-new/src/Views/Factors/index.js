import React, { useState, useEffect } from 'react';
import Header from '../AppLayout/Header';
import SearchBar from '../../components/SearchBar';
import {
  Row, Col, Button, Spin
} from 'antd'; 
import ConfigureDP from './ConfigureDP';
import CreateGoalDrawer from './CreateGoalDrawer';
import { fetchFactorsGoals, fetchFactorsModels, fetchGoalInsights, saveGoalInsightRules, fetchFactorsTrackedEvents, fetchFactorsTrackedUserProperties } from 'Reducers/factors';
import { fetchEventNames, getUserProperties } from 'Reducers/coreQuery/middleware';
import { connect } from 'react-redux';
import { fetchProjectAgents } from 'Reducers/agentActions';  
import _, { isEmpty } from 'lodash'; 
import { useHistory } from 'react-router-dom';
import SavedGoals from './SavedGoals'; 
import { Text, FaErrorComp, FaErrorLog } from 'factorsComponents';
import {ErrorBoundary} from 'react-error-boundary';

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
  , fetchGoalInsights
  , fetchFactorsTrackedEvents
  , fetchFactorsTrackedUserProperties 
  , getUserProperties
}) => {
  const [loadingTable, SetLoadingTable] = useState(true);
  const [fetchingIngishts, SetfetchingIngishts] = useState(false);
  const [showConfigureDPModal, setConfigureDPModal] = useState(false);
  const [showGoalDrawer, setGoalDrawer] = useState(false);
  const [dataSource, setdataSource] = useState(null);
  const history = useHistory();
 

  useEffect(() => {  
    const getData1 = async () => { 
      await fetchProjectAgents(activeProject.id);
      await fetchFactorsGoals(activeProject.id); 
      await fetchEventNames(activeProject.id);
      await fetchFactorsModels(activeProject.id); 
      await fetchFactorsTrackedEvents(activeProject.id);
      await fetchFactorsTrackedUserProperties(activeProject.id); 
      await getUserProperties(activeProject.id, 'events'); 
    };
    getData1(); 
  }, [activeProject]);

  const handleCancel = () => {
    setConfigureDPModal(false);
  };
  
  return (
    <>
    <ErrorBoundary fallback={<FaErrorComp size={'medium'} title={'Explain Error '} subtitle={'We are facing trouble loading Explain. Drop us a message on the in-app chat.'} />} onError={FaErrorLog}>

    {fetchingIngishts ? <Spin size={'large'} className={'fa-page-loader'} /> : 
    <>
        <Header>
              <div className="w-full h-full py-4 flex flex-col justify-center items-center">
                  <SearchBar />
              </div>
          </Header>

          <div className={'fa-container mt-24'}>
              <Row gutter={[24, 24]}>
                  
              <Col span={16}>
                  <Row gutter={[24, 24]} justify="center">
                      <Col span={20}>
                          <Text type={'title'} level={3} weight={'bold'} extraClass={'m-0'} >Explain</Text> 
                          <Text type={'title'} level={6} extraClass={'m-0 mt-2'} color={"grey"} >Periodically tracks website events, pages, user properties that are important to you and get insights that influence your goals.</Text>
                      </Col>
                      <Col span={20}>
                          <Button size={'large'} type={'primary'} onClick={() => setGoalDrawer(true)}>Create a New Goal</Button>
                          <a className={'ml-4'}>learn more</a>
                      </Col>
                  </Row> 
                  <Row gutter={[24, 24]} justify="center">
                      <Col span={20}> 
                      <SavedGoals SetfetchingIngishts={SetfetchingIngishts} />
                      </Col>
                  </Row>
              </Col>
              <Col span={8} className={'border-left--thin-2'}>
                      <Row gutter={[24, 24]} className={'p-4'}>
                          <Col span={24}>
                              <Text type={'title'} level={5} weight={'bold'} extraClass={'m-0'} >What’s being tracked?</Text>
                              <Text type={'title'} level={7} extraClass={'m-0'} >Explain periodically tracks a pre-configured set of data points for faster and efficient retrieval of insights. </Text>
                              <Button className={'m-0 mt-4'} size={'large'} onClick={() => setConfigureDPModal(true)}>Configure Data Points</Button>
                          </Col>

                          {/* <Col span={24}>
                              <Text type={'title'} level={7} weight={'bold'} extraClass={'mt-8'} >Suggestions based on your activity</Text>
                              {suggestionList.map((item, index) => {
                                return (
                                  <div key={index} className={'flex justify-between items-center mt-2'}>
                                      <Text type={'title'} level={7} weight={'thin'} extraClass={'m-0'} ><SlackOutlined className={'mr-1'} />{item.name}</Text>
                                      <Button size={'small'} type="text"><SVG name="plus" color={'grey'} /></Button>
                                  </div>
                                );
                              })}
                          </Col> */}
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
    }
  </ErrorBoundary>
    </>
  );
};
const mapStateToProps = (state) => {
  return {
    activeProject: state.global.active_project,
    goals: state.factors.goals,
    agents: state.agent.agents,
    factors_models: state.factors.factors_models,
  };
};
export default connect(mapStateToProps, { fetchFactorsGoals, fetchFactorsTrackedEvents, fetchFactorsTrackedUserProperties, fetchProjectAgents, saveGoalInsightRules, fetchGoalInsights, fetchFactorsModels , fetchEventNames, getUserProperties})(Factors);
