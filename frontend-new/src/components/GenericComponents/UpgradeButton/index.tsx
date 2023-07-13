import { SVG, Text } from 'Components/factorsComponents';
import { PathUrls } from 'Routes/pathUrls';
import React from 'react';
import { Link, useHistory } from 'react-router-dom';

const UpgradeButton = ({ extraClass }: UpgradeButtonProps) => {
  const history = useHistory();
  return (
    <div
      className='flex items-center font-semibold gap-2 flex-nowrap whitespace-no-wrap cursor-pointer'
      onClick={() => history.push(PathUrls.SettingsPricing)}
    >
      <Text type='paragraph' mini weight={'bold'} color='brand-color-6'>
        Upgrade plan
      </Text>
      <SVG size={20} name='Lock' />
    </div>
  );
};

type UpgradeButtonProps = {
  extraClass?: string;
};

export default UpgradeButton;
