import React from 'react';
import { Number as NumFormat, SVG, Text } from 'Components/factorsComponents';

interface ComparePercentProps {
  value: number;
}

export default function ComparePercent({ value }: ComparePercentProps) {
  return (
    <div className='flex gap-x-1 items-center justify-center'>
      <SVG
        color={value >= 0 ? '#5ACA89' : '#FF4D4F'}
        name={value >= 0 ? 'arrowLift' : 'arrowDown'}
        size={16}
      />
      <Text
        extraClass='mb-0'
        level={7}
        type='title'
        color={value < 0 ? 'red' : 'green'}
        weight='medium'
      >
        <NumFormat number={Math.abs(value)} />%
      </Text>
    </div>
  );
}
