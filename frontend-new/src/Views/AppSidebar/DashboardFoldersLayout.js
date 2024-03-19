import React, {
  Fragment,
  useCallback,
  useEffect,
  useMemo,
  useRef,
  useState
} from 'react';
import cx from 'classnames';
import { Popover } from 'antd';
import { useDispatch, useSelector } from 'react-redux';
import { useHistory } from 'react-router-dom';
import { SVG, Text } from 'Components/factorsComponents';
import {
  selectActiveDashboard,
  selectAllBoardsFolderId,
  selectAreDraftsSelected,
  selectDashboardFoldersListState,
  selectDashboardList,
  selectDashboardListFilteredBySearchText,
  selectDeleteFolderState,
  selectRenameFolderState
} from 'Reducers/dashboard/selectors';
import {
  addDashboardToExistingFolder,
  changeActiveDashboard,
  deleteDashboardFolder,
  renameDashboardFolder
} from 'Reducers/dashboard/services';
import { PathUrls } from 'Routes/pathUrls';
import ControlledComponent from 'Components/ControlledComponent';
import { changeActivePreDashboard } from 'Views/PreBuildDashboard/state/services';
import { ADD_DASHBOARD_MODAL_OPEN } from 'Reducers/types';
import { INITIATE_EDIT_DASHBOARD_DETAILS } from 'Reducers/dashboard/types';
import DashboardSidebarMenuItem from './DashboardSidebarMenuItem';
import styles from './index.module.scss';
import DashboardNewFolderModal from './DashboardNewFolderModal';
import DeleteDashboardFolderModal from './DeleteDashboardFolderModal';

function DashboardFolderButton({
  folder,
  onFolderStateToggle,
  expanded,
  onRenameFolder,
  onDeleteFolder
}) {
  const allBoardsFolderId = useSelector((state) =>
    selectAllBoardsFolderId(state)
  );
  const handleFolderStateToggle = () => {
    onFolderStateToggle(folder.id);
  };

  const handleRenameFolder = (e) => {
    e.stopPropagation();
    onRenameFolder(folder.id);
  };

  const handleDeleteFolder = (e) => {
    e.stopPropagation();
    onDeleteFolder(folder.id);
  };

  const content = (
    <div className='flex flex-col py-2'>
      <button
        type='button'
        className={cx(
          'px-4 py-2 text-left',
          styles['dashboard-more-options-menu']
        )}
        onClick={handleRenameFolder}
      >
        <Text type='title' extraClass='mb-0'>
          Rename folder
        </Text>
      </button>

      <button
        type='button'
        className={cx(
          'px-4 py-2 text-left',
          styles['dashboard-more-options-menu']
        )}
        onClick={handleDeleteFolder}
      >
        <Text type='title' extraClass='mb-0'>
          Delete folder
        </Text>
      </button>
    </div>
  );

  return (
    <button
      key={folder.id}
      type='button'
      onClick={handleFolderStateToggle}
      className={cx(
        'flex col-gap-1 py-2 items-center cursor-pointer',
        styles['dashboard-folder']
      )}
    >
      <SVG
        name={expanded ? 'caretDown' : 'caretRight'}
        color='#8c8c8c'
        size={16}
      />
      <div className='flex justify-between items-center w-full'>
        <Text type='title' color='character-primary' extraClass='mb-0'>
          {folder.name}
        </Text>
        <ControlledComponent controller={allBoardsFolderId !== folder.id}>
          <Popover
            overlayClassName={styles['more-actions-popover']}
            content={content}
            placement='right'
            arrow={false}
          >
            <span className='p-2 rounded hover:bg-gray-300'>
              <SVG
                size={16}
                color='#8C8C8C'
                name='more'
                extraClass={styles['more-actions-icon']}
              />
            </span>
          </Popover>
        </ControlledComponent>
      </div>
    </button>
  );
}

