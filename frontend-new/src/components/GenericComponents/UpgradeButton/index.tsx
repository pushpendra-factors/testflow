import { SVG, Text } from 'Components/factorsComponents';
import usePlanUpgrade from 'hooks/usePlanUpgrade';
import React from 'react';

function UpgradeButton({ extraClass, featureName }: UpgradeButtonProps) {
  const { handlePlanUpgradeClick } = usePlanUpgrade();
  return (
    <div
      className='flex items-center font-semibold gap-2 flex-nowrap whitespace-nowrap cursor-pointer'
      onClick={(e: React.MouseEvent<HTMLDivElement>) => {
        e.stopPropagation();
        handlePlanUpgradeClick(featureName);
      }}
    >
      <Text type='paragraph' mini weight='bold' color='brand-color-6'>
        Upgrade plan
      </Text>
      <SVG size={20} name='Lock' />
    </div>
  );
}

type UpgradeButtonProps = {
  extraClass?: string;
  featureName: string;
};

export default UpgradeButton;
