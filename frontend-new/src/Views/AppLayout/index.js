import React, { useEffect, useState, useCallback, Suspense } from 'react';
import { Layout, Modal, Spin } from 'antd';

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
  // fetchAttrContentGroups,
  fetchGroups,
  fetchQueries
  // fetchSmartPropertyRules
} from '../../reducers/coreQuery/services';
import {
  fetchAttrContentGroups,
  fetchSmartPropertyRules,
  fetchAttributionQueries
} from 'Attribution/state/services';
import {
  getUserProperties,
  getEventProperties,
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
import NavigationBar from '../../components/NavigationBar';
import SearchBar from '../../components/SearchBar';

import { EMPTY_ARRAY } from '../../utils/global';

import { fetchTemplates } from '../../reducers/dashboard_templates/services';
import { AppLayoutRoutes } from 'Routes';
import { TOGGLE_GLOBAL_SEARCH } from 'Reducers/types';
import GlobalSearch from 'Components/GlobalSearch';
import './index.css';

// customizing highcharts for project requirements
customizeHighCharts(Highcharts);

function AppLayout({
  fetchProjects,
  fetchEventNames,
  getEventProperties,
  getUserProperties,
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
  const { projects } = useSelector((state) => state.global);
  const { show_analytics_result } = useSelector((state) => state.coreQuery);
  const { currentProjectSettings } = useSelector((state) => state.global);
  const dispatch = useDispatch();
  const [sidebarCollapse, setSidebarCollapse] = useState(true);

  const activeAgent = agentState?.agent_details?.email;

  const isVisibleGlobalSearch = useSelector(
    (state) => state.globalSearch.visible
  );

  const onKeyDown = (e) => {
    if (e.metaKey && e.keyCode == 75) {
      dispatch({ type: TOGGLE_GLOBAL_SEARCH });
    }
  };
  const asyncCallOnLoad = useCallback(async () => {
    try {
      if (isAgentLoggedIn) await fetchProjects();
      setDataLoading(false);
    } catch (err) {
      console.log(err);
    }
  }, [fetchProjects, isAgentLoggedIn]);
  useEffect(() => {
    // on Mount of Component
    document.onkeydown = onKeyDown;
    return () => {
      // on Unmount of Component
      document.onkeydown = null;
    };
  }, []);
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
    if (active_project && active_project?.id && isAgentLoggedIn) {
      dispatch(fetchDashboards(active_project?.id));
      dispatch(fetchQueries(active_project?.id));
      dispatch(fetchGroups(active_project?.id));
      dispatch(fetchKPIConfig(active_project?.id));
      dispatch(fetchPageUrls(active_project?.id));
      // dispatch(deleteQueryTest())
      fetchEventNames(active_project?.id);
      getUserProperties(active_project?.id);
      getGroupProperties(active_project?.id);
      dispatch(fetchSmartPropertyRules(active_project?.id));
      fetchWeeklyIngishtsMetaData(active_project?.id);
      dispatch(fetchAttrContentGroups(active_project?.id));
      dispatch(fetchTemplates());
      fetchProjectSettings(active_project?.id);

      fetchProjectSettingsV1(active_project?.id);
      dispatch(fetchEventDisplayNames({ projectId: active_project?.id }));
      dispatch(fetchAttributionQueries(active_project?.id));
    }
  }, [dispatch, active_project]);

  return (
    // eslint-disable-next-line react/jsx-no-useless-fragment
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
                subtitle='We are facing trouble loading App Bundles. Drop us a message on the in-app chat.'
              />
            }
            onError={FaErrorLog}
          >
            {!show_analytics_result && isAgentLoggedIn ? (
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
                  <AppLayoutRoutes
                    activeAgent={activeAgent}
                    demoProjectId={demoProjectId}
                    active_project={active_project}
                    currentProjectSettings={currentProjectSettings}
                  />
                </Suspense>
              </Content>
            </Layout>
            <Modal
              zIndex={2000}
              keyboard={true}
              visible={isVisibleGlobalSearch}
              footer={null}
              closable={false}
              onCancel={() => {
                dispatch({ type: TOGGLE_GLOBAL_SEARCH });
              }}
              bodyStyle={{ padding: 0 }}
              width={'40vw'}
              className='modal-globalsearch'
            >
              <GlobalSearch />
            </Modal>
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
