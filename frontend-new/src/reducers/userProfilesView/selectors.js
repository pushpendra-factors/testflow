import { GROUP_NAME_DOMAINS } from 'Components/GlobalFilter/FilterWrapper/utils';
import { selectSegments } from 'Reducers/timelines/selectors';
import { createSelector } from 'reselect';

export const selectTimelinePayload = (state) =>
  state.userProfilesView.timelinePayload;

export const selectSegmentModalState = (state) =>
  state.userProfilesView.showSegmentModal;

export const selectSegmentsList = createSelector(selectSegments, (segments) => {
  const segmentsList = [];
  Object.entries(segments)
    .filter((segment) => segment[0] !== GROUP_NAME_DOMAINS)
    .forEach(([_, vals]) => {
      segmentsList.push(...vals);
    });
  return segmentsList;
});
