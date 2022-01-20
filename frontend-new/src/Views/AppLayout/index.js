import React, { useEffect, useState, useCallback, lazy, Suspense } from "react";
import { Layout, Spin } from "antd";
import Sidebar from "../../components/Sidebar"; 
import ProjectSettings from "../Settings/ProjectSettings";
import componentsLib from "../../Views/componentsLib";
import SetupAssist from "../Settings/SetupAssist";
import { connect, useSelector, useDispatch } from "react-redux";
import { bindActionCreators } from "redux";
import {
  BrowserRouter as Router,
  Route,
  Switch,
  useHistory,
} from 'react-router-dom';
import { fetchProjects, setActiveProject } from 'Reducers/global';
import {
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

const CoreQuery = lazyWithRetry(() => import('../CoreQuery'));
const Dashboard = lazyWithRetry(() => import('../Dashboard'));
const Factors = lazyWithRetry(() => import('../Factors'));

// import CoreQuery from "../CoreQuery";
// import Dashboard from "../Dashboard";
// import Factors from "../Factors";

function AppLayout({
  fetchProjects,
  fetchEventNames,
  getEventProperties,
  getUserProperties,
  fetchWeeklyIngishtsMetaData,
  setActiveProject,
}) {
  const [dataLoading, setDataLoading] = useState(true);
  const { Content } = Layout;
  const history = useHistory();
  const agentState = useSelector((state) => state.agent);
  const isAgentLoggedIn = agentState.isLoggedIn;
  const { active_project } = useSelector((state) => state.global);
  const { projects } = useSelector((state) => state.global);
  const { show_analytics_result } = useSelector((state) => state.coreQuery);
  const dispatch = useDispatch();

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
    if (projects.length && _.isEmpty(active_project)) {
      let activeItem = projects?.filter(
        (item) => item.id == localStorage.getItem('activeProject')
      );
      let projectDetails = _.isEmpty(activeItem) ? projects[0] : activeItem[0];
      setActiveProject(projectDetails);
    }
  }, [projects]);

  useEffect(() => {
    if (active_project && active_project.id) {
      dispatch(fetchDashboards(active_project.id));
      dispatch(fetchQueries(active_project.id));
      dispatch(fetchKPIConfig(active_project.id));
      dispatch(fetchPageUrls(active_project.id));
      fetchEventNames(active_project.id);
      getUserProperties(active_project.id);
      dispatch(fetchSmartPropertyRules(active_project.id));
      fetchWeeklyIngishtsMetaData(active_project.id);
    }
  }, [dispatch, active_project]);

  if (!isAgentLoggedIn) {
    history.push('/login');
    return null;
  }

  let contentClassName = 'fa-content-container';

  if (show_analytics_result) {
    contentClassName = 'fa-content-container no-sidebar';
  }

  return (
    <>
      {dataLoading ? (
        <Spin size={'large'} className={'fa-page-loader'} />
      ) : (
        <Layout>
          <ErrorBoundary fallback={<FaErrorComp size={'medium'} title={'Bundle Error:01'} subtitle={ "We are facing trouble loading App Bundles. Drop us a message on the in-app chat."} />} onError={FaErrorLog}> 
          {!show_analytics_result ? <Sidebar /> : null}
          <Layout className={contentClassName}>
            <Content className="bg-white min-h-screen">
              <Suspense fallback={<PageSuspenseLoader />}>
                <Switch>
                  <Route exact path="/" name="Home" component={Dashboard} />
                  <Route
                    path="/components"
                    name="componentsLib"
                    component={componentsLib}
                  />
                  <Route path="/settings" component={ProjectSettings} />
                  <Route
                    path="/analyse"
                    name="Home"
                    component={CoreQuery}
                  />
                  <Route path="/explain" name="Factors" component={Factors} />
                  {/* <Route path="/project-setup" component={SetupAssist} /> */}
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
    },
    dispatch
  );

export default connect(null, mapDispatchToProps)(AppLayout);
