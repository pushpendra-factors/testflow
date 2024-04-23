import React from 'react';
import cx from 'classnames';
import styles from './index.module.scss';

const ProfilesWrapper = ({ children }) => (
  <div
    className={cx(
      'w-full h-full overflow-scroll px-1',
      styles['profile-wrapper']
    )}
  >
    {children}
  </div>
);

export default ProfilesWrapper;
