import React from 'react';
import { Row, Col, Progress, Button } from 'antd';
import { Text, SVG } from 'factorsComponents';
import _ from 'lodash';
import MoreInsightsLines from './MoreInsightsLines';
import { numberWithCommas } from 'Utils/dataFormatter';
import { generateInsightKey } from './Utils';
import {
  CHART_COLOR_1,
  CHART_COLOR_6
} from '../../../constants/color.constants';

const ProgressColor = {
  blue: CHART_COLOR_1,
  yellow: CHART_COLOR_6
};

const InsightItem = ({
  data,
  category,
  showSubInsightsData,
  displayType = false
}) => {
  if (data) {
    // console.log('insights-data--->',data);
    const isJourney = data?.type === 'journey' ? true : false;
    return data.insights.map((dataItem, index) => {
      if (dataItem.factors_insights_type === category) {
        let insightKey = generateInsightKey(dataItem);

        let factors_insights_text = `Users who visit <a>${insightKey}</a> show ${dataItem.factors_insights_percentage}% conversion`;
        if (dataItem.factors_insights_type == 'attribute') {
          factors_insights_text = `Users with <a>${insightKey}</a> show ${dataItem.factors_insights_percentage}% conversion`;
        }
        if (dataItem.factors_insights_type == 'campaign') {
          factors_insights_text = `Users from <a>${insightKey}</a> show ${dataItem.factors_insights_percentage}% conversion`;
        }

        let insightLevel1Percentage = 100;
        let insightLevel1Journey = 100;

        if (isJourney) {
          insightLevel1Journey =
            (dataItem.factors_insights_users_count / data.total_users_count) *
            100;
          insightLevel1Percentage =
            (dataItem.factors_goal_users_count / data.total_users_count) * 100;
        } else {
          insightLevel1Percentage =
            (dataItem.factors_goal_users_count /
              dataItem.factors_insights_users_count) *
            100;
        }

        return (
          <div
            key={index}
            className={
              'relative border-bottom--thin-2 fa-insight-item--container'
            }
          >
            <Row gutter={[0, 0]} justify={'center'}>
              <Col span={16}>
                <div
                  className={
                    'relative border-left--thin-2 m-0 pl-16 py-8 cursor-pointer fa-insight-item'
                  }
                  onClick={() => {
                    if (!_.isEmpty(dataItem?.factors_sub_insights)) {
                      showSubInsightsData(dataItem, data);
                    }
                  }}
                >
                  {displayType && (
                    <Text
                      type={'paragraph'}
                      mini
                      color={'grey'}
                      weight={'bold'}
                      extraClass={'uppercase fa-insights-box--type'}
                    >
                      {category}
                    </Text>
                  )}
                  <Text
                    type={'title'}
                    level={4}
                    weight={'bold'}
                    extraClass={'m-0 pr-2'}
                  >
                    <span
                      dangerouslySetInnerHTML={{
                        __html: factors_insights_text
                      }}
                    />{' '}
                  </Text>
                  {!_.isEmpty(dataItem.factors_higher_completion_text) && (
                    <Text
                      type={'title'}
                      level={6}
                      color={'grey'}
                      extraClass={'mt-4'}
                    >
                      {dataItem.factors_higher_completion_text}
                    </Text>
                  )}
                  {!_.isEmpty(dataItem.factors_lower_completion_text) && (
                    <Text
                      type={'title'}
                      level={6}
                      color={'grey'}
                      extraClass={'mt-2'}
                    >
                      {dataItem.factors_lower_completion_text}
                    </Text>
                  )}

                  <div className={'mt-8 w-9/12'}>
                    {isJourney && (
                      <>
                        <div className={'relative flex items-end'}>
                          <SVG
                            name={'ProgressArrow'}
                            color={ProgressColor.blue}
                            extraClass={'mr-2 mb-1'}
                          />
                          <div className={'flex-grow'}>
                            <div className={'flex items-end'}>
                              {/* <div className={'flex items-center ml-4 fa-insights-box--fixed-count'}><a><Text type={'title'} weight={'regular'} level={7} extraClass={'m-0 tracking-wider'} >{numberWithCommas(data.total_users_count)}</Text></a> </div> */}
                              <div
                                className={
                                  'flex items-center fa-insights-box--animate'
                                }
                              >
                                {' '}
                                <Text
                                  type={'title'}
                                  weight={'thin'}
                                  level={7}
                                  extraClass={'m-0 ml-1'}
                                >
                                  {_.isEmpty(data.goal?.st_en)
                                    ? 'All Visitors'
                                    : data.goal?.st_en}
                                </Text>
                              </div>
                            </div>
                            <Progress
                              percent={100}
                              strokeColor={ProgressColor.blue}
                              className={
                                'fa-custom-stroke-bg fa-custom-progress-value'
                              }
                              showInfo={false}
                              value={numberWithCommas(data.total_users_count)}
                            />
                          </div>
                        </div>
                      </>
                    )}
                    <div className={'relative flex items-end'}>
                      <SVG
                        name={'ProgressArrow'}
                        color={ProgressColor.blue}
                        extraClass={'mr-2 mb-1'}
                      />
                      <div className={'flex-grow'}>
                        <div className={'flex items-end'}>
                          {/* <div className={'flex items-center ml-4 fa-insights-box--fixed-count'}><a><Text type={'title'} weight={'regular'} level={7} extraClass={'m-0 tracking-wider'} >{numberWithCommas(dataItem.factors_insights_users_count)}</Text></a> </div> */}
                          <div
                            className={
                              'flex items-center fa-insights-box--animate'
                            }
                          >
                            {' '}
                            <Text
                              type={'title'}
                              weight={'thin'}
                              level={7}
                              extraClass={'m-0 ml-1'}
                            >
                              {insightKey}
                            </Text>
                          </div>
                        </div>
                        <Progress
                          strokeWidth={10}
                          percent={isJourney ? insightLevel1Journey : 100}
                          value={numberWithCommas(
                            dataItem.factors_insights_users_count
                          )}
                          strokeColor={ProgressColor.blue}
                          className={
                            'fa-custom-stroke-bg fa-custom-progress-value'
                          }
                          showInfo={false}
                        />
                      </div>
                    </div>

                    <div className={'relative flex items-end'}>
                      <SVG
                        name={'ProgressArrow'}
                        color={ProgressColor.yellow}
                        extraClass={'mr-2 mb-1'}
                      />
                      <div className={'flex-grow'}>
                        <div className={'flex items-end'}>
                          {/* <div className={'flex items-center ml-4 fa-insights-box--fixed-count'}> <Text type={'title'} weight={'regular'} level={7} extraClass={'m-0 mt-2 tracking-wider'} >{`${numberWithCommas(dataItem.factors_goal_users_count)}`}</Text><span><Text type={'title'} weight={'thin'} level={7} extraClass={'m-0 mt-2 ml-1'} >{`(${dataItem.factors_insights_percentage}% conversion)`}</Text></span></div> */}
                          <div
                            className={
                              'flex items-center fa-insights-box--animate'
                            }
                          >
                            <Text
                              type={'title'}
                              weight={'thin'}
                              level={7}
                              extraClass={'m-0 ml-1'}
                            >{`${data?.goal?.en_en} (${dataItem.factors_insights_percentage}% conversion)`}</Text>
                          </div>
                        </div>
                        <Progress
                          strokeWidth={10}
                          percent={insightLevel1Percentage}
                          strokeColor={ProgressColor.yellow}
                          value={numberWithCommas(
                            dataItem.factors_goal_users_count
                          )}
                          className={'fa-progress fa-custom-progress-value'}
                          showInfo={false}
                        />
                      </div>
                    </div>

                    {/* {!_.isEmpty(dataItem?.factors_sub_insights) && <div className={'fa-insights-box--actions'}>
                                    <Button type={'link'} size={'large'}>
                                        <SVG name={'corequery'} size={24} color={'grey'} />
                                    </Button>
                                  </div>
                                  }  */}

                    <div className={'fa-insights-box--spike'}>
                      <div className={'flex justify-end items-center'}>
                        <div className={'flex flex-col items-end mr-4'}>
                          <Text
                            type={'title'}
                            level={5}
                            color={'grey'}
                            weight={'bold'}
                            extraClass={'m-0 fa-insights-box--multiplier pt-2'}
                          >{`${dataItem.factors_insights_multiplier}x`}</Text>
                          <Text
                            type={'title'}
                            color={'grey'}
                            level={7}
                            extraClass={'m-0 fa-insights-box--label'}
                          >
                            {dataItem.factors_multiplier_increase_flag
                              ? `Lift`
                              : `Drop`}
                          </Text>
                        </div>
                        {dataItem.factors_multiplier_increase_flag ? (
                          <SVG name={'spikeup'} size={32} color={'green'} />
                        ) : (
                          <SVG name={'spikedown'} size={32} color={'red'} />
                        )}
                      </div>
                    </div>
                  </div>
                </div>
              </Col>
            </Row>
            {!_.isEmpty(dataItem?.factors_sub_insights) && (
              <MoreInsightsLines
                onClick={() =>
                  showSubInsightsData(dataItem.factors_sub_insights)
                }
                insightCount={dataItem?.factors_sub_insights.length}
              />
            )}
          </div>
        );
      }
    });
  } else {
    return null;
  }
};

export default InsightItem;
