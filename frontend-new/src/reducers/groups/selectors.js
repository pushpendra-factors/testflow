import { getGroupList } from 'Components/Profile/AccountProfiles/accountProfiles.helpers';
import { createSelector } from 'reselect';

export const selectGroupOptions = (state) => state.groups.data;

export const selectGroupsList = createSelector(
  selectGroupOptions,
  (groupOptions) => {
    return getGroupList(groupOptions);
  }
);
