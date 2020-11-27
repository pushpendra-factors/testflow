import React from 'react';
import {
  Row, Col, Progress
} from 'antd';
import { Text } from 'factorsComponents';
import _ from 'lodash';

function numberWithCommas(x) {
  return x.toString().replace(/\B(?=(\d{3})+(?!\d))/g, ',');
}

const InsightHighlightItem = ({ data }) => {
  if (data) {
    return (
            <div className={'relative my-4'}>
                <Row gutter={[0, 0]} justify={'center'}>
                    <Col span={16}>
                        <div className={'relative border-left--thin-2 m-0 pl-16 py-2'}>
                            <div className={'w-full'}>
                            <Text type={'title'} weight={'thin'} level={7} extraClass={'m-0'} >{data.goal?.st_en}</Text>
                            <Progress percent={100} strokeColor={'#5949BC'} className={'fa-custom-stroke-bg'} showInfo={false} />

                            <Text type={'title'} level={2} weight={'bold'} extraClass={'m-0'} >{data.overall_percentage_text}</Text>

                            <Progress percent={data.overall_percentage} strokeColor={'#F9C06E'} showInfo={false} />
                            <Text type={'title'} weight={'thin'} level={7} extraClass={'m-0'} >{data.goal?.en_en}</Text>
                            </div>

                            <div className={'fa-insights-box--highlight'}>
                                <div className={'flex justify-between items-end flex-col h-full'}>
                                    <Text type={'title'} level={5} color={'blue'} weight={'bold'} extraClass={'m-0 tracking-wider'} >{numberWithCommas(data.total_users_count)}</Text>
                                    <div className={'flex flex-col items-center justify-center '}>
                                        <Text type={'title'} level={4} color={'grey'} weight={'bold'} extraClass={'m-0'} >{`${data.overall_multiplier}x`}</Text>
                                        <Text type={'title'} level={7} color={'grey'} extraClass={'m-0'} >Impact</Text>
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
