import React, { useState, useEffect, useCallback } from 'react';
import { connect, useSelector } from 'react-redux';
import { bindActionCreators } from 'redux';
import { Button, Collapse, Select, Popover } from 'antd';
import MomentTz from 'Components/MomentTz';
import { SVG, Text } from '../factorsComponents';
import styles from './index.module.scss';
import QueryBlock from './QueryBlock';
import SeqSelector from './AnalysisSeqSelector';
import GroupBlock from './GroupBlock';
import {
  QUERY_TYPE_FUNNEL,
  QUERY_TYPE_EVENT,
  QUERY_TYPE_KPI,
} from '../../utils/constants';

import FaDatepicker from '../../components/FaDatepicker';

import ComposerBlock from '../QueryCommons/ComposerBlock';

import CriteriaSection from './CriteriaSection';

import { DEFAULT_DATE_RANGE, displayRange } from './DateRangeSelector/utils';

import {
  fetchEventNames,
  getUserProperties,
  getEventProperties,
} from 'Reducers/coreQuery/middleware';

import GLobalFilter from './GlobalFilter';
import { getValidGranularityOptions } from '../../utils/dataFormatter';

const { Option } = Select;

const { Panel } = Collapse;

function KPIComposer({
  queries,
  runQuery,
  eventChange,
  queryType,
  fetchEventNames,
  getUserProperties,
  getEventProperties,
  activeProject,
  eventProperties,
  queryOptions,
  setQueryOptions,
  runFunnelQuery,
  runCampaignsQuery,
  handleRunQuery,
  collapse = false,
  setCollapse,
  selectedMainCategory,
  setSelectedMainCategory,
  KPIConfigProps,
}) {
  const [analyticsSeqOpen, setAnalyticsSeqVisible] = useState(false);
  const [calendarLabel, setCalendarLabel] = useState('Pick Dates');
  const [criteriaTabOpen, setCriteriaTabOpen] = useState(false);

  const userProperties = useSelector((state) => state.coreQuery.userProperties);

  // useEffect(() => {
  //   if (activeProject && activeProject.id) {
  //     fetchEventNames(activeProject.id);
  //     getUserProperties(activeProject.id, queryType);
  //   }
  // }, [activeProject, fetchEventNames]);

  useEffect(() => {
    setSelectedMainCategory(queries[0]);
  }, [queries]);

  useEffect(() => {
    convertToDateRange();
  }, [queryOptions]);

  const queryList = () => {
    const blockList = [];

    queries.forEach((event, index) => {
      blockList.push(
        <div key={index} className={styles.composer_body__query_block}>
          <QueryBlock
            index={index + 1}
            queryType={queryType}
            event={event}
            queries={queries}
            eventChange={eventChange}
            selectedMainCategory={selectedMainCategory}
            setSelectedMainCategory={setSelectedMainCategory}
            KPIConfigProps={KPIConfigProps}
          />
        </div>
      );
    });

    if (queries.length < 6) {
      blockList.push(
        <div key={'init'} className={styles.composer_body__query_block}>
          <QueryBlock
            queryType={queryType}
            index={queries.length + 1}
            queries={queries}
            eventChange={eventChange}
            groupBy={queryOptions.groupBy}
            selectedMainCategory={selectedMainCategory}
            setSelectedMainCategory={setSelectedMainCategory}
            KPIConfigProps={KPIConfigProps}
          />
        </div>
      );
    }

    return blockList;
  };

  const setGlobalFiltersOption = (filters) => {
    const opts = Object.assign({}, queryOptions);
    opts.globalFilters = filters;
    setQueryOptions(opts);
  };

  const renderGlobalFilterBlock = (isSameKPIgrp) => {
    const [filterBlockOpen, setFilterBlockOpen] = useState(true);
    if (!isSameKPIgrp || _.isEmpty(queries)) {
      return null;
    }
    try {
      if (queryType === QUERY_TYPE_EVENT && queries.length < 1) {
        return null;
      }
      if (queryType === QUERY_TYPE_FUNNEL && queries.length < 2) {
        return null;
      }

      return (
        <ComposerBlock
          blockTitle={'FILTER BY'}
          isOpen={filterBlockOpen}
          showIcon={true}
          onClick={() => setFilterBlockOpen(!filterBlockOpen)}
          extraClass={`no-padding-l`}
        >
          <div key={0} className={'fa--query_block borderless no-padding '}>
            <GLobalFilter
              filters={queryOptions.globalFilters}
              setGlobalFilters={setGlobalFiltersOption}
              onFiltersLoad={[
                () => {
                  getUserProperties(activeProject.id, queryType);
                },
              ]}
              selectedMainCategory={selectedMainCategory}
              KPIConfigProps={KPIConfigProps}
            ></GLobalFilter>
          </div>
        </ComposerBlock>
      );
    } catch (err) {
      console.log(err);
    }
  };

  const groupByBlock = (isSameKPIgrp) => {
    const [groupBlockOpen, setGroupBlockOpen] = useState(true);
    if (!isSameKPIgrp || _.isEmpty(queries)) {
      return null;
    }

    try {
      if (queryType === QUERY_TYPE_EVENT && queries.length < 1) {
        return null;
      }
      if (queryType === QUERY_TYPE_FUNNEL && queries.length < 2) {
        return null;
      }

      return (
        <ComposerBlock
          blockTitle={'BREAKDOWN'}
          isOpen={groupBlockOpen}
          showIcon={true}
          onClick={() => setGroupBlockOpen(!groupBlockOpen)}
          extraClass={`no-padding-l`}
        >
          <div key={0} className={'fa--query_block borderless no-padding '}>
            <GroupBlock
              queryType={queryType}
              events={queries}
              selectedMainCategory={selectedMainCategory}
              KPIConfigProps={KPIConfigProps}
            />
          </div>
        </ComposerBlock>
      );
    } catch (err) {
      console.log(err);
    }
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

  const getDateRange = () => {
    const ranges = [{ ...DEFAULT_DATE_RANGE }];
    const queryOptionsState = Object.assign({}, queryOptions);

    if (
      queryOptionsState &&
      queryOptionsState.date_range &&
      queryOptionsState.date_range.from &&
      queryOptionsState.date_range.to
    ) {
      ranges[0].startDate = MomentTz(queryOptionsState.date_range.from);
      ranges[0].endDate = MomentTz(queryOptionsState.date_range.to);
    }

    return ranges;
  };

  const setDateRange = (dates) => {
    const queryOptionsState = Object.assign({}, queryOptions);
    if (dates && dates.startDate && dates.endDate) {
      if (Array.isArray(dates.startDate)) {
        queryOptionsState.date_range.from = dates.startDate[0];
        queryOptionsState.date_range.to = dates.startDate[1];
      } else {
        queryOptionsState.date_range.from = dates.startDate;
        queryOptionsState.date_range.to = dates.endDate;
      }
      const frequency = getValidGranularityOptions({
        from: queryOptionsState.date_range.from,
        to: queryOptionsState.date_range.to,
      })[0];
      queryOptionsState.date_range.frequency = frequency;
      setQueryOptions(queryOptionsState);
    }
  };

  const convertToDateRange = () => {
    const range = getDateRange();
    setCalendarLabel(displayRange(range[0]));
  };

  // const handleRunQueryCamp = useCallback(() => {
  //   console.log('handleRunQueryCamp',queryType )
  //   // if (queryType === QUERY_TYPE_EVENT) {
  //   //   runQuery(false);
  //   // } else {
  //   //   runFunnelQuery(false);
  //   // }
  //   handleRunQuery()
  // }, [runFunnelQuery, runQuery, queryType]);

  const handleRunQueryCamp = () => {
    handleRunQuery();
  };

  const footer = () => {
    try {
      if (queryType === QUERY_TYPE_KPI && queries.length == 0) {
        return null;
      } else {
        return (
          <div
            className={
              !collapse ? styles.composer_footer : styles.composer_footer_right
            }
          >
            {!collapse ? (
              <FaDatepicker
                customPicker
                presetRange
                monthPicker
                placement='topRight'
                buttonSize={'large'}
                range={{
                  startDate: queryOptions.date_range.from,
                  endDate: queryOptions.date_range.to,
                }}
                onSelect={setDateRange}
              />
            ) : (
              <Button
                className={`mr-2`}
                size={'large'}
                type={'default'}
                onClick={() => setCollapse(false)}
              >
                <SVG name={`arrowUp`} size={20} extraClass={`mr-1`}></SVG>
                Collapse all
              </Button>
            )}
            <Button
              className={`ml-2`}
              size={'large'}
              type='primary'
              onClick={handleRunQueryCamp}
            >
              Run Analysis
            </Button>
          </div>
        );
      }
    } catch (err) {
      console.log(err);
    }
  };

  const renderEACrit = () => {
    return (
      <div>
        <CriteriaSection
          queryCount={queries.length}
          queryType={QUERY_TYPE_EVENT}
        ></CriteriaSection>
      </div>
    );
  };

  const renderSeqSel = () => {
    if (
      queryOptions.session_analytics_seq.start &&
      queryOptions.session_analytics_seq.end
    ) {
      return (
        <>
          <Text
            type={'paragraph'}
            mini
            weight={'thin'}
            extraClass={'m-0 ml-2 inline'}
          >
            Where sequence
          </Text>
          <Popover
            className='fa-event-popover'
            content={
              <SeqSelector
                seq={queryOptions.session_analytics_seq}
                queryCount={queries.length}
                setAnalysisSequence={setAnalysisSequence}
              />
            }
            trigger='click'
            visible={analyticsSeqOpen}
            onVisibleChange={(visible) => setAnalyticsSeqVisible(visible)}
          >
            <Button Button type='link' className={'ml-2'}>
              Between &nbsp;
              {queryOptions.session_analytics_seq.start}
              &nbsp; to &nbsp;
              {queryOptions.session_analytics_seq.end}
            </Button>
          </Popover>
          <Text
            type={'paragraph'}
            mini
            weight={'thin'}
            extraClass={'m-0 ml-2 inline'}
          >
            happened in the same session
          </Text>
        </>
      );
    } else {
      return (
        <>
          <Text
            type={'paragraph'}
            mini
            weight={'thin'}
            extraClass={'m-0 ml-2 inline'}
          >
            Where
          </Text>
          <Popover
            className='fa-event-popover'
            content={
              <SeqSelector
                seq={queryOptions.session_analytics_seq}
                queryCount={queries.length}
                setAnalysisSequence={setAnalysisSequence}
              />
            }
            trigger='click'
            visible={analyticsSeqOpen}
            onVisibleChange={(visible) => setAnalyticsSeqVisible(visible)}
          >
            <Button Button type='link' className={'ml-2'}>
              Select Sequence
            </Button>
          </Popover>
          <Text
            type={'paragraph'}
            mini
            weight={'thin'}
            extraClass={'m-0 ml-2 inline'}
          >
            happened in the same session
          </Text>
        </>
      );
    }
  };

  const renderFuCrit = () => {
    return (
      <div className={'flex justify-start items-center mt-2'}>
        <div className={styles.composer_body__session_analytics__options}>
          {renderSeqSel()}
        </div>
      </div>
    );
  };

  const renderCriteria = () => {
    const [criterieaBlockOpen, setCriterieaBlockOpen] = useState(true);
    try {
      if (queryType === QUERY_TYPE_EVENT) {
        if (queries.length <= 0) return null;

        return (
          <ComposerBlock
            blockTitle={'CRITERIA'}
            isOpen={criterieaBlockOpen}
            showIcon={true}
            onClick={() => {
              setCriterieaBlockOpen(!criterieaBlockOpen);
            }}
            extraClass={`no-padding-l`}
          >
            <div className={styles.criteria}>{renderEACrit()}</div>
          </ComposerBlock>
        );
      }
      if (queryType === QUERY_TYPE_FUNNEL) {
        return null;
        // if (queries.length <= 1) return null;
        // return (
        //   <ComposerBlock blockTitle={'CRITERIA'}
        //   isOpen={true} showIcon={false}>
        //     {renderFuCrit()}

        //   </ComposerBlock>
        // );
      }
    } catch (err) {
      console.log(err);
    }
  };

  const renderQueryList = () => {
    const [eventBlockOpen, setEventBlockOpen] = useState(true);
    try {
      return (
        <ComposerBlock
          blockTitle={'KPI TO ANALYSE'}
          isOpen={eventBlockOpen}
          showIcon={true}
          onClick={() => setEventBlockOpen(!eventBlockOpen)}
          extraClass={`no-padding-l`}
        >
          {queryList()}
        </ComposerBlock>
      );
    } catch (err) {
      console.log(err);
    }
  };

  const isSameKPIgrp = queries.every(
    (item, index) => queries[0].group == queries[index].group
  );
  return (
    <div className={styles.composer_body}>
      {renderQueryList()}
      {renderGlobalFilterBlock(isSameKPIgrp)}
      {groupByBlock(isSameKPIgrp)}
      {renderCriteria()}
      {footer()}
    </div>
  );
}

const mapStateToProps = (state) => ({
  activeProject: state.global.active_project,
  eventProperties: state.coreQuery.eventProperties,
});

const mapDispatchToProps = (dispatch) =>
  bindActionCreators(
    {
      fetchEventNames,
      getEventProperties,
      getUserProperties,
    },
    dispatch
  );

export default connect(mapStateToProps, mapDispatchToProps)(KPIComposer);
