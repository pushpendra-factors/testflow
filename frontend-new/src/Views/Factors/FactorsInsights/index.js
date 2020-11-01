import React from 'react';
import {
  Tabs, Row, Col, Progress
} from 'antd';
import { Text, SVG } from 'factorsComponents';

const { TabPane } = Tabs;

const InsightHighlightItem = () => {
  return (
        <div className={'relative my-4'}>
            <Row gutter={[0, 0]} justify={'center'}>
                <Col span={16}>
                    <div className={'relative border-left--thin-2 m-0 pl-16 py-2'}>
                        <div className={'w-full'}>
                        <Text type={'title'} weight={'thin'} level={7} extraClass={'m-0'} >All visitors</Text>
                        <Progress percent={100} strokeColor={'#5949BC'} showInfo={false} />

                        <Text type={'title'} level={2} weight={'bold'} extraClass={'m-0'} >8.2% of all visitors have completed this goal.</Text>

                        <Progress percent={20} strokeColor={'#F9C06E'} showInfo={false} />
                        <Text type={'title'} weight={'thin'} level={7} extraClass={'m-0'} >Deal won</Text>
                        </div>

                        <div className={'fa-insights-box--highlight'}>
                            <div className={'flex justify-between items-end flex-col h-full'}>
                                <Text type={'title'} level={5} color={'blue'} weight={'bold'} extraClass={'m-0'} >1,22,340</Text>
                                <div className={'flex flex-col items-center justify-center '}>
                                    <Text type={'title'} level={4} color={'grey'} weight={'bold'} extraClass={'m-0'} >1x</Text>
                                    <Text type={'title'} level={6} color={'grey'} weight={'bold'} extraClass={'m-0'} >Impact</Text>
                                </div>
                                <Text type={'title'} level={5} color={'yellow'} weight={'bold'} extraClass={'m-0'} >2,340</Text>
                            </div>
                        </div>
                    </div>
                </Col>
            </Row>
        </div>
  );
};
const InsightItem = () => {
  return (
        <div className={'relative border-bottom--thin'}>
            <Row gutter={[0, 0]} justify={'center'}>
                <Col span={16}>
                    <div className={'relative border-left--thin-2 m-0 pl-16 py-8'}>
                        <Text type={'title'} level={4} weight={'bold'} extraClass={'m-0'} >of which visitors coming via <a>.../product</a> shows 2x higher goal completion.</Text>
                        <Text type={'title'} level={6} color={'grey'} extraClass={'mt-4'} >{'Higher completions for time spent on page <= 1min +3 factors'} </Text>
                        <Text type={'title'} level={6} color={'grey'} extraClass={'mt-2'} >{'Lower completions for Time-Spent <= 10sec +2 factors'} </Text>

                        <div className={'mt-8 w-9/12'}>
                        <Text type={'title'} weight={'thin'} level={7} extraClass={'m-0'} >1,234</Text>
                        <Progress percent={100} strokeColor={'#5949BC'} showInfo={false} />

                        <Text type={'title'} weight={'thin'} level={7} extraClass={'m-0 mt-2'} >246 (20% goal completion)</Text>
                        <Progress percent={20} strokeColor={'#F9C06E'} showInfo={false} />
                        </div>

                        <div className={'fa-insights-box--spike'}>
                            <div className={'flex justify-end items-center'}>
                                <Text type={'title'} level={5} color={'grey'} weight={'bold'} extraClass={'m-0 mr-2'} >3x</Text>
                                <SVG name={'spikeup'} size={42} />
                            </div>
                        </div>
                    </div>
                </Col>
            </Row>
        </div>
  );
};

const FactorsInsights = () => {
  return (
    <>
           <div className={'fa-container mt-24'}>
                <Row gutter={[24, 24]}>
                    <Col span={24}>
                        <InsightHighlightItem />
                     </Col>
                </Row>
                <Row gutter={[24, 24]}>
                    <Col span={24}>

                    <div className={'fa-insights--tab'}>
                    <Tabs defaultActiveKey="1" >
                        <TabPane tab="All Insights" key="1">

                                    <InsightItem />
                                    <InsightItem />
                                    <InsightItem />
                                    <InsightItem />

                        </TabPane>
                        <TabPane tab="Attributes" key="2">
                                    <InsightItem />
                                    <InsightItem />
                        </TabPane>
                        <TabPane tab="Campaigns" key="3">
                                <InsightItem />
                        </TabPane>
                        <TabPane tab="Journeys" key="3">
                                <InsightItem />
                                    <InsightItem />
                                    <InsightItem />
                        </TabPane>
                    </Tabs>
                    </div>

                    </Col>
                </Row>
            </div>

    </>
  );
};

export default FactorsInsights;
