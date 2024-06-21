import React from 'react';
import { Avatar, Divider } from 'antd';
import { SVG, Text } from 'Components/factorsComponents';
import useFeatureLock from 'hooks/useFeatureLock';
import { AccountIdentificationProviderData } from '../integrations.constants';
import { IntegrationConfig } from '../types';
import LockedIntegrationCard from '../IntegrationCard/LockedIntegrationCard';

const IndividualCard = ({
  featureName,
  name,
  desc,
  icon,
  kbLink,
  Component
}: IntegrationConfig) => {
  const { isFeatureLocked } = useFeatureLock(featureName);
  if (isFeatureLocked) {
    return (
      <LockedIntegrationCard
        title={name}
        description={desc}
        icon={icon}
        featureName={featureName}
      />
    );
  }
  return (
    <div className='flex justify-between '>
      <div className='flex '>
        <Avatar
          size={40}
          shape='square'
          icon={<SVG name={icon} size={40} />}
          style={{ backgroundColor: '#F5F6F8' }}
        />
      </div>
      <div className='flex flex-col justify-start items-start ml-4 w-full'>
        <Text
          type='title'
          level={6}
          weight='bold'
          extraClass='m-0'
          color='character-primary'
        >
          {name}
        </Text>

        <Text
          type='title'
          level={7}
          extraClass='m-0 w-9/12'
          lineHeight='medium'
          color='character-secondary'
        >
          {desc}
        </Text>
        <div>
          <Component kbLink={kbLink} />
        </div>
      </div>
    </div>
  );
};

const ExternalProvider = () => (
  <div className='mt-5 mb-5 py-2'>
    {AccountIdentificationProviderData.map((config, i) => (
      <>
        <IndividualCard {...config} />
        {i !== AccountIdentificationProviderData.length - 1 && <Divider />}
      </>
    ))}
  </div>
);

export default ExternalProvider;
