import React from 'react';
import { Input } from 'antd';
import EnrichFeature from './EnrichFeature';
import { Text } from 'Components/factorsComponents';

const ConnectedScreen = ({ apiKey }: ConnectScreenProps) => {
  return (
    <div className='mt-4 flex flex-col border-top--thin  py-4 w-full'>
      <div>
        <Text type='title' level={7} color='grey' extraClass='mb-2'>
          API Key
        </Text>
        <Input
          size='large'
          disabled
          placeholder='API Key'
          value={apiKey}
          style={{ width: '400px', borderRadius: 6 }}
        />
      </div>
      <div className='mt-4'>
        <EnrichFeature
          type='page'
          title='Enrich for specific pages'
          subtitle='Gain insight into who is visiting your website and where they are in the buying journey'
        />
      </div>
      <div className='mt-4'>
        <EnrichFeature
          type='country'
          title='Enable country filtering'
          subtitle='Gain insight into who is visiting your website and where they are in the buying journey'
        />
      </div>
    </div>
  );
};

type ConnectScreenProps = {
  apiKey: string;
};

export default ConnectedScreen;
