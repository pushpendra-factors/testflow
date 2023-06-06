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

  if (accountPayload.source === 'All') {
    const allowedGroups = [
      '$hubspot_company',
      '$salesforce_account',
      '$6signal'
    ];

    Object.entries(segments)
      .filter(([group]) => allowedGroups.includes(group))
      .map(([group, vals]) => formatSegmentsObjToGroupSelectObj(group, vals))
      .forEach((obj) => segmentsList.push(obj));
  } else {
    const obj = formatSegmentsObjToGroupSelectObj(
      accountPayload.source,
      segments[accountPayload.source]
    );
    segmentsList.push(obj);
  }
  return segmentsList;
};
