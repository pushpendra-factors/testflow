import React, { useEffect, useState, useCallback, Suspense } from 'react';
import cx from 'classnames';
import { Layout, Spin } from 'antd';

import { connect, useSelector, useDispatch } from 'react-redux';
import { bindActionCreators } from 'redux';
import Highcharts from 'highcharts';
import {
  fetchProjects,
  setActiveProject,
  fetchDemoProject,
  fetchProjectSettings,
  fetchProjectSettingsV1
} from 'Reducers/global';
import customizeHighCharts from 'Utils/customizeHighcharts';
import {
  fetchEventDisplayNames,
  fetchQueries
} from '../../reducers/coreQuery/services';
import {
  fetchAttrContentGroups,
  fetchSmartPropertyRules,
  fetchAttributionQueries
} from 'Attribution/state/services';
import {
  getUserPropertiesV2,
  getEventPropertiesV2,
  fetchEventNames,
  getGroupProperties
} from '../../reducers/coreQuery/middleware';
import { fetchDashboards } from '../../reducers/dashboard/services';
import PageSuspenseLoader from '../../components/SuspenseLoaders/PageSuspenseLoader';
import { FaErrorComp, FaErrorLog } from 'factorsComponents';
import { ErrorBoundary } from 'react-error-boundary';
import { fetchWeeklyIngishtsMetaData } from 'Reducers/insights';
import { fetchKPIConfig, fetchPageUrls } from '../../reducers/kpi';
import FaHeader from '../../components/FaHeader';

import { EMPTY_ARRAY } from '../../utils/global';

import { fetchTemplates } from '../../reducers/dashboard_templates/services';
import { AppLayoutRoutes } from 'Routes/index';
import { TOGGLE_GLOBAL_SEARCH } from 'Reducers/types';
import './index.css';
import _ from 'lodash';
import logger from 'Utils/logger';
import { useLocation } from 'react-router-dom';
import GlobalSearchModal from './GlobalSearchModal';
import { APP_LAYOUT_ROUTES } from '../../routes/constants';
import AppSidebar from '../AppSidebar';
import styles from './index.module.scss';
import { routesWithSidebar } from './appLayout.constants';
import { selectSidebarCollapsedState } from 'Reducers/global/selectors';
import { fetchProjectAgents } from 'Reducers/agentActions';
import { fetchFeatureConfig } from 'Reducers/featureConfig/middleware';
import { selectAreDraftsSelected } from 'Reducers/dashboard/selectors';
import { PathUrls } from '../../routes/pathUrls';

// customizing highcharts for project requirements
customizeHighCharts(Highcharts);

