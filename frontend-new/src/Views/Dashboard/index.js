import React, { useState, useCallback, useEffect } from 'react';
import Header from '../AppLayout/Header';
import SearchBar from '../../components/SearchBar';
import ProjectTabs from './ProjectTabs';
// import ProjectTabs from './ProjectTabs';
import AddDashboard from './AddDashboard';
import { useDispatch } from 'react-redux';
import { DASHBOARD_UNMOUNTED } from '../../reducers/types';

function Dashboard() {
  const [addDashboardModal, setaddDashboardModal] = useState(false);
  const [editDashboard, setEditDashboard] = useState(null);
  const [durationObj, setDurationObj] = useState({
    from: '',
    to: '',
    frequency: 'date'
  });
  const dispatch = useDispatch();

  const handleEditClick = useCallback((dashboard) => {
    setaddDashboardModal(true);
    setEditDashboard(dashboard);
  }, []);

  const handleDurationChange = useCallback((dates) => {
    if (dates && dates.selected) {
      setDurationObj(currState => {
        return {
          ...currState,
          from: dates.selected.startDate,
          to: dates.selected.endDate
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

      <div className={'mt-16'}>
        <ProjectTabs
          handleEditClick={handleEditClick}
          setaddDashboardModal={setaddDashboardModal}
          durationObj={durationObj}
          handleDurationChange={handleDurationChange}
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
