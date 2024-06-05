import React, { useCallback, useEffect, useState } from 'react';
import cx from 'classnames';
import { connect, useDispatch, useSelector } from 'react-redux';
import { Button, message, notification } from 'antd';
import { getUserOptionsForDropdown } from 'Components/Profile/UserProfiles/userProfiles.helpers';
import { SVG, Text } from 'Components/factorsComponents';
import {
  selectSegmentsList,
  selectTimelinePayload
} from 'Reducers/userProfilesView/selectors';
import {
  setNewSegmentModeAction,
  setTimelinePayloadAction
} from 'Reducers/userProfilesView/actions';
import FolderStructure from 'Components/FolderStructure';
import {
  deleteSegment,
  getSavedSegments,
  getSegmentFolders,
  updateSegmentForId
} from 'Reducers/timelines/middleware';
import {
  deleteSegmentFolders,
  moveSegmentToNewFolder,
  renameSegmentFolders,
  updateSegmentToFolder
} from 'Reducers/timelines';
import DeleteSegmentModal from 'Components/Profile/AccountProfiles/DeleteSegmentModal';
import RenameSegmentModal from 'Components/Profile/AccountProfiles/RenameSegmentModal';
import { bindActionCreators } from 'redux';
import logger from 'Utils/logger';
import { ProfilesSidebarIconsMapping } from './appSidebar.constants';
import SidebarMenuItem from './SidebarMenuItem';
import styles from './index.module.scss';

function GroupItem({ group }) {
  const dispatch = useDispatch();
  const timelinePayload = useSelector((state) => selectTimelinePayload(state));
  const { newSegmentMode } = useSelector((state) => state.userProfilesView);

  const setTimelinePayload = () => {
    if (timelinePayload.source !== group[1] || newSegmentMode === true) {
      dispatch(
        setTimelinePayloadAction({
          source: group[1],
          segment: {}
        })
      );
    }
  };

  const isActive =
    timelinePayload.source === group[1] &&
    !timelinePayload.segment.id &&
    newSegmentMode === false;

  return (
    <SidebarMenuItem
      text={group[0]}
      isActive={isActive}
      onClick={setTimelinePayload}
      icon={ProfilesSidebarIconsMapping[group[1]]}
    />
  );
}

function ProfilesSidebar({
  getSegmentFolders,
  getSavedSegments,
  deleteSegment,
  updateSegmentForId
}) {
  const dispatch = useDispatch();
  const userOptions = getUserOptionsForDropdown();
  const timelinePayload = useSelector((state) => selectTimelinePayload(state));
  const { segmentFolders } = useSelector((state) => state.timelines);
  const { active_project } = useSelector((state) => state.global);
  const activeSegment = timelinePayload?.segment;
  const [modalState, setModalState] = useState({
    rename: false,
    delete: false,
    unit: null
  });

  const userSegmentsList = useSelector((state) => selectSegmentsList(state));
  useEffect(() => {
    getSegmentFolders(active_project?.id, 'user');
    // need to add segment folders for people too
  }, []);
  const handleMoveToNewFolder = (segmentID, folder_name) => {
    moveSegmentToNewFolder(
      active_project.id,
      segmentID,
      {
        name: folder_name
      },
      'user'
    )
      .then(async () => {
        getSegmentFolders(active_project.id, 'user');
        await getSavedSegments(active_project.id);
        message.success('Segment Moved to New Folder');
      })
      .catch((err) => {
        console.error(err);
        message.error('Failed to move segment');
      });
  };
  const moveSegmentToFolder = (event, folderID, segmentID) => {
    updateSegmentToFolder(
      active_project.id,
      segmentID,
      {
        folder_id: folderID
      },
      'user'
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
  const handleRenameFolder = (folderId, name) => {
    renameSegmentFolders(active_project.id, folderId, { name }, 'user')
      .then(async () => {
        getSegmentFolders(active_project.id, 'user');
        message.success('Folder Renamed');
      })
      .catch((err) => {
        console.error(err);
      });
  };
  const handleDeleteFolder = (folderId) => {
    deleteSegmentFolders(active_project.id, folderId, 'user')
      .then(async () => {
        getSegmentFolders(active_project.id, 'user');
        await getSavedSegments(active_project.id);
        message.success('Folder Deleted');
      })
      .catch((err) => {
        console.error(err);
        message.error('Folder to Delete');
      });
  };

  const changeActiveSegment = (segment) => {
    const opts = { ...timelinePayload };
    opts.source = segment?.type;
    opts.segment = segment;
    delete opts.search_filter;
    dispatch(setTimelinePayloadAction(opts));
  };

  const setActiveSegment = (segment) => {
    if (activeSegment?.id !== segment?.id) {
      changeActiveSegment(segment);
    }
  };

  const handleRenameSegment = async (name) => {
    try {
      const segmentId = modalState.unit?.id;

      await updateSegmentForId(active_project.id, segmentId, { name });
      getSavedSegments(active_project.id);

      setModalState({ rename: false, delete: false, unit: null });
      notification.success({
        message: 'Segment renamed successfully',
        duration: 5
      });
    } catch (error) {
      logger.error(error);
    }
  };
  const handleDeleteSegment = () => {
    deleteSegment({
      projectId: active_project.id,
      segmentId: modalState.unit?.id
    })
      .then(() => {
        setModalState({ rename: false, delete: false, unit: null });
        notification.success({
          message: 'Segment deleted successfully',
          duration: 5
        });
      })
      .finally(() => {
        dispatch(
          setTimelinePayloadAction({
            source: 'All',
            segment: {}
          })
        );
      });
  };
  const handleCancelModal = useCallback(() => {
    setModalState({ delete: false, rename: false, unit: null });
  }, []);
  return (
    <div className='flex flex-col gap-y-5'>
      <div
        className={cx(
          'flex flex-col gap-y-1 overflow-auto',
          styles['accounts-list-container']
        )}
      >
        <div className='px-2'>
          {userOptions.slice(1).map((option) => (
            <GroupItem key={option[0]} group={option} />
          ))}
        </div>
        <FolderStructure
          folders={segmentFolders.peoples}
          items={userSegmentsList?.sort((a, b) => a.name.localeCompare(b.name))}
          unit='segment'
          handleNewFolder={handleMoveToNewFolder}
          moveToExistingFolder={moveSegmentToFolder}
          onRenameFolder={handleRenameFolder}
          onDeleteFolder={handleDeleteFolder}
          onUnitClick={setActiveSegment}
          handleEditUnit={(unit) => {
            setModalState({ rename: true, unit, delete: false });
          }}
          handleDeleteUnit={(unit) => {
            setModalState({ rename: false, unit, delete: true });
          }}
        />
      </div>{' '}
      <div className='px-4'>
        <Button
          className={cx(
            'flex gap-x-2 items-center w-full',
            styles.sidebar_action_button
          )}
          onClick={() => {
            dispatch(setNewSegmentModeAction(true));
          }}
        >
          <SVG
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
      <DeleteSegmentModal
        segmentName={modalState?.unit?.name}
        visible={modalState.delete}
        onCancel={handleCancelModal}
        onOk={handleDeleteSegment}
      />
      <RenameSegmentModal
        segmentName={modalState?.unit?.name}
        visible={modalState.rename}
        onCancel={handleCancelModal}
        handleSubmit={handleRenameSegment}
      />
    </div>
  );
}
const mapDispatchToProps = (dispatch) =>
  bindActionCreators(
    {
      getSegmentFolders,
      getSavedSegments,
      deleteSegment,
      updateSegmentForId
    },
    dispatch
  );
export default connect(null, mapDispatchToProps)(ProfilesSidebar);
