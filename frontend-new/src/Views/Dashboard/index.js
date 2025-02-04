import React, { useState, useCallback, useEffect } from 'react';
import { useDispatch, useSelector, connect } from 'react-redux';
import { ErrorBoundary } from 'react-error-boundary';
import { Spin } from 'antd';
import { get, isEmpty } from 'lodash';
import { useHistory } from 'react-router-dom';
import { PathUrls } from 'Routes/pathUrls';
import { fetchDashboardFolders } from 'Reducers/dashboard/services';
import {
  fetchProjectSettingsV1 as fetchProjectSettingsV1Service,
  fetchBingAdsIntegration as fetchBingAdsIntegrationService,
  fetchMarketoIntegration as fetchMarketoIntegrationService,
  fetchProjectSettings as fetchProjectSettingsService
} from 'Reducers/global';

import {
  selectAreDraftsSelected,
  selectDashboardFoldersListState,
  selectDashboardList,
  selectEditDashboardDetailsState
} from 'Reducers/dashboard/selectors';
import CommonBeforeIntegrationPage from 'Components/GenericComponents/CommonBeforeIntegrationPage';
import AddDashboard from './AddDashboard';
import {
  ADD_DASHBOARD_MODAL_OPEN,
  DASHBOARD_UNMOUNTED
} from '../../reducers/types';
import { FaErrorComp, FaErrorLog } from '../../components/factorsComponents';
import { setItemToLocalStorage } from '../../utils/localStorage.helpers';
import { getDashboardDateRange } from './utils';
import DashboardAfterIntegration from './EmptyDashboard/DashboardAfterIntegration';
import ProjectDropdown from './ProjectDropdown';
import { DASHBOARD_KEYS } from '../../constants/localStorage.constants';
import Drafts from './Drafts';

const dashboardRefreshInitialState = {
  inProgress: false,
  widgetIdGettingFetched: null,
  widgetIdsLeftToBeFetched: [],
  widgetIdsAlreadyFetched: []
};

