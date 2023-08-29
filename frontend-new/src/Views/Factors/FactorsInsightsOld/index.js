import React, { useState, useEffect } from 'react';
import Header from '../../AppLayout/Header';
import SearchBar from '../../../components/SearchBar';
import { Row, Col, Button, Spin } from 'antd';
import {
  fetchFactorsGoals,
  fetchFactorsModels,
  fetchGoalInsights,
  setGoalInsight,
  saveGoalInsightRules,
  saveGoalInsightModel,
  fetchFactorsTrackedEvents,
  fetchFactorsTrackedUserProperties
} from 'Reducers/factors';
import {
  fetchEventNames,
  getUserPropertiesV2
} from 'Reducers/coreQuery/middleware';
import { connect, useSelector, useDispatch } from 'react-redux';
import { fetchProjectAgents } from 'Reducers/agentActions';
import _, { isEmpty } from 'lodash';
import { useHistory } from 'react-router-dom';
import { Text, FaErrorComp, FaErrorLog } from 'factorsComponents';
import { ErrorBoundary } from 'react-error-boundary';
import ResultsTableL1 from './Components/ResultsTableL1';
import ExplainQueryBuilderOld from '../ExplainQueryBuilderOld';
import HeaderContents from './HeaderContents';
import { SHOW_ANALYTICS_RESULT } from 'Reducers/types';
import matchEventName from './Utils/MatchEventNames';

const Factors = ({
  fetchFactorsGoals,
  activeProject,
  goals,
  agents,
  fetchProjectAgents,
  fetchEventNames,
  fetchFactorsModels,
  fetchGoalInsights,
  fetchFactorsTrackedEvents,
  fetchFactorsTrackedUserProperties,
  getUserPropertiesV2,
  goalInsights,
  saveGoalInsightRules,
  saveGoalInsightModel,
  setGoalInsight,
  eventPropNames,
  userPropNames
}) => {
  const [loadingTable, SetLoadingTable] = useState(true);
  const [fetchingIngishts, SetfetchingIngishts] = useState(false);
  const [showConfigureDPModal, setConfigureDPModal] = useState(false);
  const [showGoalDrawer, setGoalDrawer] = useState(false);
  const [dataSource, setdataSource] = useState(null);
  const history = useHistory();
  const dispatch = useDispatch();

  useEffect(() => {
    dispatch({ type: SHOW_ANALYTICS_RESULT, payload: true });
    const getData1 = async () => {
      await fetchProjectAgents(activeProject.id);
      await fetchFactorsGoals(activeProject.id);
      await fetchEventNames(activeProject.id);
      await fetchFactorsModels(activeProject.id);
      await fetchFactorsTrackedEvents(activeProject.id);
      await fetchFactorsTrackedUserProperties(activeProject.id);
      await getUserPropertiesV2(activeProject.id, 'events');
    };
    getData1();

    return () => {
      dispatch({ type: SHOW_ANALYTICS_RESULT, payload: false });
    };
  }, [activeProject]);

  const smoothScroll = (element) => {
    document.querySelector(element).scrollIntoView({
      behavior: 'smooth'
    });
  };

  useEffect(() => {
    if (goalInsights) {
      setTimeout(() => {
        smoothScroll('#explain-builder--footer');
      }, 200);
    }
    return () => {
      saveGoalInsightRules(null);
      saveGoalInsightModel(null);
      setGoalInsight(null);
    };
  }, []);

  const handleCancel = () => {
    setConfigureDPModal(false);
  };
  const explainMatchEventName = (
    eventName,
    stringOnly = false,
    color = 'grey'
  ) => {
    return matchEventName(
      eventName,
      eventPropNames,
      userPropNames,
      stringOnly,
      color
    );
  };

  return (
    <>
      <ErrorBoundary
        fallback={
          <FaErrorComp
            size={'medium'}
            title={'Explain Error '}
            subtitle={
              'We are facing trouble loading Explain. Drop us a message on the in-app chat.'
            }
          />
        }
        onError={FaErrorLog}
      >
        {fetchingIngishts ? (
          <Spin size={'large'} className={'fa-page-loader'} />
        ) : (
          <>
            <HeaderContents />
            <div className={'fa-container'}>
              <div className={'mt-24'}>
                <ExplainQueryBuilderOld />
                <div id='fa-explain-results--container'>
                  {!_.isEmpty(goalInsights?.insights) && (
                    <ResultsTableL1
                      goalInsights={goalInsights}
                      explainMatchEventName={explainMatchEventName}
                    />
                  )}
                </div>
              </div>
            </div>
          </>
        )}
      </ErrorBoundary>
    </>
  );
};
const mapStateToProps = (state) => {
  return {
    activeProject: state.global.active_project,
    goals: state.factors.goals,
    agents: state.agent.agents,
    factors_models: state.factors.factors_models,
    goalInsights: state.factors.goal_insights,
    eventPropNames: state.coreQuery.eventPropNames,
    userPropNames: state.coreQuery.userPropNames
  };
};
export default connect(mapStateToProps, {
  fetchFactorsGoals,
  setGoalInsight,
  saveGoalInsightModel,
  fetchFactorsTrackedEvents,
  fetchFactorsTrackedUserProperties,
  fetchProjectAgents,
  saveGoalInsightRules,
  fetchGoalInsights,
  fetchFactorsModels,
  fetchEventNames,
  getUserPropertiesV2
})(Factors);
