import React, { useEffect, useState } from 'react';
import OnboardingLayout from './components/OnboardingLayout';
import { bindActionCreators } from 'redux';
import { ConnectedProps, connect, useSelector } from 'react-redux';
import { setShowAnalyticsResult } from 'Reducers/coreQuery/actions';
import onboardingStep1Image from '../../../assets/images/onboarding_step1.png';
import onboardingStep2Image from '../../../assets/images/onboarding_step2.png';
import onboardingStep3Image from '../../../assets/images/onboarding_step3.png';
import onboardingStep4Image from '../../../assets/images/onboarding_step4.png';

import Step1 from './components/OnboardingSteps/CreateProject';
import Step2 from './components/OnboardingSteps/SetupSdk';
import Step4 from './components/OnboardingSteps/TypformDetails';
import Step3 from './components/OnboardingSteps/VisitorIdentificationSetup';
import Step5 from './components/OnboardingSteps/AfterSetupScreen';
import useQuery from 'hooks/useQuery';
import { OnboardingStepsConfig } from './types';
import { getCurrentStep } from '../utils';

const getIllustrationImage = (currentStep: number): string => {
  switch (currentStep) {
    case 1:
      return onboardingStep1Image;
    case 2:
      return onboardingStep2Image;
    case 3:
      return onboardingStep3Image;
    case 4:
      return onboardingStep4Image;
    default:
      return onboardingStep1Image;
  }
};

const Onboarding = ({ setShowAnalyticsResult }: OnboardingComponentProps) => {
  const [currentStep, setCurrentStep] = useState<number>(1);
  const routerQuery = useQuery();
  const paramTarget = routerQuery.get('target');
  const paramSetup = routerQuery.get('setup');

  const onboarding_steps: OnboardingStepsConfig = useSelector(
    (state) => state?.global?.currentProjectSettings?.onboarding_steps
  );
  const incrementStepCount = () => {
    if (currentStep < 5) setCurrentStep(currentStep + 1);
  };

  const decrementStepCount = () => {
    if (currentStep > 0) setCurrentStep(currentStep - 1);
  };

  //Effect for hiding the side panel and menu
  useEffect(() => {
    setShowAnalyticsResult(true);

    return () => {
      setShowAnalyticsResult(false);
    };
  }, [setShowAnalyticsResult]);

  useEffect(() => {
    if (paramSetup === 'new') {
      setCurrentStep(1);
      return;
    }
    let step = getCurrentStep(onboarding_steps);
    setCurrentStep(step);
  }, [onboarding_steps, paramSetup]);

  if (paramTarget === 'invited_user') {
    return (
      <OnboardingLayout
        showStepsCounter={false}
        stepImage={onboardingStep4Image}
        currentStep={0}
        totalSteps={0}
      >
        <Step4
          variant='invitedUser'
          incrementStepCount={incrementStepCount}
          decrementStepCount={decrementStepCount}
        />
      </OnboardingLayout>
    );
  }

  if (currentStep === 5) {
    return <Step5 />;
  }
  return (
    <OnboardingLayout
      currentStep={currentStep}
      stepImage={getIllustrationImage(currentStep)}
      totalSteps={5}
      showCloseButton={currentStep === 1 && paramSetup === 'new'}
    >
      {currentStep === 1 && (
        <Step1
          incrementStepCount={incrementStepCount}
          decrementStepCount={decrementStepCount}
        />
      )}
      {currentStep === 2 && (
        <Step2
          incrementStepCount={incrementStepCount}
          decrementStepCount={decrementStepCount}
        />
      )}
      {currentStep === 3 && (
        <Step3
          incrementStepCount={incrementStepCount}
          decrementStepCount={decrementStepCount}
        />
      )}
      {currentStep === 4 && (
        <Step4
          variant='admin'
          incrementStepCount={incrementStepCount}
          decrementStepCount={decrementStepCount}
        />
      )}
      {currentStep === 5 && <Step5 />}
    </OnboardingLayout>
  );
};

const mapDispatchToProps = (dispatch) =>
  bindActionCreators(
    {
      setShowAnalyticsResult
    },
    dispatch
  );
const connector = connect(null, mapDispatchToProps);
type OnboardingComponentProps = ConnectedProps<typeof connector>;

export default connector(Onboarding);
