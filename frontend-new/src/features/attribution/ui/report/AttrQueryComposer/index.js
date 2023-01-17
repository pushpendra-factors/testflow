import React, { useState, useEffect } from 'react';
import { connect, useSelector, useDispatch } from 'react-redux';
import { bindActionCreators } from 'redux';
import styles from './index.module.scss';
import { SVG, Text } from 'Components/factorsComponents';
import ConversionGoalBlock from './ConversionGoalBlock';
import FaDatepicker from 'Components/FaDatepicker';
import ComposerBlock from 'Components/QueryCommons/ComposerBlock';
import GlobalFilterBlock from 'Components/KPIComposer/GlobalFilter/GlobalFilterBlock';
import { PropTextFormat } from 'Utils/dataFormatter';
import { INITIALIZE_GROUPBY } from 'Reducers/coreQuery/actions';
import FaSelect from 'Components/FaSelect';

import {
  fetchEventNames,
  getUserProperties,
  getEventProperties,
  getCampaignConfigData
} from 'Reducers/coreQuery/middleware';
import {
  setGoalEvent,
  setTouchPoint,
  setModels,
  setWindow,
  setLinkedEvents,
  setAttrDateRange
} from 'Attribution/state/actions';
import { Button, Tooltip } from 'antd';
import MarkTouchpointBlock from './MarkTouchpointBlock';
import AttributionOptions from './AttributionOptions';
import LinkedEventsBlock from './LinkedEventsBlock';
import { SET_ATTR_QUERIES } from 'Attribution/state/action.constants';

