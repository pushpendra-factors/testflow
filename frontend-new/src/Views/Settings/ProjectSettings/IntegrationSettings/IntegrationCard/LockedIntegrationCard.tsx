import UpgradeButton from 'Components/GenericComponents/UpgradeButton';
import { SVG, Text } from 'Components/factorsComponents';
import { FEATURES } from 'Constants/plans.constants';
import { Avatar } from 'antd';
import usePlanUpgrade from 'hooks/usePlanUpgrade';
import React from 'react';

const LockedIntegrationCard = ({
  title,
  description,
  featureName,
  icon,
  ...props
}: LockedIntegrationCardProps) => {
  const { handlePlanUpgradeClick } = usePlanUpgrade();
  const handleCardClick = () => {
    handlePlanUpgradeClick(featureName);
  };

  return (
    <div
      className='fa-intergration-card'
      {...props}
      style={{ background: '#FAFAFA' }}
    >
      <div
        className='flex justify-between cursor-pointer'
        onClick={handleCardClick}
      >
        <div className='flex items-center justify-center '>
          <Avatar
            size={60}
            shape='square'
            icon={<SVG name={icon} size={40} color='purple' />}
            style={{
              backgroundColor: '#fff',
              borderRadius: 10,
              border: '1px solid #f0f0f0',
              display: 'flex'
            }}
            className='flex items-center justify-center'
          />
        </div>
        <div className='flex flex-col justify-start items-start ml-4 w-full'>
          <div className='flex flex-row items-center justify-start'>
            <Text type='title' level={5} weight='bold' extraClass='m-0'>
              {title}
            </Text>
          </div>

          <Text
            type='paragraph'
            mini
            extraClass='m-0 w-9/12'
            color='grey'
            lineHeight='medium'
          >
            {description}
          </Text>
        </div>
        <div className='flex justify-center items-center'>
          <UpgradeButton featureName={featureName} />;
        </div>
      </div>
    </div>
  );
};

interface LockedIntegrationCardProps {
  title: string;
  icon: string;
  description?: string;
  featureName: (typeof FEATURES)[keyof typeof FEATURES];
}

export default LockedIntegrationCard;
