import { ProfileMapper, profileOptions } from 'Utils/constants';

export const getUserOptions = () => {
  const userOptions = [...profileOptions.users].map((item) => [
    item,
    ProfileMapper[item]
  ]);
  return userOptions;
};

export const getUserOptionsForDropdown = () => {
  return [['All Users', 'All'], ...getUserOptions()];
};
