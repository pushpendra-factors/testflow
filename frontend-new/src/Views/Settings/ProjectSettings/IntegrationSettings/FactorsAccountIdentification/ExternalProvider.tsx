import React from 'react';
import { Avatar, Divider } from 'antd';
import { SVG, Text } from 'Components/factorsComponents';
import { AccountIdentificationProviderData } from '../integrations.constants';

const ExternalProvider = () => (
  <div className='mt-8'>
    {AccountIdentificationProviderData.map((config, i) => (
      <>
        <div className='flex justify-between '>
          <div className='flex '>
            <Avatar
              size={50}
              shape='square'
              icon={<SVG name={config.icon} size={50} color='purple' />}
              style={{ backgroundColor: '#F5F6F8' }}
            />
          </div>
          <div className='flex flex-col justify-start items-start ml-4 w-full'>
            <Text
              type='title'
              level={5}
              weight='bold'
              extraClass='m-0'
              color='character-primary'
            >
              {config.name}
            </Text>

            <Text
              type='title'
              level={7}
              extraClass='m-0 w-9/12'
              lineHeight='medium'
              color='character-secondary'
            >
              {config.desc}
            </Text>
            <div>
              <config.Component kbLink={config.kbLink} />
            </div>
          </div>
        </div>

        {i !== AccountIdentificationProviderData.length - 1 && <Divider />}
      </>
    ))}
    {/* <SixSignal />
    <Divider />
    <Reveal /> */}
  </div>
);

export default ExternalProvider;
