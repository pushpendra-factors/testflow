import React, { useEffect, useState } from 'react';
import {
  Tabs, Row, Col, Progress, Modal, Button
} from 'antd';
import { Text, SVG } from 'factorsComponents';
import { fetchGoalInsights } from 'Reducers/factors';
import { connect } from 'react-redux';

const { TabPane } = Tabs;

const InsightHighlightItem = ({ data }) => {
  if (data) {
    return (
            <div className={'relative my-4'}>
                <Row gutter={[0, 0]} justify={'center'}>
                    <Col span={16}>
                        <div className={'relative border-left--thin-2 m-0 pl-16 py-2'}>
                            <div className={'w-full'}>
                            <Text type={'title'} weight={'thin'} level={7} extraClass={'m-0'} >{data.goal?.st_en}</Text>
                            <Progress percent={100} strokeColor={'#5949BC'} showInfo={false} />

                            <Text type={'title'} level={2} weight={'bold'} extraClass={'m-0'} >{data.overall_percentage_text}</Text>

                            <Progress percent={data.overall_percentage} strokeColor={'#F9C06E'} showInfo={false} />
                            <Text type={'title'} weight={'thin'} level={7} extraClass={'m-0'} >{data.goal?.en_en}</Text>
                            </div>

                            <div className={'fa-insights-box--highlight'}>
                                <div className={'flex justify-between items-end flex-col h-full'}>
                                    <Text type={'title'} level={5} color={'blue'} weight={'bold'} extraClass={'m-0'} >{data.total_users_count}</Text>
                                    <div className={'flex flex-col items-center justify-center '}>
                                        <Text type={'title'} level={4} color={'grey'} weight={'bold'} extraClass={'m-0'} >1x</Text>
                                        <Text type={'title'} level={7} color={'grey'} extraClass={'m-0'} >Impact</Text>
                                    </div>
                                    <Text type={'title'} level={5} color={'yellow'} weight={'bold'} extraClass={'m-0'} >{data.goal_user_count}</Text>
                                </div>
                            </div>
                        </div>
                    </Col>
                </Row>
            </div>

    );
  } else return null;
};
const SubInsightItem = ({ SubInsightsData, showModal, handleClose }) => {
  if (SubInsightsData) {
    return (
        <Modal
        className={'fa-modal--regular'}
        visible={showModal}
        onOk={handleClose}
        onCancel={handleClose}
        width={750}
        footer={null}
        title={null}
      >
      {SubInsightsData.map((dataItem, index) => {
        return (
            <Row key={index} gutter={[0, 0]} justify={'center'}>
            <Col span={22}>
              <div className={'relative border-bottom--thin'}>
                    <Row gutter={[0, 0]} justify={'center'}>
                        <Col span={24}>
                            <div className={'relative border-left--thin-2 m-0 pl-10 py-6'}>
                                <Text type={'title'} level={4} extraClass={'m-0'} >{dataItem.factors_insights_text}</Text>
                                <Text type={'title'} level={3} weight={'bold'} extraClass={'m-0'} >{`${dataItem.factors_insights_multiplier}x`}</Text>

                                <div className={'mt-8 w-9/12'}>
                                <Text type={'title'} weight={'thin'} level={7} extraClass={'m-0'} >{dataItem.factors_insights_users_count}</Text>
                                <Progress percent={100} strokeColor={'#5949BC'} showInfo={false} />

                                <Text type={'title'} weight={'thin'} level={7} extraClass={'m-0 mt-2'} >{`${dataItem.factors_goal_users_count} (${dataItem.factors_insights_percentage}% goal completion)`}</Text>
                                <Progress percent={dataItem.factors_insights_percentage} strokeColor={'#F9C06E'} showInfo={false} />
                                </div>

                                <div className={'fa-sub-insights-box--spike'}>
                                    <div className={'flex justify-end items-center'}>
                                        {dataItem.factors_multiplier_increase_flag ? <SVG name={'spikeup'} size={42} /> : <SVG name={'spikedown'} size={42} />}
                                    </div>
                                </div>
                            </div>
                        </Col>
                    </Row>
                  </div>
                </Col>
            </Row>

        );
      })}

      </Modal>

    );
  } else return null;
};

