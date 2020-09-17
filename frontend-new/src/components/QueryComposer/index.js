import React, { useState } from 'react';
import {
  Button, Collapse, Select, Popover
} from 'antd';
import { SVG, Text } from 'factorsComponents';
import styles from './index.module.scss';
import QueryBlock from './QueryBlock';
import SeqSelector from './AnalysisSeqSelector';
import GroupBlock from './GroupBlock';

const { Option } = Select;

const { Panel } = Collapse;

function QueryComposer({
  queries, runQuery, eventChange, queryType
}) {
  const [analyticsSeqOpen, setAnalyticsSeqVisible] = useState(false);

  const [queryOptions, setQueryOptions] = useState({
    groupBy: {
      property: '',
      eventValue: ''
    },
    event_analysis_seq: '',
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
                    <QueryBlock index={index + 1} event={event} queries={queries} eventChange={eventChange}></QueryBlock>
                </div>
      );
    });

    if (queries.length < 6) {
      blockList.push(
                <div className={styles.composer_body__query_block}>
                    <QueryBlock index={queries.length + 1} queries={queries} eventChange={eventChange}></QueryBlock>
                </div>
      );
    }

    return blockList;
  };

  const groupByBlock = () => {
    if (queryType === 'event' && queries.length < 1) { return null; }
    if (queryType === 'funnel' && queries.length < 2) { return null; }

    return (
                <div className={'fa--query_block bordered '}>
                    <GroupBlock groupBy={queryOptions.groupBy} events={queries}></GroupBlock>
                </div>
    );
  };

  const setEventSequence = (value) => {
    const options = Object.assign({}, queryOptions);
    options.event_analysis_seq = value;
    setQueryOptions(options);
  };

  const setAnalysisSequence = (seq) => {
    const options = Object.assign({}, queryOptions);
    options.session_analytics_seq = seq;
    setQueryOptions(options);
  };

  const moreOptionsBlock = () => {
    if (queries.length >= 2) {
      return (
                <div className={' fa--query_block bordered '}>
                <Collapse bordered={false} expandIcon={() => {}} expandIconPosition={'right'}>
                    <Panel header={<div className={'flex justify-between items-center'}>
                        <Text type={'title'} level={6} weight={'bold'} extraClass={'m-0 mb-2 inline'}>More options</Text>
                        <SVG name="plus" />
                        </div>
                        }>
                        <div className={'flex justify-start items-center'}>
                            <span className={'mr-2'}>
                                <SVG name="sortdown" size={16} color={'purple'}></SVG>
                            </span>
                            <Text type={'title'} level={7} extraClass={'m-0 mr-2 inline'}>Analyse events in the</Text>
                            <div>
                                <Select
                                    style={{ width: 170 }}
                                    value="same_sequence" onChange={setEventSequence}
                                    className={'no-ant-border'}
                                    >
                                    <Option value="same_sequence"> Same Sequence</Option>
                                    <Option value="exact_sequence"> Exact Sequence</Option>
                                </Select>
                            </div>
                        </div>

                        <div className={'flex flex-col justify-start items-start mt-4'}>
                            <div className={'flex justify-start items-center'}>
                                <span className={'mr-2'}>
                                    <SVG name="sortdown" size={16} color={'purple'}></SVG>
                                </span>
                                <Text type={'title'} level={7} extraClass={'m-0 mr-2 inline'}>In Session Analytics</Text>
                            </div>

                            <div className={'flex justify-start items-center mt-2'}>
                                <div className={styles.composer_body__session_analytics__options}>
                                    <Popover
                                        className="fa-event-popover"
                                        content={
                                            <SeqSelector
                                                seq={queryOptions.session_analytics_seq}
                                                queryCount={queries.length}
                                                setAnalysisSequence={setAnalysisSequence}
                                            />
                                        }
                                        trigger="click"
                                        visible={analyticsSeqOpen}
                                        onVisibleChange={(visible) => setAnalyticsSeqVisible(visible)}
                                    >
                                        <Button Button type="link" className={'ml-4'} size={'small'}>
                                            Between &nbsp;
                                            {queryOptions.session_analytics_seq.start}
                                            &nbsp;
                                                to
                                                &nbsp;
                                            {queryOptions.session_analytics_seq.end}
                                        </Button>
                                    </Popover>
                                    <Text type={'paragraph'} mini weight={'thin'} extraClass={'m-0 ml-2 inline'}>happened in the same session</Text>

                                </div>
                            </div>
                        </div>
                    </Panel>
                </Collapse>
            </div>
      );
    }
  };

  const footer = () => {
    if (queryType === 'event' && queries.length < 1) { return null; }
    if (queryType === 'funnel' && queries.length < 2) { return null; } else {
      return (
                <div className={styles.composer_footer}>
                    <Button><SVG name={'calendar'} extraClass={'mr-1'} />Last Week </Button>
                    <Button type="primary" onClick={runQuery}>Run Query</Button>
                </div>
      );
    }
  };

  return (
        <div className={styles.composer_body}>
            {queryList()}
            {groupByBlock()}
            {queryType === 'funnel' ? moreOptionsBlock() : null}
            {footer()}
        </div>
  );
}

export default QueryComposer;
