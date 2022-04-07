import React, { useState, useEffect } from 'react';
import {
    Row, Col, Button, Spin, Tag, Modal
} from 'antd';
import _, { isEmpty } from 'lodash';
import { Text, SVG, FaErrorComp, FaErrorLog, Number } from 'factorsComponents';
import EventTag from './EventTag';


const CardInsight = ({ title, count, arrow = false, conv, flag = false, tagTitle = false, showflag = false }) => {
    return (

        <div className={`px-6 py-2 flex flex-col background-color--brand-color-1 explain-insight--item w-full ${arrow ? 'arrow-right' : ''}`}>
            <div className={`flex items-center`}>
                {tagTitle && <EventTag text={tagTitle} color={tagTitle == 'A' ? 'blue' : 'yellow'} />}
                <Text type={'title'} level={7} extraClass={'m-0 truncate'} truncate={true} charLimit={35}>{title}</Text>
            </div>
            <div className={`flex items-center`}>
                <Text type={'title'} level={3} weight={'bold'} extraClass={'m-0 mt-1 capitalize mr-2'}>
                    <Number number={count} />
                </Text>
                {showflag ? (flag ? <SVG name={'spikeup'} size={16} color={'green'} /> : <SVG name={'spikedown'} size={16} color={'red'} />) : null}
                {conv ? <Text type={'title'} level={7} color={`${showflag ? (flag ? 'green' : 'red') : 'grey'}`} weight={flag ? 'bold' : 'normal'} extraClass={'m-0'}>
                    <Number suffix={'%'} number={conv} />{` conv.`}
                </Text> : ""}
            </div>
        </div>
    )
}
export default CardInsight