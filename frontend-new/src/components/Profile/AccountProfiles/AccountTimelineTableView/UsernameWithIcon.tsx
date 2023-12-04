// UsernameWithIcon.tsx
import React from 'react';
import { Avatar, Tooltip } from 'antd';
import { SVG } from 'Components/factorsComponents';
import { ALPHANUMSTR, iconColors } from '../../utils';
import { UsernameWithIconProps } from './types';

const UsernameWithIcon: React.FC<UsernameWithIconProps> = ({
  title,
  userID,
  isAnonymous
}) => (
  <>
    {isAnonymous ? (
      <SVG name={`TrackedUser${userID.match(/\d/g)?.[0] || 0}`} size={24} />
    ) : (
      <Avatar
        size={24}
        style={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          backgroundColor: `${
            iconColors[ALPHANUMSTR.indexOf(title.charAt(0).toUpperCase()) % 8]
          }`,
          fontSize: '16px'
        }}
      >
        {title.charAt(0).toUpperCase()}
      </Avatar>
    )}
    <Tooltip
      title={
        title === 'group_user'
          ? 'Account Activity'
          : isAnonymous
          ? 'Anonymous User'
          : title
      }
    >
      <span className='ml-2'>
        {title === 'group_user'
          ? 'Account Activity'
          : isAnonymous
          ? 'Anonymous User'
          : title}
      </span>
    </Tooltip>
  </>
);

export default UsernameWithIcon;
