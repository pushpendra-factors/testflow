import React from 'react';
import { SVG, Text } from 'Components/factorsComponents';
import ProgressBar from 'Components/GenericComponents/Progress';
import { useSelector } from 'react-redux';
import { FeatureConfigState } from 'Reducers/featureConfig/types';
import { useHistory } from 'react-router-dom';
import { PathUrls } from 'Routes/pathUrls';
import { PRICING_PAGE_TABS } from '../../Pricing/utils';

const ConnectedScreen = () => {
  const { sixSignalInfo } = useSelector(
    (state: any) => state.featureConfig
  ) as FeatureConfigState;
  const history = useHistory();
  const sixSignalLimit = sixSignalInfo?.limit || 0;
  const sixSignalUsage = sixSignalInfo?.usage || 0;
  return (
    <div className='mt-4 flex flex-col border-top--thin  py-4 w-full'>
      <div>
        <div className='flex justify-between items-center'>
          <div className='flex items-center justify-start gap-2'>
            <Text type={'paragraph'} mini>
              Default Monthly Quota
            </Text>
            <div
              className='flex items-center justify-start gap-1 cursor-pointer'
              onClick={() =>
                history.push(
                  `${PathUrls.SettingsPricing}?activeTab=${PRICING_PAGE_TABS.ENRICHMENT_RULES}`
                )
              }
            >
              <SVG name='ArrowUpRightSquare' color='#40A9FF' />

              <Text type={'paragraph'} mini color='brand-color'>
                Enrichment rules
              </Text>
            </div>
          </div>

          <Text type={'paragraph'} mini>
            {`${sixSignalUsage} / ${sixSignalLimit}`}
          </Text>
        </div>
        <ProgressBar percentage={(sixSignalUsage / sixSignalLimit) * 100} />
      </div>
    </div>
  );
};

export default ConnectedScreen;
