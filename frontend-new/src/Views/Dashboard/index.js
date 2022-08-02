import React, { useState, useCallback, useEffect } from 'react';
import { useDispatch, useSelector, connect } from 'react-redux';
import { ErrorBoundary } from 'react-error-boundary';
import { useHistory } from 'react-router-dom';
import { Spin } from 'antd';

import {
  fetchProjectSettingsV1,
  fetchDemoProject,
  fetchBingAdsIntegration,
  fetchMarketoIntegration,
  fetchProjectSettings
} from 'Reducers/global';

import AddDashboard from './AddDashboard';
import { DASHBOARD_UNMOUNTED } from '../../reducers/types';
import { FaErrorComp, FaErrorLog } from '../../components/factorsComponents';
import { setItemToLocalStorage } from '../../utils/localStorage.helpers';
import { getDashboardDateRange } from './utils';
import EmptyDashboard from './EmptyDashboard';
import DashboardAfterIntegration from './EmptyDashboard/DashboardAfterIntegration';
import ProjectDropdown from './ProjectDropdown';
import { DASHBOARD_KEYS } from '../../constants/localStorage.constants';
import DashboardBeforeIntegration from './DashboardBeforeIntegration';

function Dashboard({
  fetchProjectSettingsV1,
  fetchBingAdsIntegration,
  fetchMarketoIntegration,
  fetchProjectSettings
}) {
  const [addDashboardModal, setaddDashboardModal] = useState(false);
  const [editDashboard, setEditDashboard] = useState(null);
  const [durationObj, setDurationObj] = useState(getDashboardDateRange());
  const [refreshClicked, setRefreshClicked] = useState(false);
  const [sdkCheck, setSdkCheck] = useState(false);
  const { dashboards } = useSelector((state) => state.dashboard);
  const integration = useSelector((state) => state.global.currentProjectSettings);
  const integrationV1 = useSelector((state) => state.global.projectSettingsV1);
  const activeProject = useSelector((state) => state.global.active_project);
  const { bingAds, marketo } = useSelector((state) => state.global);
  const currentAgent = useSelector((state) => state.agent.agent_details);
  const dispatch = useDispatch();
  const history = useHistory();

  useEffect(() => {
    fetchProjectSettingsV1(activeProject?.id)
    .then((res) => {
      setSdkCheck(res?.data?.int_completed);
    })
    .catch((err) => {
      console.log(err);
    });

    fetchProjectSettings(activeProject?.id);

    if (dashboards?.data?.length == 0) {
      fetchBingAdsIntegration(activeProject?.id);
      fetchMarketoIntegration(activeProject?.id);
    }
  }, [activeProject, sdkCheck]);

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
    marketo?.status || integrationV1?.int_slack;

  const handleEditClick = useCallback((dashboard) => {
    setaddDashboardModal(true);
    setEditDashboard(dashboard);
  }, []);

  const handleDurationChange = useCallback((dates) => {
    let from, to;
    if (Array.isArray(dates.startDate)) {
      from = dates.startDate[0];
      to = dates.startDate[1];
    } else {
      from = dates.startDate;
      to = dates.endDate;
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
  }, []);

  useEffect(() => {
    return () => {
      dispatch({ type: DASHBOARD_UNMOUNTED });
    };
  }, [dispatch]);

  if (dashboards.loading) {
    return (
      <div className="flex justify-center items-center w-full h-64">
        <Spin size="large" />
      </div>
    );
  }

  if (dashboards.data.length) {
    return (
      <>
        <ErrorBoundary
          fallback={
            <FaErrorComp
              size={'medium'}
              title={'Dashboard Overview Error'}
              subtitle={
                'We are facing trouble loading dashboards overview. Drop us a message on the in-app chat.'
              }
            />
          }
          onError={FaErrorLog}
        >
          {/* <FaHeader>
            <SearchBar />
          </FaHeader> */}
          <div className="mt-20 flex-1 flex flex-col">
            <ProjectDropdown
              handleEditClick={handleEditClick}
              setaddDashboardModal={setaddDashboardModal}
              durationObj={durationObj}
              handleDurationChange={handleDurationChange}
              refreshClicked={refreshClicked}
              setRefreshClicked={setRefreshClicked}
            />
          </div>

          <AddDashboard
            setEditDashboard={setEditDashboard}
            editDashboard={editDashboard}
            addDashboardModal={addDashboardModal}
            setaddDashboardModal={setaddDashboardModal}
          />
        </ErrorBoundary>
      </>
    );
  } else {
    return (
      <>
        {checkIntegration ? (
          <>
            <DashboardAfterIntegration
              setaddDashboardModal={setaddDashboardModal}
            />
            <AddDashboard
              setEditDashboard={setEditDashboard}
              editDashboard={editDashboard}
              addDashboardModal={addDashboardModal}
              setaddDashboardModal={setaddDashboardModal}
            />
          </>
        ) : (
          // <EmptyDashboard />
          <DashboardBeforeIntegration />
        )}
      </>
    );
  }
}

export default connect(null, {
  fetchProjectSettingsV1,
  fetchDemoProject,
  fetchBingAdsIntegration,
  fetchMarketoIntegration,
  fetchProjectSettings
})(Dashboard);
