import React, {
  useEffect,
  useState,
  useCallback,
  Suspense,
  useRef,
  useMemo
} from 'react';
import cx from 'classnames';
import { Layout, Spin } from 'antd';

import { connect, useSelector, useDispatch } from 'react-redux';
import { bindActionCreators } from 'redux';
import Highcharts from 'highcharts';
import factorsai from 'factorsai';

import {
  fetchProjectsList,
  fetchProjectSettings,
  fetchProjectSettingsV1,
  triggerHubspotCustomFormFillEvent,
  getActiveProjectDetails,
  fetchDomainList
} from 'Reducers/global';

import customizeHighCharts from 'Utils/customizeHighcharts';
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

import { fetchTemplates } from '../../reducers/dashboard_templates/services';
import { AppLayoutRoutes } from 'Routes/AppLayoutRoutes';
import { TOGGLE_GLOBAL_SEARCH } from 'Reducers/types';
import './index.css';
import _, { isEmpty } from 'lodash';
import logger from 'Utils/logger';
import { matchPath, useLocation } from 'react-router-dom';
import { selectSidebarCollapsedState } from 'Reducers/global/selectors';
import { fetchProjectAgents, fetchAgentInfo } from 'Reducers/agentActions';
import { fetchFeatureConfig } from 'Reducers/featureConfig/middleware';
import {
  fetchCurrentSubscriptionDetail,
  fetchDifferentialPricing,
  fetchPlansDetail
} from 'Reducers/plansConfig/middleware';
import { selectAreDraftsSelected } from 'Reducers/dashboard/selectors';
import OnboardingRouting from 'Onboarding/ui/OnboardingRouting';
import moment from 'moment';
import useAgentInfo from 'hooks/useAgentInfo';
import { fetchAlertTemplates } from 'Reducers/alertTemplates';
import {
  checkMatchPath,
  isSettingsUrl
} from 'Views/AppSidebar/appSidebar.helpers';
import {
  SIGNUP_HS_FORM_ID,
  SIGNUP_HS_PORTAL_ID,
  routesWithSidebar
} from './appLayout.constants';
import styles from './index.module.scss';
import AppSidebar from '../AppSidebar';
import GlobalSearchModal from './GlobalSearchModal';
import {
  fetchEventDisplayNames,
  fetchQueries
} from '../../reducers/coreQuery/services';
import { PathUrls } from '../../routes/pathUrls';

// customizing highcharts for project requirements
customizeHighCharts(Highcharts);

