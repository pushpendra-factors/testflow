import React from 'react';
import { Checkbox } from 'antd';
import { Text } from '../../factorsComponents';

export default function CustomCheckbox({ key, name, checked, onChange }) {
  return (
    <div key={key} className='flex justify-start items-center px-4 py-2'>
      <div className='mr-2'>
        <Checkbox checked={checked} onChange={onChange} />
      </div>
      <Text type='title' level={7} extraClass='mb-0' truncate charLimit={25}>
        {name}
      </Text>
    </div>
  );
}
