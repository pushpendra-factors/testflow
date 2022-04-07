import React, { useState, useEffect } from 'react';
import {
    Row, Col, Button, Spin, Tag, Modal
} from 'antd';
import _, { isEmpty } from 'lodash';
import { connect } from 'react-redux';
import { Text, SVG, FaErrorComp, FaErrorLog, Number } from 'factorsComponents';

const InsightItem = ({ data, modalData, explainMatchEventName }) => {

    const renderAttributeValue = (data) => {
        let attributeValue = '';
        if (_.isEmpty(data?.factors_insights_attribute[0]?.factors_attribute_value)) {
            let attributeBoundKey = '';
            if (data?.factors_insights_attribute[0]?.factors_attribute_use_bound == 'OnlyUpper') {
                attributeBoundKey = `<= ${data?.factors_insights_attribute[0]?.factors_attribute_upper_bound}`
            }
            if (data?.factors_insights_attribute[0]?.factors_attribute_use_bound == 'OnlyLower') {
                attributeBoundKey = `>= ${data?.factors_insights_attribute[0]?.factors_attribute_lower_bound}`
            }
            if (data?.factors_insights_attribute[0]?.factors_attribute_use_bound == 'Both') {
                attributeBoundKey = `>= ${data?.factors_insights_attribute[0]?.factors_attribute_lower_bound} and <= ${data?.factors_insights_attribute[0]?.factors_attribute_upper_bound}`
            }
            attributeValue = attributeBoundKey;
        }
        else {
            attributeValue = `= ${data?.factors_insights_attribute[0]?.factors_attribute_value}`;
        }
        return attributeValue
    }

    if (data) {
        return data?.map((item) => {
            if (item?.factors_insights_type == "attribute") {
                return (
                    <div className={`flex items-center justify-between cursor-pointer explain-table--row`}>
                        <div className={`py-2 px-4 flex items-center `}>
                            <Tag className={'m-0 mx-2'} className={'fa-tag--regular fa-tag--highlight truncate'}> {explainMatchEventName(item?.factors_insights_attribute[0]?.factors_attribute_key, false, 'blue')}</Tag>
                            {/* <Text type={'title'} level={7} weight={'bold'} color={'grey'} extraClass={'m-0 mr-3'}>{generateInsightKey(item)}</Text> */}
                            <Text type={'title'} level={7} weight={'bold'} color={'grey'} extraClass={'m-0 mr-3'}>{renderAttributeValue(item)}</Text>

                        </div>

                        <div className={`flex items-center justify-end`}>
                            <div className={'py-2 px-4 flex justify-end column_right'}>
                                <Number number={item?.factors_insights_users_count} />
                            </div>
                            <div className={'py-2 px-4 flex justify-end column_right'}>
                                <Number number={item?.factors_goal_users_count} />
                            </div>
                            <div className={'py-2 px-4 flex justify-end column_right'}>
                                <Tag color={item?.factors_multiplier_increase_flag ? 'green' : "red"} className={`m-0 mx-1 ${item?.factors_multiplier_increase_flag ? 'fa-tag--green' : "fa-tag--red"}`}>
                                    <Number suffix={'%'} number={item?.factors_insights_percentage} />
                                </Tag>
                            </div>
                        </div>
                    </div>
                )
            }
            else return null
        })
    }
    else return null
}




const ModalTable = ({ data, modalData, explainMatchEventName }) => {
    if (data) {
        return (
            <>
                <div className={'border--thin-2  border-radius--sm mt-10'}>
                    <div className={'py-4 pl-6 background-color--brand-color-1 border-radius--sm flex items-center justify-between'}>
                        <Text type={'title'} level={7} weight={'bold'} extraClass={'m-0 capitalize'}>{`Sub Segments`}</Text>
                        <div className={`flex items-center justify-end explain-table--row`}>
                            <div className={'py-2 px-4 flex justify-end column_right'}>
                                <Text type={'title'} level={7} weight={'bold'} color={'grey'} extraClass={'m-0'}>{`A`}</Text>
                            </div>
                            <div className={'py-2 px-4 flex justify-end column_right'}>
                                <Text type={'title'} level={7} weight={'bold'} color={'grey'} extraClass={'m-0'}>{`B`}</Text>
                            </div>
                            <div className={'py-2 px-4 flex justify-end column_right'}>
                                <Text type={'title'} level={7} weight={'bold'} color={'grey'} extraClass={'m-0'}>{`Conversion`}</Text>
                            </div>
                        </div>
                    </div>
                    <Row gutter={[0, 0]}>
                        <Col span={24}>
                            <InsightItem
                                showIncrease={true}
                                data={data}
                                modalData={modalData}
                                explainMatchEventName={explainMatchEventName}
                            />
                        </Col>
                    </Row>
                </div>
            </>
        )
    }
    else return null
}
export default ModalTable


