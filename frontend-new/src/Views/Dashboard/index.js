import React, { useState, useCallback } from 'react';
import Header from '../AppLayout/Header';
import SearchBar from '../../components/SearchBar';
import ProjectTabs from './ProjectTabs';
import AddDashboard from './AddDashboard';

function Dashboard() {
  const [addDashboardModal, setaddDashboardModal] = useState(false);
  const [editDashboard, setEditDashboard] = useState(null);

  const handleEditClick = useCallback((dashboard) => {
    setaddDashboardModal(true);
    setEditDashboard(dashboard);
  }, []);

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
