import React, { useState } from 'react';
import {
  Row, Col, Progress, Modal, Button
} from 'antd';
import { Text, SVG } from 'factorsComponents';
import _ from 'lodash';
import MoreInsightsLines from './MoreInsightsLines';

function numberWithCommas(x) {
  return x.toString().replace(/\B(?=(\d{3})+(?!\d))/g, ',');
}

const SubInsightItem = ({ SubInsightsData, showModal, handleClose }) => {
  const [SubLevel2Data, SetSubLevel2Data] = useState(null);
  const [SubLevel1Data, SetSubLevel1Data] = useState(null);
  if (SubInsightsData) {
    console.log('SubInsightsData', SubInsightsData);
    return (
          <Modal
          className={'fa-modal--regular'}
          visible={showModal}
          onOk={handleClose}
          onCancel={() => {
            handleClose();
            SetSubLevel2Data(null);
            SetSubLevel1Data(null);
          }
          }
          width={750}
          footer={null}
          title={null}
        >

        {!SubLevel2Data && SubInsightsData.factors_sub_insights.map((dataItem, index) => {
          let insightKeyLevel1 = '';
          if (_.isEmpty(dataItem.factors_insights_key)) {
            insightKeyLevel1 = `${SubInsightsData.factors_insights_attribute[0].factors_attribute_key} = ${SubInsightsData.factors_insights_attribute[0].factors_attribute_value}`;
          } else {
            insightKeyLevel1 = SubInsightsData.factors_insights_key;
          }

          let insightKeyLevel2 = '';
          if (_.isEmpty(dataItem.factors_insights_key)) {
            insightKeyLevel2 = `${dataItem.factors_insights_attribute[0].factors_attribute_key} = ${dataItem.factors_insights_attribute[0].factors_attribute_value}`;
          } else {
            insightKeyLevel2 = dataItem.factors_insights_key;
          }

          const insightLevel2Percentage = (dataItem.factors_insights_users_count / SubInsightsData.factors_insights_users_count) * 100;
          const insightLevel3Percentage = (dataItem.factors_goal_users_count / SubInsightsData.factors_insights_users_count) * 100;

          return (
              <Row key={index} gutter={[0, 0]} justify={'center'}>
              <Col span={22}>
                <div className={'relative border-bottom--thin-2 fa-insight-item--sub-container'}>
                      <Row gutter={[0, 0]} justify={'center'}>
                          <Col span={24}>
                              <div className={'relative border-left--thin-2 m-0 pl-10 py-6 cursor-pointer fa-insight-item'} onClick={() => {
                                SetSubLevel2Data(dataItem);
                                SetSubLevel1Data(SubInsightsData);
                              }}>
                                  <Text type={'title'} level={4} extraClass={'m-0'} >{dataItem.factors_insights_text}</Text>
                                  <Text type={'title'} level={3} weight={'bold'} extraClass={'m-0'} >{`${dataItem.factors_insights_multiplier}x`}</Text>
                                  {!_.isEmpty(dataItem.factors_higher_completion_text) && <Text type={'title'} level={6} color={'grey'} extraClass={'mt-2'} >{dataItem.factors_higher_completion_text}</Text>}
                                  {!_.isEmpty(dataItem.factors_lower_completion_text) && <Text type={'title'} level={6} color={'grey'} extraClass={'mt-2'} >{dataItem.factors_lower_completion_text}</Text>}

                                <div className={'mt-8 w-9/12'}>

                                    <div className={'flex items-end'}>
                                      <div className={'flex items-center ml-4 fa-insights-box--fixed-count'}><a><Text type={'title'} weight={'thin'} level={7} extraClass={'m-0'} >{numberWithCommas(SubInsightsData.factors_insights_users_count)}</Text></a> </div>
                                      <div className={'flex items-center ml-4 fa-insights-box--animate'}>  <SVG name={'arrowdown'} size={12} color={'grey'} /> <Text type={'title'} weight={'thin'} level={7} extraClass={'m-0 ml-1'} >{insightKeyLevel1}</Text></div>
                                    </div>
                                    <Progress percent={100} strokeColor={'#5949BC'} className={'fa-custom-stroke-bg'} showInfo={false} />

                                    <div className={'flex items-end'}>
                                      <div className={'flex items-center ml-4 fa-insights-box--fixed-count'}><a><Text type={'title'} weight={'thin'} level={7} extraClass={'m-0'} >{numberWithCommas(dataItem.factors_insights_users_count)}</Text></a> </div>
                                      <div className={'flex items-center ml-4 fa-insights-box--animate'}>  <SVG name={'arrowdown'} size={12} color={'grey'} /> <Text type={'title'} weight={'thin'} level={7} extraClass={'m-0 ml-1'} >{insightKeyLevel2}</Text></div>
                                    </div>
                                    <Progress percent={insightLevel2Percentage} strokeColor={'#5949BC'} className={'fa-custom-stroke-bg'} showInfo={false} />

                                    <div className={'flex items-end'}>
                                      <div className={'flex items-center ml-4 fa-insights-box--fixed-count'}> <Text type={'title'} weight={'thin'} level={7} extraClass={'m-0 mt-2'} >{`${numberWithCommas(dataItem.factors_goal_users_count)}`}</Text><span><Text type={'title'} weight={'thin'} level={7} extraClass={'m-0 mt-2 ml-1'} >{`(${dataItem.factors_insights_percentage}% goal completion)`}</Text></span></div>
                                      <div className={'flex items-center ml-4 fa-insights-box--animate'}><SVG name={'arrowdown'} size={12} color={'grey'} /><Text type={'title'} weight={'thin'} level={7} extraClass={'m-0 ml-1'} >{`(${dataItem.factors_insights_percentage}% goal completion)`}</Text></div>
                                    </div>
                                    <Progress percent={insightLevel3Percentage} strokeColor={'#F9C06E'} showInfo={false} />

                                </div>

                                  <div className={'fa-sub-insights-box--spike'}>
                                      <div className={'flex justify-end items-center'}>
                                          {dataItem.factors_multiplier_increase_flag ? <SVG name={'spikeup'} size={42} /> : <SVG name={'spikedown'} size={42} />}
                                      </div>
                                  </div>
                              </div>
                          </Col>
                      </Row>
                      {dataItem?.factors_sub_insights && <MoreInsightsLines onClick={() => SetSubLevel2Data(dataItem.factors_sub_insights)} insightCount={dataItem?.factors_sub_insights.length} /> }
                  </div>
                  </Col>
              </Row>
          );
        })}
        {SubLevel2Data &&
        <>
            <Row gutter={[0, 0]} justify={'center'}>
              <Col span={24}>
                  <div className={'w-full p-4 background-color--brand-color-1 '}>
                      <Button className={'fa-button-ghost'} type={'text'} onClick={() => { SetSubLevel2Data(false); }}>Back</Button>
                  </div>
              </Col>
          </Row>
            {SubLevel2Data.factors_sub_insights.map((dataItem, index) => {
              let insightKeyLevel1 = '';
              if (_.isEmpty(dataItem.factors_insights_key)) {
                insightKeyLevel1 = `${SubLevel1Data.factors_insights_attribute[0].factors_attribute_key} = ${SubLevel1Data.factors_insights_attribute[0].factors_attribute_value}`;
              } else {
                insightKeyLevel1 = SubLevel1Data.factors_insights_key;
              }

              let insightKeyLevel2 = '';
              if (_.isEmpty(dataItem.factors_insights_key)) {
                insightKeyLevel2 = `${SubLevel2Data.factors_insights_attribute[0].factors_attribute_key} = ${SubLevel2Data.factors_insights_attribute[0].factors_attribute_value}`;
              } else {
                insightKeyLevel2 = SubLevel2Data.factors_insights_key;
              }

              let insightKeyLevel3 = '';
              if (_.isEmpty(dataItem.factors_insights_key)) {
                insightKeyLevel3 = `${dataItem.factors_insights_attribute[0].factors_attribute_key} = ${dataItem.factors_insights_attribute[0].factors_attribute_value}`;
              } else {
                insightKeyLevel3 = dataItem.factors_insights_key;
              }

              const insightLevel2Percentage = (SubLevel2Data.factors_insights_users_count / SubLevel1Data.factors_insights_users_count) * 100;
              const insightLevel3Percentage = (dataItem.factors_insights_users_count / SubLevel1Data.factors_insights_users_count) * 100;
              const insightLevel4Percentage = (dataItem.factors_goal_users_count / SubLevel1Data.factors_insights_users_count) * 100;

              return (
                  <Row key={index} gutter={[0, 0]} justify={'center'}>
                  <Col span={22}>
                    <div className={'relative border-bottom--thin-2 fa-insight-item--sub-container'}>
                          <Row gutter={[0, 0]} justify={'center'}>
                              <Col span={24}>
                                  <div className={'relative border-left--thin-2 m-0 pl-10 py-6 fa-insight-item'}>
                                      <Text type={'title'} level={4} extraClass={'m-0'} >{dataItem.factors_insights_text}</Text>
                                      <Text type={'title'} level={3} weight={'bold'} extraClass={'m-0'} >{`${dataItem.factors_insights_multiplier}x`}</Text>

                                      <div className={'mt-8 w-9/12'}>

                                    <div className={'flex items-end'}>
                                      <div className={'flex items-center ml-4 fa-insights-box--fixed-count'}><a><Text type={'title'} weight={'thin'} level={7} extraClass={'m-0'} >{numberWithCommas(SubLevel1Data.factors_insights_users_count)}</Text></a> </div>
                                      <div className={'flex items-center ml-4 fa-insights-box--animate'}>  <SVG name={'arrowdown'} size={12} color={'grey'} /> <Text type={'title'} weight={'thin'} level={7} extraClass={'m-0 ml-1'} >{insightKeyLevel1}</Text></div>
                                    </div>
                                    <Progress percent={100} strokeColor={'#5949BC'} className={'fa-custom-stroke-bg'} showInfo={false} />

                                    <div className={'flex items-end'}>
                                      <div className={'flex items-center ml-4 fa-insights-box--fixed-count'}><a><Text type={'title'} weight={'thin'} level={7} extraClass={'m-0'} >{numberWithCommas(SubLevel2Data.factors_insights_users_count)}</Text></a> </div>
                                      <div className={'flex items-center ml-4 fa-insights-box--animate'}>  <SVG name={'arrowdown'} size={12} color={'grey'} /> <Text type={'title'} weight={'thin'} level={7} extraClass={'m-0 ml-1'} >{insightKeyLevel2}</Text></div>
                                    </div>
                                    <Progress percent={insightLevel2Percentage} strokeColor={'#5949BC'} className={'fa-custom-stroke-bg'} showInfo={false} />

                                    <div className={'flex items-end'}>
                                      <div className={'flex items-center ml-4 fa-insights-box--fixed-count'}><a><Text type={'title'} weight={'thin'} level={7} extraClass={'m-0'} >{numberWithCommas(dataItem.factors_insights_users_count)}</Text></a> </div>
                                      <div className={'flex items-center ml-4 fa-insights-box--animate'}>  <SVG name={'arrowdown'} size={12} color={'grey'} /> <Text type={'title'} weight={'thin'} level={7} extraClass={'m-0 ml-1'} >{insightKeyLevel3}</Text></div>
                                    </div>
                                    <Progress percent={insightLevel3Percentage} strokeColor={'#5949BC'} className={'fa-custom-stroke-bg'} showInfo={false} />

                                    <div className={'flex items-end'}>
                                      <div className={'flex items-center ml-4 fa-insights-box--fixed-count'}> <Text type={'title'} weight={'thin'} level={7} extraClass={'m-0 mt-2'} >{`${numberWithCommas(dataItem.factors_goal_users_count)}`}</Text><span><Text type={'title'} weight={'thin'} level={7} extraClass={'m-0 mt-2 ml-1'} >{`(${dataItem.factors_insights_percentage}% goal completion)`}</Text></span></div>
                                      <div className={'flex items-center ml-4 fa-insights-box--animate'}><SVG name={'arrowdown'} size={12} color={'grey'} /><Text type={'title'} weight={'thin'} level={7} extraClass={'m-0 ml-1'} >{`(${dataItem.factors_insights_percentage}% goal completion)`}</Text></div>
                                    </div>
                                    <Progress percent={insightLevel4Percentage} strokeColor={'#F9C06E'} showInfo={false} />

                                    </div>

                                      <div className={'fa-sub-insights-box--spike'}>
                                          <div className={'flex justify-end items-center'}>
                                              {dataItem.factors_multiplier_increase_flag ? <SVG name={'spikeup'} size={42} /> : <SVG name={'spikedown'} size={42} />}
                                          </div>
                                      </div>
                                  </div>
                              </Col>
                          </Row>
                          {dataItem?.factors_sub_insights && <MoreInsightsLines insightCount={dataItem?.factors_sub_insights.length} /> }
                        </div>
                      </Col>
                  </Row>
              );
            })}
        </>
        }

        </Modal>

    );
  } else return null;
};

export default SubInsightItem;
