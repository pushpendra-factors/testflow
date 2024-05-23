import React, { useContext, useEffect, useRef } from 'react';
import { Alert, Avatar, Button, Divider } from 'antd';
import { SVG, Text } from 'Components/factorsComponents';
import useFeatureLock from 'hooks/useFeatureLock';
import UpgradeButton from 'Components/GenericComponents/UpgradeButton';
import usePlanUpgrade from 'hooks/usePlanUpgrade';
import { useHistory } from 'react-router-dom';
import { PathUrls } from 'Routes/pathUrls';
import { IntegrationInfoInterface } from 'hooks/useIntegrationCheck';
import { CloseCircleFilled, ExclamationCircleFilled } from '@ant-design/icons';
import useAgentInfo from 'hooks/useAgentInfo';
import { IntegrationConfig } from '../types';
import { INTEGRATION_ID } from '../integrations.constants';
import { IntegrationContext } from '../IntegrationContext';
import {
  getIntegrationActionText,
  getIntegrationStatus,
  showIntegrationStatus
} from '../util';

const IntegrationCard = ({
  integrationConfig,
  integrationInfo
}: CommonIntegrationCardProps) => {
  const history = useHistory();
  const { email: userEmail } = useAgentInfo();
  const cardRef = useRef<HTMLDivElement>(null);
  const showIntegrationStatusFlag = showIntegrationStatus(userEmail);
  const { icon, name, desc, featureName } = integrationConfig;

  const { isFeatureLocked } = useFeatureLock(featureName);

  const { handlePlanUpgradeClick } = usePlanUpgrade();

  const { integrationStatus } = useContext(IntegrationContext);

  const integrationStatusState = integrationStatus?.[featureName]?.state;
  const integrationStatusValue = getIntegrationStatus(
    integrationStatus?.[featureName]
  );
  const integrationActionText = getIntegrationActionText(
    integrationStatus?.[featureName]
  );
  const integrationStatusMessage = integrationStatus?.[featureName]?.message;

  const isErrorState = integrationStatusValue === 'error';

  const isPendingState = integrationStatusValue === 'pending';
  const isConnectedState = integrationStatusValue === 'connected';
  const isNotConnectedState = integrationStatusValue === 'not_connected';
  const isFeatureIntegrated = integrationInfo?.[featureName];

  const handleCardClick = () => {
    if (isFeatureLocked && featureName !== 'sdk') {
      handlePlanUpgradeClick(featureName);
      return;
    }
    const path = `${PathUrls.SettingsIntegration}/${integrationConfig.id}`;

    history.push(path);
  };

  const renderActionButton = () => {
    if (isFeatureLocked && featureName !== 'sdk')
      return <UpgradeButton featureName={integrationConfig.featureName} />;

    const ConnectNowButton = (
      <Button type='text' onClick={handleCardClick}>
        <div className='flex items-center gap-1'>
          {integrationConfig?.id !== INTEGRATION_ID.sdk && (
            <Text type='paragraph' mini weight='bold' color='brand-color-6'>
              Connect Now{' '}
            </Text>
          )}
          <SVG name='ChevronRight' size={20} />
        </div>
      </Button>
    );

    if (!showIntegrationStatusFlag) {
      if (!isFeatureIntegrated) {
        return ConnectNowButton;
      }
      return (
        <Button type='text' onClick={handleCardClick}>
          <div className='flex items-center gap-2'>
            <SVG
              size={18}
              name='CheckCircle'
              extraClass='inline'
              color='#52C41A'
            />

            <Text type='paragraph' mini color='grey' extraClass='m-0'>
              Connected
            </Text>

            <SVG name='ChevronRight' size={20} />
          </div>
        </Button>
      );
    }

    if (
      !isFeatureIntegrated ||
      !integrationStatus?.[featureName] ||
      integrationStatusState === ''
    ) {
      return ConnectNowButton;
    }
    if (isNotConnectedState) {
      return ConnectNowButton;
    }
    return (
      <Button type='text' onClick={handleCardClick}>
        <div className='flex items-center gap-2'>
          {isErrorState && (
            <CloseCircleFilled style={{ color: '#EA6262', fontSize: 15 }} />
          )}
          {isPendingState && (
            <ExclamationCircleFilled
              style={{ color: '#DEA069', fontSize: 15 }}
            />
          )}
          {isConnectedState && (
            <SVG
              size={18}
              name='CheckCircle'
              extraClass='inline'
              color='#52C41A'
            />
          )}

          <Text type='paragraph' mini color='grey' extraClass='m-0'>
            {integrationActionText}
          </Text>

          <SVG name='ChevronRight' size={20} />
        </div>
      </Button>
    );
  };

  let border;
  let backgroundColor;
  if (isFeatureLocked && featureName !== 'sdk') {
    backgroundColor = '#FAFAFA';
  }
  if (
    (featureName === 'sdk' &&
      (isErrorState || isPendingState) &&
      showIntegrationStatusFlag) ||
    (isFeatureLocked &&
      (isErrorState || isPendingState) &&
      showIntegrationStatusFlag)
  ) {
    border = `1px solid ${isPendingState ? '#DEA069' : '#EA6262'}`;
  }

  useEffect(() => {
    if (!cardRef?.current) return;
    if (sessionStorage.getItem('integration-card') === integrationConfig.id) {
      cardRef.current.scrollIntoView({ behavior: 'auto', block: 'center' });
      sessionStorage.setItem('integration-card', '');
    }
  }, []);

  return (
    <div
      className='fa-intergration-card'
      style={{
        background: backgroundColor || undefined,
        border: border || undefined
      }}
      ref={cardRef}
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
              {name}
            </Text>
          </div>

          <Text
            type='paragraph'
            mini
            extraClass='m-0 w-9/12'
            color='grey'
            lineHeight='medium'
          >
            {desc}
          </Text>
        </div>
        <div className='flex justify-center items-center'>
          {renderActionButton()}
        </div>
      </div>
      {showIntegrationStatusFlag &&
        isErrorState &&
        integrationStatusMessage && (
          <>
            <Divider />
            <Alert message={integrationStatusMessage} type='error' showIcon />
          </>
        )}
    </div>
  );
};

interface CommonIntegrationCardProps {
  integrationConfig: IntegrationConfig;
  integrationInfo: IntegrationInfoInterface;
}

export default IntegrationCard;
