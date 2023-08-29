import React from 'react';
import StepsCounter from 'Components/GenericComponents/StepsCounter';
import { SVG, Text } from 'Components/factorsComponents';
import { Button, Dropdown, Menu } from 'antd';
import useMobileView from 'hooks/useMobileView';

const OnboardingHeader = ({
  currentStep,
  totalSteps,
  showStepsCounter = true,
  showCloseButton = false,
  handleCloseClick
}: OnboardingHeaderProps) => {
  const isMobileView = useMobileView();

  const handleHelpClick = () => {
    if (window?.Intercom) {
      window.Intercom(
        'showNewMessage',
        `Hey, I have few doubts regarding onboarding! Can you guys help me out?`
      );
    }
  };

  const getMobileMenu = () => {
    return (
      <Menu>
        <Menu.Item key='1'>
          <StepsCounter currentStep={currentStep} totalSteps={totalSteps} />
        </Menu.Item>
        <Menu.Item key='2'>
          <Button
            icon={<SVG name='Headset' size='16' color='#8C8C8C' />}
            onClick={handleHelpClick}
          >
            Need help?
          </Button>
        </Menu.Item>
        {showCloseButton && (
          <Menu.Item key='3'>
            <Button onClick={handleCloseClick}>Close</Button>
          </Menu.Item>
        )}
      </Menu>
    );
  };
  return (
    <div
      className='flex justify-between items-center'
      style={{
        boxShadow: '0px 1px 1px 0px rgba(0, 0, 0, 0.10)',
        padding: isMobileView ? '0px 4px' : '0px 36px',
        height: '64px'
      }}
    >
      <div className='flex items-center gap-1'>
        <SVG name='brand' size='32' />
        <Text
          type={'title'}
          level={5}
          color='character-title'
          extraClass='m-0'
          weight={'bold'}
        >
          Project Setup Wizard
        </Text>
      </div>
      {isMobileView && (
        <>
          <Dropdown placement='bottomRight' overlay={getMobileMenu()}>
            <Button icon={<SVG name='Bars' size='16' />} />
          </Dropdown>
        </>
      )}

      {!isMobileView && (
        <div className='flex flex-row-reverse gap-10'>
          <div className='flex gap-2'>
            <Button
              icon={<SVG name='Headset' size='16' color='#8C8C8C' />}
              onClick={handleHelpClick}
            >
              Need help?
            </Button>
            {showCloseButton && (
              <Button onClick={handleCloseClick}>Close</Button>
            )}
          </div>

          {showStepsCounter && (
            <StepsCounter currentStep={currentStep} totalSteps={totalSteps} />
          )}
        </div>
      )}
    </div>
  );
};

interface OnboardingHeaderProps {
  totalSteps: number;
  currentStep: number;
  showStepsCounter?: boolean;
  showCloseButton?: boolean;
  handleCloseClick?: () => void;
}

export default OnboardingHeader;
