import {
  FaErrorComp,
  FaErrorLog,
  SVG,
  Text
} from 'Components/factorsComponents';
import { Avatar, Button, Tag, Tooltip } from 'antd';
import React, { useEffect, useState } from 'react';
import { ErrorBoundary } from 'react-error-boundary';
import { IntegrationConfig } from '../types';
import useFeatureLock from 'hooks/useFeatureLock';
import LockedIntegrationCard from '../LockedIntegrationCard';

function IntegrationCard({
  integrationConfig,
  defaultOpen
}: IntegrationCardProps) {
  const [isActive, setIsActive] = useState(false);
  const [toggle, setToggle] = useState(false);
  const [isStatus, setIsStatus] = useState('');
  const { isFeatureLocked } = useFeatureLock(integrationConfig.featureName);

  const loadIntegrationForm = () => {
    const { Component } = integrationConfig;
    if (Component) {
      return (
        <Component
          kbLink={integrationConfig.kbLink}
          setIsActive={setIsActive}
          setIsStatus={setIsStatus}
        />
      );
    }
    return (
      <>
        <Tag color='orange' style={{ marginTop: '8px' }}>
          Enable from{' '}
          <a
            href='https://app-old.factors.ai/'
            target='_blank'
            rel='noreferrer'
          >
            here
          </a>
        </Tag>{' '}
      </>
    );
  };

  useEffect(() => {
    setToggle(!(isActive || isStatus === 'Active'));

    if (defaultOpen) {
      setToggle(true);
    }
  }, [isActive, isStatus]);

  if (isFeatureLocked) {
    return <LockedIntegrationCard integrationConfig={integrationConfig} />;
  }

  return (
    <div className='fa-intergration-card'>
      <ErrorBoundary
        fallback={
          <FaErrorComp
            size='medium'
            title='Bundle Error:02'
            subtitle='We are facing trouble loading App Bundles. Drop us a message on the in-app chat.'
          />
        }
        onError={FaErrorLog}
      >
        <div>
          <div
            className='flex justify-between cursor-pointer'
            onClick={() =>
              isActive || isStatus === 'Active' ? setToggle(!toggle) : null
            }
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
                {(isActive || isStatus === 'Active') && (
                  <Tag color='green' style={{ marginLeft: '8px' }}>
                    Active
                  </Tag>
                )}
              </div>

              {isStatus === 'Pending' && (
                <Tooltip
                  title={
                    integrationConfig.name === 'Google Ads'
                      ? 'Account(s) Selection Pending.'
                      : 'URL(s) Selection Pending.'
                  }
                >
                  <Tag color='orange' style={{ marginLeft: '8px' }}>
                    Pending!
                  </Tag>
                </Tooltip>
              )}
              <Text
                type='paragraph'
                mini
                extraClass='m-0 w-9/12'
                color='grey'
                lineHeight='medium'
              >
                {integrationConfig.desc}
              </Text>
            </div>
            {(isActive || isStatus === 'Active') && (
              <Button
                type='text'
                onClick={() => setToggle(!toggle)}
                icon={
                  toggle ? (
                    <SVG size={16} name='ChevronDown' />
                  ) : (
                    <SVG size={16} name='ChevronRight' />
                  )
                }
              />
            )}
          </div>
          <div className='ml-16 flex flex-col items-start'>
            {toggle && loadIntegrationForm()}
          </div>
        </div>
      </ErrorBoundary>
    </div>
  );
}

type IntegrationCardProps = {
  defaultOpen: boolean;
  integrationConfig: IntegrationConfig;
};

export default IntegrationCard;
