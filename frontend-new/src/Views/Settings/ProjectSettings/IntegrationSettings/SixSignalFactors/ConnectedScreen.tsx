import React from 'react';
import { SVG, Text } from 'Components/factorsComponents';
import ProgressBar from 'Components/GenericComponents/Progress';
import { useSelector } from 'react-redux';
import { FeatureConfigState } from 'Reducers/featureConfig/types';
import { useHistory, useLocation } from 'react-router-dom';
import styles from './index.module.scss';

function ConnectedScreen() {
  const { sixSignalInfo } = useSelector(
    (state: any) => state.featureConfig
  ) as FeatureConfigState;
  const { currentProjectSettings } = useSelector((state) => state?.global);
  const location = useLocation();
  const history = useHistory();
  const sixSignalLimit = sixSignalInfo?.limit || 0;
  const sixSignalUsage = sixSignalInfo?.usage || 0;
  const isProviderClearbit =
    currentProjectSettings?.factors_deanon_config?.clearbit
      ?.traffic_fraction === 1;
  const renderProviderCard = () => (
    <div className={styles.providerCard}>
      <div className='flex items-center justify-center'>
        <SVG
          name={isProviderClearbit ? 'ClearbitLogo' : 'SixSignalLogo'}
          size={44}
          color='purple'
        />
      </div>

      <div style={{ width: 178 }}>
        <Text
          type='title'
          level={7}
          weight='bold'
          extraClass='m-0'
          color='character-primary'
        >
          {isProviderClearbit ? 'Clearbit Reveal' : '6Signal by 6Sense'}
        </Text>
        <Text
          type='title'
          level={8}
          extraClass='m-0'
          color='character-secondary'
        >
          Using {isProviderClearbit ? 'Clearbit Reveal' : '6Signal by 6Sense'}{' '}
          to identify accounts
        </Text>
      </div>
    </div>
  );

  const handleCustomerSupportClick = () => {
    if (window?.Intercom) {
      window.Intercom(
        'showNewMessage',
        'Hi, I want to change my account identification provider. Can you share details about this?'
      );
    }
  };
  return (
    <div className='mt-5 flex flex-col  w-full'>
      {/* <div>
        <Text
          type='title'
          level={6}
          color='character-primary'
          extraClass='m-0'
          weight='bold'
        >
          Integration Details
        </Text>
        <Text
          type='title'
          level={7}
          extraClass='m-0 mt-1'
          color='character-secondary'
        >
          Gain insight into accounts visiting your website and where they are in
          the buying journey.
        </Text>
      </div> */}
      <div>{renderProviderCard()}</div>
      <div className='mt-2'>
        <Text
          type='title'
          level={7}
          extraClass='m-0 mt-2'
          color='character-secondary'
        >
          Please contact our{' '}
          <span
            onClick={handleCustomerSupportClick}
            className='cursor-pointer'
            style={{ color: '#1890FF' }}
          >
            customer support
          </span>{' '}
          team if you want to change your de-anonymisation provider
        </Text>
      </div>
      <div className='mt-4'>
        <div className='flex justify-between items-center'>
          <div className='flex items-center justify-start gap-2'>
            <Text type='paragraph' mini>
              Default Monthly Quota
            </Text>
            <div
              className='flex items-center justify-start gap-1 cursor-pointer'
              onClick={() =>
                history.replace(
                  `${location.pathname}?activeTab=enrichmentRules`
                )
              }
            >
              <SVG name='ArrowUpRightSquare' size={14} color='#40A9FF' />

              <Text type='paragraph' mini color='brand-color'>
                Enrichment rules
              </Text>
            </div>
          </div>

          <Text type='paragraph' mini>
            {`${sixSignalUsage} / ${sixSignalLimit}`}
          </Text>
        </div>
        <ProgressBar percentage={(sixSignalUsage / sixSignalLimit) * 100} />
      </div>
    </div>
  );
}

export default ConnectedScreen;
