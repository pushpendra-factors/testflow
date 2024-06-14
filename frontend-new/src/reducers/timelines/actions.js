import {
  LOADING_SEGMENT_FOLDER,
  SEGMENT_DELETED,
  SET_ACCOUNT_SEGMENT_FOLDERS,
  SET_PEOPLE_SEGMENT_FOLDERS
} from './types';

export const deleteSegmentAction = ({ segmentId }) => ({
  type: SEGMENT_DELETED,
  payload: segmentId
});

export const setLoadingSegmentFolders = () => ({
  type: LOADING_SEGMENT_FOLDER
});

export const setSegmentFolders = ({ folder_type = 'account', data = [] }) => ({
  type:
    folder_type === 'account'
      ? SET_ACCOUNT_SEGMENT_FOLDERS
      : SET_PEOPLE_SEGMENT_FOLDERS,
  payload: data
});
