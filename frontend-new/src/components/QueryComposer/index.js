import React, { useState, useEffect, useCallback, useMemo } from 'react';
import { connect, useDispatch, useSelector } from 'react-redux';
import { bindActionCreators } from 'redux';
import { Button, Tooltip } from 'antd';
import {
  fetchEventNames,
  getUserProperties,
  getGroupProperties,
  getEventProperties
} from '../../reducers/coreQuery/middleware';
import { SVG, Text } from '../factorsComponents';
import styles from './index.module.scss';
import QueryBlock from './QueryBlock';
import GroupBlock from './GroupBlock';
import {
  QUERY_TYPE_FUNNEL,
  QUERY_TYPE_EVENT,
  FunnelEventsConditionMap,
  RevFunnelEventsConditionMap
} from 'Utils/constants';
import FaDatepicker from '../FaDatepicker';
import ComposerBlock from '../QueryCommons/ComposerBlock';
import CriteriaSection from './CriteriaSection';
import GLobalFilter from './GlobalFilter';
import { getValidGranularityOptions } from 'Utils/dataFormatter';
import FaSelect from '../FaSelect';
import { TOOLTIP_CONSTANTS } from '../../constants/tooltips.constans';
import {
  INITIALIZE_GROUPBY,
  setEventGroupBy
} from 'Reducers/coreQuery/actions';
import { ReactSortable } from 'react-sortablejs';
import { isEqual } from 'lodash';
import { fetchGroups } from 'Reducers/coreQuery/services';

