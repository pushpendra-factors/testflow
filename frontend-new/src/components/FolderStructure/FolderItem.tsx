import React, { useContext, useMemo, useState } from 'react';
import { Divider, Input, Popover, Tooltip } from 'antd';
import { SVG, Text } from 'Components/factorsComponents';
import {
  DeleteOutlined,
  EditOutlined,
  FolderOpenFilled,
  LoadingOutlined,
  PlusOutlined,
  RightOutlined
} from '@ant-design/icons';
import { SegmentIcon } from 'Views/AppSidebar/AccountsSidebar';
import { getSegmentColorCode } from 'Views/AppSidebar/appSidebar.helpers';
import AppModal from 'Components/AppModal';
import styles from './index.module.scss';

import { FolderContext } from './FolderContext';
import { FolderItemOptionsType, FolderItemPropType } from './type';

const PopoverOptionWrapper = ({
  children,
  submenu,
  unitid,
  folder_id,
  handleNewFolder = null,
  moveToExistingFolder = null
}: any) => {
  const [newFolderState, setNewFolderState] = useState({
    visible: false,
    name: ''
  });
  const contextValue = useContext(FolderContext);
  if (submenu && submenu.length > 0) {
    const submenuList = submenu

      ?.sort((a: any, b: any) => {
        if (a.title < b.title) {
          return -1;
        }
        if (a.title > b.title) {
          return 1;
        }
        return 0;
      })
      // eslint-disable-next-line eqeqeq
      ?.filter((e: any) => e.id != folder_id);
    return (
      <Popover
        content={
          <div className={styles.submenu_list}>
            {' '}
            <div className={styles.submenu_list_group}>
              {submenuList.map((eachSubMenu: any) => (
                <div
                  key={eachSubMenu.id}
                  className={styles.submenu_item}
                  onClick={
                    moveToExistingFolder
                      ? () => moveToExistingFolder(eachSubMenu.id, unitid)
                      : (e: any) => {
                          contextValue.moveToExistingFolder(
                            e,
                            eachSubMenu.id,
                            unitid
                          );
                        }
                  }
                >
                  {false ? <LoadingOutlined /> : <FolderOpenFilled />}
                  <div>
                    <Tooltip title={eachSubMenu.title} placement='right'>
                      {eachSubMenu.title}
                    </Tooltip>
                  </div>
                </div>
              ))}
            </div>
            {submenuList.length ? (
              <Divider style={{ margin: '4px 0' }} />
            ) : null}
            <div
              className={styles.submenu_item}
              onClick={
                handleNewFolder
                  ? () => setNewFolderState({ visible: true, name: '' })
                  : (e) =>
                      contextValue.setFolderModalState &&
                      contextValue.setFolderModalState((prev: any) => ({
                        ...prev,
                        visible: true,
                        action: 'create',
                        segmentId: unitid,
                        unit: null
                      }))
                // contextValue.handleNewFolder &&
                // contextValue.handleNewFolder(e, unitid)
              }
            >
              <PlusOutlined /> New Folder
            </div>
          </div>
        }
        trigger='hover'
        overlayClassName={styles.submenu_list_container}
        placement='right'
      >
        <div className={styles.submenu}>
          {' '}
          <div>{children}</div>
          <RightOutlined />
          <AppModal
            visible={newFolderState.visible}
            title='Create New Folder'
            maskClosable
            onCancel={() => setNewFolderState({ visible: false, name: '' })}
            onOk={() => {
              handleNewFolder(unitid, newFolderState.name);
              setNewFolderState({ visible: false, name: '' });
            }}
          >
            <div className='flex flex-col gap-y-2'>
              <Text type='title' color='character-primary' extraClass='mb-0'>
                Enter Folder name
              </Text>
              <Input
                onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
                  const { value } = e.target;
                  setNewFolderState((prev) => ({ ...prev, name: value }));
                }}
                value={newFolderState.name}
                className={`fa-input ${styles.input}`}
                size='large'
                placeholder='Eg- Folder 1'
              />
            </div>
          </AppModal>
        </div>
      </Popover>
    );
  }
  return children;
};

export function FolderItemOptions(props: FolderItemOptionsType) {
  const {
    id,
    folders,
    hideDefaultOptions,
    unit,
    handleEditUnit,
    handleDeleteUnit,
    handleNewFolder,
    moveToExistingFolder,
    extraOptions,
    data,
    folder_id
  } = props;
  const actionsMenu = useMemo(() => {
    let tmpActions: Array<any> = [];
    if (!hideDefaultOptions) {
      tmpActions = [
        {
          id: '1',
          title: 'Move to',
          icon: <SVG name='AddFromDraft' />,
          // {id: '11', title: }
          submenu: folders?.map((eachFolder) => ({
            id: eachFolder.id,
            title: eachFolder.name,
            icon: <EditOutlined />
          }))
        },
        {
          id: '2',
          title: `Edit ${unit} Details`,
          icon: <EditOutlined />,
          onClick: handleEditUnit
        },
        {
          id: '3',
          title: `Delete ${unit}`,
          icon: <DeleteOutlined />,
          onClick: handleDeleteUnit
        }
      ];
    }
    if (extraOptions && Array.isArray(extraOptions)) {
      tmpActions = [...tmpActions, ...extraOptions];
    }
    return tmpActions;
  }, [folders, extraOptions]);
  const popoverContent = (
    <div
      onClick={(e) => {
        e.stopPropagation();
      }}
    >
      {actionsMenu.map((eachAction) => (
        <div
          key={eachAction.id}
          className={styles.popover_list}
          onClick={() => eachAction.onClick && eachAction.onClick(data)}
        >
          <PopoverOptionWrapper
            submenu={eachAction?.submenu}
            unitid={id}
            folder_id={folder_id}
            handleNewFolder={handleNewFolder}
            moveToExistingFolder={moveToExistingFolder}
          >
            <div className='flex items-center'>
              {eachAction.icon}
              {eachAction.title}
            </div>
          </PopoverOptionWrapper>
        </div>
      ))}
    </div>
  );
  return (
    <div>
      <Popover
        content={popoverContent}
        placement='right'
        trigger='hover'
        arrowContent={<RightOutlined />}
        overlayClassName={styles.popover_list_container}
      >
        <span>
          {' '}
          <SVG size={16} color='#8C8C8C' name='more' />
        </span>{' '}
      </Popover>
    </div>
  );
}
function FolderItem(props: FolderItemPropType) {
  const { id, data, folder_id, folders } = props;
  const contextValue = useContext(FolderContext);

  const iconColor = getSegmentColorCode(data?.name);
  return (
    <div
      className={`${styles.folder_item} ${
        id === contextValue.active_item ? styles.active_item : ''
      }`}
      onClick={() => contextValue.onUnitClick(data)}
    >
      <div className='flex justify-left gap-2'>
        {contextValue.showItemIcons && (
          <div>
            <SVG name={SegmentIcon(data?.name)} size={20} color={iconColor} />
          </div>
        )}
        <div className={styles.folder_item_name}>
          <Tooltip title={data.name}>{data.name}</Tooltip>
        </div>
      </div>
      <div className={styles.folder_item_actions}>
        <FolderItemOptions
          id={id}
          handleDeleteUnit={contextValue.handleDeleteUnit}
          handleEditUnit={contextValue.handleEditUnit}
          folder_id={folder_id}
          folders={folders}
          data={data}
          unit={contextValue.unit}
          moveToExistingFolder={null}
          handleNewFolder={null}
        />
      </div>
    </div>
  );
}
export default FolderItem;
