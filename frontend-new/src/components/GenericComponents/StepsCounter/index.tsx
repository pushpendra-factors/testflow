import { Text } from 'Components/factorsComponents';
import React from 'react';

const StepsCounter = ({ totalSteps, currentStep }: StepCounterProps) => {
  const stepWidth = 24;
  const stepHeight = 8;
  const stepGap = 2;
  return (
    <div className='flex items-center gap-2.5'>
      <div className='flex'>
        {Array.from({ length: totalSteps }).map((_, index) => (
          <div
            key={index}
            style={{
              width: stepWidth,
              height: stepHeight,
              marginRight: stepGap,
              background: index < currentStep ? '#1890FF' : '#F5F5F5'
            }}
          />
        ))}
      </div>
      <Text
        type={'title'}
        level={7}
        color='character-secondary'
        extraClass='m-0 ml-2'
      >
        {currentStep} / {totalSteps} steps
      </Text>
    </div>
  );
};

interface StepCounterProps {
  totalSteps: number;
  currentStep: number;
}

export default StepsCounter;
