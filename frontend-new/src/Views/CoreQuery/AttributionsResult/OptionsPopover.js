import React, { memo } from 'react';
import { Checkbox } from 'antd';
import { Text } from '../../../components/factorsComponents';

function OptionsPopover({ options, onChange }) {
  return (
    <>
      {options.map((option) => {
        return (
          <div
            key={option.title}
            className='flex justify-start items-center p-2'
          >
            <div className='mr-2'>
              <Checkbox
                checked={option.enabled}
                onChange={onChange.bind(this, option)}
                disabled={option.disabled}
              />
            </div>
            <Text mini extraClass='mb-0' type='paragraph'>
              {option.title}
            </Text>
          </div>
        );
      })}
    </>
  );
}

export default memo(OptionsPopover);
