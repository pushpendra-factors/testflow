import React from 'react';
import { Avatar } from 'antd';
import { SVG } from 'Components/factorsComponents';
import TextWithOverflowTooltip from 'Components/GenericComponents/TextWithOverflowTooltip';
import { ALPHANUMSTR, iconColors } from 'Components/Profile/constants';
import { UsernameWithIconProps } from 'Components/Profile/types';

function UsernameWithIcon({
  title,
  userID,
  isGroupUser,
  isAnonymous = true
}: UsernameWithIconProps): JSX.Element {
  const getUsernameInitial = () => title.charAt(0).toUpperCase();

  const getBackgroundColor = () => {
    const index = ALPHANUMSTR.indexOf(getUsernameInitial()) % 8;
    return iconColors[index];
  };

  const renderUsername = () => {
    if (isGroupUser) {
      return 'Account Activity';
    }
    if (isAnonymous) {
      return 'Anonymous User';
    }
    return title;
  };

  return (
    <>
      <div className='grow-0 shrink-0'>
        {isAnonymous ? (
          <SVG name={`TrackedUser${userID.match(/\d/g)?.[0] || 0}`} size={24} />
        ) : (
          <Avatar
            size={24}
            style={{
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              backgroundColor: getBackgroundColor(),
              fontSize: '16px'
            }}
          >
            {getUsernameInitial()}
          </Avatar>
        )}
      </div>
      <TextWithOverflowTooltip text={renderUsername()} extraClass='text' />
    </>
  );
}

export default UsernameWithIcon;
