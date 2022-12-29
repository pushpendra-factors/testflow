import React, { useState, useEffect } from 'react';
import { connect, useSelector, useDispatch } from 'react-redux';
import { bindActionCreators } from 'redux';
import styles from './index.module.scss';
import { SVG, Text } from '../../components/factorsComponents';
import ConversionGoalBlock from './ConversionGoalBlock';
import FaDatepicker from '../../components/FaDatepicker';
import ComposerBlock from '../QueryCommons/ComposerBlock';
import { PropTextFormat } from '../../utils/dataFormatter';
import { INITIALIZE_GROUPBY } from 'Reducers/coreQuery/actions';
import FaSelect from '../FaSelect';

import {
  fetchEventNames,
  getUserProperties,
  getEventProperties,
  setGoalEvent,
  setTouchPoint,
  setModels,
  setWindow,
  setLinkedEvents,
  setAttrDateRange,
  getCampaignConfigData
} from '../../reducers/coreQuery/middleware';
import { Button, Tooltip } from 'antd';
import MarkTouchpointBlock from './MarkTouchpointBlock';
import AttributionOptions from './AttributionOptions';
import LinkedEventsBlock from './LinkedEventsBlock';
import { QUERY_TYPE_EVENT } from '../../utils/constants';
import { SET_ATTR_QUERIES } from '../../reducers/coreQuery/actions';

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
  const [isGroupDDVisible, setGroupDDVisible] = useState(false);

  const { attrQueries } = useSelector((state) => state.coreQuery);

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
    if (
      !queryOptions.group_analysis ||
      queryOptions.group_analysis === 'users'
    ) {
      if (eventGoal) {
        return (
          <ConversionGoalBlock
            eventGoal={eventGoal}
            eventGoalChange={goalChange}
            delEvent={goalDel}
            showDerivedKPI={false}
          ></ConversionGoalBlock>
        );
      } else {
        return <ConversionGoalBlock></ConversionGoalBlock>;
      }
    } else {
      const qs = queries.map((query, index) => {
        return (
          <ConversionGoalBlock
            eventGoal={query}
            eventGoalChange={(val) => setToQueries(val, index)}
            delEvent={() => delQuery(index)}
            group_analysis={queryOptions.group_analysis}
            showDerivedKPI={false}
          ></ConversionGoalBlock>
        );
      });

      if (qs.length < 5) {
        qs.push(
          <ConversionGoalBlock
            eventGoalChange={(val) => setToQueries(val, -1)}
            group_analysis={queryOptions.group_analysis}
            showDerivedKPI={false}
          ></ConversionGoalBlock>
        );
      }

      return qs;
    }
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

  const renderGroupSection = () => {
    try {
      return (
        <div className={`flex items-center pt-6`}>
          <Text
            type={'title'}
            level={6}
            weight={'normal'}
            extraClass={`m-0 mr-3`}
          >
            Analyse
          </Text>{' '}
          <div className={`${styles.groupsection}`}>
            <Tooltip title='Attribute at a User, Deal, or Opportunity level'>
              <Button
                className={`${styles.groupsection_button}`}
                type='text'
                onClick={triggerDropDown}
              >
                <div className={`flex items-center`}>
                  <Text
                    type={'title'}
                    level={6}
                    weight={'bold'}
                    extraClass={`m-0 mr-1`}
                  >
                    {PropTextFormat(queryOptions.group_analysis)}
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

  try {
    return (
      <div className={`${styles.composer}`}>
        {renderGroupSection()}
        <ComposerBlock
          blockTitle={'CONVERSION GOAL'}
          isOpen={convGblockOpen}
          showIcon={true}
          onClick={() => setConvGblockOpen(!convGblockOpen)}
          extraClass={`no-padding-l no-padding-r`}
        >
          {renderConversionBlock()}
        </ComposerBlock>

        {eventGoal?.label?.length || queries.length ? (
          <ComposerBlock
            blockTitle={'MARKETING TOUCHPOINTS'}
            isOpen={tchPointblockOpen}
            showIcon={true}
            onClick={() => setTchPointblockOpen(!tchPointblockOpen)}
            extraClass={`no-padding-l no-padding-r`}
          >
            {renderMarkTouchpointBlock()}
          </ComposerBlock>
        ) : null}

        {eventGoal?.label?.length || queries.length ? (
          <ComposerBlock
            blockTitle={'CRITERIA'}
            isOpen={criteriablockOpen}
            showIcon={true}
            onClick={() => setCriteriablockOpen(!criteriablockOpen)}
            extraClass={`no-padding-l no-padding-r`}
          >
            {renderAttributionOptions()}
          </ComposerBlock>
        ) : null}

        {eventGoal?.label?.length && (
          <ComposerBlock
            blockTitle={'LINKED EVENTS'}
            isOpen={linkEvExpansion}
            showIcon={true}
            onClick={() => toggleLinkEvExpansion()}
            extraClass={`no-padding-l no-padding-r`}
          >
            {linkEvExpansion && models.length <= 1 && renderLinkedEvents()}
          </ComposerBlock>
        )}

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
  eventGoal: state.coreQuery.eventGoal,
  touchPoint: state.coreQuery.touchpoint,
  models: state.coreQuery.models,
  window: state.coreQuery.window,
  linkedEvents: state.coreQuery.linkedEvents,
  dateRange: state.coreQuery.attr_dateRange
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
