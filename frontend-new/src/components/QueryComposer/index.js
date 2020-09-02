import React, { useState } from 'react';
import { Drawer, Button, Collapse, Select, Popover } from 'antd';
import {SVG, Text} from 'factorsComponents'; 
import styles from './index.module.scss';
import QueryBlock from './QueryBlock';
import SeqSelector from './AnalysisSeqSelector';
import GroupBlock from './GroupBlock';

const { Option } = Select;

const { Panel } = Collapse;

function QueryComposer({ queries, runQuery, eventChange}) {

    const [analyticsSeqOpen, setAnalyticsSeqVisible] = useState(false);

    const [queryOptions, setQueryOptions] = useState({
        groupBy: {
            property: "",
            eventValue: ""
        },
        event_analysis_seq: "",
        session_analytics_seq: {
            start: 1,
            end: 2
        }
    }); 



    const queryList = () => {
        const blockList = [];

        queries.forEach((event, index) => {
            blockList.push(
                <div className={styles.composer_body__query_block}>
                    <QueryBlock index={index+1} event={event} eventChange={eventChange}></QueryBlock>
                </div>
            )
        });

        if(queries.length < 6) {
            blockList.push(
                <div className={styles.composer_body__query_block}>
                    <QueryBlock index={queries.length+1} eventChange={eventChange}></QueryBlock>
                </div>
            )
        }

        return blockList;
    }

    const groupByBlock = () => {
        if(queries.length >= 2) {
            return ( 
                <div className={`fa--query_block `}>
                    <GroupBlock groupBy={queryOptions.groupBy} events={queries}></GroupBlock>
                </div>
            )
        }
    }

    const setEventSequence = (value) => {
        const options = Object.assign({}, queryOptions);
        options.event_analysis_seq = value;
        setQueryOptions(options);
    }

    const setAnalysisSequence = (seq) => {
        const options = Object.assign({}, queryOptions);
        options.session_analytics_seq = seq;
        setQueryOptions(options);
    }

    const moreOptionsBlock = () => {
        if(queries.length >= 2) {
            return (
                <div className={` fa--query_block `}>
                <Collapse bordered={false} expandIcon={()=>{}} expandIconPosition={`right`}>
                    <Panel header={<div className={`flex justify-between items-center`}>
                        <Text type={'title'} level={6} weight={'bold'} extraClass={`m-0 mb-2 inline`}>More options</Text>
                        <SVG name="plus" />
                        </div>
                        }>
                        <div className={styles.composer_body__event_sequence}>
                            <span className={styles.composer_body__event_sequence__logo}>
                                <SVG name="play"></SVG>
                            </span>
                            <span className={styles.composer_body__event_sequence__text}> Analyse events in the</span>
                            <div className={styles.composer_body__event_sequence__select}>
                                <Select 
                                    showArrow={false} 
                                    style={{ width: 200}} 
                                    value="same_sequence" onChange={setEventSequence}>
                                    <Option value="same_sequence"> Same Sequence</Option>
                                    <Option value="exact_sequence"> Exact Sequence</Option>
                                </Select>
                            </div>
                        </div>

                        <div className={styles.composer_body__session_analytics}>
                            <span className={styles.composer_body__session_analytics__logo}>
                                <SVG name="play"></SVG>
                            </span>

                            <div className={styles.composer_body__session_analytics__selection}>
                                <span className={styles.composer_body__session_analytics__text}> 
                                    In Session Analytics
                                </span>

                                <div className={styles.composer_body__session_analytics__options}>
                                    <Popover
                                        content={
                                            <SeqSelector 
                                                seq={queryOptions.session_analytics_seq} 
                                                queryCount={queries.length}
                                                setAnalysisSequence={setAnalysisSequence}
                                            >
                                            </SeqSelector>
                                        }
                                        trigger="click"
                                        visible={analyticsSeqOpen}
                                        onVisibleChange={(visible) => setAnalyticsSeqVisible(visible)}
                                    >
                                        <Button type="secondary">
                                            Between 
                                            {queryOptions.session_analytics_seq.start} 
                                                to 
                                            {queryOptions.session_analytics_seq.end} 
                                        </Button>
                                    </Popover>
                                    <span>happened in the same session</span>

                                </div>
                            </div>
                        </div>
                    </Panel>
                </Collapse>
            </div>
            );
        }
    }

    const footer = () => {
        if(queries.length < 2) {return null}
        else {

            return (
                <div className={styles.composer_footer}>
                    <Button><SVG name={`calendar`} extraClass={`mr-1`} />Last Week </Button>
                    <Button type="primary" onClick={runQuery}>Run Query</Button> 
                </div>
            )
        }
    }

    return(  
        <div className={styles.composer_body}>
            {queryList()}
            {groupByBlock()}
            {moreOptionsBlock()}
            {footer()} 
        </div> 
    )
}

export default QueryComposer;