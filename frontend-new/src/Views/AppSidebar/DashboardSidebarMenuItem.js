import React, { useState } from 'react';
import cx from 'classnames';
import { Popover } from 'antd';
import { SVG, SVG as Svg, Text } from 'Components/factorsComponents';
import { useDispatch, useSelector } from 'react-redux';
import { toggleNewFolderModal } from 'Reducers/dashboard/actions';
import { selectDashboardFoldersListState } from 'Reducers/dashboard/selectors';
import styles from './index.module.scss';

function DashboardSidebarMenuItem({
  text,
  isActive,
  onClick,
  onAdditionToNewFolder,
  onAddDashboardToExistingFolder
}) {
  const [showExistingFoldersList, setShowExistingFoldersList] = useState(false);
  const [showMoreActions, setShowMoreActions] = useState(false);
  const dashboardFoldersList = useSelector((state) =>
    selectDashboardFoldersListState(state)
  );
  const { data: foldersList } = dashboardFoldersList;
  const dispatch = useDispatch();

  const existingFoldersList = () => (
    <div className='flex flex-col py-2'>
      {foldersList.map((folder) => (
        <button
          type='button'
          key={folder.id}
          onClick={(e) => {
            e.stopPropagation();
            onAddDashboardToExistingFolder(folder.id);
          }}
          className={cx(
            'py-2 px-4 cursor-pointer text-left',
            styles['dashboard-more-options-menu']
          )}
        >
          <Text type='title' extraClass='mb-0'>
            {folder.name}
          </Text>
        </button>
      ))}
    </div>
  );

  const content = (
    <div className='flex flex-col py-2'>
      <Popover
        placement='right'
        visible={showExistingFoldersList}
        onVisibleChange={(visible) => {
          setShowExistingFoldersList(visible);
        }}
        onClick={() => {
          setShowExistingFoldersList(true);
        }}
        trigger='hover'
        content={existingFoldersList}
        overlayClassName={styles['more-actions-popover']}
      >
        <button
          type='button'
          className={cx(
            'px-4 py-2 text-left flex col-gap-4 items-center',
            styles['dashboard-more-options-menu']
          )}
        >
          <Text type='title' extraClass='mb-0'>
            Move to existing folder
          </Text>
          <SVG name='caretRight' size={16} color='#8c8c8c' />
        </button>
      </Popover>

      <button
        type='button'
        className={cx(
          'px-4 py-2 text-left',
          styles['dashboard-more-options-menu']
        )}
        onClick={(e) => {
          e.stopPropagation();
          onAdditionToNewFolder();
          dispatch(toggleNewFolderModal(true));
        }}
      >
        <Text type='title' extraClass='mb-0'>
          Move to new folder
        </Text>
      </button>

      <button
        type='button'
        onClick={(e) => {
          e.stopPropagation();
        }}
        className={cx(
          'px-4 py-2 text-left',
          styles['dashboard-more-options-menu']
        )}
      >
        <Text type='title' extraClass='mb-0'>
          Edit dashboard details
        </Text>
      </button>

      <button
        type='button'
        onClick={(e) => {
          e.stopPropagation();
        }}
        className={cx(
          'px-4 py-2 text-left',
          styles['dashboard-more-options-menu']
        )}
      >
        <Text type='title' extraClass='mb-0'>
          Delete dashboard
        </Text>
      </button>
    </div>
  );

  return (
    <div
      onClick={onClick}
      className={cx(
        'cursor-pointer rounded-md p-2 flex justify-between col-gap-2 items-center',
        {
          [styles.active]: isActive
        },
        styles['sidebar-menu-item'],
        {
          [styles.hovered]: showMoreActions
        }
      )}
    >
      <div className={cx('flex col-gap-1 items-center w-full')}>
        <Text
          type='title'
          level={7}
          extraClass='mb-0 text-with-ellipsis w-40'
          weight='medium'
        >
          {text}
        </Text>
      </div>
      <Popover
        overlayClassName={styles['more-actions-popover']}
        content={content}
        placement='right'
        arrow={false}
        onVisibleChange={(visible) => {
          setShowMoreActions(visible);
        }}
      >
        <span>
          <Svg
            size={16}
            color='#8C8C8C'
            name='more'
            extraClass={styles['more-actions-icon']}
          />
        </span>
      </Popover>
    </div>
  );
}

export default DashboardSidebarMenuItem;