function AppLayout({
  fetchProjects,
  fetchEventNames,
  getEventPropertiesV2,
  getUserPropertiesV2,
  getGroupProperties,
  fetchWeeklyIngishtsMetaData,
  setActiveProject,
  fetchDemoProject,
  fetchProjectSettings,
  fetchProjectSettingsV1
}) {
  const [dataLoading, setDataLoading] = useState(true);
  const [demoProjectId, setDemoProjectId] = useState(EMPTY_ARRAY);
  const { Content } = Layout;
  const agentState = useSelector((state) => state.agent);
  const isAgentLoggedIn = agentState.isLoggedIn;
  const { active_project } = useSelector((state) => state.global);
  const isSidebarCollapsed = useSelector((state) =>
    selectSidebarCollapsedState(state)
  );
  const { projects } = useSelector((state) => state.global);
  const { show_analytics_result } = useSelector((state) => state.coreQuery);
  const { currentProjectSettings } = useSelector((state) => state.global);
  const areDraftsSelected = useSelector((state) =>
    selectAreDraftsSelected(state)
  );
  const dispatch = useDispatch();
  const location = useLocation();
  const { pathname } = location;

  const activeAgent = agentState?.agent_details?.email;

  const fetchProjectsOnLoad = useCallback(async () => {
    try {
      if (isAgentLoggedIn) await fetchProjects();
      else setDataLoading(false);
    } catch (err) {
      console.log(err);
    }
  }, [fetchProjects, isAgentLoggedIn]);

  useEffect(() => {
    const onKeyDown = (e) => {
      if (e.metaKey && e.keyCode === 75) {
        dispatch({ type: TOGGLE_GLOBAL_SEARCH });
      }
    };
    // on Mount of Component
    document.onkeydown = onKeyDown;
    return () => {
      // on Unmount of Component
      document.onkeydown = null;
    };
  }, [dispatch]);

  useEffect(() => {
    fetchProjectsOnLoad();
  }, [fetchProjectsOnLoad]);

  useEffect(() => {
    fetchDemoProject().then((res) => {
      setDemoProjectId(res.data);
    });
  }, [fetchDemoProject, setDemoProjectId]);

  useEffect(() => {
    if (projects.length && _.isEmpty(active_project)) {
      let activeItem = projects?.filter(
        (item) => item.id === localStorage.getItem('activeProject')
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
      localStorage.setItem('activeProject', projectDetails?.id);
      setActiveProject(projectDetails);
    }
  }, [projects]);

  const handleRedirection = async () => {
    try {
      if (active_project && active_project?.id && isAgentLoggedIn) {
        await fetchProjectSettings(active_project?.id);
        // if (
        //   location?.state?.navigatedFromLoginPage &&
        //   (res?.data?.int_factors_six_signal_key ||
        //     res?.data?.int_client_six_signal_key)
        // ) {
        //   history.push(APP_LAYOUT_ROUTES.VisitorIdentificationReport.path);
        // }
      }
      setDataLoading(false);
    } catch (error) {
      logger.error('Error in fetching project settings', error);
      setDataLoading(false);
    }
  };

  useEffect(() => {
    if (active_project && active_project?.id && isAgentLoggedIn) {
      dispatch(fetchDashboards(active_project?.id));
      dispatch(fetchQueries(active_project?.id));
      dispatch(fetchKPIConfig(active_project?.id));
      dispatch(fetchPageUrls(active_project?.id));
      // dispatch(deleteQueryTest())
      fetchEventNames(active_project?.id);
      getUserPropertiesV2(active_project?.id);
      dispatch(fetchSmartPropertyRules(active_project?.id));
      fetchWeeklyIngishtsMetaData(active_project?.id);
      dispatch(fetchAttrContentGroups(active_project?.id));
      dispatch(fetchTemplates());
      handleRedirection();

      fetchProjectSettingsV1(active_project?.id);
      dispatch(fetchEventDisplayNames({ projectId: active_project?.id }));
      dispatch(fetchAttributionQueries(active_project?.id));
      dispatch(fetchProjectAgents(active_project.id));
      dispatch(fetchFeatureConfig(active_project?.id));
    }
  }, [dispatch, active_project]);

  if (dataLoading) {
    return <Spin size={'large'} className={'fa-page-loader'} />;
  }

  const hasSidebar = routesWithSidebar.includes(pathname);

  return (
    <Layout>
      <ErrorBoundary
        fallback={
          <FaErrorComp
            size={'medium'}
            title={'Bundle Error:01'}
            subtitle='We are facing trouble loading App Bundles. Drop us a message on the in-app chat.'
          />
        }
        onError={FaErrorLog}
      >
        {!show_analytics_result && isAgentLoggedIn ? <FaHeader /> : null}
        <Layout
          className={cx(styles['content-layout'], {
            [styles['no-header']]: show_analytics_result === true
          })}
        >
          {hasSidebar && <AppSidebar />}
          <Layout
            className={cx(
              {
                [styles['layout-with-sidebar']]: hasSidebar
              },
              {
                [styles['enable-transition']]:
                  hasSidebar &&
                  (pathname !== PathUrls.Dashboard ||
                    areDraftsSelected === false)
              },
              {
                [styles['collapsed-sidebar']]: isSidebarCollapsed
              }
            )}
          >
            <Content
              className={cx('bg-white', {
                'py-6 px-10': !show_analytics_result
              })}
            >
              <Suspense fallback={<PageSuspenseLoader />}>
                <AppLayoutRoutes
                  activeAgent={activeAgent}
                  demoProjectId={demoProjectId}
                  active_project={active_project}
                  currentProjectSettings={currentProjectSettings}
                />
              </Suspense>
            </Content>
          </Layout>
        </Layout>
        <GlobalSearchModal />
      </ErrorBoundary>
    </Layout>
  );
}

const mapDispatchToProps = (dispatch) =>
  bindActionCreators(
    {
      fetchProjects,
      fetchDashboards,
      fetchEventNames,
      getEventPropertiesV2,
      getUserPropertiesV2,
      getGroupProperties,
      fetchWeeklyIngishtsMetaData,
      setActiveProject,
      fetchDemoProject,
      fetchProjectSettings,
      fetchProjectSettingsV1
    },
    dispatch
  );

export default connect(null, mapDispatchToProps)(AppLayout);
