import { getGroupList } from 'Components/Profile/AccountProfiles/accountProfiles.helpers';
import { createSelector } from 'reselect';

export const selectGroupOptions = (state) => state.coreQuery.groups;

export const selectGroupsList = createSelector(
  selectGroupOptions,
  (groupOptions) => {
    return getGroupList(groupOptions?.account_groups);
  }
);
