import React, { useCallback, useEffect, useMemo, useState } from 'react';
import { connect, useDispatch, useSelector } from 'react-redux';
import cx from 'classnames';
import { Button, message, notification } from 'antd';
import { SVG as Svg, Text } from 'Components/factorsComponents';
import {
  setAccountPayloadAction,
  setDrawerVisibleAction,
  setNewSegmentModeAction,
  toggleAccountsTab
} from 'Reducers/accountProfilesView/actions';
import { selectAccountPayload } from 'Reducers/accountProfilesView/selectors';
import { selectSegments } from 'Reducers/timelines/selectors';
import { useHistory } from 'react-router-dom';
import { reorderDefaultDomainSegmentsToTop } from 'Components/Profile/AccountProfiles/accountProfiles.helpers';
import { GROUP_NAME_DOMAINS } from 'Components/GlobalFilter/FilterWrapper/utils';
import { PathUrls } from 'Routes/pathUrls';
import FolderStructure from 'Components/FolderStructure';
import RenameSegmentModal from 'Components/Profile/AccountProfiles/RenameSegmentModal';
import {
  deleteSegment,
  getSavedSegments,
  getSegmentFolders,
  updateSegmentForId
} from 'Reducers/timelines/middleware';
import DeleteSegmentModal from 'Components/Profile/AccountProfiles/DeleteSegmentModal';
import { INITIAL_ACCOUNT_PAYLOAD } from 'Reducers/accountProfilesView';
import {
  deleteSegmentFolders,
  moveSegmentToNewFolder,
  renameSegmentFolders,
  updateSegmentToFolder
} from 'Reducers/timelines';
import { LoadingOutlined } from '@ant-design/icons';
import { bindActionCreators } from 'redux';
import { defaultSegmentIconsMapping } from './appSidebar.constants';
import styles from './index.module.scss';

export const SegmentIcon = (name) =>
  defaultSegmentIconsMapping[name] || 'pieChart';

