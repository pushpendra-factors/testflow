import React, { memo, useCallback, useEffect, useState } from 'react';
import { useHistory } from 'react-router-dom';
import { useDispatch, useSelector } from 'react-redux';
import cx from 'classnames';
import { Button } from 'antd';
import { SVG, Text } from 'Components/factorsComponents';
import {
  selectActiveDashboard,
  selectAreDraftsSelected,
  selectShowDashboardNewFolderModal,
  selectNewFolderCreationState,
  selectDeleteDashboardState
} from 'Reducers/dashboard/selectors';

import {
  addDashboardToNewFolder,
  deleteDashboardAction
} from 'Reducers/dashboard/services';
import { NEW_DASHBOARD_TEMPLATES_MODAL_OPEN } from 'Reducers/types';
import {
  makeDraftsActiveAction,
  toggleNewFolderModal
} from 'Reducers/dashboard/actions';
import ConfirmationModal from 'Components/ConfirmationModal';
import { PathUrls } from 'Routes/pathUrls';
import SidebarSearch from './SidebarSearch';
import DashboardNewFolderModal from './DashboardNewFolderModal';
import styles from './index.module.scss';
import DashboardFoldersLayout from './DashboardFoldersLayout';
import { PlusOutlined } from '@ant-design/icons';

function DashboardSidebar() {
  const dispatch = useDispatch();
  const history = useHistory();
  const [searchText, setSearchText] = useState('');
  const [deleteDashboardModal, setDeleteDashboardModal] = useState(false);
  const [activeDashboardForFolder, setActiveDashboardForFolder] =
    useState(null);
  const areDraftsSelected = useSelector((state) =>
    selectAreDraftsSelected(state)
  );
  const deleteDashboardState = useSelector((state) =>
    selectDeleteDashboardState(state)
  );
  const newFolderCreationState = useSelector((state) =>
    selectNewFolderCreationState(state)
  );
  const { active_project } = useSelector((state) => state.global);
  const showNewFolderModal = useSelector((state) =>
    selectShowDashboardNewFolderModal(state)
  );
  const queries = useSelector((state) => state.queries.data);

  const activeDashboard = useSelector((state) => selectActiveDashboard(state));

  const hideDashboardNewFolderModal = useCallback(() => {
    setActiveDashboardForFolder(null);
    dispatch(toggleNewFolderModal(false));
  });

  const handleDraftsClick = () => {
    if (activeDashboard?.class === 'predefined') {
      history.replace(`${PathUrls.Dashboard}`);
    }
    dispatch(makeDraftsActiveAction());
  };

  const onNewFolderCreation = useCallback(
    (folderName) => {
      dispatch(
        addDashboardToNewFolder(
          active_project.id,
          activeDashboardForFolder,
          folderName
        )
      );
    },
    [activeDashboardForFolder, active_project.id]
  );

  const handleDeleteDashboardClick = useCallback((dashboard) => {
    setDeleteDashboardModal(true);
    setActiveDashboardForFolder(dashboard);
  }, []);

  const confirmDeleteDashboard = useCallback(() => {
    dispatch(
      deleteDashboardAction(active_project.id, activeDashboardForFolder.id)
    );
  }, [active_project?.id, activeDashboardForFolder?.id]);

  const closeDeleteModal = useCallback(() => {
    setDeleteDashboardModal(false);
    setActiveDashboardForFolder(null);
  }, []);

  useEffect(() => {
    if (deleteDashboardState.completed === true) {
      closeDeleteModal();
    }
  }, [closeDeleteModal, deleteDashboardState.completed]);

  useEffect(() => {
    if (newFolderCreationState.completed === true) {
      hideDashboardNewFolderModal();
    }
  }, [newFolderCreationState.completed, hideDashboardNewFolderModal]);

  return (
    <div className='flex flex-col gap-y-2'>
      <button
        type='button'
        onClick={handleDraftsClick}
        className='px-4 w-full cursor-pointer'
      >
        <div
          className={cx(
            'flex gap-x-1 cursor-pointer py-2 rounded-md items-center w-full px-2',
            styles['draft-title'],
            {
              [styles['item-active']]: areDraftsSelected
            }
          )}
        >
          <SVG name='drafts' />
          <Text
            color='character-primary'
            type='title'
            level={7}
            extraClass='mb-0'
          >
            Drafts
          </Text>
          <Text
            type='title'
            extraClass='mb-0'
            level={8}
            color='character-secondary'
          >
            {queries.length}
          </Text>
        </div>
      </button>
      <div className='flex flex-col gap-y-5 px-4'>
        <SidebarSearch
          placeholder='Search board'
          setSearchText={setSearchText}
          searchText={searchText}
        />
        <div
          className={cx(
            'flex flex-col gap-y-1 overflow-auto',
            styles['dashboard-list-container']
          )}
        >
          <DashboardFoldersLayout
            searchText={searchText}
            setActiveDashboardForFolder={setActiveDashboardForFolder}
            onDeleteDashboardClick={handleDeleteDashboardClick}
          />
        </div>
        <Button
          onClick={() => {
            dispatch({ type: NEW_DASHBOARD_TEMPLATES_MODAL_OPEN });
          }}
          type='dashed'
          icon={<PlusOutlined />}
        >
          New Dashboard
        </Button>
      </div>
      <DashboardNewFolderModal
        handleCancel={hideDashboardNewFolderModal}
        visible={showNewFolderModal}
        handleSubmit={onNewFolderCreation}
        isLoading={newFolderCreationState.loading}
      />
      <ConfirmationModal
        visible={deleteDashboardModal}
        confirmationText='Are you sure you want to delete this Dashboard?'
        onOk={confirmDeleteDashboard}
        onCancel={closeDeleteModal}
        title={`Delete Dashboard - ${activeDashboardForFolder?.name}`}
        okText='Confirm'
        cancelText='Cancel'
        confirmLoading={deleteDashboardState.loading}
      />
    </div>
  );
}

export default memo(DashboardSidebar);
