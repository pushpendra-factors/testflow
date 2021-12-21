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

function Dashboard() {
  const [addDashboardModal, setaddDashboardModal] = useState(false);
  const [editDashboard, setEditDashboard] = useState(null);
  const [durationObj, setDurationObj] = useState(getDashboardDateRange());
  const [refreshClicked, setRefreshClicked] = useState(false);
  const { dashboards } = useSelector(
    (state) => state.dashboard
  );
  const dispatch = useDispatch();

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
        <Header>
          <div className='w-full h-full py-4 flex flex-col justify-center items-center'>
            <SearchBar />
          </div>
        </Header>

        <div className={'mt-20'}>
          <ProjectTabs
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
      <EmptyDashboard />
      // <>
      //   <DashboardAfterIntegration setaddDashboardModal={setaddDashboardModal}/>
      //   {dashboards.data.length?
      //     <ProjectTabs
      //           handleEditClick={handleEditClick}
      //           setaddDashboardModal={setaddDashboardModal}
      //           durationObj={durationObj}
      //           handleDurationChange={handleDurationChange}
      //           refreshClicked={refreshClicked}
      //           setRefreshClicked={setRefreshClicked}
      //         />
      //   : null}
      //     <AddDashboard
      //         setEditDashboard={setEditDashboard}
      //         editDashboard={editDashboard}
      //         addDashboardModal={addDashboardModal}
      //         setaddDashboardModal={setaddDashboardModal}
      //       />
      //   </>
    );
  }
}

export default Dashboard;
