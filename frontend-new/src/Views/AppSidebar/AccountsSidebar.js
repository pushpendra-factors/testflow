import React, { useCallback, useEffect, useMemo, useState } from 'react';
import { connect, useDispatch, useSelector } from 'react-redux';
import cx from 'classnames';
import { Button, Spin, message, notification } from 'antd';
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
import {
  defaultSegmentsList,
  reorderDefaultDomainSegmentsToTop
} from 'Components/Profile/AccountProfiles/accountProfiles.helpers';
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
    () =>
      reorderDefaultDomainSegmentsToTop(segments?.[GROUP_NAME_DOMAINS]) || [],
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
    const loadingMessageHandle = message.loading('Renaming Segment', 0);
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
      loadingMessageHandle();
      setModalState((prev) => ({
        ...prev,
        rename: false,
        unit: null
      }));
    }
  };

  const handleDeleteSegment = async () => {
    const loadingMessageHandle = message.loading('Deleting Segment', 0);
    try {
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
    } catch (error) {
      message.error('Failed to Delete Segment');
    } finally {
      loadingMessageHandle();
    }
  };
  const moveSegmentToFolder = async (event, folderID, segmentID) => {
    const loadingMessageHandle = message.loading('Moving Segment to Folder', 0);
    try {
      await updateSegmentToFolder(
        active_project.id,
        segmentID,
        {
          folder_id: folderID
        },
        'account'
      );
      await getSavedSegments(active_project.id);
      message.success('Segment Moved');
    } catch (err) {
      console.error(err);
      message.error('Segment failed to move');
    } finally {
      loadingMessageHandle();
    }
  };

  const handleMoveToNewFolder = async (segmentID, folder_name) => {
    const loadingMessageHandle = message.loading(
      `Moving Segment to \`${folder_name}\` Folder`,
      0
    );
    try {
      await moveSegmentToNewFolder(
        active_project.id,
        segmentID,
        {
          name: folder_name
        },
        'account'
      );
      getSegmentFolders(active_project.id, 'account');
      await getSavedSegments(active_project.id);
      message.success('Segment Moved to New Folder');
    } catch (err) {
      console.error(err);
      getSegmentFolders(active_project.id, 'account');
      message.error('Failed to move segment');
    } finally {
      loadingMessageHandle();
    }
  };
  const handleRenameFolder = async (folderId, name) => {
    const loadingMessageHandle = message.loading('Renaming Folder', 0);
    try {
      await renameSegmentFolders(
        active_project.id,
        folderId,
        { name },
        'account'
      );
      getSegmentFolders(active_project?.id, 'account');
      message.success('Folder Renamed');
    } catch (error) {
      console.error(error);
      message.error(
        error?.data?.err?.code || 'Failed to Rename Segment Folders'
      );
    } finally {
      loadingMessageHandle();
    }
  };

  const handleDeleteFolder = async (folderId) => {
    const loadingMessageHandle = message.loading('Deleting Folder', 0);
    try {
      await deleteSegmentFolders(active_project.id, folderId, 'account');
      await getSegmentFolders(active_project.id, 'account');
      await getSavedSegments(active_project.id);
      message.success('Folder Deleted');
    } catch (error) {
      message.success(error?.data?.err?.code || 'Failed to Delete Folder');
    } finally {
      loadingMessageHandle();
    }
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
        {segmentFolders?.isLoading ? (
          <Spin />
        ) : (
          <FolderStructure
            folders={segmentFolders?.accounts || []}
            items={segmentsList}
            active_item={activeSegment?.id}
            unit='segment'
            onRenameFolder={handleRenameFolder}
            onDeleteFolder={handleDeleteFolder}
            onUnitClick={setActiveSegment}
            handleNewFolder={handleMoveToNewFolder}
            handleEditUnit={handleEditUnit}
            handleDeleteUnit={handleDeleteUnit}
            moveToExistingFolder={moveSegmentToFolder}
            showItemIcons
            hideItemOptionsList={defaultSegmentsList}
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
