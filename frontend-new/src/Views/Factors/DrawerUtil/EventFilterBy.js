import React, { useEffect, useState } from 'react';
import { Drawer, Button, Row, Col, Select, message } from 'antd';
import { SVG, Text } from 'factorsComponents';
import {
  fetchEventNames,
  getUserPropertiesV2,
  getEventPropertiesV2
} from 'Reducers/coreQuery/middleware';
import {
  fetchGoalInsights,
  fetchFactorsModels,
  saveGoalInsightRules,
  saveGoalInsightModel
} from 'Reducers/factors';
import { connect } from 'react-redux';
import { useHistory } from 'react-router-dom';
import FilterBlock from '../../../components/QueryComposer/FilterBlock';
import { fetchUserPropertyValues } from 'Reducers/coreQuery/services';
// import MomentTz from 'Components/MomentTz';
import moment from 'moment-timezone';

const EventFilterBy = (props) => {
  const [TrackedEventNames, SetTrackedEventNames] = useState([]);
  const [filterLoader, setfilterLoader] = useState(false);
  const [filters, setfilters] = useState([]);
  const [filterProps, setFilterProperties] = useState({
    event: {},
    user: {}
  });
  // const [filterDD, setFilterDD] = useState(false);

  const timeZone = localStorage.getItem('project_timeZone') || 'Asia/Kolkata';
  moment.tz.setDefault(timeZone);

  const readableTimstamp = (unixTime) => {
    return moment.unix(unixTime).utc().format('MMM DD, YYYY');
  };

  useEffect(() => {
    if (props.activeProject && props.activeProject.id) {
      props.getUserPropertiesV2(props.activeProject.id, 'channel');
    }
    if (props.tracked_events) {
      const fromatterTrackedEvents = props.tracked_events.map((item) => {
        return [item.name];
      });
      SetTrackedEventNames(fromatterTrackedEvents);
    }
  }, [
    props.activeProject,
    props.tracked_events,
    props.factors_models,
    props.goal_insights
  ]);

  useEffect(() => {
    if (!props.eventPropertiesV2[props.event]) {
      props.getEventPropertiesV2(props.activeProject.id, props.event);
    }
    setfilters([]);
  }, [props.event]);

  useEffect(() => {
    const assignFilterProps = Object.assign({}, filterProps);
    //removing numerical type for both events and user properties
    const eventPropertiesFiltered = {};
    if (props.event && props.eventPropertiesV2?.[props.event]) {
      for (const key in props.eventPropertiesV2[props.event]) {
        if (props.eventPropertiesV2[props.event].hasOwnProperty(key)) {
          eventPropertiesFiltered[key] = props.eventPropertiesV2[props.event][
            key
          ].filter((item) => item?.[2] == 'categorical');
        }
      }
    }
    const eventUserPropertiesFiltered = {};
    if (props.eventUserPropertiesV2) {
      for (const key in props.eventUserPropertiesV2) {
        if (props.eventUserPropertiesV2.hasOwnProperty(key)) {
          eventUserPropertiesFiltered[key] = props.eventUserPropertiesV2[
            key
          ].filter((item) => item?.[2] == 'categorical');
        }
      }
    }
    assignFilterProps.event = eventPropertiesFiltered;
    assignFilterProps.user = eventUserPropertiesFiltered;
    setFilterProperties(assignFilterProps);
  }, [props.eventUserPropertiesV2, props.eventPropertiesV2]);

  const delFilter = (index) => {
    const fltrs = filters.filter((v, i) => i !== index);
    setfilters(fltrs);
    props.setfiltersParent(fltrs);
  };

  const addFilter = (val) => {
    const filterState = [...filters];
    filterState.push(val);
    setfilters(filterState);
    props.setfiltersParent(filterState);
  };

  const closeFilter = () => {
    props.setEventFilterDD(false);
  };

  const renderFilterBlock = () => {
    if (filterProps) {
      const filtrs = [];

      filters.forEach((filt, id) => {
        filtrs.push(
          <div key={id} className={`mt-0 relative flex flex-grow w-full`}>
            <FilterBlock
              activeProject={props.activeProject}
              index={id}
              blockType={'event'}
              // filterType={'channel'}
              filter={filt}
              extraClass={'filter-block--row'}
              delBtnClass={'filter-block--delete--mini'}
              delIcon={`times`}
              deleteFilter={delFilter}
              event={{ label: props.event }}
              // typeProps={{channel: channel}}
              filterProps={filterProps}
              propsConstants={Object.keys(filterProps)}
            ></FilterBlock>
          </div>
        );
      });

      if (props.showEventFilterDD) {
        filtrs.push(
          <div
            key={filtrs.length}
            className={`mt-0 relative flex flex-grow w-full`}
          >
            <FilterBlock
              activeProject={props.activeProject}
              blockType={'event'}
              // extraClass={styles.filterSelect}
              extraClass={'filter-block--row'}
              delBtnClass={'filter-block--delete--mini'}
              // typeProps={{channel: channel}}
              filterProps={filterProps}
              propsConstants={Object.keys(filterProps)}
              insertFilter={addFilter}
              closeFilter={closeFilter}
              event={{ label: props.event }}
              operatorProps={{
                categorical: ['=', '!='],
                numerical: ['=', '<=', '>='],
                datetime: ['=']
              }}
            ></FilterBlock>
          </div>
        );
      }

      return <div className={`relative flex flex-col w-full`}>{filtrs}</div>;
    }
  };

  return (
    <div className={'relative flex flex-grow w-full'}>
      {renderFilterBlock()}
    </div>
  );
};

const mapStateToProps = (state) => {
  return {
    activeProject: state.global.active_project,
    GlobalEventNames: state.coreQuery?.eventOptions[0]?.values,
    factors_models: state.factors.factors_models,
    goal_insights: state.factors.goal_insights,
    tracked_events: state.factors.tracked_events,
    eventUserPropertiesV2: state.coreQuery.eventUserPropertiesV2,
    eventPropertiesV2: state.coreQuery.eventPropertiesV2
  };
};
export default connect(mapStateToProps, {
  fetchEventNames,
  fetchGoalInsights,
  fetchFactorsModels,
  saveGoalInsightRules,
  saveGoalInsightModel,
  getUserPropertiesV2,
  fetchUserPropertyValues,
  getEventPropertiesV2
})(EventFilterBy);
