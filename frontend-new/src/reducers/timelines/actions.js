import { SEGMENT_DELETED } from './types';

export const deleteSegmentAction = ({ segmentId }) => {
  return { type: SEGMENT_DELETED, payload: segmentId };
};
