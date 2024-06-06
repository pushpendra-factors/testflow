import { createContext } from 'react';
import { FolderType } from './Folder';

export type folderModalInitStateType = {
  visible: boolean;
  action: '' | 'rename' | 'create' | 'delete';
  unit: {
    id: string | number;
    name: string;
  };
  segmentId: '';
};
export const folderModalInitState: folderModalInitStateType = {
  visible: false,
  action: '',
  unit: {
    id: '',
    name: ''
  },
  segmentId: ''
};
type folderContextType = {
  handleNewFolder: (e: any, id: number | string) => void;
  moveToExistingFolder: (
    e: any,
    folderId: string | number,
    unitid: string | number
  ) => void;
  handleEditUnit: (unit: any) => void;
  handleDeleteUnit: (unit: any) => void;
  onUnitClick: (unit: any) => void;
  onRenameFolder: (unit_id: any, name: string) => void;
  onDeleteFolder: (unit_id: any) => void;
  folderModalState: folderModalInitStateType;
  setFolderModalState: React.SetStateAction<any>;
  unit: string;
  folders: Array<FolderType>;
  active_item: string | number | null;
  showItemIcons: boolean;
};
export const FolderContext = createContext<folderContextType>({
  handleNewFolder: (e: any, id: number | string) => undefined,
  moveToExistingFolder: (
    e: any,
    folderId: string | number,
    unitid: string | number
  ) => undefined,
  handleEditUnit: (unit: any) => undefined,
  handleDeleteUnit: (unit: any) => undefined,
  onUnitClick: (unit: any) => undefined,
  onRenameFolder: (unit_id: any, name: string) => undefined,
  onDeleteFolder: (unit_id: any) => undefined,
  folderModalState: folderModalInitState,
  setFolderModalState: (prevState: any) => undefined,
  unit: 'segment',
  folders: Array(0),
  active_item: null,
  showItemIcons: false
});
