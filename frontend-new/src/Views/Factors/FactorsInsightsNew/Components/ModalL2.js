import React, { useState, useEffect } from 'react';
import {
    Row, Col, Button, Spin, Tag, Modal
} from 'antd';
import _ from 'lodash';
import { Text, SVG, FaErrorComp, FaErrorLog, Number } from 'factorsComponents';
import CardInsight from './CardInsight';
import ModalTable from './ModalTable';

const L2Modal = ({ data, setModalL2, showModalL2, modalData, explainMatchEventName }) => {

    console.log('Debug Mode: Insights L1 -->', data)
    console.log('Debug Mode: Subinsights L2+ -->', modalData)
    const isAttribute = modalData?.factors_insights_type == "attribute";
    return (
        <>
            <Modal
                visible={showModalL2}
                onCancel={() => setModalL2(false)}
                onOk={() => setModalL2(false)}
                title={null}
                footer={null}
                className='explain-insight--modal'
                title={<Text type={'title'} level={6} weight={'bold'} extraClass={'m-0 capitalize'}>{`${isAttribute ? 'Significant Segments' : `Significant Engagements`}`}</Text>}
            >
                <div className='p-4'>
                    {isAttribute && <div className={`py-2 px-2 flex items-center mb-2`}>
                        <Tag className={'m-0 mx-2'} className={'fa-tag--regular fa-tag--highlight truncate'}>
                            {explainMatchEventName(modalData?.factors_insights_attribute[0]?.factors_attribute_key, false, 'blue')}
                        </Tag>
                        <Text type={'title'} level={7} weight={'bold'} color={'grey'} extraClass={'m-0 mr-3'}>
                            {`= ${modalData?.factors_insights_attribute[0]?.factors_attribute_value}`}
                        </Text>
                    </div>}

                    <div className={'flex items-center justify-center w-full explain-insight--wrapper'}>

                        {/* first insight card */}

                        {!isAttribute ?
                            <CardInsight
                                title={data?.goal?.st_en ? data?.goal?.st_en : "All Visitors"}
                                count={data?.total_users_count}
                                arrow={true}
                                tagTitle={isAttribute ? 'A' : ''}
                            /> :
                            <CardInsight
                                title={`${data?.goal?.st_en ? data?.goal?.st_en : "All visitors"} with ${explainMatchEventName(modalData?.factors_insights_attribute[0]?.factors_attribute_key, true)} = ${modalData?.factors_insights_attribute[0]?.factors_attribute_value} `}
                                count={modalData?.factors_insights_users_count}
                                arrow={true}
                                tagTitle={isAttribute ? 'A' : ''}
                            />
                        }

                        {/* second insight card only for journeys */}
                        {!isAttribute && <>
                            <CardInsight
                                title={modalData?.factors_insights_attribute ? modalData?.factors_insights_attribute[0]?.factors_attribute_key : modalData?.factors_insights_key}
                                count={modalData?.factors_insights_users_count}
                                arrow={true}
                            // conv={modalData?.factors_insights_percentage}
                            /></>}

                        {/* third insight card for both */}
                        <CardInsight
                            title={data?.goal?.en_en}
                            count={modalData?.factors_goal_users_count}
                            arrow={false}
                            conv={modalData?.factors_insights_percentage}
                            tagTitle={isAttribute ? 'B' : ''}
                        />

                    </div>

                    {isAttribute && <div className={'flex items-center justify-center mt-4'}>
                        <Text type={'title'} level={7} extraClass={'m-0 mr-1'}>
                            {`From A, ${modalData?.factors_insights_users_count} were of `}
                        </Text>
                        <Tag className={'m-0 mx-2'} className={'fa-tag--regular fa-tag--highlight truncate'}>
                            {explainMatchEventName(modalData?.factors_insights_attribute[0]?.factors_attribute_key, false, 'blue')}
                        </Tag>
                        <Text type={'title'} level={7} extraClass={'m-0 mr-1'}>
                            {`= ${modalData?.factors_insights_attribute[0]?.factors_attribute_value},`}
                        </Text>
                        <Text type={'title'} level={7} extraClass={'m-0 mr-1'}>
                            {`Out of which ${modalData?.factors_goal_users_count} converted to B`}
                        </Text>
                        <Text type={'title'} level={7} weight={'thin'} color={'grey'} extraClass={'m-0'}>
                            {`(`}
                        </Text>
                        <Text type={'title'} level={7} weight={'thin'} color={'grey'} extraClass={'m-0 mr-1'}>
                            <Number suffix={'%'} number={modalData?.factors_insights_percentage} />
                        </Text>
                        <Text type={'title'} level={7} weight={'thin'} color={'grey'} extraClass={'m-0'}>
                            {`conversion)`}
                        </Text>
                    </div>}

                    {(isAttribute) ? (!_.isEmpty(modalData?.factors_sub_insights) && <ModalTable data={modalData?.factors_sub_insights} modalData={modalData} explainMatchEventName={explainMatchEventName} />) : !_.isEmpty(modalData?.factors_sub_insights) && <div className={'mt-8'}>

                        <Text type={'title'} level={7} weight={'bold'} color={'grey'} extraClass={'m-0 mb-1 capitalize'}>{'Sub Insights'}</Text>

                        {modalData?.factors_sub_insights?.map((item) => {
                            return (
                                <div className={'flex items-center justify-center w-full explain-insight--wrapper mt-4'}>

                                    <CardInsight
                                        title={modalData?.factors_insights_attribute ? explainMatchEventName(modalData?.factors_insights_attribute[0]?.factors_attribute_key, true) : explainMatchEventName(modalData?.factors_insights_key, true)}
                                        count={modalData?.factors_insights_users_count}
                                        arrow={true}
                                    />

                                    <CardInsight
                                        title={item?.factors_insights_attribute ? explainMatchEventName(item?.factors_insights_attribute[0]?.factors_attribute_key, true) : explainMatchEventName(item?.factors_insights_key, true)}
                                        count={item?.factors_insights_users_count}
                                        arrow={true}
                                    />

                                    <CardInsight
                                        title={data?.goal?.en_en}
                                        count={item?.factors_goal_users_count}
                                        arrow={false}
                                        conv={item?.factors_insights_percentage}
                                        flag={item?.factors_multiplier_increase_flag}
                                        showflag={true}
                                    />
                                </div>
                            )
                        })}


                    </div>
                    }
                </div>
            </Modal>
        </>
    )
}
export default L2Modal