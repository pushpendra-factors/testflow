import React, { useEffect, useState, useCallback, Suspense } from 'react';
import { Layout, Spin } from 'antd';
// import ProjectSettings from '../Settings/ProjectSettings';
import componentsLib from '../../Views/componentsLib';
import SetupAssist from '../Settings/SetupAssist';
import { connect, useSelector, useDispatch } from 'react-redux';
import { bindActionCreators } from 'redux';
import { Route, Switch, useHistory } from 'react-router-dom';
import {
  fetchProjects,
  setActiveProject,
  fetchDemoProject,
} from 'Reducers/global';
import {
  fetchAttrContentGroups,
  fetchGroups,
  fetchQueries,
  fetchSmartPropertyRules,
} from '../../reducers/coreQuery/services';
import {
  getUserProperties,
  getEventProperties,
  fetchEventNames,
} from '../../reducers/coreQuery/middleware';
import { fetchDashboards } from '../../reducers/dashboard/services';
import PageSuspenseLoader from '../../components/SuspenseLoaders/PageSuspenseLoader';
import lazyWithRetry from 'Utils/lazyWithRetry';
import { FaErrorComp, FaErrorLog } from 'factorsComponents';
import { ErrorBoundary } from 'react-error-boundary';
import { fetchWeeklyIngishtsMetaData } from 'Reducers/insights';
import { fetchKPIConfig, fetchPageUrls } from '../../reducers/kpi';
import Welcome from '../Settings/SetupAssist/Welcome';
import FaHeader from '../../components/FaHeader';
import NavigationBar from '../../components/NavigationBar';
import SearchBar from '../../components/SearchBar';
import AttributionSettings from '../Settings/ProjectSettings/AttributionSettings';
import BasicSettings from '../Settings/ProjectSettings/BasicSettings';
import SDKSettings from '../Settings/ProjectSettings/SDKSettings';
import UserSettings from '../Settings/ProjectSettings/UserSettings';
import IntegrationSettings from '../Settings/ProjectSettings/IntegrationSettings';
import Alerts from '../Settings/ProjectSettings/Alerts';
import ExplainDataPoints from '../Settings/ProjectConfigure/ExplainDataPoints';
import Events from '../Settings/ProjectConfigure/Events';
import PropertySettings from '../Settings/ProjectConfigure/PropertySettings';
import ContentGroups from '../Settings/ProjectConfigure/ContentGroups';
import Touchpoints from '../Settings/ProjectConfigure/Touchpoints';
import CustomKPI from '../Settings/ProjectConfigure/CustomKPI';
import { EMPTY_ARRAY } from '../../utils/global';

const FactorsInsights = lazyWithRetry(() =>
  import('../Factors/FactorsInsightsNew')
);
const CoreQuery = lazyWithRetry(() => import('../CoreQuery'));
const Dashboard = lazyWithRetry(() => import('../Dashboard'));
const Factors = lazyWithRetry(() => import('../Factors'));

