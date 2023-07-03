import { formatSegmentsObjToGroupSelectObj } from 'Components/Profile/utils';
import { selectSegments } from 'Reducers/timelines/selectors';
import { ReverseProfileMapper } from 'Utils/constants';
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
      .filter((segment) =>
        Object.keys(ReverseProfileMapper).includes(segment[0])
      )
      .forEach(([group, vals]) => {
        const obj = formatSegmentsObjToGroupSelectObj(group, vals);
        segmentsList.push(obj);
      });

    return segmentsList;
  }
);
