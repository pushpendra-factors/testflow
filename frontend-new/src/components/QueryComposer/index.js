import React, { useState, useEffect, useCallback, useMemo } from 'react';
import { connect, useDispatch, useSelector } from 'react-redux';
import { bindActionCreators } from 'redux';
import { Button, Dropdown, Menu, Tooltip } from 'antd';
import {
  fetchEventNames,
  getUserPropertiesV2,
  getGroupProperties,
  getEventPropertiesV2
} from '../../reducers/coreQuery/middleware';
import { SVG, Text } from '../factorsComponents';
import styles from './index.module.scss';
import QueryBlock from './QueryBlock';
import GroupBlock from './GroupBlock';
import {
  QUERY_TYPE_FUNNEL,
  QUERY_TYPE_EVENT,
  FunnelEventsConditionMap,
  RevFunnelEventsConditionMap,
  TOTAL_EVENTS_CRITERIA,
  TOTAL_USERS_CRITERIA
} from 'Utils/constants';
import FaDatepicker from '../FaDatepicker';
import ComposerBlock from '../QueryCommons/ComposerBlock';
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
import GlobalFilter from '../GlobalFilter';
import FunnelsConversionDurationBlock from './FunnelsConversionDurationBlock/FunnelsConversionDurationBlock';
import { setShowCriteria } from 'Reducers/analyticsQuery';
import CriteriaSection from './CriteriaSection';

function QueryComposer({
  queries = [],
  setQueries,
  runQuery,
  eventChange,
  queryType,
  fetchGroups,
  fetchEventNames,
  getUserPropertiesV2,
  getGroupProperties,
  getEventPropertiesV2,
  activeProject,
  groupOpts,
  eventPropertiesV2,
  queryOptions,
  setQueryOptions,
  runFunnelQuery,
  collapse = false,
  setCollapse,
  setShowCriteria
}) {
  const [filterBlockOpen, setFilterBlockOpen] = useState(true);
  const [groupBlockOpen, setGroupBlockOpen] = useState(true);
  const [criteriaBlockOpen, setCriteriaBlockOpen] = useState(true);
  const [eventBlockOpen, setEventBlockOpen] = useState(true);
  const [isOrderDDVisible, setOrderDDVisible] = useState(false);
  const eventBreakdowns = useSelector((state) => state.coreQuery.groupBy.event);
  const criteria = useSelector((state) => state.analyticsQuery.show_criteria);

  const dispatch = useDispatch();

  useEffect(() => {
    fetchGroups(activeProject?.id);
  }, [activeProject]);

  const groupsList = useMemo(() => {
    let groups = [['Users', 'users']];
    if (queryType === QUERY_TYPE_EVENT) {
      groups.unshift(['Events', 'events']);
    }
    Object.entries(groupOpts || {}).forEach(([group_name, display_name]) => {
      groups.push([display_name, group_name]);
    });
    return groups;
  }, [groupOpts]);

  useEffect(() => {
    if (activeProject && activeProject.id) {
      getUserPropertiesV2(activeProject.id, queryType);
    }
  }, [activeProject, fetchEventNames, getUserPropertiesV2, queryType]);

  useEffect(() => {
    if (queryOptions.group_analysis === 'users') return;
    getGroupProperties(activeProject.id, queryOptions.group_analysis);
  }, [activeProject.id, queryOptions.group_analysis]);

  useEffect(() => {
    queries.forEach((ev) => {
      if (!eventPropertiesV2[ev.label]) {
        getEventPropertiesV2(activeProject.id, ev.label);
      }
    });
  }, [activeProject?.id, eventPropertiesV2, getEventPropertiesV2, queries]);

  const setEventsCondition = (condition) => {
    setQueryOptions((prevOptions) => ({
      ...prevOptions,
      events_condition: condition
    }));
  };

  const onOrderChange = (value) => {
    setEventsCondition(value);
    setOrderDDVisible(false);
  };

  const selectEventsCondition = () => (
    <div className={`${styles.toplevel_select_dropdown}`}>
      {isOrderDDVisible ? (
        <FaSelect
          options={Object.entries(RevFunnelEventsConditionMap)}
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
    if (!['users', 'events'].includes(group)) {
      getGroupProperties(activeProject.id, group);
    }

    const criteria =
      group === 'events' ? TOTAL_EVENTS_CRITERIA : TOTAL_USERS_CRITERIA;
    setShowCriteria(criteria);

    const opts = {
      ...queryOptions,
      group_analysis: group,
      globalFilters: []
    };

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
    if (value.key !== queryOptions.group_analysis) {
      setGroupAnalysis(value.key);
    }
  };

  const groupsMenuItems = groupsList.map((opt) => ({
    label: opt[0],
    key: opt[1],
    lineBreak: opt[1] === 'users'
  }));

  const groupsMenu = (
    <Menu className='dropdown-menu' onClick={onGroupChange}>
      {groupsMenuItems.map((item) => (
        <>
          <Menu.Item key={item.key} className='dropdown-menu-item'>
            <Text color='black' level={7} type='title' extraClass='mb-0'>
              {item.label}
            </Text>
          </Menu.Item>
          {item.lineBreak && <hr />}
        </>
      ))}
    </Menu>
  );

  const renderGroupSection = () => {
    try {
      const activeGroup = groupsList.find(
        ([_, groupName]) => groupName === queryOptions?.group_analysis
      )?.[0];
      return (
        <div className='flex items-center pt-4'>
          <Text type='title' level={6} weight='normal' extraClass='m-0 mr-3'>
            Analyse
          </Text>
          <Dropdown
            trigger={['click']}
            placement='bottomLeft'
            overlay={groupsMenu}
          >
            <div className='cursor-pointer flex items-center text-base font-semibold'>
              {activeGroup}
              <SVG name='caretDown' />
            </div>
          </Dropdown>
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
            <GlobalFilter
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

  const renderCriteria = () => {
    try {
      if (
        ((queryType === QUERY_TYPE_EVENT &&
          criteria === TOTAL_USERS_CRITERIA) ||
          queryType === QUERY_TYPE_FUNNEL) &&
        queries.length > 1
      ) {
        return (
          <ComposerBlock
            blockTitle={
              queryType === QUERY_TYPE_FUNNEL ? 'FUNNEL CRITERIA' : 'CRITERIA'
            }
            isOpen={criteriaBlockOpen}
            showIcon
            onClick={() => {
              setCriteriaBlockOpen(!criteriaBlockOpen);
            }}
            extraClass='no-padding-l no-padding-r'
          >
            <div className={styles.criteria}>
              {<CriteriaSection queryType={queryType} />}
            </div>
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
      {renderGroupSection()}
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
  eventPropertiesV2: state.coreQuery.eventPropertiesV2
});

const mapDispatchToProps = (dispatch) =>
  bindActionCreators(
    {
      setShowCriteria,
      fetchGroups,
      fetchEventNames,
      getEventPropertiesV2,
      getUserPropertiesV2,
      getGroupProperties
    },
    dispatch
  );

export default connect(mapStateToProps, mapDispatchToProps)(QueryComposer);
