import React, { useState } from 'react';
import { Checkbox } from 'antd';
import { SVG, Text } from '../../factorsComponents';
import { HolderOutlined } from '@ant-design/icons';

export default function CustomCheckbox({ key, name, checked, onChange }) {
  const [isHovered, setIsHovered] = useState(false);

  const handleMouseEnter = () => {
    setIsHovered(true);
  };

  const handleMouseLeave = () => {
    setIsHovered(false);
  };

  return (
    <div
      key={key}
      className='inline-flex justify-between items-center gap--4 min-w-full py-2 px-3 cursor-pointer'
      onClick={onChange}
      style={{
        background: checked ? '#F5F5F5' : null,
        cursor: isHovered ? 'grab' : 'pointer'
      }}
      onMouseEnter={handleMouseEnter}
      onMouseLeave={handleMouseLeave}
    >
      <div className='inline-flex'>
        {checked && (
          <div style={{ cursor: isHovered ? 'grab' : 'pointer' }}>
            {checked && <HolderOutlined />}
          </div>
        )}

        <Text
          type='title'
          level={7}
          extraClass='mb-0 truncate not-draggable pl-1'
          truncate
          charLimit={25}
        >
          {name}
        </Text>
      </div>

      {checked && (
        <div className='not-draggable'>
          <SVG
            name='checkmark'
            extraClass={'self-center not-draggable'}
            size={17}
            color={'purple'}
          />
        </div>
      )}
    </div>
  );
}
