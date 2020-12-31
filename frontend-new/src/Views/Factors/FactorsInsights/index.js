import React, { useEffect, useState } from 'react';
import {
  Tabs, Row, Col, Spin
} from 'antd';
import { Text } from 'factorsComponents'; 
import { connect } from 'react-redux';
import _ from 'lodash';
import InsightHighlightItem from './InsightHighlightItem';
import SubInsightItem from './SubInsightItem';
import InsightItem from './InsightItem';
import HeaderContents from './HeaderContents';
import SubHeaderContents from './SubHeaderContents';

const { TabPane } = Tabs;

const FactorsInsights = ({ activeProject, goal_insights }) => {
  const [showModal, SetShowModal] = useState(false);
  const [SubInsightsData, setSubInsightsData] = useState(null);
  const [ParentData, setParentData] = useState(null);

  const handleClose = () => {
    SetShowModal(false);
  };
  const showSubInsightsData = (data,parentData=null) => {
    setParentData(parentData);
    setSubInsightsData(data);
    SetShowModal(true);
  };

  return (
    <>
           {!goal_insights ? <Spin size={'large'} className={'fa-page-loader'} /> :  <> 
            
            <HeaderContents />

           <div className={'fa-container mt-24'}>
                <Row gutter={[24, 24]}>
                    <Col span={24}>
                        <SubHeaderContents />
                     </Col>
                </Row>
                <Row gutter={[24, 24]}>
                    <Col span={24}>
                        <InsightHighlightItem data={goal_insights} />
                     </Col>
                </Row>
                <Row gutter={[24, 24]}>
                    <Col span={24}>

                    <div className={'fa-insights--tab'}>
                    <Tabs defaultActiveKey="1" >
                        <TabPane tab="All Insights" key="1">
                                    <InsightItem data={goal_insights} showSubInsightsData={showSubInsightsData} displayType={true} category={'journey'} />
                                    <InsightItem data={goal_insights} showSubInsightsData={showSubInsightsData} displayType={true} category={'attribute'} />
                                    <InsightItem data={goal_insights} showSubInsightsData={showSubInsightsData} displayType={true} category={'campaign'} />
                        </TabPane>
                        <TabPane tab="Attributes" key="2"> 
                            <InsightItem data={goal_insights} showSubInsightsData={showSubInsightsData} category={'attribute'} />
                        </TabPane>
                        <TabPane tab="Campaigns" key="3">
                            <div className={'w-full p-4 background-color--brand-color-1'}>
                                    <Text type={'title'} level={7} weight={'regular'} align={'center'} extraClass={'m-0'} >Show insights with reference to <a>Campaigns</a></Text>
                            </div>
                                    <InsightItem data={goal_insights} showSubInsightsData={showSubInsightsData} category={'campaign'} />
                        </TabPane>
                        <TabPane tab="Journeys" key="4"> 
                                  <InsightItem data={goal_insights} showSubInsightsData={showSubInsightsData} category={'journey'} />
                        </TabPane>
                    </Tabs>
                    </div>

                    </Col>
                </Row>

                <SubInsightItem showModal={showModal} ParentData={ParentData} SubInsightsData={SubInsightsData} handleClose={handleClose} />

            </div>
            </>
           }

    </>
  );
};

const mapStateToProps = (state) => {
  return {
    activeProject: state.global.active_project,
    goal_insights: state.factors.goal_insights
  };
};
export default connect(mapStateToProps)(FactorsInsights);
