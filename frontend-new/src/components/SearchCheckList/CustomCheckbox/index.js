import React from 'react';
import { Checkbox } from 'antd';
import { Text } from '../../factorsComponents';

export default function CustomCheckbox({ key, name, checked, onChange }) {
  return (
    <div key={key} className="flex justify-start items-center px-4 py-2">
      <div className="mr-2">
        <Checkbox checked={checked} onChange={onChange} />
      </div>
      <Text mini extraClass="mb-0 truncate" type="paragraph">
        {name}
      </Text>
    </div>
  );
}
