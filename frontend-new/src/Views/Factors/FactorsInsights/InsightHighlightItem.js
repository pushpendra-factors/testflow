import React from 'react';
import {
  Row, Col, Progress
} from 'antd';
import { Text, SVG } from 'factorsComponents';
import _ from 'lodash';
import { numberWithCommas } from 'Utils/dataFormatter';

const ProgressColor = {
    blue: '#4D7DB4',
    yellow: '#F9C06E'
  }

  const InsightHighlightItem = ({ data }) => {
  if (data) {
    return (
            <div className={'relative my-4'}>
                <Row gutter={[0, 0]} justify={'center'}>
                    <Col span={16}>
                        <div className={'relative m-0 pl-16 py-0'}>
                            <div className={'w-full'}>

                            <div className={'relative flex items-end'}>
                                    <SVG name={'ProgressArrow'} color={ProgressColor.blue} extraClass={'mr-2 mb-1'} />
                                     <div className={'flex-grow'}> 
                                    <Text type={'title'} weight={'thin'} level={7} extraClass={'m-0 ml-2'} >{_.isEmpty(data.goal?.st_en) ? 'All Visitors' : data.goal?.st_en }</Text>
                                    <Progress strokeWidth={12} percent={100} strokeColor={'#5949BC'} className={'fa-custom-stroke-bg'} showInfo={false} />
                                    </div>
                                </div>
                            
                            <Text type={'title'} level={1} weight={'bold'} extraClass={'m-0 ml-5 my-4 progressArrow--extraline'} lineHeight={'small'}>{`${data.overall_percentage}% of all users have completed this goal`}</Text>

                            <div className={'relative flex items-end'}>
                                    <SVG name={'ProgressArrow'} color={ProgressColor.yellow} extraClass={'mr-2 mb-1'} />
                                     <div className={'flex-grow'}> 
                                        <Text type={'title'} weight={'thin'} level={7} extraClass={'m-0'} >{data.goal?.en_en}</Text>
                                        <Progress strokeWidth={12} percent={data.overall_percentage} strokeColor={'#F9C06E'} className={'fa-progress'} showInfo={false} />
                                    </div>
                            </div>

                            </div>

                            <div className={'fa-insights-box--highlight'}>
                                <div className={'flex justify-between items-end flex-col h-full py-2'}>
                                    <Text type={'title'} level={5} color={'blue'} weight={'bold'} extraClass={'m-0 tracking-wider'} >{numberWithCommas(data.total_users_count)}</Text>
                                    <div className={'flex flex-col items-end justify-center '}>
                                        {/* <Text type={'title'} level={4} color={'grey'} weight={'bold'} extraClass={'m-0'} >{`${data.overall_multiplier}x`}</Text> */}
                                        <Text type={'title'} level={7} color={'grey'} extraClass={'m-0'} >Baseline</Text>
                                    </div>
                                    <Text type={'title'} level={5} color={'yellow'} weight={'bold'} extraClass={'m-0 tracking-wider'} >{numberWithCommas(data.goal_user_count)}</Text>
                                </div>
                            </div>
                        </div>
                    </Col>
                </Row>
            </div>

    );
  } else return null;
};
export default InsightHighlightItem;