function Dashboard({
  fetchProjectSettingsV1,
  fetchBingAdsIntegration,
  fetchMarketoIntegration,
  fetchProjectSettings
}) {
  const [editDashboard, setEditDashboard] = useState(null);
  const [durationObj, setDurationObj] = useState(getDashboardDateRange());
  const [sdkCheck, setSdkCheck] = useState(false);
  const [oldestRefreshTime, setOldestRefreshTime] = useState(null);

  const [dashboardRefreshState, setDashboardRefreshState] = useState(
    dashboardRefreshInitialState
  );

  const { activeDashboardUnits, activeDashboard } = useSelector(
    (state) => state.dashboard
  );
  const editDashboardDetailsState = useSelector((state) =>
    selectEditDashboardDetailsState(state)
  );
  const dashboards = useSelector((state) => selectDashboardList(state));

  const areDraftsSelected = useSelector((state) =>
    selectAreDraftsSelected(state)
  );

  const integration = useSelector(
    (state) => state.global.currentProjectSettings
  );
  const queries = useSelector((state) => state.queries);
  const integrationV1 = useSelector((state) => state.global.projectSettingsV1);
  const activeProject = useSelector((state) => state.global.active_project);
  const foldersList = useSelector((state) =>
    selectDashboardFoldersListState(state)
  );
  const { bingAds, marketo } = useSelector((state) => state.global);
  const dispatch = useDispatch();
  const history = useHistory();

  const checkIntegration =
    integration?.int_segment ||
    integration?.int_adwords_enabled_agent_uuid ||
    integration?.int_linkedin_agent_uuid ||
    integration?.int_facebook_user_id ||
    integration?.int_hubspot ||
    integration?.int_salesforce_enabled_agent_uuid ||
    integration?.int_drift ||
    integration?.int_google_organic_enabled_agent_uuid ||
    integration?.int_clear_bit ||
    sdkCheck ||
    bingAds?.accounts ||
    marketo?.status ||
    integrationV1?.int_slack ||
    integration?.lead_squared_config !== null ||
    integration?.int_client_six_signal_key ||
    integration?.int_factors_six_signal_key ||
    integration?.int_rudderstack;

  const handleEditClick = useCallback((dashboard) => {
    dispatch({ type: ADD_DASHBOARD_MODAL_OPEN });
    setEditDashboard(dashboard);
  }, []);

  const handleRefreshClick = useCallback(() => {
    if (
      dashboardRefreshState.inProgress ||
      activeDashboardUnits.data.length === 0
    ) {
      return false;
    }
    setOldestRefreshTime(null);
    setDashboardRefreshState({
      inProgress: true,
      widgetIdsLeftToBeFetched: activeDashboardUnits.data
        .slice(1)
        .map((unit) => unit.id),
      widgetIdGettingFetched: activeDashboardUnits.data[0].id,
      widgetIdsAlreadyFetched: []
    });
    return null;
  }, [dashboardRefreshState.inProgress, activeDashboardUnits.data]);

  const handleWidgetRefresh = useCallback(
    (widgetId) => {
      if (dashboardRefreshState.inProgress) {
        return false;
      }
      setDashboardRefreshState({
        inProgress: true,
        widgetIdsLeftToBeFetched: [],
        widgetIdGettingFetched: widgetId,
        widgetIdsAlreadyFetched: []
      });
      return false;
    },
    [dashboardRefreshState.inProgress]
  );

  const resetDashboardRefreshState = useCallback(() => {
    setDashboardRefreshState(dashboardRefreshInitialState);
  }, []);

  const onDataLoadSuccess = useCallback(({ unitId }) => {
    setDashboardRefreshState((currState) => {
      if (currState.inProgress) {
        return {
          inProgress: currState.widgetIdsLeftToBeFetched.length > 0,
          widgetIdsAlreadyFetched: [
            ...currState.widgetIdsAlreadyFetched,
            unitId
          ],
          widgetIdGettingFetched:
            currState.widgetIdsLeftToBeFetched.length > 0
              ? currState.widgetIdsLeftToBeFetched[0]
              : null,
          widgetIdsLeftToBeFetched: currState.widgetIdsLeftToBeFetched.slice(1)
        };
      }
      return dashboardRefreshInitialState;
    });
  }, []);

  const handleDurationChange = useCallback(
    (dates) => {
      let from;
      let to;
      const { startDate, endDate } = dates;
      setOldestRefreshTime(null);
      resetDashboardRefreshState();
      if (Array.isArray(dates.startDate)) {
        from = get(startDate, 0);
        to = get(startDate, 1);
      } else {
        from = startDate;
        to = endDate;
      }

      setDurationObj((currState) => {
        const newState = {
          ...currState,
          from,
          to,
          dateType: dates.dateType
        };
        setItemToLocalStorage(
          DASHBOARD_KEYS.DASHBOARD_DURATION,
          JSON.stringify(newState)
        );
        return newState;
      });
    },
    [resetDashboardRefreshState]
  );

  useEffect(
    () => () => {
      dispatch({ type: DASHBOARD_UNMOUNTED });
    },
    [dispatch]
  );

  useEffect(() => {
    fetchProjectSettingsV1(activeProject?.id)
      .then((res) => {
        setSdkCheck(res?.data?.int_completed);
      })
      .catch((err) => {
        console.log(err);
      });

    fetchProjectSettings(activeProject?.id);

    if (isEmpty(dashboards)) {
      fetchBingAdsIntegration(activeProject?.id);
      fetchMarketoIntegration(activeProject?.id);
    }

    if (activeDashboard?.class === 'predefined') {
      history.replace(`${PathUrls.PreBuildDashboard}`);
    }
  }, [activeProject, sdkCheck]);

  useEffect(() => {
    if (foldersList.completed === false) {
      dispatch(fetchDashboardFolders(activeProject.id));
    }
  }, [activeProject.id, foldersList.completed]);

  useEffect(() => {
    if (editDashboardDetailsState.initiated === true) {
      setEditDashboard(editDashboardDetailsState.editDashboard);
    }
  }, [editDashboardDetailsState]);

  if (dashboards.loading || queries.loading || foldersList.completed !== true) {
    return (
      <div className='flex justify-center items-center w-full h-64'>
        <Spin size='large' />
      </div>
    );
  }
  const addDashboardModal = (
    <AddDashboard
      setEditDashboard={setEditDashboard}
      editDashboard={editDashboard}
    />
  );
  if (areDraftsSelected) {
    return (
      <>
        <Drafts /> {addDashboardModal}
      </>
    );
  }
  if (dashboards.length) {
    return (
      <ErrorBoundary
        fallback={
          <FaErrorComp
            size='medium'
            title='Dashboard Overview Error'
            subtitle='We are facing trouble loading dashboards overview. Drop us a message on the in-app chat.'
          />
        }
        onError={FaErrorLog}
      >
        {areDraftsSelected === false && (
          <div className='flex-1 flex flex-col'>
            <ProjectDropdown
              handleEditClick={handleEditClick}
              durationObj={durationObj}
              handleDurationChange={handleDurationChange}
              oldestRefreshTime={oldestRefreshTime}
              setOldestRefreshTime={setOldestRefreshTime}
              handleRefreshClick={handleRefreshClick}
              dashboardRefreshState={dashboardRefreshState}
              onDataLoadSuccess={onDataLoadSuccess}
              resetDashboardRefreshState={resetDashboardRefreshState}
              handleWidgetRefresh={handleWidgetRefresh}
            />
          </div>
        )}

        {addDashboardModal}
      </ErrorBoundary>
    );
  }

  if (checkIntegration) {
    return (
      <>
        <DashboardAfterIntegration />
        {/* <AddDashboard
          setEditDashboard={setEditDashboard}
          editDashboard={editDashboard}
          addDashboardModal={addDashboardModal}
          setaddDashboardModal={setaddDashboardModal}
        /> */}
      </>
    );
  }

  return <CommonBeforeIntegrationPage />;
}

export default connect(null, {
  fetchProjectSettingsV1: fetchProjectSettingsV1Service,
  fetchBingAdsIntegration: fetchBingAdsIntegrationService,
  fetchMarketoIntegration: fetchMarketoIntegrationService,
  fetchProjectSettings: fetchProjectSettingsService
})(Dashboard);