function AccountsSidebar({
  getSegmentFolders,
  getSavedSegments,
  updateSegmentForId,
  deleteSegment
}) {
  const history = useHistory();
  const dispatch = useDispatch();
  const segments = useSelector(selectSegments);
  const { segmentFolders } = useSelector((state) => state.timelines);
  const { active_project } = useSelector((state) => state.global);
  const activeAccountPayload = useSelector(selectAccountPayload);
  const activeSegment = activeAccountPayload?.segment;

  const segmentsList = useMemo(
    () => reorderDefaultDomainSegmentsToTop(segments[GROUP_NAME_DOMAINS]) || [],
    [segments]
  );
  const [modalState, setModalState] = useState({
    rename: false,
    delete: false,
    unit: null
  });

  useEffect(() => {
    getSegmentFolders(active_project?.id, 'account');
    // need to add segment folders for people too
  }, []);

  const setAccountPayload = (payload) => {
    dispatch(setAccountPayloadAction(payload));
    if (payload?.segment?.id) {
      dispatch(setNewSegmentModeAction(false));
    }
  };

  const handleRenameSegment = async (name) => {
    try {
      await updateSegmentForId(active_project.id, modalState?.unit.id, {
        name
      });
      await getSavedSegments(active_project.id);
      notification.success({
        message: 'Segment renamed successfully',
        duration: 3
      });
    } catch (error) {
      notification.error({
        message: 'Segment rename failed',
        duration: 3
      });
    } finally {
      setModalState((prev) => ({
        ...prev,
        rename: false,
        unit: null
      }));
    }
  };

  const handleDeleteSegment = async () => {
    await deleteSegment({
      projectId: active_project.id,
      segmentId: modalState.unit?.id
    });
    await getSavedSegments(active_project.id);

    notification.success({
      message: 'Segment deleted successfully',
      duration: 5
    });

    setModalState((prev) => ({
      ...prev,
      delete: false,
      unit: null
    }));
    setAccountPayload(INITIAL_ACCOUNT_PAYLOAD);
    history.replace(PathUrls.ProfileAccounts);
  };

  const moveSegmentToFolder = (event, folderID, segmentID) => {
    updateSegmentToFolder(
      active_project.id,
      segmentID,
      {
        folder_id: folderID
      },
      'account'
    )
      .then(async () => {
        await getSavedSegments(active_project.id);
        message.success('Segment Moved');
      })
      .catch((err) => {
        console.error(err);
        message.error('Segment failed to move');
      });
  };

  const handleMoveToNewFolder = (segmentID, folder_name) => {
    moveSegmentToNewFolder(
      active_project.id,
      segmentID,
      {
        name: folder_name
      },
      'account'
    )
      .then(async () => {
        getSegmentFolders(active_project.id, 'account');
        await getSavedSegments(active_project.id);
        message.success('Segment Moved to New Folder');
      })
      .catch((err) => {
        console.error(err);
        message.error('Failed to move segment');
      });
  };
  const handleRenameFolder = (folderId, name) => {
    renameSegmentFolders(active_project.id, folderId, { name }, 'account')
      .then(async () => {
        getSegmentFolders(active_project.id, 'account');
        message.success('Folder Renamed');
      })
      .catch((err) => {
        console.error(err);
      });
  };

  const handleDeleteFolder = (folderId) => {
    deleteSegmentFolders(active_project.id, folderId, 'account')
      .then(async () => {
        getSegmentFolders(active_project.id, 'account');
        await getSavedSegments(active_project.id);
        message.success('Folder Deleted');
      })
      .catch((err) => {
        console.error(err);
        message.success('Folder to Delete');
      });
  };

  const changeActiveSegment = (segment) => {
    dispatch(setDrawerVisibleAction(false));
    dispatch(setNewSegmentModeAction(false));
    dispatch(setAccountPayloadAction({ source: GROUP_NAME_DOMAINS, segment }));
    history.replace({ pathname: `/accounts/segments/${segment.id}` });
  };

  const setActiveSegment = (segment) => {
    if (activeSegment?.id !== segment?.id) {
      changeActiveSegment(segment);
    }
  };
  const handleEditUnit = useCallback((unit) => {
    setModalState((prev) => ({ ...prev, rename: true, unit }));
  }, []);
  const handleDeleteUnit = useCallback((unit) => {
    setModalState((prev) => ({ ...prev, delete: true, unit }));
  }, []);
  const cancelRenameModal = useCallback(() => {
    setModalState((prev) => ({ ...prev, rename: false, unit: null }));
  }, []);
  const cancelDeleteModal = useCallback(() => {
    setModalState((prev) => ({ ...prev, delete: false, unit: null }));
  }, []);
  return (
    <div className='flex flex-col gap-y-5'>
      <div
        className={cx(
          'flex flex-col gap-y-6 overflow-auto',
          styles['accounts-list-container']
        )}
      >
        {segmentFolders.loading ? (
          <LoadingOutlined />
        ) : (
          <FolderStructure
            folders={segmentFolders.accounts}
            items={segmentsList}
            unit='segment'
            onRenameFolder={handleRenameFolder}
            onDeleteFolder={handleDeleteFolder}
            onUnitClick={setActiveSegment}
            handleNewFolder={handleMoveToNewFolder}
            handleEditUnit={handleEditUnit}
            handleDeleteUnit={handleDeleteUnit}
            moveToExistingFolder={moveSegmentToFolder}
          />
        )}
      </div>
      <div className='px-4'>
        <Button
          className={cx(
            'flex gap-x-2 items-center w-full',
            styles.sidebar_action_button
          )}
          onClick={() => {
            history.replace(PathUrls.ProfileAccounts);
            dispatch(toggleAccountsTab('accounts'));
            dispatch(setNewSegmentModeAction(true));
            dispatch(setAccountPayloadAction({}));
          }}
        >
          <Svg
            name='plus'
            size={16}
            extraClass={styles.sidebar_action_button__content}
            isFill={false}
          />
          <Text
            level={6}
            type='title'
            extraClass={cx('m-0', styles.sidebar_action_button__content)}
          >
            New Segment
          </Text>
        </Button>
      </div>

      <RenameSegmentModal
        segmentName={modalState?.unit?.name}
        visible={modalState.rename}
        onCancel={cancelRenameModal}
        handleSubmit={handleRenameSegment}
      />
      <DeleteSegmentModal
        segmentName={modalState?.unit?.name}
        visible={modalState.delete}
        onCancel={cancelDeleteModal}
        onOk={handleDeleteSegment}
      />
    </div>
  );
}
const mapDispatchToProps = (dispatch) =>
  bindActionCreators(
    {
      getSegmentFolders,
      updateSegmentForId,
      getSavedSegments,
      deleteSegment
    },
    dispatch
  );
export default connect(null, mapDispatchToProps)(AccountsSidebar);
