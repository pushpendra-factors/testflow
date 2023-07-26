import React from 'react';
import { IntegrationConfig } from '../types';
import { Avatar, Button } from 'antd';
import { SVG, Text } from 'Components/factorsComponents';
import UpgradeButton from 'Components/GenericComponents/UpgradeButton';
import { useHistory } from 'react-router-dom';
import { PathUrls } from 'Routes/pathUrls';

const LockedIntegrationCard = ({
  integrationConfig
}: LockedIntegrationCardProps) => {
  const history = useHistory();
  return (
    <div className='fa-intergration-card' style={{ background: '#FAFAFA' }}>
      <div
        className='flex justify-between cursor-pointer'
        onClick={() => history.push(PathUrls.SettingsPricing)}
      >
        <div className='flex'>
          <Avatar
            size={40}
            shape='square'
            icon={
              <SVG name={integrationConfig.icon} size={40} color='purple' />
            }
            style={{ backgroundColor: '#F5F6F8' }}
          />
        </div>
        <div className='flex flex-col justify-start items-start ml-4 w-full'>
          <div className='flex flex-row items-center justify-start'>
            <Text type='title' level={5} weight='bold' extraClass='m-0'>
              {integrationConfig.name}
            </Text>
          </div>

          <Text
            type='paragraph'
            mini
            extraClass='m-0 w-9/12'
            color='grey'
            lineHeight='medium'
          >
            {integrationConfig.desc}
          </Text>
          <div className={'mt-4 flex gap-2'} data-tour='step-11'>
            <Button type={'primary'} disabled={true}>
              Connect Now
            </Button>
            <Button disabled>View documentation</Button>
          </div>
        </div>
        <UpgradeButton />
      </div>
    </div>
  );
};

type LockedIntegrationCardProps = {
  integrationConfig: IntegrationConfig;
};

export default LockedIntegrationCard;
