import { SearchOutlined } from '@ant-design/icons';
import { Input } from 'antd';
import React, { useCallback, useEffect, useMemo, useState } from 'react';
import { Text } from 'Components/factorsComponents';
import AppModal from 'Components/AppModal';
import styles from './index.module.scss';
import FolderItem from './FolderItem';
import Folder from './Folder';
import {
  FolderContext,
  folderModalInitState,
  folderModalInitStateType
} from './FolderContext';
import { FolderType, ItemType } from './type';

type FolderStructurePropType = {
  folders: Array<FolderType>;
  items: Array<ItemType>;
  active_item: string | number | null;
  showItemIcons: boolean;

  unit: string;

  handleNewFolder: (e: any, id: number | string) => void;
  moveToExistingFolder: () => void;
  handleEditUnit: (unit: any) => void;
  handleDeleteUnit: (unit: any) => void;
  onUnitClick: (unit: any) => void;

  onRenameFolder: (unit_id: any, name: string) => void;
  onDeleteFolder: (unit_id: any) => void;
};

type FolderStructureType = {
  [key: string | number]: {
    can_be_deleted: boolean | undefined;
    id: number | string;
    name: string;
    data: Array<ItemType>;

    isAllboard?: boolean;
  };
};
function FolderStructure(props: FolderStructurePropType) {
  const {
    folders,
    items,
    active_item,
    showItemIcons,
    unit,
    handleNewFolder,
    moveToExistingFolder,
    handleEditUnit,
    handleDeleteUnit,
    onUnitClick,
    onRenameFolder,
    onDeleteFolder
  } = props;
  const [foldersState, setFoldersState] = useState<FolderType[]>([]);
  const [folderStructure, setFolderStructure] = useState<FolderStructureType>(
    {}
  );
  const [searchString, setSearchString] = useState('');

  const [folderModalState, setFolderModalState] =
    useState<folderModalInitStateType>({
      visible: false,
      action: '',
      unit: { id: '', name: '' },
      segmentId: ''
    });
  useEffect(() => {
    const tmpFolderStructure: FolderStructureType = {};
    folders.forEach((eachFolder) => {
      tmpFolderStructure[eachFolder.id] = {
        id: eachFolder.id,
        name: eachFolder.name,
        data: [],
        can_be_deleted: undefined
      };
      if (unit === 'dashboard') {
        tmpFolderStructure[eachFolder.id].can_be_deleted =
          eachFolder.can_be_deleted;
      }
    });
    if (unit !== 'dashboard')
      tmpFolderStructure[0] = {
        id: 0,
        name: `All ${unit}s`,
        data: [],
        can_be_deleted: true // for dashboard // ideally it should be -ve of it, but it was prev used in DashboardFolders, so kept it like this
      };
    items.forEach((eachSegment) => {
      if (tmpFolderStructure[eachSegment.folder_id])
        tmpFolderStructure[eachSegment.folder_id]?.data?.push(eachSegment);
      else tmpFolderStructure[0]?.data?.push(eachSegment);
    });
    setFolderStructure(tmpFolderStructure);

    setFoldersState(
      unit === 'dashboard'
        ? folders
        : [{ ...tmpFolderStructure[0], items: [] }, ...folders]
    );
  }, [folders, items]);

  const ContextValue = useMemo(
    () => ({
      handleNewFolder, // This creates a newfolder and moves selected item inside it.
      moveToExistingFolder,
      handleEditUnit,
      handleDeleteUnit,
      onUnitClick,
      onRenameFolder,
      onDeleteFolder,
      folderModalState,
      setFolderModalState,
      unit,
      folders,
      active_item,
      showItemIcons
    }),
    [folders, unit, folderModalState, active_item]
  );
  const handleModalCancel = () => {
    setFolderModalState(folderModalInitState);
  };
  const handleModalSubmit = () => {
    if (folderModalState.action === 'rename') {
      // rename handle
      if (onRenameFolder)
        onRenameFolder(folderModalState.unit?.id, folderModalState.unit.name);
    } else if (folderModalState.action === 'create') {
      // move to new folder handle
      if (handleNewFolder)
        handleNewFolder(
          folderModalState.segmentId,
          folderModalState.unit?.name
        );
    } else if (folderModalState.action === 'delete') {
      // delete handle
      onDeleteFolder(folderModalState.unit?.id);
    }
    setFolderModalState({
      visible: false,
      action: '',
      unit: { id: '', name: '' },
      segmentId: ''
    });
  };
  // This is important to Memoise the Title in Modal
  const RenderModalTitle = useMemo(
    () => (
      <Text
        type='title'
        level={4}
        color='character-primary'
        extraClass='mb-0'
        weight='bold'
      >
        {folderModalState.action === 'rename'
          ? `Rename folder - ${folderModalState.unit?.name}`
          : folderModalState.action === 'create'
            ? 'Create new folder'
            : folderModalState.action === 'delete'
              ? `Are you sure you want to delete "${folderModalState.unit?.name}" Folder?`
              : ''}
      </Text>
    ),
    [folderModalState.action]
  );
  const folderModal = (
    <AppModal
      visible={folderModalState.visible}
      onCancel={handleModalCancel}
      onOk={handleModalSubmit}
      okText={folderModalState.action === 'delete' ? 'Confirm' : 'Save'}
      maskClosable
    >
      <div className='flex flex-col gap-y-5'>
        {RenderModalTitle}
        <div className='flex flex-col gap-y-2'>
          {folderModalState.action !== 'delete' && (
            <>
              <Text type='title' color='character-primary' extraClass='mb-0'>
                Enter Folder name
              </Text>
              <Input
                onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
                  const { value } = e.target;

                  setFolderModalState((prev) => ({
                    ...prev,
                    unit: {
                      ...prev.unit,
                      name: value
                    }
                  }));
                }}
                value={folderModalState?.unit?.name}
                className={`fa-input ${styles.input}`}
                size='large'
                placeholder='Eg- Folder 1'
              />
            </>
          )}
        </div>
      </div>
    </AppModal>
  );
  const searchResults = (
    <div className={styles.search_results_container}>
      {items
        .sort((a, b) => (a.updated_at < b.updated_at ? -1 : 1))
        .filter(
          (e) =>
            e.name?.toLowerCase().includes(searchString.trim().toLowerCase())
        )
        .map((eachItem) => (
          <FolderItem
            key={eachItem.id}
            id={eachItem.id}
            folder_id={eachItem.folder_id}
            data={eachItem}
            folders={folders}
          />
        ))}
    </div>
  );
  return (
    <FolderContext.Provider value={ContextValue}>
      <div>
        <div
          className='p-2 bg-white'
          style={{ position: 'sticky', top: 0, zIndex: 1 }}
        >
          <Input
            placeholder={`Search ${unit}`}
            prefix={<SearchOutlined />}
            style={{ borderRadius: '8px' }}
            onChange={(e) => {
              setSearchString(e.target.value);
            }}
          />
        </div>

        {searchString.length === 0 &&
          Object.keys(folderStructure)
            .sort((a: string, b: string) => {
              if (folderStructure[a].name < folderStructure[b].name) {
                return -1;
              }
              if (folderStructure[a].name > folderStructure[b].name) {
                return 1;
              }
              return 0;
            })
            .map((eachFolderID) => (
              <Folder
                key={eachFolderID}
                name={folderStructure[eachFolderID].name}
                id={eachFolderID}
                folders={foldersState}
                items={folderStructure[eachFolderID]?.data}
                isAllBoard={folderStructure[eachFolderID]?.can_be_deleted}
              />
            ))}
        {searchString.length > 0 && searchResults}
      </div>
      {folderModal}
    </FolderContext.Provider>
  );
}

export default FolderStructure;