const InsightItem = ({
  data, type, showSubInsightsData, displayType = false
}) => {
  if (data) {
    return data.insights.map((dataItem, index) => {
      if (dataItem.factors_insights_type === type) {
        return (
                  <div key={index} className={'relative border-bottom--thin'}>
                      <Row gutter={[0, 0]} justify={'center'}>
                          <Col span={16}>
                              <div className={'relative border-left--thin-2 m-0 pl-16 py-8 cursor-pointer fa-insight-item'} onClick={() => { showSubInsightsData(dataItem.factors_sub_insights); }}>
                                 {displayType && <Text type={'paragraph'} mini color={'grey'} weight={'bold'} extraClass={'uppercase fa-insights-box--type'} >{type}</Text>}
                                  <Text type={'title'} level={4} weight={'bold'} extraClass={'m-0'} >{dataItem.factors_insights_text}</Text>
                                  <Text type={'title'} level={6} color={'grey'} extraClass={'mt-4'} >{'Higher completions for time spent on page <= 1min +3 factors'} </Text>
                                  <Text type={'title'} level={6} color={'grey'} extraClass={'mt-2'} >{'Lower completions for Time-Spent <= 10sec +2 factors'} </Text>

                                  <div className={'mt-8 w-9/12'}>
                                  <div className={'flex items-center'}>
                                    <Text type={'title'} weight={'thin'} level={7} extraClass={'m-0'} >{dataItem.factors_insights_users_count}</Text>
                                    <div className={'flex items-center ml-4 fa-insights-box--animate'}>  <SVG name={'corequery'} size={12} color={'grey'} /> <Text type={'title'} weight={'thin'} level={7} extraClass={'m-0 ml-1'} >{dataItem.factors_insights_key}</Text></div>
                                  </div>
                                  <Progress percent={100} strokeColor={'#5949BC'} showInfo={false} />

                                  <Text type={'title'} weight={'thin'} level={7} extraClass={'m-0 mt-2'} >{`${dataItem.factors_goal_users_count} (${dataItem.factors_insights_percentage}% goal completion)`}</Text>
                                  <Progress percent={dataItem.factors_insights_percentage} strokeColor={'#F9C06E'} showInfo={false} />
                                  </div>

                                  {dataItem?.factors_sub_insights?.length > 0 && <div className={'fa-insights-box--actions'}>
                                    <Button size={'large'}>
                                        <SVG name={'corequery'} size={24} color={'grey'} />
                                    </Button>
                                  </div>
                                  }

                                  <div className={'fa-insights-box--spike'}>
                                      <div className={'flex justify-end items-center'}>
                                        <Text type={'title'} level={5} color={'grey'} weight={'bold'} extraClass={'m-0 mr-2'} >{`${dataItem.factors_insights_multiplier}x`}</Text>
                                        {dataItem.factors_multiplier_increase_flag ? <SVG name={'spikeup'} size={42} /> : <SVG name={'spikedown'} size={42} />}
                                      </div>
                                  </div>

                              </div>
                          </Col>
                      </Row>
                  </div>
        );
      }
    });
  } else {
    return null;
  }
};

const FactorsInsights = ({ fetchGoalInsights, activeProject, goal_insights }) => {
  const [showModal, SetShowModal] = useState(false);
  const [SubInsightsData, setSubInsightsData] = useState(null);

  useEffect(() => {
    if (!goal_insights) {
      fetchGoalInsights(activeProject.id);
    }
  }, [goal_insights]);

  const handleClose = () => {
    SetShowModal(false);
  };
  const showSubInsightsData = (data) => {
    setSubInsightsData(data);
    SetShowModal(true);
  };

  return (
    <>
           <div className={'fa-container mt-24'}>
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
                                    <InsightItem data={goal_insights} showSubInsightsData={showSubInsightsData} displayType={true} type={'journey'} />
                                    <InsightItem data={goal_insights} showSubInsightsData={showSubInsightsData} displayType={true} type={'attribute'} />
                                    <InsightItem data={goal_insights} showSubInsightsData={showSubInsightsData} displayType={true} type={'campaign'} />
                        </TabPane>
                        <TabPane tab="Attributes" key="2">
                            <div className={'w-full p-4 background-color--brand-color-1'}>
                                    <Text type={'title'} level={7} weight={'regular'} align={'center'} extraClass={'m-0'} >Show insights with reference to <a>Attributes</a></Text>
                            </div>
                            <InsightItem data={goal_insights} showSubInsightsData={showSubInsightsData} type={'attribute'} />
                        </TabPane>
                        <TabPane tab="Campaigns" key="3">
                            <div className={'w-full p-4 background-color--brand-color-1'}>
                                    <Text type={'title'} level={7} weight={'regular'} align={'center'} extraClass={'m-0'} >Show insights with reference to <a>Campaigns</a></Text>
                            </div>
                                    <InsightItem data={goal_insights} showSubInsightsData={showSubInsightsData} type={'campaign'} />
                        </TabPane>
                        <TabPane tab="Journeys" key="4">
                             <div className={'w-full p-4 background-color--brand-color-1'}>
                                    <Text type={'title'} level={7} weight={'regular'} align={'center'} extraClass={'m-0'} >Show insights with reference to <a>Journeys</a></Text>
                             </div>
                                <InsightItem data={goal_insights} showSubInsightsData={showSubInsightsData} type={'journey'} />
                        </TabPane>
                    </Tabs>
                    </div>

                    </Col>
                </Row>

                <SubInsightItem showModal={showModal} SubInsightsData={SubInsightsData} handleClose={handleClose} />

            </div>

    </>
  );
};

const mapStateToProps = (state) => {
  return {
    activeProject: state.global.active_project,
    goal_insights: state.factors.goal_insights
  };
};
export default connect(mapStateToProps, { fetchGoalInsights })(FactorsInsights);
