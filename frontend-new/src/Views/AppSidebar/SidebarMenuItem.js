import { Text } from 'Components/factorsComponents';
import { Tooltip } from 'antd';
import React from 'react';

const SidebarMenuItem = ({ text }) => {
  return (
    <Tooltip title={text}>
      <Text
        type='title'
        level={7}
        color='character-primary'
        extraClass='mb-0 text-with-ellipsis'
        weight='medium'
      >
        {text}
      </Text>
    </Tooltip>
  );
};

export default SidebarMenuItem;
