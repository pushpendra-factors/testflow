import { createSelector } from 'reselect';
import { GROUP_NAME_DOMAINS } from 'Components/GlobalFilter/FilterWrapper/utils';
import { reorderDefaultDomainSegmentsToTop } from 'Components/Profile/AccountProfiles/accountProfiles.helpers';

export const selectSegments = (state) => state.timelines.segments;

export const selectSegmentBySegmentId = createSelector(
  selectSegments,
  (state, segmentId) => segmentId,
  (segments, segmentId) => {
    const segmentsList =
      reorderDefaultDomainSegmentsToTop(segments[GROUP_NAME_DOMAINS]) || [];
    return segmentsList.find((segment) => segment.id === segmentId);
  }
);
