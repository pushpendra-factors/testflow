export interface FolderType {
  id: number | string;
  name: string;
  created_at?: string;
  updated_at?: string;
  items: Array<ItemType>;
  can_be_deleted?: boolean;
}

export interface FolderPropType {
  id: number | string;
  name: string;
  items: Array<ItemType>;
  folders: Array<FolderType>;

  isAllBoard?: boolean;
}

export type ItemType = {
  icon?: JSX.Element | string;
  id: string | number;
  folder_id: string;
  name: string;
  created_at: string;
  updated_at: string;
  // This contains all the rest of props
  [x: string]: any;
};
export type FolderItemPropType = {
  data: ItemType;
  folders: Array<FolderType>;
  folder_id: number | string;
  id: string | number;
  [x: string]: any;
};

export type FolderItemOptionsType = {
  data: ItemType;
  folder_id: number | string;
  id: string | number;
  unit: any;
  handleEditUnit: any;
  handleDeleteUnit: any;
  folders: Array<FolderType>;
  handleNewFolder: null | ((e: any, id: number | string) => undefined);
  moveToExistingFolder:
    | null
    | ((
        e: any,
        folderId: string | number,
        unitid: string | number
      ) => undefined);
  extraOptions?: Array<any>;
  hideDefaultOptions?: boolean;
};
