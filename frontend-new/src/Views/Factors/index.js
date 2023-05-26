import React, { useState, useEffect } from 'react';
import { Row, Col, Button, Spin, Tag, Tabs, Alert } from 'antd';
import {
  fetchSavedExplainGoals,
  fetchFactorsGoals,
  fetchFactorsModels,
  fetchGoalInsights,
  saveGoalInsightRules,
  fetchFactorsTrackedEvents,
  fetchFactorsTrackedUserProperties
} from 'Reducers/factors';
import {
  fetchEventNames,
  getUserProperties
} from 'Reducers/coreQuery/middleware';
import { connect, useSelector } from 'react-redux';
import { fetchProjectAgents } from 'Reducers/agentActions';
import _ from 'lodash';
import SavedGoals from './savedGoals';
// import SavedGoalsOld from './savedGoalsOld';
import { Text, SVG, FaErrorComp, FaErrorLog } from 'factorsComponents';
import { ErrorBoundary } from 'react-error-boundary';
import { useHistory } from 'react-router-dom';
import {
  fetchDemoProject,
  getHubspotContact,
  fetchProjectSettingsV1,
  fetchProjectSettings,
  fetchMarketoIntegration,
  fetchBingAdsIntegration
} from '../../reducers/global';
import NewProject from '../Settings/SetupAssist/Modals/NewProject';
import userflow from 'userflow.js';
import ExplainBeforeIntegration from './ExplainBeforeIntegration';

// const whiteListedAccounts = [
//   'baliga@factors.ai',
//   'vinith@factors.ai',
//   'sonali@factors.ai',
//   'praveenr@factors.ai',
//   'solutions@factors.ai'
// ];

const ExplainTypeList = [
  {
    title: 'Conversions Explorer',
    desc: 'How can I improve conversions between any two milestones of a user journey?',
    icon: 'organisation',
    active: true
  },
  {
    title: 'Behavior Explorer',
    desc: 'What do users do before and after visiting the Pricing Page?',
    icon: '',
    active: false
  },
  {
    title: 'Segment Explorer ',
    desc: 'What are the differences between users in California and London?',
    icon: '',
    active: false
  }
];

