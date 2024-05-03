import { SEGMENT_DELETED } from './types';

export const deleteSegmentAction = ({ segmentId }) => ({
  type: SEGMENT_DELETED,
  payload: segmentId
});
