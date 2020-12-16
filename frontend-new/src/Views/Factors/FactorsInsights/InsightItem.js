import React from 'react';
import {
  Row, Col, Progress, Button
} from 'antd';
import { Text, SVG } from 'factorsComponents';
import _ from 'lodash';
import MoreInsightsLines from './MoreInsightsLines';
import { numberWithCommas } from 'Utils/dataFormatter';

const InsightItem = ({
  data, category, showSubInsightsData, displayType = false
}) => {
  if (data) { 
    const isJourney = data?.type === 'journey' ? true : false; 
    return data.insights.map((dataItem, index) => {
      if (dataItem.factors_insights_type === category) {

        let insightKey = '';
        if (_.isEmpty(dataItem.factors_insights_key)) {
          insightKey = `${dataItem.factors_insights_attribute[0].factors_attribute_key} = ${dataItem.factors_insights_attribute[0].factors_attribute_value}`;
        } else {
          insightKey = dataItem.factors_insights_key;
        } 
        const factors_insights_text = `of which users who perform <a>${insightKey}</a> show  ${dataItem.factors_insights_multiplier}x goal completion`
 
        let insightLevel1Percentage = 100; 
        let insightLevel1Journey = 100; 

        if(isJourney){
          insightLevel1Journey = (dataItem.factors_insights_users_count / data.total_users_count) * 100; 
           insightLevel1Percentage = (dataItem.factors_goal_users_count / data.total_users_count) * 100;
        }
        else{
          insightLevel1Percentage = (dataItem.factors_goal_users_count / dataItem.factors_insights_users_count) * 100; 
        }



        return (
                  <div key={index} className={'relative border-bottom--thin-2 fa-insight-item--container'}>
                      <Row gutter={[0, 0]} justify={'center'}>
                          <Col span={16}>
                              <div className={'relative border-left--thin-2 m-0 pl-16 py-8 cursor-pointer fa-insight-item'} onClick={() => {
                                  if(!_.isEmpty(dataItem?.factors_sub_insights)){
                                      showSubInsightsData(dataItem, data); 
                                  }
                              }}>
                                 {displayType && <Text type={'paragraph'} mini color={'grey'} weight={'bold'} extraClass={'uppercase fa-insights-box--type'} >{category}</Text>}
                                  <Text type={'title'} level={4} weight={'bold'} extraClass={'m-0 pr-2'} ><span dangerouslySetInnerHTML={{__html:factors_insights_text}}/> </Text>
                                    {!_.isEmpty(dataItem.factors_higher_completion_text) && <Text type={'title'} level={6} color={'grey'} extraClass={'mt-4'} >{dataItem.factors_higher_completion_text}</Text>}
                                    {!_.isEmpty(dataItem.factors_lower_completion_text) && <Text type={'title'} level={6} color={'grey'} extraClass={'mt-2'} >{dataItem.factors_lower_completion_text}</Text>}

                               
                                  <div className={'mt-8 w-9/12'}>

                                  {
                                    isJourney && <>
                                        <div className={'flex items-end'}>
                                          <div className={'flex items-center ml-4 fa-insights-box--fixed-count'}><a><Text type={'title'} weight={'regular'} level={7} extraClass={'m-0 tracking-wider'} >{numberWithCommas(data.total_users_count)}</Text></a> </div>
                                          <div className={'flex items-center ml-4 fa-insights-box--animate'}>  <SVG name={'arrowdown'} size={12} color={'grey'} /> <Text type={'title'} weight={'thin'} level={7} extraClass={'m-0 ml-1'} >{_.isEmpty(data.goal?.st_en) ? 'All Visitors' : data.goal?.st_en }</Text></div>
                                        </div>
                                        <Progress percent={100} strokeColor={'#5949BC'} className={'fa-custom-stroke-bg'} showInfo={false} /> 
                                    </>
                                  } 

                                  <div className={'flex items-end'}>
                                    <div className={'flex items-center ml-4 fa-insights-box--fixed-count'}><a><Text type={'title'} weight={'regular'} level={7} extraClass={'m-0 tracking-wider'} >{numberWithCommas(dataItem.factors_insights_users_count)}</Text></a> </div>
                                    <div className={'flex items-center ml-4 fa-insights-box--animate'}>  <SVG name={'arrowdown'} size={12} color={'grey'} /> <Text type={'title'} weight={'thin'} level={7} extraClass={'m-0 ml-1'} >{insightKey}</Text></div>
                                  </div>
                                  <Progress percent={isJourney? insightLevel1Journey : 100} strokeColor={'#5949BC'} className={'fa-custom-stroke-bg'} showInfo={false} />

                                  <div className={'flex items-end'}>
                                    <div className={'flex items-center ml-4 fa-insights-box--fixed-count'}> <Text type={'title'} weight={'regular'} level={7} extraClass={'m-0 mt-2 tracking-wider'} >{`${numberWithCommas(dataItem.factors_goal_users_count)}`}</Text><span><Text type={'title'} weight={'thin'} level={7} extraClass={'m-0 mt-2 ml-1'} >{`(${dataItem.factors_insights_percentage}% goal completion)`}</Text></span></div>
                                    <div className={'flex items-center ml-4 fa-insights-box--animate'}><SVG name={'arrowdown'} size={12} color={'grey'} /><Text type={'title'} weight={'thin'} level={7} extraClass={'m-0 ml-1'} >{`${data?.goal?.en_en} (${dataItem.factors_insights_percentage}% goal completion)`}</Text></div>
                                  </div>
                                  <Progress percent={insightLevel1Percentage} strokeColor={'#F9C06E'} className={'fa-progress'} showInfo={false} />
                                  </div>

                                  {!_.isEmpty(dataItem?.factors_sub_insights) && <div className={'fa-insights-box--actions'}>
                                    <Button type={'link'} size={'large'}>
                                        <SVG name={'corequery'} size={24} color={'grey'} />
                                    </Button>
                                  </div>
                                  }

                                  <div className={'fa-insights-box--spike'}>
                                      <div className={'flex justify-end items-center'}>
                                        <Text type={'title'} level={5} color={'grey'} weight={'bold'} extraClass={'m-0 mr-4'} >{`${dataItem.factors_insights_multiplier}x`}</Text>
                                        {dataItem.factors_multiplier_increase_flag ? <SVG name={'spikeup'} size={42} /> : <SVG name={'spikedown'} size={42} />}
                                      </div>
                                  </div>

                              </div>
                          </Col>
                      </Row>
                      {!_.isEmpty(dataItem?.factors_sub_insights) && <MoreInsightsLines onClick={() => showSubInsightsData(dataItem.factors_sub_insights)} insightCount={dataItem?.factors_sub_insights.length} /> }
                  </div>
        );
      }
    });
  } else {
    return null;
  }
};

export default InsightItem;
