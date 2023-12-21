import React from 'react';
import { Avatar } from 'antd';
import { SVG } from 'Components/factorsComponents';
import { ALPHANUMSTR, iconColors } from '../../utils';
import { UsernameWithIconProps } from './types';
import TextWithOverflowTooltip from 'Components/GenericComponents/TextWithOverflowTooltip';

const UsernameWithIcon: React.FC<UsernameWithIconProps> = ({
  title,
  userID,
  isAnonymous
}) => {
  const getUsernameInitial = (title: string) => {
    return title.charAt(0).toUpperCase();
  };

  const getBackgroundColor = (title: string) => {
    const index = ALPHANUMSTR.indexOf(getUsernameInitial(title)) % 8;
    return iconColors[index];
  };

  const renderUsername = (title: string, isAnonymous: boolean) => {
    if (title === 'group_user') {
      return 'Account Activity';
    } else if (isAnonymous) {
      return 'Anonymous User';
    } else {
      return title;
    }
  };

  return (
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
            backgroundColor: getBackgroundColor(title),
            fontSize: '16px'
          }}
        >
          {getUsernameInitial(title)}
        </Avatar>
      )}
      <TextWithOverflowTooltip
        text={renderUsername(title, isAnonymous)}
        extraClass='ml-2'
      />
    </>
  );
};

export default UsernameWithIcon;
