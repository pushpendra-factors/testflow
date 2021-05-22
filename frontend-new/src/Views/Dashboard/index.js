import React, { useState, useCallback, useEffect } from 'react';
import moment from 'moment';
import Header from '../AppLayout/Header';
import SearchBar from '../../components/SearchBar';
import ProjectTabs from './ProjectTabs';
import AddDashboard from './AddDashboard';
import { useDispatch } from 'react-redux';
import { DASHBOARD_UNMOUNTED } from '../../reducers/types';
import { DefaultDateRangeFormat } from '../CoreQuery/utils';
import { FaErrorComp, FaErrorLog } from '../../components/factorsComponents';
import { ErrorBoundary } from 'react-error-boundary';
import { connect } from 'react-redux';
import { bindActionCreators } from 'redux';
import { fetchEventNames } from '../../reducers/coreQuery/middleware';

function Dashboard({ fetchEventNames, activeProject }) {
  const [addDashboardModal, setaddDashboardModal] = useState(false);
  const [editDashboard, setEditDashboard] = useState(null);
  const [durationObj, setDurationObj] = useState({ ...DefaultDateRangeFormat });
  const [refreshClicked, setRefreshClicked] = useState(false);
  const dispatch = useDispatch();

  useEffect(() => {
    if (activeProject && activeProject.id) {
      fetchEventNames(activeProject.id);
    }
  }, [activeProject, fetchEventNames]);

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
    if (moment(to).diff(from, 'hours') < 24) {
      frequency = 'hour';
    }
    setDurationObj((currState) => {
      return {
        ...currState,
        from,
        to,
        frequency,
        ...dates,
      };
    });
  }, []);

  useEffect(() => {
    return () => {
      dispatch({ type: DASHBOARD_UNMOUNTED });
    };
  }, [dispatch]);

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
}

const mapStateToProps = (state) => ({
  activeProject: state.global.active_project,
});

const mapDispatchToProps = (dispatch) =>
  bindActionCreators(
    {
      fetchEventNames,
    },
    dispatch
  );

export default connect(mapStateToProps, mapDispatchToProps)(Dashboard);
