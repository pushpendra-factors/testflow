import React, { useState, useEffect, useCallback, useMemo } from 'react';
import { connect, useDispatch, useSelector } from 'react-redux';
import { bindActionCreators } from 'redux';
import { Button, Dropdown, Menu, Tooltip } from 'antd';
import {
  fetchEventNames,
  getUserPropertiesV2,
  getGroupProperties,
  getEventPropertiesV2
} from 'Reducers/coreQuery/middleware';
import { SVG, Text } from 'factorsComponents';
import styles from './index.module.scss';
import QueryBlock from './QueryBlock';
import {
  QUERY_TYPE_FUNNEL,
  QUERY_TYPE_EVENT,
  QUERY_OPTIONS_DEFAULT_VALUE,
  INITIAL_SESSION_ANALYTICS_SEQ
} from 'Utils/constants';
import { fetchGroups } from 'Reducers/coreQuery/services';
import { setShowCriteria } from 'Reducers/analyticsQuery';
import { generateRandomKey } from 'Utils/global';
import { deleteGroupByForEvent } from 'Reducers/coreQuery/middleware';
import { DefaultDateRangeFormat } from 'Views/CoreQuery/utils';
import { findGroupNameUsingOptionValue } from './utils';

function EventQueryBlock({
  selEventName,
  setEventName,
  fetchGroups,
  fetchEventNames,
  getUserPropertiesV2,
  getGroupProperties,
  getEventPropertiesV2,
  activeProject,
  groupOpts,
  eventPropertiesV2,
  setShowCriteria
}) {
  const [queries, setQueries] = useState([]);
  const [queryType, setQueryType] = useState(QUERY_TYPE_EVENT);
  const { eventOptions } = useSelector((state) => state.coreQuery);
  const [queryOptions, setQueryOptions] = useState({
    ...QUERY_OPTIONS_DEFAULT_VALUE,
    session_analytics_seq: INITIAL_SESSION_ANALYTICS_SEQ,
    date_range: { ...DefaultDateRangeFormat }
  });

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
    if (selEventName) {
      setQueries([
        {
          label: selEventName,
          group: findGroupNameUsingOptionValue(eventOptions, selEventName)
        }
      ]);
    }
  }, [selEventName]);

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

  useEffect(() => {
    setEventName(queries?.[0]?.label);
  }, [queries]);

  const eventChange = useCallback(
    (newEvent, index, changeType = 'add', flag = null) => {
      const queryupdated = [...queries];
      if (queryupdated[index]) {
        if (changeType === 'add') {
          if (
            JSON.stringify(queryupdated[index]) !== JSON.stringify(newEvent)
          ) {
            deleteGroupByForEvent(newEvent, index);
          }
          queryupdated[index] = newEvent;
        } else if (changeType === 'filters_updated') {
          // dont remove group by if filter is changed
          queryupdated[index] = newEvent;
        } else {
          deleteGroupByForEvent(newEvent, index);
          queryupdated.splice(index, 1);
        }
      } else {
        if (flag) {
          Object.assign(newEvent, { pageViewVal: flag });
        }
        queryupdated.push(newEvent);
      }
      setQueries(
        queryupdated.map((q) => {
          return {
            ...q,
            key: q.key || generateRandomKey()
          };
        })
      );
    },
    [queries, deleteGroupByForEvent]
  );

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

    if (queries.length < 1) {
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

  return <div className={styles.composer_body}>{queryList()}</div>;
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

export default connect(mapStateToProps, mapDispatchToProps)(EventQueryBlock);
