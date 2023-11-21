import { GROUP_NAME_DOMAINS } from 'Components/GlobalFilter/FilterWrapper/utils';
import { formatSegmentsObjToGroupSelectObj } from 'Components/Profile/utils';
import { selectSegments } from 'Reducers/timelines/selectors';
import { createSelector } from 'reselect';

export const selectTimelinePayload = (state) =>
  state.userProfilesView.timelinePayload;

export const selectActiveSegment = (state) =>
  state.userProfilesView.activeSegment;

export const selectSegmentModalState = (state) =>
  state.userProfilesView.showSegmentModal;

export const selectSegmentsList = createSelector(
  selectTimelinePayload,
  selectSegments,
  (timelinePayload, segments) => {
    const segmentsList = [];
    Object.entries(segments)
      .filter((segment) => segment[0] !== GROUP_NAME_DOMAINS)
      .forEach(([group, vals]) => {
        const obj = formatSegmentsObjToGroupSelectObj(group, vals);
        segmentsList.push(obj);
      });

    return segmentsList;
  }
);
