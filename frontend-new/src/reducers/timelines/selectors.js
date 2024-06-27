import { createSelector } from 'reselect';

export const selectSegments = (state) => state.timelines.accountSegments;

export const selectSegmentBySegmentId = createSelector(
  selectSegments,
  (state, segmentId) => segmentId,
  (accountSegments, segmentId) => {
    return accountSegments.find((segment) => segment.id === segmentId);
  }
);