const Factors = ({
  fetchSavedExplainGoals,
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
  getUserProperties,
  fetchDemoProject,
  getHubspotContact,
  fetchProjectSettingsV1,
  fetchProjectSettings,
  fetchMarketoIntegration,
  fetchBingAdsIntegration,
  currentProjectSettings
}) => {
  const [fetchingIngishts, SetfetchingIngishts] = useState(false);
  const [showConfigureDPModal, setConfigureDPModal] = useState(false);
  const history = useHistory();
  const [loading, setLoading] = useState(true);

  const [demoProjectId, setDemoProjectId] = useState(null);
  const [tabID, setTabID] = useState(1);
  const [showProjectModal, setShowProjectModal] = useState(false);
  const { projects } = useSelector((state) => state.global);
  const integration = useSelector(
    (state) => state.global.currentProjectSettings
  );
  const integrationV1 = useSelector((state) => state.global.projectSettingsV1);
  const { bingAds, marketo } = useSelector((state) => state.global);
  const { dashboards } = useSelector((state) => state.dashboard);

  useEffect(() => {
    fetchDemoProject()
      .then((res) => {
        setDemoProjectId(res.data[0]);
      })
      .catch((err) => {
        console.log(err.data.error);
      });
  }, [activeProject]);

  useEffect(() => {
    setTimeout(() => {
      setLoading(false);
    }, 1000);
  }, [activeProject]);

  const handleTour = () => {
    history.push('/');
    userflow.start('c162ed75-0983-41f3-ae56-8aedd7dbbfbd');
  };

  useEffect(() => {
    fetchProjectSettingsV1(activeProject.id);
    fetchProjectSettings(activeProject.id);
    if (_.isEmpty(dashboards?.data)) {
      fetchBingAdsIntegration(activeProject?.id);
      fetchMarketoIntegration(activeProject?.id);
    }
  }, [activeProject]);

  const isIntegrationEnabled =
    integration?.int_segment ||
    integration?.int_adwords_enabled_agent_uuid ||
    integration?.int_linkedin_agent_uuid ||
    integration?.int_facebook_user_id ||
    integration?.int_hubspot ||
    integration?.int_salesforce_enabled_agent_uuid ||
    integration?.int_drift ||
    integration?.int_google_organic_enabled_agent_uuid ||
    integration?.int_clear_bit ||
    integrationV1?.int_completed ||
    bingAds?.accounts ||
    marketo?.status ||
    integrationV1?.int_slack ||
    integration?.lead_squared_config !== null ||
    integration?.int_client_six_signal_key ||
    integration?.int_factors_six_signal_key ||
    integration?.int_rudderstack;

  useEffect(() => {
    const getData1 = async () => {
      await fetchFactorsGoals(activeProject.id);
      await fetchEventNames(activeProject.id);
      await fetchFactorsModels(activeProject.id);
      await fetchFactorsTrackedEvents(activeProject.id);
      await fetchFactorsTrackedUserProperties(activeProject.id);
      await getUserProperties(activeProject.id, 'events');
      await fetchProjectAgents(activeProject.id);
      // await fetchSavedExplainGoals(activeProject.id);
    };
    getData1();
  }, []);

  const onChangeTab = (key) => {
    setTabID(key);
    if (key == 1) {
      fetchFactorsGoals(activeProject.id);
    }
  };

  const ExplainDescNew = () => {
    return (
      <Row gutter={[24, 24]}>
        <Col span={24}>
          <div className='flex items-center justify-between'>
            <Text type={'title'} level={6} extraClass={'m-0 mt-2'} color={'grey'}>
              Investigate the impact of various user segments and their behaviors
              on your marketing efforts.
            </Text>
            <Button type="primary" size='large' onClick={() => history.push('/explainV2/insights')}>Create New</Button>
          </div>
        </Col>
      </Row>
    )

  }
  const ExplainCards = () => {
    return (
      <Row gutter={[24, 24]}>
        <Col span={24}>
          <Text type={'title'} level={6} extraClass={'m-0 mt-2'} color={'grey'}>
            Investigate the impact of various user segments and their behaviors
            on your marketing efforts.
          </Text>
        </Col>
        <Col span={24}>
          <div className={`flex items-stretch justify-between mb-6`}>
            {ExplainTypeList?.map((item) => {
              return (
                <div
                  onClick={
                    item.active
                      ? () => {
                        if (tabID == 1) {
                          history.push('/explain/insights');
                        } else {
                          history.push('/explainV2/insights');
                        }
                      }
                      : null
                  }
                  className={`relative inline-flex items-stretch justify-start border-radius--sm border--thin-2 cursor-pointer mr-6 ${item.active
                    ? 'cursor-pointer'
                    : 'fa-template--card cursor-not-allowed'
                    }`}
                >
                  <div className='px-6 py-4 flex flex-col items-center justify-center background-color--brand-color-1'>
                    <SVG
                      name={item?.icon ? item.icon : 'organisation'}
                      size={32}
                      color={'blue'}
                      extraClass={'mr-2'}
                    />
                  </div>
                  <div className='px-4 py-4 flex flex-col items-start justify-start'>
                    {!item.active && (
                      <Tag color='red' className={'fai--custom-card--badge'}>
                        {' '}
                        Coming Soon{' '}
                      </Tag>
                    )}
                    <Text
                      type={'title'}
                      level={7}
                      weight={'bold'}
                      extraClass={'m-0'}
                    >
                      {item.title}
                    </Text>
                    <Text
                      type={'title'}
                      level={8}
                      color={'grey'}
                      extraClass={'m-0 mb-2'}
                    >
                      {item.desc}
                    </Text>
                  </div>
                </div>
              );
            })}
          </div>
        </Col>
      </Row>
    );
  };

  if (loading) {
    return (
      <div className='flex justify-center items-center w-full h-64'>
        <Spin size='large' />
      </div>
    );
  }

  if (isIntegrationEnabled || activeProject.id === demoProjectId) {
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
          {activeProject.id === demoProjectId ? (
            <div className={'rounded-lg border-2 h-20 mx-20'}>
              <Row justify={'space-between'} className={'m-0 p-3'}>
                <Col span={projects.length === 1 ? 12 : 18}>
                  <img
                    alt='Welcome'
                    src='assets/icons/welcome.svg'
                    style={{ float: 'left', marginRight: '20px' }}
                  />
                  <Text
                    type={'title'}
                    level={6}
                    weight={'bold'}
                    extraClass={'m-0'}
                  >
                    Welcome! You just entered a Factors demo project
                  </Text>
                  {projects.length === 1 ? (
                    <Text type={'title'} level={7} extraClass={'m-0'}>
                      These reports have been built with a sample dataset. Use
                      this to start exploring!
                    </Text>
                  ) : (
                    <Text type={'title'} level={7} extraClass={'m-0'}>
                      To jump back into your Factors project, click on your
                      account card on the{' '}
                      <span className={'font-bold'}>top right</span> of the
                      screen.
                    </Text>
                  )}
                </Col>
                <Col className={'mr-2 mt-2'}>
                  {projects.length === 1 ? (
                    <Button
                      type={'default'}
                      style={{
                        background: 'white',
                        border: '1px solid #E7E9ED',
                        height: '40px'
                      }}
                      className={'m-0 mr-2'}
                      onClick={() => setShowProjectModal(true)}
                    >
                      Set up my own Factors project
                    </Button>
                  ) : null}

                  <Button
                    type={'link'}
                    style={{
                      background: 'white',
                      // border: '1px solid #E7E9ED',
                      height: '40px'
                    }}
                    className={'m-0 mr-2'}
                    onClick={() => handleTour()}
                  >
                    Take the tour{' '}
                    <SVG
                      name={'Arrowright'}
                      size={16}
                      extraClass={'ml-1'}
                      color={'blue'}
                    />
                  </Button>
                </Col>
              </Row>
            </div>
          ) : null}

          {fetchingIngishts ? (
            <Spin size={'large'} className={'fa-page-loader'} />
          ) : (
            <>
              {/* <FaHeader>
                <SearchBar />
              </FaHeader> */}

              <div className={'fa-container'}>
                <Row gutter={[24, 24]} justify='center'>
                  <Col span={20}>
                    <Row gutter={[24, 24]}>





                      <Col span={24}>
                        {/* <Col span={24}>
                        <Alert description="🎉  Explain is faster and better now." type="warning" closable />
                      </Col> */}

                        <div className='flex justify-between items-center'>
                          <div className='flex flex-col'>
                            <Text
                              type={'title'}
                              level={3}
                              weight={'bold'}
                              extraClass={'m-0'}
                            >
                              Explain
                            </Text>
                            <Text
                              type={'title'}
                              level={6}
                              extraClass={'m-0 mt-2'}
                              color={'grey'}
                            >
                              Investigate the impact of various user segments and their behaviors on your marketing efforts.

                            </Text>
                          </div>
                          <Button
                            type='primary'
                            size='large'
                            onClick={() => history.push('/explainV2/insights')}
                          > Create New
                          </Button>
                        </div>

                      </Col>

                      <Col span={24}>
                      <SavedGoals
                                  SetfetchingIngishts={SetfetchingIngishts}
                                />
                      </Col>
                    </Row>
                  </Col>
                </Row>
              </div>
            </>
          )}

          {/* create project modal */}
          <NewProject
            visible={showProjectModal}
            handleCancel={() => setShowProjectModal(false)}
          />
        </ErrorBoundary>
      </>
    );
  } else {
    return (
      <>
        <ExplainBeforeIntegration />
      </>
    );
  }
};
const mapStateToProps = (state) => {
  return {
    activeProject: state.global.active_project,
    goals: state.factors.goals,
    agents: state.agent.agents,
    factors_models: state.factors.factors_models,
    currentProjectSettings: state.global?.currentProjectSettings
  };
};
export default connect(mapStateToProps, {
  fetchSavedExplainGoals,
  fetchFactorsGoals,
  fetchFactorsTrackedEvents,
  fetchFactorsTrackedUserProperties,
  fetchProjectAgents,
  saveGoalInsightRules,
  fetchGoalInsights,
  fetchFactorsModels,
  fetchEventNames,
  getUserProperties,
  fetchDemoProject,
  getHubspotContact,
  fetchProjectSettingsV1,
  fetchProjectSettings,
  fetchMarketoIntegration,
  fetchBingAdsIntegration
})(Factors);
