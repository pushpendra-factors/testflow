import React from 'react';
import { Avatar } from 'antd';
import { SVG } from 'Components/factorsComponents';
import TextWithOverflowTooltip from 'Components/GenericComponents/TextWithOverflowTooltip';
import { ALPHANUMSTR, iconColors } from 'Components/Profile/constants';
import { UsernameWithIconProps } from 'Components/Profile/types';

function UsernameWithIcon({
  title,
  userID,
  isAnonymous = true
}: UsernameWithIconProps): JSX.Element {
  const getUsernameInitial = (user: string) => user.charAt(0).toUpperCase();

  const getBackgroundColor = (user: string) => {
    const index = ALPHANUMSTR.indexOf(getUsernameInitial(user)) % 8;
    return iconColors[index];
  };

  const renderUsername = (user: string, isAnon: boolean) => {
    if (user === 'group_user') {
      return 'Account Activity';
    }
    if (isAnon) {
      return 'Anonymous User';
    }
    return user;
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
              backgroundColor: getBackgroundColor(title),
              fontSize: '16px'
            }}
          >
            {getUsernameInitial(title)}
          </Avatar>
        )}
      </div>
      <TextWithOverflowTooltip
        text={renderUsername(title, isAnonymous)}
        extraClass='text'
      />
    </>
  );
}

export default UsernameWithIcon;
