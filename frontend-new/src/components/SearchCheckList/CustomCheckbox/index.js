import React from 'react';
import { Checkbox } from 'antd';
import { SVG, Text } from '../../factorsComponents';

export default function CustomCheckbox({ key, name, checked, onChange }) {
  return (
    <div
      key={key}
      className='inline-flex justify-between items-center gap--4 min-w-full py-2 px-3 cursor-pointer'
      onClick={onChange}
      style={{ background: checked ? '#F5F5F5' : null }}
    >
      {/* <div className='mr-2'>
        <Checkbox checked={checked} onChange={onChange} />
      </div> */}
      <Text
        type='title'
        level={7}
        extraClass='mb-0 truncate'
        truncate
        charLimit={25}
      >
        {name}
      </Text>
      {checked && (
        <SVG
          name='checkmark'
          extraClass={'self-center'}
          size={17}
          color={'purple'}
        />
      )}
    </div>
  );
}
