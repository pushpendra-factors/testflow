import React from 'react';
import EnrichFeature from './EnrichFeature';
import { SVG, Text } from 'Components/factorsComponents';
import ProgressBar from 'Components/GenericComponents/Progress';
import { useSelector } from 'react-redux';
import { FeatureConfigState } from 'Reducers/featureConfig/types';

const ConnectedScreen = () => {
  const { sixSignalInfo } = useSelector(
    (state: any) => state.featureConfig
  ) as FeatureConfigState;
  const sixSignalLimit = sixSignalInfo?.limit || 0;
  const sixSignalUsage = sixSignalInfo?.usage || 0;
  return (
    <div className='mt-4 flex flex-col border-top--thin  py-4 w-full'>
      <div>
        <div className='flex justify-between items-center'>
          <div>
            <Text type={'paragraph'} mini>
              Default Monthly Quota
            </Text>
            {/* <div>
            <SVG name='ArrowUpRightSquare' color='#40A9FF' />

              <Text type={'paragraph'} mini color='brand-color'>
                Buy add on
              </Text>
            </div> */}
          </div>

          <Text type={'paragraph'} mini>
            {`${sixSignalUsage} / ${sixSignalLimit}`}
          </Text>
        </div>
        <ProgressBar percentage={(sixSignalUsage / sixSignalLimit) * 100} />
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

export default ConnectedScreen;