function AppLayout({
  fetchProjects,
  fetchEventNames,
  getEventProperties,
  getUserProperties,
  fetchWeeklyIngishtsMetaData,
  setActiveProject,
  fetchDemoProject,
}) {
  const [dataLoading, setDataLoading] = useState(true);
  const [demoProjectId, setDemoProjectId] = useState(EMPTY_ARRAY);
  const { Content } = Layout;
  const history = useHistory();
  const agentState = useSelector((state) => state.agent);
  const isAgentLoggedIn = agentState.isLoggedIn;
  const { active_project } = useSelector((state) => state.global);
  const { projects } = useSelector((state) => state.global);
  const { show_analytics_result } = useSelector((state) => state.coreQuery);
  const dispatch = useDispatch();
  const [sidebarCollapse, setSidebarCollapse] = useState(true);

  const asyncCallOnLoad = useCallback(async () => {
    try {
      await fetchProjects();
      setDataLoading(false);
    } catch (err) {
      console.log(err);
    }
  }, [fetchProjects]);

  useEffect(() => {
    asyncCallOnLoad();
  }, [asyncCallOnLoad]);

  useEffect(() => {
    fetchDemoProject().then((res) => {
      setDemoProjectId(res.data);
    });
  }, [fetchDemoProject, setDemoProjectId]);

  useEffect(() => {
    if (projects.length && _.isEmpty(active_project)) {
      let activeItem = projects?.filter(
        (item) => item.id == localStorage.getItem('activeProject')
      );
      //handling Saas factors demo project
      let default_project = demoProjectId.includes(projects[0].id)
        ? projects[1]
          ? projects[1]
          : projects[0]
        : projects[0];
      let projectDetails = _.isEmpty(activeItem)
        ? default_project
        : activeItem[0];
      setActiveProject(projectDetails);
    }
  }, [projects]);

  useEffect(() => {
    if (active_project && active_project.id) {
      dispatch(fetchDashboards(active_project.id));
      dispatch(fetchQueries(active_project.id));
      dispatch(fetchGroups(active_project.id));
      dispatch(fetchKPIConfig(active_project.id));
      dispatch(fetchPageUrls(active_project.id));
      // dispatch(deleteQueryTest())
      fetchEventNames(active_project.id);
      getUserProperties(active_project.id);
      dispatch(fetchSmartPropertyRules(active_project.id));
      fetchWeeklyIngishtsMetaData(active_project.id);
      dispatch(fetchAttrContentGroups(active_project.id));
    }
  }, [dispatch, active_project]);

  if (!isAgentLoggedIn) {
    history.push('/login');
    return null;
  }

  return (
    <>
      {dataLoading ? (
        <Spin size={'large'} className={'fa-page-loader'} />
      ) : (
        <Layout>
          <ErrorBoundary
            fallback={
              <FaErrorComp
                size={'medium'}
                title={'Bundle Error:01'}
                subtitle={
                  'We are facing trouble loading App Bundles. Drop us a message on the in-app chat.'
                }
              />
            }
            onError={FaErrorLog}
          >
            {!show_analytics_result ? (
              <>
                <FaHeader
                  collapse={sidebarCollapse}
                  setCollapse={setSidebarCollapse}
                >
                  <SearchBar />
                </FaHeader>
                <NavigationBar
                  collapse={sidebarCollapse}
                  setCollapse={setSidebarCollapse}
                />
              </>
            ) : null}
            <Layout>
              <Content className='bg-white'>
                <Suspense fallback={<PageSuspenseLoader />}>
                  <Switch>
                    <Route exact path='/' name='Home' component={Dashboard} />
                    <Route
                      path='/components'
                      name='componentsLib'
                      component={componentsLib}
                    />
                    {/* <Route path='/settings' component={ProjectSettings} /> */}
                    <Route path='/analyse' name='Home' component={CoreQuery} />

                    <Route
                      exact
                      path='/explain'
                      name='Factors'
                      component={Factors}
                    />
                    <Route
                      exact
                      path='/explain/insights'
                      name='Factors'
                      component={FactorsInsights}
                    />

                    <Route path='/project-setup' component={SetupAssist} />
                    <Route path='/welcome' component={Welcome} />

                    {/* settings */}
                    <Route path='/settings/general' component={BasicSettings} />
                    <Route path='/settings/user' component={UserSettings} />
                    <Route
                      path='/settings/attribution'
                      component={AttributionSettings}
                    />
                    <Route path='/settings/sdk' component={SDKSettings} />
                    <Route
                      path='/settings/integration'
                      component={IntegrationSettings}
                    />

                    {/* configure */}
                    <Route path='/configure/events' component={Events} />
                    <Route
                      path='/configure/properties'
                      component={PropertySettings}
                    />
                    <Route
                      path='/configure/contentgroups'
                      component={ContentGroups}
                    />
                    <Route
                      path='/configure/touchpoints'
                      component={Touchpoints}
                    />
                    <Route path='/configure/customkpi' component={CustomKPI} />
                    <Route
                      path='/configure/explaindp'
                      component={ExplainDataPoints}
                    />
                    <Route path='/configure/alerts' component={Alerts} />
                    {/* <Route path='/configure/goals' component={goals} /> */}
                  </Switch>
                </Suspense>
              </Content>
            </Layout>
          </ErrorBoundary>
        </Layout>
      )}
    </>
  );
}

const mapDispatchToProps = (dispatch) =>
  bindActionCreators(
    {
      fetchProjects,
      fetchDashboards,
      fetchEventNames,
      getEventProperties,
      getUserProperties,
      fetchWeeklyIngishtsMetaData,
      setActiveProject,
      fetchDemoProject,
    },
    dispatch
  );

export default connect(null, mapDispatchToProps)(AppLayout);
