import React, { useState, useCallback, useEffect } from 'react';
import moment from 'moment';
import Header from '../AppLayout/Header';
import SearchBar from '../../components/SearchBar';
import ProjectTabs from './ProjectTabs';
// import ProjectTabs from './ProjectTabs';
import AddDashboard from './AddDashboard';
import { useDispatch } from 'react-redux';
import { DASHBOARD_UNMOUNTED } from '../../reducers/types';
import { DefaultDateRangeFormat } from '../CoreQuery/utils';

function Dashboard() {
  const [addDashboardModal, setaddDashboardModal] = useState(false);
  const [editDashboard, setEditDashboard] = useState(null);
  const [durationObj, setDurationObj] = useState({ ...DefaultDateRangeFormat });
  const [refreshClicked, setRefreshClicked] = useState(false);
  const dispatch = useDispatch();

  const handleEditClick = useCallback((dashboard) => {
    setaddDashboardModal(true);
    setEditDashboard(dashboard);
  }, []);

  const handleDurationChange = useCallback((dates) => {
    if (dates && dates.selected) {
      let frequency = 'date';
      if(moment(dates.selected.endDate).diff(dates.selected.startDate, 'hours') <= 24) {
        frequency = 'hour';
      }
      setDurationObj(currState => {
        return {
          ...currState,
          from: dates.selected.startDate,
          to: dates.selected.endDate,
          frequency
        };
      });
    }
  }, []);

  useEffect(() => {
    return () => {
      dispatch({ type: DASHBOARD_UNMOUNTED });
    };
  }, [dispatch]);

  return (
    <>
      <Header>
        <div className="w-full h-full py-4 flex flex-col justify-center items-center">
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

    </>
  );
}

export default Dashboard;
