import React, { useState, useCallback, useEffect } from 'react';
import MomentTz from 'Components/MomentTz';
import Header from '../AppLayout/Header';
import SearchBar from '../../components/SearchBar';
import ProjectTabs from './ProjectTabs';
import AddDashboard from './AddDashboard';
import { useDispatch, useSelector } from 'react-redux';
import { DASHBOARD_UNMOUNTED } from '../../reducers/types';
import { FaErrorComp, FaErrorLog } from '../../components/factorsComponents';
import { ErrorBoundary } from 'react-error-boundary';
import { setItemToLocalStorage } from '../../utils/dataFormatter';
import { getDashboardDateRange } from './utils';
import { LOCAL_STORAGE_ITEMS } from '../../utils/constants';
import EmptyDashboard from './EmptyDashboard';
import DashboardAfterIntegration from './EmptyDashboard/DashboardAfterIntegration'
import ProjectDropdown from './ProjectDropdown';
import { connect } from 'react-redux';
import { fetchProjectSettingsV1, fetchDemoProject } from 'Reducers/global';
import { useHistory } from 'react-router-dom';

function Dashboard({ fetchProjectSettingsV1, fetchDemoProject }) {
  const [addDashboardModal, setaddDashboardModal] = useState(false);
  const [editDashboard, setEditDashboard] = useState(null);
  const [durationObj, setDurationObj] = useState(getDashboardDateRange());
  const [refreshClicked, setRefreshClicked] = useState(false);
  const [sdkCheck, setsdkCheck] = useState();
  const { dashboards } = useSelector((state) => state.dashboard);
  let integration = useSelector((state) => state.global.currentProjectSettings);
  const activeProject = useSelector((state) => state.global.active_project) 
  const dispatch = useDispatch();
  const history = useHistory();

  useEffect(() => {
      fetchProjectSettingsV1(activeProject.id).then((res) => {
          console.log('fetch project settings success');
          setsdkCheck(res.data.int_completed);
      }).catch((err) => {
        console.log(err.data.error)
        history.push('/');
    })
  }, [activeProject, sdkCheck]);

  integration = integration?.project_settings || integration;

  const checkIntegration = integration?.int_segment || 
  integration?.int_adwords_enabled_agent_uuid ||
  integration?.int_linkedin_agent_uuid ||
  integration?.int_facebook_user_id ||
  integration?.int_hubspot ||
  integration?.int_salesforce_enabled_agent_uuid ||
  integration?.int_drift ||
  integration?.int_google_organic_enabled_agent_uuid ||
  integration?.int_clear_bit || sdkCheck;

  useEffect(() => {
    fetchDemoProject().then((res) => {
      const projectId = res.data[0];
      console.log(res.data[0]);
      if(activeProject.id === projectId) {
        history.push('/')
      } else if (!checkIntegration) {
        history.push('/project-setup')
      }
    }).catch((err) => {
      console.log(err.data.error);
    })
  },[checkIntegration, activeProject])

  const handleEditClick = useCallback((dashboard) => {
    setaddDashboardModal(true);
    setEditDashboard(dashboard);
  }, []);

  const handleDurationChange = useCallback((dates) => {
    let from,
      to,
      frequency = 'date';
    if (Array.isArray(dates.startDate)) {
      from = dates.startDate[0];
      to = dates.startDate[1];
    } else {
      from = dates.startDate;
      to = dates.endDate;
    }
    if (MomentTz(to).diff(from, 'hours') < 24) {
      frequency = 'hour';
    }

    setDurationObj((currState) => {
      const newState = {
        ...currState,
        from,
        to,
        frequency,
        dateType: dates.dateType,
      };
      setItemToLocalStorage(
        LOCAL_STORAGE_ITEMS.DASHBOARD_DURATION,
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
          <div className='flex flex-col h-full'>
            <Header>
              <div className='w-full h-full py-4 flex flex-col justify-center items-center'>
                <SearchBar />
              </div>
            </Header>

            <div className={`mt-20 flex-1 flex flex-col`}>
              <ProjectDropdown
                handleEditClick={handleEditClick}
                setaddDashboardModal={setaddDashboardModal}
                durationObj={durationObj}
                handleDurationChange={handleDurationChange}
                refreshClicked={refreshClicked}
                setRefreshClicked={setRefreshClicked}
              />
            </div>
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
        {checkIntegration ?
          <>
            <DashboardAfterIntegration setaddDashboardModal={setaddDashboardModal} />
            <AddDashboard
              setEditDashboard={setEditDashboard}
              editDashboard={editDashboard}
              addDashboardModal={addDashboardModal}
              setaddDashboardModal={setaddDashboardModal}
            />
          </>
          :
          <EmptyDashboard />
        }
      </>
    );
  }
}

export default connect(null,{ fetchProjectSettingsV1, fetchDemoProject })(Dashboard);
