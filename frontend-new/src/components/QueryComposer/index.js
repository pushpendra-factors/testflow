/* eslint-disable */
import React, { useState, useEffect, useCallback } from 'react';
import { connect } from 'react-redux';
import { bindActionCreators } from 'redux';
import {
  Button, Collapse, Select, Popover
} from 'antd';
import moment from 'moment';
import { SVG, Text } from '../factorsComponents';
import styles from './index.module.scss';
import QueryBlock from './QueryBlock';
import SeqSelector from './AnalysisSeqSelector';
import GroupBlock from './GroupBlock';
import DateRangeSelector from './DateRangeSelector';

import {
  DEFAULT_DATE_RANGE,
  displayRange
} from './DateRangeSelector/utils';

import { fetchEventNames, getUserProperties, getEventProperties } from '../../reducers/coreQuery/middleware';

const { Option } = Select;

const { Panel } = Collapse;

function QueryComposer({
  queries, runQuery, eventChange, queryType,
  fetchEventNames,
  getUserProperties,
  getEventProperties,
  activeProject,
  eventProperties,
  queryOptions,
  setQueryOptions,
  runFunnelQuery
}) {
  const [analyticsSeqOpen, setAnalyticsSeqVisible] = useState(false);
  const [dateRangeOpen, setDateRangeVisibile] = useState(false);
  const [calendarLabel, setCalendarLabel] = useState('Pick Dates');

  useEffect(() => {
    if (activeProject && activeProject.id) {
      fetchEventNames(activeProject.id);
      getUserProperties(activeProject.id, queryType);
    }
  }, [activeProject, fetchEventNames]);

  useEffect(() => {
    queries.forEach(ev => {
      if (!eventProperties[ev.label]) {
        getEventProperties(activeProject.id, ev.label);
      }
    });
  }, [queries]);

  useEffect(() => {
    convertToDateRange();
  }, [queryOptions]);

  const queryList = () => {
    const blockList = [];

    queries.forEach((event, index) => {
      blockList.push(
        <div key={index} className={styles.composer_body__query_block}>
          <QueryBlock index={index + 1}
            queryType={queryType}
            event={event}
            queries={queries}
            eventChange={eventChange}
          ></QueryBlock>
        </div>
      );
    });

    if (queries.length < 6) {
      blockList.push(
        <div key={'init'} className={styles.composer_body__query_block}>
          <QueryBlock queryType={queryType} index={queries.length + 1}
            queries={queries} eventChange={eventChange}
            groupBy={queryOptions.groupBy}
          ></QueryBlock>
        </div>
      );
    }

    return blockList;
  };

  const groupByBlock = () => {
    if (queryType === 'event' && queries.length < 1) { return null; }
    if (queryType === 'funnel' && queries.length < 2) { return null; }

    return (
      <div key={0} className={'fa--query_block bordered '}>
        <GroupBlock
          queryType={queryType}
          events={queries}>
        </GroupBlock>
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
          <Collapse bordered={false} expandIcon={() => { }} expandIconPosition={'right'}>
            <Panel header={<div className={'flex justify-between items-center'}>
              <Text type={'title'} level={6} weight={'bold'} extraClass={'m-0 mb-2 inline'}>More options</Text>
              <SVG name="plus" color={'grey'} />
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

  const getDateRange = () => {
    const ranges = [{...DEFAULT_DATE_RANGE}];
    const queryOptionsState = Object.assign({}, queryOptions);

    if (
      queryOptionsState &&
      queryOptionsState.date_range &&
      queryOptionsState.date_range.from &&
      queryOptionsState.date_range.to
    ) {
      ranges[0].startDate = moment(queryOptionsState.date_range.from).toDate();
      ranges[0].endDate = moment(queryOptionsState.date_range.to).toDate();
    }

    return ranges;
  };

  const setDateRange = (dates) => {
    const queryOptionsState = Object.assign({}, queryOptions);
    if (dates && dates.selected) {
      queryOptionsState.date_range.from = dates.selected.startDate;
      queryOptionsState.date_range.to = dates.selected.endDate;
      setQueryOptions(queryOptionsState);
    }
    setDateRangeVisibile(false);
  };

  const convertToDateRange = () => {
    const range = getDateRange();
    setCalendarLabel(displayRange(range[0]));
  };

  const handleRunQuery = useCallback(() => {
    if (queryType === 'event') {
      runQuery('0', true);
    } else {
      runFunnelQuery();
    }
  }, [runFunnelQuery, runQuery, queryType]);

  const footer = () => {
    if (queryType === 'event' && queries.length < 1) { return null; }
    if (queryType === 'funnel' && queries.length < 2) { return null; } else {
      return (
        <div className={styles.composer_footer}>
          <Popover
            className="fa-event-popover"
            trigger="click"
            visible={dateRangeOpen}
            content={
            <DateRangeSelector
                ranges={getDateRange()}
                pickerVisible={dateRangeOpen} setDates={setDateRange} 
                closeDatePicker={() => setDateRangeVisibile(false)}
              />}
            onVisibleChange={(visible) => setDateRangeVisibile(visible)}
          >
            <Button size={'large'}><SVG name={'calendar'} extraClass={'mr-1'} /> {calendarLabel} </Button>
          </Popover>
          <Button size={'large'}type="primary" onClick={handleRunQuery}>Run Query</Button>
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

const mapStateToProps = (state) => ({
  activeProject: state.global.active_project,
  eventProperties: state.coreQuery.eventProperties
});

const mapDispatchToProps = dispatch => bindActionCreators({
  fetchEventNames,
  getEventProperties,
  getUserProperties
}, dispatch);

export default connect(mapStateToProps, mapDispatchToProps)(QueryComposer);
