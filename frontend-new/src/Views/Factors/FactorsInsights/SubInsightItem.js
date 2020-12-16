/* eslint-disable */
import React, { useState } from 'react';
import {
  Row, Col, Progress, Modal, Button
} from 'antd';
import { Text, SVG } from 'factorsComponents';
import _ from 'lodash';
import MoreInsightsLines from './MoreInsightsLines';
import { numberWithCommas } from 'Utils/dataFormatter';


const ModalHeader = (SubInsightsData,handleClose) => { 
    let insightKey = '';
          if (_.isEmpty(SubInsightsData.factors_insights_key)) {
            insightKey = `${SubInsightsData.factors_insights_attribute[0].factors_attribute_key} = ${SubInsightsData.factors_insights_attribute[0].factors_attribute_value}`;
          } else {
            insightKey = SubInsightsData.factors_insights_key;
          }

    return (
      <div className={'flex justify-between items-center px-4 py-3'}>
        <div className={'flex flex-col'}> 
         <Text type={'title'} level={7} weight={'bold'} color={'grey'} extraClass={'ml-2 m-0 capitalize'}>{SubInsightsData.factors_insights_type}</Text>
         <Text type={'title'} level={4} weight={'bold'} extraClass={'ml-2 m-0'}>{insightKey}</Text>
        </div>
        <div className={'flex justify-end items-center'}>
          <Text type={'title'} level={3} weight={'bold'} extraClass={'m-0'} >{`${SubInsightsData.factors_insights_multiplier}x`}</Text>
          {SubInsightsData.factors_multiplier_increase_flag ? <SVG name={'spikeup'} size={24} /> : <SVG name={'spikedown'} size={24} />}
          <Button size={'large'} type="text" className={'ml-2'} onClick={handleClose}><SVG name="times"></SVG></Button>
        </div>
      </div>
    );
  };

