import React, { useState, useEffect } from 'react';
import Header from '../../AppLayout/Header';
import SearchBar from '../../../components/SearchBar';
import { Row, Col, Button, Spin, Empty } from 'antd';
import {
  fetchSavedExplainGoals,
  fetchFactorsModels,
  setGoalInsight,
  saveGoalInsightRules,
  saveGoalInsightModel,
  fetchFactorsTrackedEvents,
  fetchFactorsTrackedUserProperties,
  setActiveExplainQuery
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
import ExplainQueryBuilder from '../ExplainQueryBuilder';
import HeaderContents from './HeaderContents';
import { SHOW_ANALYTICS_RESULT } from 'Reducers/types';
import matchEventName from './Utils/MatchEventNames';

const Factors = ({
  fetchSavedExplainGoals,
  activeProject,
  goals,
  agents,
  fetchProjectAgents,
  fetchEventNames,
  fetchFactorsModels,
  fetchFactorsTrackedEvents,
  fetchFactorsTrackedUserProperties,
  getUserPropertiesV2,
  goalInsights,
  saveGoalInsightRules,
  saveGoalInsightModel,
  setGoalInsight,
  eventPropNames,
  userPropNames,
  activeQuery,
  setActiveExplainQuery
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
      await fetchSavedExplainGoals(activeProject.id);
      await fetchEventNames(activeProject.id);
      await fetchFactorsModels(activeProject.id);
      await fetchFactorsTrackedEvents(activeProject.id);
      await fetchFactorsTrackedUserProperties(activeProject.id);
      await getUserPropertiesV2(activeProject.id, 'events');
    };
    getData1();

    return () => { 
      dispatch({ type: SHOW_ANALYTICS_RESULT, payload: false }); 
      setActiveExplainQuery(false)
    };
  }, [activeProject]);

  const smoothScroll = (element) => {
    document.querySelector(element).scrollIntoView({
      behavior: 'smooth'
    });
  };

  useEffect(() => {
    if (goalInsights || activeQuery) {
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

  const DataBuildMessage = () => { 
    return (
      <div className='flex flex-col items-center justify-center mt-20'>
        <img
          style={{ maxWidth: '200px', height: 'auto' }}
          src='https://s3.amazonaws.com/www.factors.ai/assets/img/product/report-building.png'
        />
        <Text type={'title'} weight={'bold'} extraClass={'mt-4'} level={6}>
          Your report is being built
        </Text>
        <Text type={'title'} weight={'thin'} level={7}>
          This might take a while.
        </Text>
      </div>
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
            <HeaderContents activeQuery={activeQuery} />
            <div className={'fa-container'}>
              <div className={'mt-24 '}>
                <ExplainQueryBuilder />
                <div id='fa-explain-results--container' className='px-20'>
                  {/* {!_.isEmpty(goalInsights?.insights) && (
                    <ResultsTableL1
                      goalInsights={goalInsights}
                      explainMatchEventName={explainMatchEventName}
                    />
                  )}  */}

                  {!_.isEmpty(activeQuery) && _.isEmpty(goalInsights?.insights) ? (
                            activeQuery?.status == 'building' ||
                            activeQuery?.status == 'saved' ? (
                              <div className='mb-20'>
                                <DataBuildMessage />
                              </div>
                            ) : (
                              <div className='mb-20'>
                                <Empty /> 
                              </div>
                            )
                          ) : (
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
    userPropNames: state.coreQuery.userPropNames,
    activeQuery: state.factors.activeQuery,
  };
};
export default connect(mapStateToProps, {
  fetchSavedExplainGoals,
  setGoalInsight,
  saveGoalInsightModel,
  fetchFactorsTrackedEvents,
  fetchFactorsTrackedUserProperties,
  fetchProjectAgents,
  saveGoalInsightRules,
  fetchFactorsModels,
  fetchEventNames,
  getUserPropertiesV2,
  setActiveExplainQuery
})(Factors);