function DashboardItem({
  dashboard,
  setActiveDashboardForFolder,
  onAddDashboardToExistingFolder,
  onDeleteDashboardClick
}) {
  const dispatch = useDispatch();
  const activeDashboardRef = useRef(null);
  const history = useHistory();
  const activeDashboard = useSelector((state) => selectActiveDashboard(state));
  const dashboards = useSelector((state) => selectDashboardList(state));
  const areDraftsSelected = useSelector((state) =>
    selectAreDraftsSelected(state)
  );

  const handleActiveDashboardChange = useCallback(() => {
    const selectedDashboard = dashboards.find((d) => d.id === dashboard.id);
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
  }, [dashboard, dashboards, dispatch, activeDashboard?.id]);

  const isActive =
    activeDashboard?.id === dashboard?.id && areDraftsSelected === false;

  const handleAdditionToNewFolder = useCallback(() => {
    setActiveDashboardForFolder(dashboard.id);
  }, [dashboard.id]);

  const handleAddDashboardToExistingFolder = useCallback(
    (folderId) => {
      onAddDashboardToExistingFolder(folderId, dashboard.id);
    },
    [dashboard.id, onAddDashboardToExistingFolder]
  );

  const handleEditDashboardDetails = useCallback(() => {
    dispatch({ type: INITIATE_EDIT_DASHBOARD_DETAILS, payload: { dashboard } });
    dispatch({ type: ADD_DASHBOARD_MODAL_OPEN });
  }, [dashboard]);

  const handleDeleteDashboardClick = useCallback(() => {
    onDeleteDashboardClick(dashboard);
  }, [dashboard]);

  useEffect(() => {
    if (!isActive && activeDashboard.class === 'predefined') {
      const preDashboard = dashboards.filter((db) => db.class === 'predefined');
      dispatch(changeActivePreDashboard(preDashboard[0]));
      dispatch(changeActiveDashboard(preDashboard[0]));
    }
  }, [dashboards, dispatch, isActive]);

  useEffect(() => {
    activeDashboardRef.current?.scrollIntoView();
  }, []);

  return (
    <DashboardSidebarMenuItem
      activeDashboardRef={activeDashboardRef}
      text={dashboard.name}
      onClick={handleActiveDashboardChange}
      isActive={isActive}
      onAdditionToNewFolder={handleAdditionToNewFolder}
      onAddDashboardToExistingFolder={handleAddDashboardToExistingFolder}
      onEditDashboardDetails={handleEditDashboardDetails}
      onDeleteDashboardClick={handleDeleteDashboardClick}
    />
  );
}