const SubInsightItem = ({ SubInsightsData, showModal, handleClose, ParentData=null }) => {
  
  const [SubLevel2Data, SetSubLevel2Data] = useState(null);
  const [SubLevel1Data, SetSubLevel1Data] = useState(null);
  if (SubInsightsData) {  
    const isJourney = ParentData?.type === 'journey' ? true : false; 
    return (
          <Modal
          className={'fa-modal--regular fa-modal-body--no-padding fa-modal-header--no-padding'}
          visible={showModal}
          onOk={handleClose}
          title={ModalHeader(SubInsightsData,handleClose)}
          closable={false}
          style={{ top: 30 }}
          onCancel={() => {
            handleClose();
            SetSubLevel2Data(null);
            SetSubLevel1Data(null);
          }
          }
          width={900}
          footer={null} 
        >
        
        {!SubLevel2Data && <div className={'fa-modal-body--custom-scrollable'}>
         { SubInsightsData.factors_sub_insights.map((dataItem, index) => {

          let insightKeyLevel1 = '';
          if (_.isEmpty(SubInsightsData.factors_insights_key)) {
            insightKeyLevel1 = `${SubInsightsData.factors_insights_attribute[0].factors_attribute_key} = ${SubInsightsData.factors_insights_attribute[0].factors_attribute_value}`;
          } else {
            insightKeyLevel1 = SubInsightsData.factors_insights_key;
          }
 
        const factors_insights_text = `and then <a>${insightKeyLevel1}</a> show  ${dataItem.factors_insights_multiplier}x goal completion`

          let insightKeyLevel2 = '';
          if (_.isEmpty(dataItem.factors_insights_key)) {
            insightKeyLevel2 = `${dataItem.factors_insights_attribute[0].factors_attribute_key} = ${dataItem.factors_insights_attribute[0].factors_attribute_value}`;
          } else {
            insightKeyLevel2 = dataItem.factors_insights_key;
          }
 

          
          let insightLevel1Percentage = 100; 
          let insightLevel2Percentage = 100;
          let insightLevel3Percentage = 100;

          if(isJourney){ 
            insightLevel1Percentage = (SubInsightsData.factors_insights_users_count / ParentData.total_users_count) * 100;
            insightLevel2Percentage = (dataItem.factors_insights_users_count / ParentData.total_users_count) * 100;
           insightLevel3Percentage = (dataItem.factors_goal_users_count / ParentData.total_users_count) * 100;
          } 
          else{
           insightLevel2Percentage = (dataItem.factors_insights_users_count / SubInsightsData.factors_insights_users_count) * 100;
           insightLevel3Percentage = (dataItem.factors_goal_users_count / SubInsightsData.factors_insights_users_count) * 100;
          }


          return (
              <Row key={index} gutter={[0, 0]} justify={'center'}>
              <Col span={22}>
                <div className={'relative border-bottom--thin-2 fa-insight-item--sub-container px-4'}>
                      <Row gutter={[0, 0]} justify={'center'}>
                          <Col span={24}>
                              <div className={'relative border-left--thin-2 m-0 pl-10 py-6 cursor-pointer fa-insight-item'} onClick={() => {
                                if(dataItem?.factors_sub_insights && !_.isEmpty(dataItem?.factors_sub_insights)){
                                    SetSubLevel2Data(dataItem);
                                    SetSubLevel1Data(SubInsightsData); 
                                }
                              }}>
                                  <Text type={'title'} level={4} extraClass={'m-0'} > <span dangerouslySetInnerHTML={{__html: factors_insights_text}}/> </Text>
                                  <Text type={'title'} level={3} weight={'bold'} extraClass={'m-0'} >{`${dataItem.factors_insights_multiplier}x`}</Text>
                                  {!_.isEmpty(dataItem.factors_higher_completion_text) && <Text type={'title'} level={6} color={'grey'} extraClass={'mt-2'} >{dataItem.factors_higher_completion_text}</Text>}
                                  {!_.isEmpty(dataItem.factors_lower_completion_text) && <Text type={'title'} level={6} color={'grey'} extraClass={'mt-2'} >{dataItem.factors_lower_completion_text}</Text>}

                                <div className={'mt-8 w-9/12'}>
                                  
                                  {
                                    isJourney && <>
                                      <div className={'flex items-end'}>
                                        <div className={'flex items-center ml-4 fa-insights-box--fixed-count'}><a><Text type={'title'} weight={'regular'} level={7} extraClass={'m-0 tracking-wider'} >{numberWithCommas(ParentData.total_users_count)}</Text></a> </div>
                                        <div className={'flex items-center ml-4 fa-insights-box--animate'}>  <SVG name={'arrowdown'} size={12} color={'grey'} /> <Text type={'title'} weight={'thin'} level={7} extraClass={'m-0 ml-1'} >{_.isEmpty(ParentData.goal?.st_en) ? 'All Visitors' : ParentData.goal?.st_en } </Text></div>
                                      </div>
                                      <Progress percent={100} strokeColor={'#5949BC'} className={'fa-custom-stroke-bg'} showInfo={false} />
                                    </>
                                  } 

                                    <div className={'flex items-end'}>
                                      <div className={'flex items-center ml-4 fa-insights-box--fixed-count'}><a><Text type={'title'} weight={'regular'} level={7} extraClass={'m-0 tracking-wider'} >{numberWithCommas(SubInsightsData.factors_insights_users_count)}</Text></a> </div>
                                      <div className={'flex items-center ml-4 fa-insights-box--animate'}>  <SVG name={'arrowdown'} size={12} color={'grey'} /> <Text type={'title'} weight={'thin'} level={7} extraClass={'m-0 ml-1'} >{insightKeyLevel1}</Text></div>
                                    </div>
                                    <Progress percent={insightLevel1Percentage} strokeColor={'#5949BC'} className={'fa-custom-stroke-bg'} showInfo={false} />

                                    <div className={'flex items-end'}>
                                      <div className={'flex items-center ml-4 fa-insights-box--fixed-count'}><a><Text type={'title'} weight={'regular'} level={7} extraClass={'m-0 tracking-wider'} >{numberWithCommas(dataItem.factors_insights_users_count)}</Text></a> </div>
                                      <div className={'flex items-center ml-4 fa-insights-box--animate'}>  <SVG name={'arrowdown'} size={12} color={'grey'} /> <Text type={'title'} weight={'thin'} level={7} extraClass={'m-0 ml-1'} >{insightKeyLevel2}</Text></div>
                                    </div>
                                    <Progress percent={insightLevel2Percentage} strokeColor={'#5949BC'} className={'fa-custom-stroke-bg'} showInfo={false} />

                                    <div className={'flex items-end'}>
                                      <div className={'flex items-center ml-4 fa-insights-box--fixed-count'}> <Text type={'title'} weight={'regular'} level={7} extraClass={'m-0 mt-2 tracking-wider'} >{`${numberWithCommas(dataItem.factors_goal_users_count)}`}</Text><span><Text type={'title'} weight={'thin'} level={7} extraClass={'m-0 mt-2 ml-1'} >{`(${dataItem.factors_insights_percentage}% goal completion)`}</Text></span></div>
                                      <div className={'flex items-center ml-4 fa-insights-box--animate'}><SVG name={'arrowdown'} size={12} color={'grey'} /><Text type={'title'} weight={'thin'} level={7} extraClass={'m-0 ml-1'} >{`(${dataItem.factors_insights_percentage}% goal completion)`}</Text></div>
                                    </div>
                                    <Progress percent={insightLevel3Percentage} strokeColor={'#F9C06E'} className={'fa-progress'} showInfo={false} />

                                </div>

                                  <div className={'fa-sub-insights-box--spike'}>
                                      <div className={'flex justify-end items-center'}>
                                          {dataItem.factors_multiplier_increase_flag ? <SVG name={'spikeup'} size={42} /> : <SVG name={'spikedown'} size={42} />}
                                      </div>
                                  </div>
                              </div>
                          </Col>
                      </Row>
                      {!_.isEmpty(dataItem?.factors_sub_insights) && <MoreInsightsLines onClick={() => SetSubLevel2Data(dataItem.factors_sub_insights)} insightCount={dataItem?.factors_sub_insights.length} /> }
                  </div>
                  </Col>
              </Row>
          );
        })}
        </div>
  }
        
        {SubLevel2Data &&
        <>
            <Row gutter={[0, 0]} justify={'center'}>
              <Col span={24}>
                  <div className={'w-full px-4 py-2 background-color--brand-color-1 flex items-center'}>  
                      <Button className={'fa-button-ghost'} type={'text'} onClick={() => { SetSubLevel2Data(false); }}><SVG name={'doubleArrowLeft'} size={16} color={"#8692A3"} extraClass={'mr-2'}/> Back</Button>
                  </div>
              </Col>
          </Row>
          {console.log("SubLevel2Data",SubLevel2Data)}
          <div className={'fa-modal-body--custom-scrollable fa-modal-body--custom-scrollable-1'}>
            {SubLevel2Data?.factors_sub_insights?.map((dataItem, index) => {
              console.log("dataItem",dataItem);
              let insightKeyLevel1 = '';
              if (_.isEmpty(SubLevel1Data.factors_insights_key)) {
                insightKeyLevel1 = `${SubLevel1Data.factors_insights_attribute[0].factors_attribute_key} = ${SubLevel1Data.factors_insights_attribute[0].factors_attribute_value}`;
              } else {
                insightKeyLevel1 = SubLevel1Data.factors_insights_key;
              }

              let insightKeyLevel2 = '';
              if (_.isEmpty(SubLevel2Data.factors_insights_key)) {
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
                    <div className={'relative border-bottom--thin-2 fa-insight-item--sub-container px-4'}>
                          <Row gutter={[0, 0]} justify={'center'}>
                              <Col span={24}>
                                  <div className={'relative border-left--thin-2 m-0 pl-10 py-6 fa-insight-item'}>
                                      <Text type={'title'} level={4} extraClass={'m-0'} > <span dangerouslySetInnerHTML={{__html: `and then <a>${insightKeyLevel2}</a> ${dataItem.factors_insights_text}` }}/>  </Text>
                                      <Text type={'title'} level={3} weight={'bold'} extraClass={'m-0'} >{`${dataItem.factors_insights_multiplier}x`}</Text>

                                      <div className={'mt-8 w-9/12'}>

                                    <div className={'flex items-end'}>
                                      <div className={'flex items-center ml-4 fa-insights-box--fixed-count'}><a><Text type={'title'} weight={'regular'} level={7} extraClass={'m-0 tracking-wider'} >{numberWithCommas(SubLevel1Data.factors_insights_users_count)}</Text></a> </div>
                                      <div className={'flex items-center ml-4 fa-insights-box--animate'}>  <SVG name={'arrowdown'} size={12} color={'grey'} /> <Text type={'title'} weight={'thin'} level={7} extraClass={'m-0 ml-1'} >{insightKeyLevel1}</Text></div>
                                    </div>
                                    <Progress percent={100} strokeColor={'#5949BC'} className={'fa-custom-stroke-bg'} showInfo={false} />

                                    <div className={'flex items-end'}>
                                      <div className={'flex items-center ml-4 fa-insights-box--fixed-count'}><a><Text type={'title'} weight={'regular'} level={7} extraClass={'m-0 tracking-wider'} >{numberWithCommas(SubLevel2Data.factors_insights_users_count)}</Text></a> </div>
                                      <div className={'flex items-center ml-4 fa-insights-box--animate'}>  <SVG name={'arrowdown'} size={12} color={'grey'} /> <Text type={'title'} weight={'thin'} level={7} extraClass={'m-0 ml-1'} >{insightKeyLevel2}</Text></div>
                                    </div>
                                    <Progress percent={insightLevel2Percentage} strokeColor={'#5949BC'} className={'fa-custom-stroke-bg'} showInfo={false} />

                                    <div className={'flex items-end'}>
                                      <div className={'flex items-center ml-4 fa-insights-box--fixed-count'}><a><Text type={'title'} weight={'regular'} level={7} extraClass={'m-0 tracking-wider'} >{numberWithCommas(dataItem.factors_insights_users_count)}</Text></a> </div>
                                      <div className={'flex items-center ml-4 fa-insights-box--animate'}>  <SVG name={'arrowdown'} size={12} color={'grey'} /> <Text type={'title'} weight={'thin'} level={7} extraClass={'m-0 ml-1'} >{insightKeyLevel3}</Text></div>
                                    </div>
                                    <Progress percent={insightLevel3Percentage} strokeColor={'#5949BC'} className={'fa-custom-stroke-bg'} showInfo={false} />

                                    <div className={'flex items-end'}>
                                      <div className={'flex items-center ml-4 fa-insights-box--fixed-count'}> <Text type={'title'} weight={'regular'} level={7} extraClass={'m-0 mt-2 tracking-wider'} >{`${numberWithCommas(dataItem.factors_goal_users_count)}`}</Text><span><Text type={'title'} weight={'thin'} level={7} extraClass={'m-0 mt-2 ml-1'} >{`(${dataItem.factors_insights_percentage}% goal completion)`}</Text></span></div>
                                      <div className={'flex items-center ml-4 fa-insights-box--animate'}><SVG name={'arrowdown'} size={12} color={'grey'} /><Text type={'title'} weight={'thin'} level={7} extraClass={'m-0 ml-1'} >{`(${dataItem.factors_insights_percentage}% goal completion)`}</Text></div>
                                    </div>
                                    <Progress percent={insightLevel4Percentage} strokeColor={'#F9C06E'} className={'fa-progress'} showInfo={false} />

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
            </div>
        </>
        }

        </Modal>

    );
  } else return null;
};

export default SubInsightItem;
