import { ReverseProfileMapper } from 'Utils/constants';
import { formatSegmentsObjToGroupSelectObj } from '../utils';

export const getGroupList = (groupOptions) => {
  const groups = Object.entries(groupOptions || {}).map(
    ([group_name, display_name]) => [display_name, group_name]
  );
  groups.unshift(['All Accounts', 'All']);
  return groups;
};

export const generateSegmentsList = ({ accountPayload, segments }) => {
  const segmentsList = [];

  Object.entries(segments)
    .filter(
      (segment) => !Object.keys(ReverseProfileMapper).includes(segment[0])
    )
    .map(([group, vals]) => formatSegmentsObjToGroupSelectObj(group, vals))
    .forEach((obj) => segmentsList.push(obj));
  return segmentsList;
};