function AppLayout({
  fetchProjectsList,
  fetchEventNames,
  getEventPropertiesV2,
  getUserPropertiesV2,
  getGroupProperties,
  fetchWeeklyIngishtsMetaData,
  getActiveProjectDetails,
  fetchProjectSettings,
  fetchProjectSettingsV1,
  fetchAgentInfo
}) {
  const [dataLoading, setDataLoading] = useState(true);
  const { Content } = Layout;
  const agentState = useSelector((state) => state.agent);
  const isAgentLoggedIn = agentState.isLoggedIn;
  const {
    projects,
    active_project,
    activeProjectLoading,
    currentProjectSettings
  } = useSelector((state) => state.global);
  const isSidebarCollapsed = useSelector((state) =>
    selectSidebarCollapsedState(state)
  );
  const { show_analytics_result } = useSelector((state) => state.coreQuery);
  const areDraftsSelected = useSelector((state) =>
    selectAreDraftsSelected(state)
  );
  const dispatch = useDispatch();
  const location = useLocation();
  const agentInfo = useAgentInfo();
  const agentInfoRef = useRef(agentInfo);

  const { pathname } = location;

  const activeAgent = agentState?.agent_details?.email;

  const activeAgentUUID = agentState?.agent_details?.uuid;

  const isChecklistEnabled = useMemo(() => {
    const agent = agentState.agents.filter(
      (data) => data.uuid === activeAgentUUID
    );
    return agent[0]?.checklist_dismissed;
  }, [agentState, agentState?.agents]);

  const fetchProjectsOnLoad = useCallback(async () => {
    try {
      if (isAgentLoggedIn) {
        const res = await fetchProjectsList();
        // handling when no project is present
        if (isEmpty(res?.status === 200)) {
          setDataLoading(false);
        }
      } else setDataLoading(false);
    } catch (err) {
      console.log(err);
    }
  }, [fetchProjectsList, isAgentLoggedIn]);

  useEffect(() => {
    const onKeyDown = (e) => {
      if ((e.metaKey || e.ctrlKey) && e.keyCode === 75) {
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
    if (projects.length > 0 && _.isEmpty(active_project)) {
      const storedActiveProjectId = localStorage.getItem('activeProject');
      const activeItem = projects.find(
        (item) => item.id === storedActiveProjectId
      );

      const projectID = _.isEmpty(activeItem)
        ? projects[0]?.id
        : activeItem?.id;

      localStorage.setItem('activeProject', projectID);
      getActiveProjectDetails(projectID);
    }
  }, [projects, active_project]);

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
      dispatch(fetchDomainList(active_project?.id));
      // dispatch(deleteQueryTest())
      fetchEventNames(active_project?.id);
      getUserPropertiesV2(active_project?.id);
      dispatch(fetchSmartPropertyRules(active_project?.id));
      fetchWeeklyIngishtsMetaData(active_project?.id);
      dispatch(fetchAttrContentGroups(active_project?.id));
      dispatch(fetchTemplates());
      dispatch(fetchAlertTemplates());
      handleRedirection();

      fetchProjectSettingsV1(active_project?.id);
      dispatch(fetchEventDisplayNames({ projectId: active_project?.id }));
      dispatch(fetchAttributionQueries(active_project?.id));
      dispatch(fetchProjectAgents(active_project.id));
      dispatch(fetchFeatureConfig(active_project?.id));

      // calling V2 pricing API's only if flag is enabled.
      if (active_project?.enable_billing) {
        dispatch(fetchCurrentSubscriptionDetail(active_project?.id));
        dispatch(fetchPlansDetail(active_project?.id));
        dispatch(fetchDifferentialPricing(active_project?.id));
      }
    }
  }, [dispatch, active_project]);

  // fetching agent info -> not dependent on active project
  useEffect(() => {
    if (isAgentLoggedIn) fetchAgentInfo();
  }, [isAgentLoggedIn, fetchAgentInfo]);

  // for handling signup event for the first time logged in user
  useEffect(() => {
    const login_count = agentState?.agent_details?.login_count;
    // using last login time so that for existing logged in users with login count 1 we dont trigger the signup event
    const lastLoggeInTime =
      moment(agentState?.agent_details?.last_logged_in) || moment();
    const currentTime = moment();
    const timeDifference = currentTime.diff(lastLoggeInTime, 'hours');
    if (activeAgent && login_count) {
      const signupEventlocalStoragekey = `${activeAgent}-signup_event_sent`;
      const isSignUpEventSent = localStorage.getItem(
        signupEventlocalStoragekey
      );
      if (login_count === 1 && !isSignUpEventSent && timeDifference < 24) {
        // triggering inside settimeout to prevent triggering event before sdk is initialised
        setTimeout(() => {
          factorsai.track('SIGNUP', {
            first_name: agentInfo?.firstName || '',
            email: agentInfo?.email,
            last_name: agentInfo?.lastName || ''
          });
          triggerHubspotCustomFormFillEvent(
            SIGNUP_HS_PORTAL_ID,
            SIGNUP_HS_FORM_ID,
            [
              {
                name: 'email',
                value: agentInfo?.email
              },
              {
                name: 'firstname',
                value: agentInfo?.firstName || ''
              },
              {
                name: 'lastname',
                value: agentInfo?.lastName || ''
              },
              {
                name: 'invited_user',
                value: !!agentInfo?.isAgentInvited
              },
              {
                name: 'phone',
                value: agentInfo?.phone
              }
            ]
          );
        }, 3000);

        localStorage.setItem(signupEventlocalStoragekey, 'true');
      }
    }
  }, [activeAgent, agentState]);

  if (dataLoading || activeProjectLoading) {
    return <Spin size='large' className='fa-page-loader' />;
  }

  const hasSidebar = routesWithSidebar.find((route) => {
    if (
      matchPath(pathname, {
        path: PathUrls.VisitorIdentificationReport,
        exact: true,
        strict: false
      })
    )
      return false;
    return matchPath(pathname, {
      path: route,
      exact: true,
      strict: false
    });
  });
  // 3.5rem is used because Top Navbar is 3.5rem
  return (
    <Layout className={styles['parent-layout']}>
      <ErrorBoundary
        fallback={
          <FaErrorComp
            size='medium'
            title='Bundle Error:01'
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
              'bg-white',
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
              style={{ minHeight: 'auto' }}
              className={cx(
                'bg-white',
                {
                  'py-6 px-10':
                    pathname !== PathUrls.ProfileAccounts &&
                    !checkMatchPath(
                      pathname,
                      PathUrls.ProfileAccountsSegmentsURL
                    ) &&
                    !checkMatchPath(
                      pathname,
                      PathUrls.ProfileAccountDetailsURL
                    ) &&
                    !checkMatchPath(
                      pathname,
                      PathUrls.ProfilePeopleDetailsURL
                    ) &&
                    !show_analytics_result &&
                    !isSettingsUrl(pathname)
                },
                {
                  'py-2':
                    pathname === PathUrls.ProfileAccounts ||
                    checkMatchPath(
                      pathname,
                      PathUrls.ProfileAccountsSegmentsURL
                    )
                },
                {
                  'px-8 py-4': isSettingsUrl(pathname)
                }
              )}
            >
              <Suspense fallback={<PageSuspenseLoader />}>
                <AppLayoutRoutes
                  activeAgent={activeAgent}
                  active_project={active_project}
                  currentProjectSettings={currentProjectSettings}
                />
              </Suspense>
            </Content>
          </Layout>
        </Layout>
        <GlobalSearchModal />
        {!activeProjectLoading && <OnboardingRouting />}
      </ErrorBoundary>
    </Layout>
  );
}

const mapDispatchToProps = (dispatch) =>
  bindActionCreators(
    {
      fetchProjectsList,
      fetchDashboards,
      fetchEventNames,
      getEventPropertiesV2,
      getUserPropertiesV2,
      getGroupProperties,
      fetchWeeklyIngishtsMetaData,
      getActiveProjectDetails,
      fetchProjectSettings,
      fetchProjectSettingsV1,
      fetchAgentInfo
    },
    dispatch
  );

export default connect(null, mapDispatchToProps)(AppLayout);
