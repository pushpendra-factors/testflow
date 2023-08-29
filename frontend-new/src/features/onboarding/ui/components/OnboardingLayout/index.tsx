import React from 'react';
import OnboardingHeader from '../OnboardingHeader';
import useMobileView from 'hooks/useMobileView';
import style from './index.module.scss';

const OnboardingLayout: React.FC<OnboardingLayoutProps> = ({
  children,
  totalSteps,
  currentStep,
  stepImage,
  showStepsCounter,
  showCloseButton,
  handleCloseClick
}) => {
  const isMobileView = useMobileView();
  const renderIllustrationImage = () => (
    <div
      style={{
        marginTop: !isMobileView ? 118 : 24,
        width: isMobileView ? 208 : '100%'
      }}
    >
      <img src={stepImage} alt='illustration' />
    </div>
  );
  return (
    <div style={{ height: isMobileView ? '100%' : '100vh' }}>
      <OnboardingHeader
        totalSteps={totalSteps}
        currentStep={currentStep}
        showStepsCounter={showStepsCounter}
        showCloseButton={showCloseButton}
        handleCloseClick={handleCloseClick}
      />

      {isMobileView && (
        <div className={style.contentContainerMobile}>
          <div
            className={`flex justify-center items-center ${style.contentImageMobile}`}
          >
            {renderIllustrationImage()}
          </div>
          <div className={style.contentMobileChild}>{children}</div>
        </div>
      )}

      {!isMobileView && (
        <div className={`flex ${style.content}`}>
          <div
            className='h-full px-5'
            style={{ background: '#E5EEFF', width: 398 }}
          >
            {renderIllustrationImage()}
          </div>
          <div className={style.contentDesktopChild}>{children}</div>
        </div>
      )}
    </div>
  );
};

interface OnboardingLayoutProps {
  stepImage: string;
  currentStep: number;
  totalSteps: number;
  showStepsCounter?: boolean;
  showCloseButton?: boolean;
  handleCloseClick?: () => void;
}

export default OnboardingLayout;