function QueryComposer({
  queries = [],
  setQueries,
  runQuery,
  eventChange,
  queryType,
  fetchGroups,
  fetchEventNames,
  getUserProperties,
  getGroupProperties,
  getEventProperties,
  activeProject,
  groupOpts,
  eventProperties,
  queryOptions,
  setQueryOptions,
  runFunnelQuery,
  collapse = false,
  setCollapse
}) {
  const [filterBlockOpen, setFilterBlockOpen] = useState(true);
  const [groupBlockOpen, setGroupBlockOpen] = useState(true);
  const [criterieaBlockOpen, setCriterieaBlockOpen] = useState(true);
  const [eventBlockOpen, setEventBlockOpen] = useState(true);
  const [isOrderDDVisible, setOrderDDVisible] = useState(false);
  const [isGroupDDVisible, setGroupDDVisible] = useState(false);
  const eventBreakdowns = useSelector((state) => state.coreQuery.groupBy.event);

  const dispatch = useDispatch();

  useEffect(() => {
    fetchGroups(activeProject.id, true);
  }, [activeProject]);

  const groupsList = useMemo(() => {
    let groups = [['Users', 'users']];
    groupOpts?.forEach((elem) => {
      groups.push([elem.display_name, elem.group_name]);
    });
    return groups;
  }, [groupOpts]);

  useEffect(() => {
    if (activeProject && activeProject.id) {
      getUserProperties(activeProject.id, queryType);
    }
  }, [activeProject, fetchEventNames, getUserProperties, queryType]);

  useEffect(() => {
    queries.forEach((ev) => {
      if (!eventProperties[ev.label]) {
        getEventProperties(activeProject.id, ev.label);
      }
    });
  }, [activeProject?.id, eventProperties, getEventProperties, queries]);

  const setEventsCondition = (condition) => {
    const opts = { ...queryOptions };
    opts.events_condition = condition;
    setQueryOptions(opts);
  };

  const onOrderChange = (value) => {
    setEventsCondition(value);
    setOrderDDVisible(false);
  };

  const showConditions = () => {
    const retArray = [];
    Object.entries(RevFunnelEventsConditionMap).forEach(([key, value]) =>
      retArray.push([key, value])
    );
    return retArray;
  };

  const selectEventsCondition = () => (
    <div className={`${styles.toplevel_select_dropdown}`}>
      {isOrderDDVisible ? (
        <FaSelect
          options={showConditions()}
          onClickOutside={() => setOrderDDVisible(false)}
          optionClick={(val) => onOrderChange(val[1])}
        />
      ) : null}
    </div>
  );

  const renderEventsConditionSection = () => (
    <div className='flex items-center pt-6'>
      <Text type='title' level={6} weight='bold' extraClass='m-0 mr-3'>
        EVENTS PERFORMED IN
      </Text>
      <div className={`${styles.toplevel_select}`}>
        <Tooltip title='Select Events Condition' color={TOOLTIP_CONSTANTS.DARK}>
          <Button
            className={`${styles.toplevel_select_button}`}
            type='text'
            onClick={() => setOrderDDVisible(true)}
          >
            <div className='flex items-center'>
              <Text
                type='title'
                level={7}
                weight='bold'
                color='brand-color-6'
                extraClass='m-0 mr-1'
              >
                {FunnelEventsConditionMap[queryOptions.events_condition]}
              </Text>
              <SVG name='caretDown' color='blue' />
            </div>
          </Button>
        </Tooltip>
        {selectEventsCondition()}
      </div>
    </div>
  );

  const setGroupAnalysis = (group) => {
    getGroupProperties(activeProject.id, group);
    const opts = Object.assign({}, queryOptions);
    opts.group_analysis = group;
    opts.globalFilters = [];
    dispatch({
      type: INITIALIZE_GROUPBY,
      payload: {
        global: [],
        event: []
      }
    });
    setQueries([]);
    setQueryOptions(opts);
  };

  const onGroupChange = (value) => {
    setGroupAnalysis(value);
    setGroupDDVisible(false);
  };

  const selectGroup = () => {
    return (
      <div className={`${styles.toplevel_select_dropdown}`}>
        {isGroupDDVisible ? (
          <FaSelect
            options={groupsList}
            onClickOutside={() => setGroupDDVisible(false)}
            optionClick={(val) => onGroupChange(val[1])}
          />
        ) : null}
      </div>
    );
  };

  const renderGroupSection = () => {
    try {
      return (
        <div className={`flex items-center pt-4`}>
          <Text
            type={'title'}
            level={6}
            weight={'normal'}
            extraClass={`m-0 mr-3`}
          >
            Analyse
          </Text>
          <div className={`${styles.toplevel_select}`}>
            <Tooltip
              title='Select profile type to analyse'
              color={TOOLTIP_CONSTANTS.DARK}
            >
              <Button
                className={`${styles.groupsection_button}`}
                type='text'
                onClick={() => setGroupDDVisible(true)}
              >
                <div className={`flex items-center`}>
                  <Text
                    type={'title'}
                    level={6}
                    weight={'bold'}
                    extraClass={`m-0 mr-1`}
                  >
                    {
                      groupsList?.find(
                        ([_, groupName]) =>
                          groupName === queryOptions?.group_analysis
                      )?.[0]
                    }
                  </Text>
                  <SVG name='caretDown' />
                </div>
              </Button>
            </Tooltip>
            {selectGroup()}
          </div>
        </div>
      );
    } catch (err) {
      console.log(err);
    }
  };

  const queryList = () => {
    const blockList = [];
    queries.forEach((event, index) => {
      blockList.push(
        <div key={index} className={styles.composer_body__query_block}>
          <QueryBlock
            availableGroups={groupsList}
            index={index + 1}
            queryType={queryType}
            event={event}
            queries={queries}
            eventChange={eventChange}
            groupAnalysis={queryOptions.group_analysis}
          />
        </div>
      );
    });

    if (
      (queryType === QUERY_TYPE_FUNNEL && queries.length < 10) ||
      (queryType === QUERY_TYPE_EVENT && queries.length < 6)
    ) {
      blockList.push(
        <div key='init' className={styles.composer_body__query_block}>
          <QueryBlock
            availableGroups={groupsList}
            queryType={queryType}
            index={queries.length + 1}
            queries={queries}
            eventChange={eventChange}
            groupBy={queryOptions.groupBy}
            groupAnalysis={queryOptions.group_analysis}
          />
        </div>
      );
    }

    return blockList;
  };

  const setGlobalFiltersOption = (filters) => {
    const opts = { ...queryOptions };
    opts.globalFilters = filters;
    setQueryOptions(opts);
  };

  const renderGlobalFilterBlock = () => {
    try {
      if (queryType === QUERY_TYPE_EVENT && queries.length < 1) {
        return null;
      }
      if (queryType === QUERY_TYPE_FUNNEL && queries.length < 2) {
        return null;
      }

      return (
        <ComposerBlock
          blockTitle='FILTER BY'
          isOpen={filterBlockOpen}
          showIcon
          onClick={() => setFilterBlockOpen(!filterBlockOpen)}
          extraClass='no-padding-l no-padding-r'
        >
          <div key={0} className='fa--query_block borderless no-padding '>
            <GLobalFilter
              filters={queryOptions.globalFilters}
              setGlobalFilters={setGlobalFiltersOption}
              groupName={queryOptions.group_analysis}
            />
          </div>
        </ComposerBlock>
      );
    } catch (err) {
      console.log(err);
    }
  };

  const groupByBlock = () => {
    try {
      if (queryType === QUERY_TYPE_EVENT && queries.length < 1) {
        return null;
      }
      if (queryType === QUERY_TYPE_FUNNEL && queries.length < 2) {
        return null;
      }

      return (
        <ComposerBlock
          blockTitle='BREAKDOWN'
          isOpen={groupBlockOpen}
          showIcon
          onClick={() => setGroupBlockOpen(!groupBlockOpen)}
          extraClass='no-padding-l no-padding-r'
        >
          <div key={0} className='fa--query_block borderless no-padding '>
            <GroupBlock groupName={queryOptions.group_analysis} />
          </div>
        </ComposerBlock>
      );
    } catch (err) {
      console.log(err);
    }
  };

  const setDateRange = (dates) => {
    const queryOptionsState = { ...queryOptions };
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
        to: queryOptionsState.date_range.to
      })[0];
      queryOptionsState.date_range.frequency = frequency;
      setQueryOptions(queryOptionsState);
    }
  };

  const handleRunQuery = useCallback(() => {
    if (queryType === QUERY_TYPE_EVENT) {
      runQuery(false);
    } else {
      runFunnelQuery(false);
    }
  }, [runFunnelQuery, runQuery, queryType]);

  const footer = () => {
    try {
      if (queryType === QUERY_TYPE_EVENT && queries.length < 1) {
        return null;
      }
      if (queryType === QUERY_TYPE_FUNNEL && queries.length < 2) {
        return null;
      }
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
              quarterPicker
              placement='topRight'
              buttonSize='large'
              range={{
                startDate: queryOptions.date_range.from,
                endDate: queryOptions.date_range.to
              }}
              onSelect={setDateRange}
            />
          ) : (
            <Button
              className='mr-2'
              size='large'
              type='default'
              onClick={() => setCollapse(false)}
            >
              <SVG name='arrowUp' size={20} extraClass='mr-1' />
              Collapse all
            </Button>
          )}
          <Button
            className='ml-2'
            size='large'
            type='primary'
            onClick={handleRunQuery}
          >
            Run Analysis
          </Button>
        </div>
      );
    } catch (err) {
      console.log(err);
    }
  };

  const renderEACrit = () => (
    <CriteriaSection queryCount={queries.length} queryType={queryType} />
  );

  const renderCriteria = () => {
    try {
      if (
        (queryType === QUERY_TYPE_EVENT && queries.length > 0) ||
        (queryType === QUERY_TYPE_FUNNEL && queries.length > 1)
      ) {
        return (
          <ComposerBlock
            blockTitle={
              queryType === QUERY_TYPE_FUNNEL ? 'FUNNEL CRITERIA' : 'CRITERIA'
            }
            isOpen={criterieaBlockOpen}
            showIcon
            onClick={() => {
              setCriterieaBlockOpen(!criterieaBlockOpen);
            }}
            extraClass='no-padding-l no-padding-r'
          >
            <div className={styles.criteria}>{renderEACrit()}</div>
          </ComposerBlock>
        );
      }
      return null;
    } catch (err) {
      console.log(err);
    }
  };

  const renderQueryList = () => {
    try {
      return (
        <ComposerBlock
          blockTitle={queryType === QUERY_TYPE_FUNNEL ? null : 'EVENTS'}
          isOpen={eventBlockOpen}
          showIcon
          onClick={() => setEventBlockOpen(!eventBlockOpen)}
          extraClass={`no-padding-l no-padding-r ${
            queryType === QUERY_TYPE_FUNNEL ? 'no-padding-t' : ''
          }`}
        >
          <ReactSortable
            list={queries}
            setList={(newQueriesState) => {
              if (!isEqual(queries, newQueriesState)) {
                const indexMapping = newQueriesState.map((elem) =>
                  queries.findIndex((q) => q.key === elem.key)
                );
                setQueries(newQueriesState);
                const newEventBreakdowns = eventBreakdowns.map((b) => {
                  const newEventIndex = indexMapping.findIndex(
                    (m) => m === b.eventIndex - 1
                  );
                  if (newEventIndex !== b.eventIndex - 1) {
                    return {
                      ...b,
                      eventIndex: newEventIndex + 1
                    };
                  }
                  return b;
                });
                if (!isEqual(newEventBreakdowns, eventBreakdowns)) {
                  dispatch(setEventGroupBy(newEventBreakdowns));
                }
              }
            }}
          >
            {queryList()}
          </ReactSortable>
        </ComposerBlock>
      );
    } catch (err) {
      console.log(err);
    }
  };

  return (
    <div className={styles.composer_body}>
      {queryType === QUERY_TYPE_FUNNEL && renderEventsConditionSection()}
      {queryType === QUERY_TYPE_FUNNEL && renderGroupSection()}
      {renderQueryList()}
      {renderGlobalFilterBlock()}
      {groupByBlock()}
      {renderCriteria()}
      {footer()}
    </div>
  );
}

const mapStateToProps = (state) => ({
  activeProject: state.global.active_project,
  groupOpts: state.groups.data,
  eventProperties: state.coreQuery.eventProperties
});

const mapDispatchToProps = (dispatch) =>
  bindActionCreators(
    {
      fetchGroups,
      fetchEventNames,
      getEventProperties,
      getUserProperties,
      getGroupProperties
    },
    dispatch
  );

export default connect(mapStateToProps, mapDispatchToProps)(QueryComposer);
