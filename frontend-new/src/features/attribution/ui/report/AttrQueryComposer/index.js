import React, { useState, useEffect } from 'react';
import { connect, useSelector, useDispatch } from 'react-redux';
import { bindActionCreators } from 'redux';
import { SVG, Text } from 'Components/factorsComponents';
import FaDatepicker from 'Components/FaDatepicker';
import ComposerBlock from 'Components/QueryCommons/ComposerBlock';
import { INITIALIZE_GROUPBY } from 'Reducers/coreQuery/actions';
import FaSelect from 'Components/FaSelect';

import {
  fetchEventNames,
  getUserPropertiesV2,
  getEventPropertiesV2,
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
import { SET_ATTR_QUERIES } from 'Attribution/state/action.constants';
import MarkTouchpointBlock from './MarkTouchpointBlock';
import AttributionOptions from './AttributionOptions';
import LinkedEventsBlock from './LinkedEventsBlock';
import ConversionGoalBlock from './ConversionGoalBlock';
import styles from './index.module.scss';

const AttrQueryComposer = ({
  activeProject,
  fetchEventNames,
  getEventPropertiesV2,
  eventUserPropertiesV2,
  eventPropertiesV2,
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
  const [criteriablockOpen, setCriteriablockOpen] = useState(true);

  const { attrQueries } = useSelector((state) => state.attributionDashboard);

  const dispatch = useDispatch();

  useEffect(() => {
    if (activeProject && activeProject.id) {
      getCampaignConfigData(activeProject.id, 'all_ads');
      fetchEventNames(activeProject.id);
      if (!eventUserPropertiesV2.length) {
        getUserPropertiesV2(activeProject.id, 'analysis');
      }
    }
  }, [activeProject]);

  useEffect(() => {
    if (!eventPropertiesV2[eventGoal?.label]) {
      getEventPropertiesV2(activeProject.id, eventGoal.label);
    }
  }, [eventGoal]);

  useEffect(() => {
    linkedEvents.forEach((ev, index) => {
      if (!eventPropertiesV2[ev.label]) {
        getEventPropertiesV2(activeProject.id, ev.label);
      }
    });
  }, [linkedEvents]);

  const setToQueries = (val, index) => {
    const qs = [...attrQueries];
    if (qs[index]) {
      qs[index] = val;
    } else {
      qs.push(val);
    }
    dispatch({
      type: SET_ATTR_QUERIES,
      payload: qs
    });
    setGoalEvent({ filters: [] });
  };

  const delQuery = (index) => {
    const qs = [...attrQueries].filter((v, i) => i != index);
    dispatch({
      type: SET_ATTR_QUERIES,
      payload: qs
    });
  };

  const renderConversionBlock = () => {
    const qs = attrQueries.map((query, index) => (
      <ConversionGoalBlock
        eventGoal={query}
        eventGoalChange={(val) => setToQueries(val, index)}
        delEvent={() => delQuery(index)}
        group_analysis='all'
        showDerivedKPI={false}
      />
    ));

    if (qs.length < 5) {
      qs.push(
        <ConversionGoalBlock
          eventGoalChange={(val) => setToQueries(val, -1)}
          group_analysis='all'
          showDerivedKPI={false}
        />
      );
    }

    return qs;
  };

  const renderMarkTouchpointBlock = () => (
    <MarkTouchpointBlock
      touchPoint={touchPoint}
      setTouchpoint={(tchPoint) => setTouchPoint(tchPoint)}
    />
  );

  const renderAttributionOptions = () => (
    <AttributionOptions
      models={models}
      setModelOpt={(val) => setModels(val)}
      window={window}
      setWindowOpt={(win) => setWindow(win)}
    />
  );

  const toggleLinkEvExpansion = () => {
    if (models.length > 1) return null;
    setLinkEvExpansion(!linkEvExpansion);
  };

  const handleRunQuery = () => {
    runAttributionQuery(false);
  };

  const setDateRange = (ranges) => {
    const dtRange = { ...dateRange };
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
      (!attrQueries || !attrQueries.length)
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
            buttonSize='large'
            className='mr-2'
            range={{
              startDate: dateRange.from,
              endDate: dateRange.to
            }}
            placement='topRight'
            onSelect={setDateRange}
            withoutYesterday
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
  };

  const setGroupAnalysis = (group) => {
    const opts = { ...queryOptions };
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

  try {
    return (
      <div className={`${styles.composer}`}>
        {/* renderGroupSection() */}
        <ComposerBlock
          blockTitle='CONVERSION GOALS'
          isOpen={convGblockOpen}
          showIcon
          onClick={() => setConvGblockOpen(!convGblockOpen)}
          extraClass='no-padding-l no-padding-r'
        >
          {renderConversionBlock()}
        </ComposerBlock>

        {eventGoal?.label?.length || attrQueries.length ? (
          <div className='no-padding-l no-padding-r'>
            {renderMarkTouchpointBlock()}
          </div>
        ) : null}

        {eventGoal?.label?.length || attrQueries.length ? (
          <ComposerBlock
            blockTitle='Attribution Model'
            isOpen={criteriablockOpen}
            showIcon
            onClick={() => setCriteriablockOpen(!criteriablockOpen)}
            extraClass='no-padding-l no-padding-r'
          >
            {renderAttributionOptions()}
          </ComposerBlock>
        ) : null}

        {eventGoal?.label?.length || attrQueries.length ? footer() : null}
      </div>
    );
  } catch (err) {
    console.log(err);
  }
};

const mapStateToProps = (state) => ({
  activeProject: state.global.active_project,
  eventUserPropertiesV2: state.coreQuery.eventUserPropertiesV2,
  eventPropertiesV2: state.coreQuery.eventPropertiesV2,
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
      getEventPropertiesV2,
      getUserPropertiesV2,
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