const AttrQueryComposer = ({
  activeProject,
  fetchEventNames,
  getEventProperties,
  userProperties,
  eventProperties,
  runAttributionQuery,
  eventGoal,
  setGoalEvent,
  touchPoint,
  setTouchPoint,
  models,
  setModels,
  window,
  setWindow,
  linkedEvents,
  setLinkedEvents,
  setAttrDateRange,
  dateRange,
  collapse = false,
  setCollapse,
  queryOptions,
  setQueryOptions
}) => {
  const [linkEvExpansion, setLinkEvExpansion] = useState(true);
  const [convGblockOpen, setConvGblockOpen] = useState(true);
  const [tchPointblockOpen, setTchPointblockOpen] = useState(true);
  const [criteriablockOpen, setCriteriablockOpen] = useState(true);
  const [filterResultsOpen, setFilterResultsOpen] = useState(true);
  const [isGroupDDVisible, setGroupDDVisible] = useState(false);

  const { attrQueries } = useSelector((state) => state.attributionDashboard);

  const [queries, setQueries] = useState(attrQueries);

  const dispatch = useDispatch();

  const enabledGroups = () => {
    let groups = [
      ['Users', 'users'],
      ['Hubspot Deals', 'hubspot_deals'],
      ['Salesforce Opportunity', 'salesforce_opportunities']
    ];
    return groups;
  };
  useEffect(() => {
    if (activeProject && activeProject.id) {
      getCampaignConfigData(activeProject.id, 'all_ads');
      fetchEventNames(activeProject.id);
      if (!userProperties.length) {
        getUserProperties(activeProject.id, 'analysis');
      }
    }
  }, [activeProject]);

  useEffect(() => {
    if (!eventProperties[eventGoal?.label]) {
      getEventProperties(activeProject.id, eventGoal.label);
    }
  }, [eventGoal]);

  useEffect(() => {
    linkedEvents.forEach((ev, index) => {
      if (!eventProperties[ev.label]) {
        getEventProperties(activeProject.id, ev.label);
      }
    });
  }, [linkedEvents]);

  const goalChange = (eventGoal) => {
    setGoalEvent(eventGoal);
    dispatch({
      type: SET_ATTR_QUERIES,
      payload: []
    });
  };

  const linkEventChange = (linkEvent, index) => {
    const currLinkedEvs = [...linkedEvents];
    if (index === undefined || index < 0) {
      currLinkedEvs.push(linkEvent);
    } else {
      currLinkedEvs[index] = linkEvent;
    }
    setLinkedEvents(currLinkedEvs);
  };

  const goalDel = () => {
    setGoalEvent({});
  };

  const linkEventDel = (index) => {
    const currLinkedEvs = linkedEvents.filter((ev, i) => i !== index);
    setLinkedEvents(currLinkedEvs);
  };

  const setToQueries = (val, index) => {
    const qs = [...queries];
    if (qs[index]) {
      qs[index] = val;
    } else {
      qs.push(val);
    }
    dispatch({
      type: SET_ATTR_QUERIES,
      payload: qs
    });
    setQueries(qs);
    setGoalEvent({ filters: [] });
  };

  const delQuery = (index) => {
    const qs = [...queries].filter((v, i) => i != index);
    setQueries(qs);
    dispatch({
      type: SET_ATTR_QUERIES,
      payload: qs
    });
  };

  const renderConversionBlock = () => {
    
      const qs = queries.map((query, index) => {
        return (
          <ConversionGoalBlock
            eventGoal={query}
            eventGoalChange={(val) => setToQueries(val, index)}
            delEvent={() => delQuery(index)}
            group_analysis={'all'}
            showDerivedKPI={false}
          ></ConversionGoalBlock>
        );
      });

      if (qs.length < 5) {
        qs.push(
          <ConversionGoalBlock
            eventGoalChange={(val) => setToQueries(val, -1)}
            group_analysis={'all'}
            showDerivedKPI={false}
          ></ConversionGoalBlock>
        );
      }

      return qs;
  };

  const renderMarkTouchpointBlock = () => {
    return (
      <MarkTouchpointBlock
        touchPoint={touchPoint}
        setTouchpoint={(tchPoint) => setTouchPoint(tchPoint)}
      ></MarkTouchpointBlock>
    );
  };

  const renderAttributionOptions = () => {
    return (
      <AttributionOptions
        models={models}
        setModelOpt={(val) => setModels(val)}
        window={window}
        setWindowOpt={(win) => setWindow(win)}
      ></AttributionOptions>
    );
  };

  const renderLinkedEvents = () => {
    const linkEventsList = [];
    if (linkedEvents && linkedEvents.length) {
      linkedEvents.forEach((ev, index) => {
        linkEventsList.push(
          <LinkedEventsBlock
            linkEvent={ev}
            linkEventChange={(ev) => linkEventChange(ev, index)}
            delLinkEvent={() => linkEventDel(index)}
          ></LinkedEventsBlock>
        );
      });
    }

    linkEventsList.push(
      <LinkedEventsBlock
        linkEventChange={(ev) => linkEventChange(ev, -1)}
      ></LinkedEventsBlock>
    );

    return linkEventsList;
  };

  const toggleLinkEvExpansion = () => {
    if (models.length > 1) return null;
    setLinkEvExpansion(!linkEvExpansion);
  };

  const handleRunQuery = () => {
    runAttributionQuery(false);
  };

  const setDateRange = (ranges) => {
    const dtRange = Object.assign({}, dateRange);
    if (ranges && ranges.startDate) {
      if (Array.isArray(ranges.startDate)) {
        dtRange.from = ranges.startDate[0];
        dtRange.to = ranges.startDate[1];
      } else {
        dtRange.from = ranges.startDate;
        dtRange.to = ranges.endDate;
      }
    }
    setAttrDateRange(dtRange);
  };

  const footer = () => {
    if (
      (!eventGoal || !eventGoal?.label?.length) &&
      (!queries || !queries.length)
    ) {
      return null;
    }

    return (
      <div
        className={`${
          !collapse ? styles.composer__footer : styles.composer_footer_right
        }`}
      >
        {!collapse ? (
          <FaDatepicker
            customPicker
            presetRange
            quarterPicker
            monthPicker
            buttonSize={`large`}
            className={`mr-2`}
            range={{
              startDate: dateRange.from,
              endDate: dateRange.to
            }}
            placement='topRight'
            onSelect={setDateRange}
          />
        ) : (
          <Button
            className={`mr-2`}
            size={'large'}
            type={'default'}
            onClick={() => setCollapse(false)}
          >
            <SVG name={`arrowUp`} size={20} extraClass={`mr-1`}></SVG>Collapse
            all
          </Button>
        )}
        <Button
          className={`ml-2`}
          size={'large'}
          type='primary'
          onClick={handleRunQuery}
        >
          Run Analysis
        </Button>
      </div>
    );
  };

  const setGroupAnalysis = (group) => {
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
    setQueryOptions(opts);
  };

  const onGroupSelect = (val) => {
    setGroupAnalysis(val);
    setGroupDDVisible(false);
  };

  const selectGroup = () => {
    return (
      <div className={`${styles.groupsection_dropdown}`}>
        {isGroupDDVisible ? (
          <FaSelect
            extraClass={`${styles.groupsection_dropdown_menu}`}
            options={enabledGroups()}
            onClickOutside={() => setGroupDDVisible(false)}
            optionClick={(val) => onGroupSelect(val[1])}
          ></FaSelect>
        ) : null}
      </div>
    );
  };

  const triggerDropDown = () => {
    setGroupDDVisible(true);
  };

  const renderFilterResults = () => {
    return null;

    // return (
    //   <GlobalFilterBlock
    //     queries={queries}
    //     queryOptions={queryOptions}
    //     activeProject={activeProject}
    //     selectedMainCategory={eventGoal[0]}
    //     setQueryOptions={setQueryOptions}
    //   />
    // )
  }

  try {
    return (
      <div className={`${styles.composer}`}>
        {/*renderGroupSection()*/}
        <ComposerBlock
          blockTitle={'CONVERSION GOALS'}
          isOpen={convGblockOpen}
          showIcon={true}
          onClick={() => setConvGblockOpen(!convGblockOpen)}
          extraClass={`no-padding-l no-padding-r`}
        >
          {renderConversionBlock()}
        </ComposerBlock>

        {eventGoal?.label?.length || queries.length ? (
          <div
            className={`no-padding-l no-padding-r`}
          >
            {renderMarkTouchpointBlock()}
          </div>
        ) : null}

        {eventGoal?.label?.length || queries.length ? (
          <ComposerBlock
            blockTitle={'Attribution Model'}
            isOpen={criteriablockOpen}
            showIcon={true}
            onClick={() => setCriteriablockOpen(!criteriablockOpen)}
            extraClass={`no-padding-l no-padding-r`}
          >
            {renderAttributionOptions()}
          </ComposerBlock>
        ) : null}

      {/* {eventGoal?.label?.length || queries.length ? (
          <ComposerBlock
            blockTitle={'Filter Results'}
            isOpen={criteriablockOpen}
            showIcon={true}
            onClick={() => setFilterResultsOpen(!filterResultsOpen)}
            extraClass={`no-padding-l no-padding-r`}
          >
            {renderFilterResults()}
          </ComposerBlock>
        ) : null} */}


        {eventGoal?.label?.length || queries.length ? footer() : null}
      </div>
    );
  } catch (err) {
    console.log(err);
  }
};

const mapStateToProps = (state) => ({
  activeProject: state.global.active_project,
  userProperties: state.coreQuery.userProperties,
  eventProperties: state.coreQuery.eventProperties,
  eventGoal: state.attributionDashboard.eventGoal,
  touchPoint: state.attributionDashboard.touchpoint,
  models: state.attributionDashboard.models,
  window: state.attributionDashboard.window,
  linkedEvents: state.attributionDashboard.linkedEvents,
  dateRange: state.attributionDashboard.attr_dateRange
});

const mapDispatchToProps = (dispatch) =>
  bindActionCreators(
    {
      fetchEventNames,
      getEventProperties,
      getUserProperties,
      getCampaignConfigData,
      setGoalEvent,
      setTouchPoint,
      setAttrDateRange,
      setModels,
      setWindow,
      setLinkedEvents
    },
    dispatch
  );

export default connect(mapStateToProps, mapDispatchToProps)(AttrQueryComposer);