function DashboardFoldersLayout({
  searchText,
  setActiveDashboardForFolder,
  onDeleteDashboardClick
}) {
  const dispatch = useDispatch();
  const dashboardFolders = useSelector((state) =>
    selectDashboardFoldersListState(state)
  );
  const { active_project } = useSelector((state) => state.global);
  const allBoardsFolderId = useSelector((state) =>
    selectAllBoardsFolderId(state)
  );
  const filteredDashboardList = useSelector((state) =>
    selectDashboardListFilteredBySearchText(state, searchText)
  );
  const renameFolderState = useSelector((state) =>
    selectRenameFolderState(state)
  );
  const deleteFolderState = useSelector((state) =>
    selectDeleteFolderState(state)
  );
  const activeDashboard = useSelector((state) => selectActiveDashboard(state));
  const { data: dashboardFoldersList } = dashboardFolders;

  const [expandedFolders, setExpandedFolders] = useState({});
  const [renameFolderId, setRenameFolderId] = useState(null);
  const [deleteFolderId, setDeleteFolderId] = useState(null);

  const handleFolderStateToggle = useCallback((folderId) => {
    setExpandedFolders((curr) => ({
      ...curr,
      [folderId]: Boolean(curr[folderId]) !== true
    }));
  }, []);

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

  const hideRenameFolder = useCallback(() => {
    setRenameFolderId(null);
  }, []);

  const hideDeleteFolderModal = useCallback(() => {
    setDeleteFolderId(null);
  }, []);

  const handleRenameFolder = useCallback((folderId) => {
    setRenameFolderId(folderId);
  }, []);

  const handleRenameFolderSubmit = useCallback(
    (name) => {
      dispatch(renameDashboardFolder(active_project.id, renameFolderId, name));
    },
    [active_project.id, renameFolderId]
  );

  const handleDeleteFolder = useCallback((folderId) => {
    setDeleteFolderId(folderId);
  }, []);

  const handleFolderDeleteSubmit = useCallback(() => {
    dispatch(deleteDashboardFolder(active_project.id, deleteFolderId));
  }, [deleteFolderId, active_project.id]);

  useEffect(() => {
    const currentActiveDashboardFolder = dashboardFoldersList.find((folder) =>
      folder.dashboardIds.find((dId) => dId === activeDashboard.id)
    );
    if (currentActiveDashboardFolder != null) {
      setExpandedFolders({ [currentActiveDashboardFolder.id]: true });
    } else {
      setExpandedFolders({ [allBoardsFolderId]: true });
    }
  }, [allBoardsFolderId, dashboardFoldersList, activeDashboard?.id]);

  useEffect(() => {
    if (renameFolderState.completed === true) {
      setRenameFolderId(null);
    }
  }, [renameFolderState.completed]);

  useEffect(() => {
    if (deleteFolderState.completed === true) {
      setDeleteFolderId(null);
    }
  }, [deleteFolderState.completed]);

  if (dashboardFoldersList.length === 0) {
    return (
      <>
        {filteredDashboardList.map((dashboard) => (
          <DashboardItem
            setActiveDashboardForFolder={setActiveDashboardForFolder}
            dashboard={dashboard}
            key={dashboard.id}
            onAddDashboardToExistingFolder={handleAddDashboardToExistingFolder}
            onDeleteDashboardClick={onDeleteDashboardClick}
          />
        ))}
      </>
    );
  }

  const allBoardsFolderExpanded = expandedFolders[allBoardsFolderId] === true;
  const deleteFolderName =
    deleteFolderId != null
      ? foldersListWithoutAllBoards.find(
          (folder) => folder.id === deleteFolderId
        )?.name
      : '';

  return (
    <>
      <DashboardFolderButton
        folder={allBoardsFolder}
        onFolderStateToggle={handleFolderStateToggle}
        expanded={allBoardsFolderExpanded}
      />
      <ControlledComponent controller={allBoardsFolderExpanded}>
        {dashboardsByFolderId[allBoardsFolderId]?.map((dashboard) => (
          <DashboardItem
            setActiveDashboardForFolder={setActiveDashboardForFolder}
            dashboard={dashboard}
            key={dashboard.id}
            onAddDashboardToExistingFolder={handleAddDashboardToExistingFolder}
            onDeleteDashboardClick={onDeleteDashboardClick}
          />
        ))}
      </ControlledComponent>
      {foldersListWithoutAllBoards.map((folder) => (
        <Fragment key={folder.id}>
          <DashboardFolderButton
            folder={folder}
            onFolderStateToggle={handleFolderStateToggle}
            expanded={expandedFolders[folder.id]}
            onRenameFolder={handleRenameFolder}
            onDeleteFolder={handleDeleteFolder}
          />
          <ControlledComponent controller={expandedFolders[folder.id] === true}>
            {dashboardsByFolderId[folder.id]?.map((dashboard) => (
              <DashboardItem
                setActiveDashboardForFolder={setActiveDashboardForFolder}
                dashboard={dashboard}
                key={dashboard.id}
                onAddDashboardToExistingFolder={
                  handleAddDashboardToExistingFolder
                }
                onDeleteDashboardClick={onDeleteDashboardClick}
              />
            ))}
          </ControlledComponent>
          <ControlledComponent
            controller={
              expandedFolders[folder.id] === true &&
              dashboardsByFolderId[folder.id].length === 0
            }
          >
            <Text
              level={8}
              color='character-secondary'
              type='title'
              extraClass='mb-0 text-center'
            >
              No dashboards in this Folder
            </Text>
          </ControlledComponent>
        </Fragment>
      ))}
      <DashboardNewFolderModal
        handleCancel={hideRenameFolder}
        visible={renameFolderId != null}
        handleSubmit={handleRenameFolderSubmit}
        isLoading={renameFolderState.loading}
        renameFolder
      />
      <DeleteDashboardFolderModal
        visible={deleteFolderId !== null}
        onOk={handleFolderDeleteSubmit}
        onCancel={hideDeleteFolderModal}
        folderName={deleteFolderName}
        isLoading={deleteFolderState.loading}
      />
    </>
  );
}

export default DashboardFoldersLayout;
