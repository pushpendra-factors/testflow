import { Text } from 'Components/factorsComponents';
import { Tag } from 'antd';
import React from 'react';

interface HeaderProps {
  title: string;
  tag?: string;
  description: string;
}

const Header = ({ title, tag, description }: HeaderProps) => (
  <div className='w-full'>
    <div className='flex gap-2'>
      <Text
        type='title'
        level={5}
        weight='bold'
        color='character-primary'
        extraClass='m-0'
      >
        {title}
      </Text>
      {tag && (
        <div>
          <Tag color='success'>{tag}</Tag>
        </div>
      )}
    </div>
    <Text type='title' level={6} color='character-secondary' extraClass='m-0'>
      {description}
    </Text>
  </div>
);

export default Header;
