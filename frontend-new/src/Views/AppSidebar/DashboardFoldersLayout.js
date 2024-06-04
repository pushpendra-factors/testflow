import React, { useCallback, useEffect, useMemo, useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { useHistory } from 'react-router-dom';
import {
  selectActiveDashboard,
  selectAllBoardsFolderId,
  selectDashboardFoldersListState,
  selectDashboardList,
  selectDeleteFolderState
} from 'Reducers/dashboard/selectors';
import {
  addDashboardToExistingFolder,
  addDashboardToNewFolder,
  changeActiveDashboard,
  deleteDashboardFolder,
  renameDashboardFolder
} from 'Reducers/dashboard/services';
import { PathUrls } from 'Routes/pathUrls';
import { changeActivePreDashboard } from 'Views/PreBuildDashboard/state/services';
import { ADD_DASHBOARD_MODAL_OPEN } from 'Reducers/types';
import { INITIATE_EDIT_DASHBOARD_DETAILS } from 'Reducers/dashboard/types';
import FolderStructure from 'Components/FolderStructure';
import { LoadingOutlined } from '@ant-design/icons';

function DashboardFoldersLayout({ onDeleteDashboardClick }) {
  const dispatch = useDispatch();
  const history = useHistory();
  const { foldersList } = useSelector((state) => state.dashboard);
  const dashboardsState = useSelector((state) => state.dashboard.dashboards);
  const dashboards = useSelector((state) => selectDashboardList(state));
  const dashboardFolders = useSelector((state) =>
    selectDashboardFoldersListState(state)
  );
  const { active_project } = useSelector((state) => state.global);
  const allBoardsFolderId = useSelector((state) =>
    selectAllBoardsFolderId(state)
  );
  const filteredDashboardList = useSelector(
    (state) => state.dashboard.dashboards.data
  );

  const deleteFolderState = useSelector((state) =>
    selectDeleteFolderState(state)
  );
  const activeDashboard = useSelector((state) => selectActiveDashboard(state));
  const { data: dashboardFoldersList } = dashboardFolders;

  const [deleteFolderId, setDeleteFolderId] = useState(null);

  const onNewFolderCreation = useCallback(
    (dashboardID, folderName) => {
      dispatch(
        addDashboardToNewFolder(active_project.id, dashboardID, folderName)
      );
    },
    [active_project.id]
  );

  const handleAddDashboardToExistingFolder = useCallback(
    (folderId, dashboardId) => {
      dispatch(
        addDashboardToExistingFolder(active_project.id, folderId, dashboardId)
      );
    },
    [active_project.id]
  );

  const allBoardsFolder = useMemo(
    () =>
      dashboardFoldersList.find((folder) => folder.id === allBoardsFolderId),
    [dashboardFoldersList, allBoardsFolderId]
  );

  const foldersListWithoutAllBoards = useMemo(
    () =>
      dashboardFoldersList.filter((folder) => folder.id !== allBoardsFolderId),
    [dashboardFoldersList, allBoardsFolderId]
  );

  const dashboardsByFolderId = useMemo(
    () =>
      dashboardFoldersList.reduce((prev, folder) => {
        const folderId = folder.id;
        const { dashboardIds } = folder;
        const dashboardsList = dashboardIds.map((dashboardId) =>
          filteredDashboardList.find(
            (dashboard) => dashboard.id === dashboardId
          )
        );
        return {
          ...prev,
          [folderId]: dashboardsList.filter((dashboard) => dashboard)
        };
      }, {}),
    [dashboardFoldersList, filteredDashboardList]
  );

  const handleRenameFolderSubmit = useCallback(
    (folderId, name) => {
      dispatch(renameDashboardFolder(active_project.id, folderId, name));
    },
    [active_project.id]
  );

  const handleFolderDeleteSubmit = useCallback(
    (folderId) => {
      dispatch(deleteDashboardFolder(active_project.id, folderId));
    },
    [deleteFolderId, active_project.id]
  );

  useEffect(() => {
    if (deleteFolderState.completed === true) {
      setDeleteFolderId(null);
    }
  }, [deleteFolderState.completed]);

  const getAllDashboardsList = () => {
    const allUnits = [];
    Object.keys(dashboardsByFolderId).forEach((eachFolderId) => {
      let folder = foldersListWithoutAllBoards.find(
        (e) => e.id === eachFolderId
      );
      if (!folder && eachFolderId === allBoardsFolder.id) {
        folder = allBoardsFolder;
      }
      if (eachFolderId in dashboardsByFolderId)
        dashboardsByFolderId[eachFolderId].forEach((eachUnit) => {
          allUnits.push({ ...eachUnit, folder_id: eachFolderId });
        });
    });
    return allUnits;
  };
  const allDashboardsList = useMemo(
    () => getAllDashboardsList(),
    [dashboardsByFolderId]
  );

  const handleActiveDashboardChange = (dashboardId) => {
    const selectedDashboard = dashboards.find((d) => d.id === dashboardId);
    if (selectedDashboard.id === activeDashboard.id) {
      return;
    }
    if (selectedDashboard.class === 'predefined') {
      history.replace(`${PathUrls.PreBuildDashboard}`);
      dispatch(changeActivePreDashboard(selectedDashboard));
    } else {
      history.replace(`${PathUrls.Dashboard}/${selectedDashboard.id}`);
    }
    dispatch(changeActiveDashboard(selectedDashboard));
  };

  return foldersList.loading ||
    dashboardsState.loading ||
    !foldersList.completed ? (
    <LoadingOutlined style={{ fontSize: '24px', margin: '24px 0' }} />
  ) : (
    <FolderStructure
      folders={[
        ...foldersListWithoutAllBoards,
        { ...allBoardsFolder, isAllBoard: true }
      ]}
      items={allDashboardsList}
      unit='dashboard'
      handleNewFolder={onNewFolderCreation}
      moveToExistingFolder={(e, folderid, dashboardid) => {
        handleAddDashboardToExistingFolder(folderid, dashboardid);
      }}
      handleEditUnit={(unit) => {
        dispatch({
          type: INITIATE_EDIT_DASHBOARD_DETAILS,
          payload: { dashboard: unit }
        });
        dispatch({ type: ADD_DASHBOARD_MODAL_OPEN });
      }}
      handleDeleteUnit={onDeleteDashboardClick}
      onUnitClick={(unit) => {
        handleActiveDashboardChange(unit.id);
      }}
      onRenameFolder={handleRenameFolderSubmit}
      onDeleteFolder={handleFolderDeleteSubmit}
    />
  );
}

export default DashboardFoldersLayout;
