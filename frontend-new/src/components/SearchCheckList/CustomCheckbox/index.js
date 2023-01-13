import React from 'react';
import { Checkbox } from 'antd';
import { Text } from '../../factorsComponents';

export default function CustomCheckbox({ key, name, checked, onChange }) {
  return (
    <div
      key={key}
      className='inline-flex-gap--4 min-w-full p-2'
    >
      <div className='mr-2'>
        <Checkbox checked={checked} onChange={onChange} />
      </div>
      <Text
        type='title'
        level={7}
        extraClass='mb-0 truncate'
        truncate
        charLimit={25}
      >
        {name}
      </Text>
    </div>
  );
}
