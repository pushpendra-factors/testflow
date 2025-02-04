import React, { useEffect, useState } from 'react';
import { bindActionCreators } from 'redux';
import { ConnectedProps, connect, useSelector } from 'react-redux';
import { setShowAnalyticsResult } from 'Reducers/coreQuery/actions';
import useQuery from 'hooks/useQuery';
import { useProductFruitsApi } from 'react-product-fruits';
import { useHistory } from 'react-router-dom';
import { fetchProjectSettings, getActiveProjectDetails } from 'Reducers/global';
import { isEmpty } from 'lodash';
import { PathUrls } from 'Routes/pathUrls';
import OnboardingLayout from './components/OnboardingLayout';
import onboardingStep1Image from '../../../assets/images/onboarding_step1.png';
import onboardingStep2Image from '../../../assets/images/onboarding_step2.png';
import onboardingStep3Image from '../../../assets/images/onboarding_step3.png';
import onboardingStep4Image from '../../../assets/images/onboarding_step4.png';

import Step1 from './components/OnboardingSteps/CreateProject';
import Step2 from './components/OnboardingSteps/SetupSdk';
import Step4 from './components/OnboardingSteps/TypformDetails';
import Step3 from './components/OnboardingSteps/VisitorIdentificationSetup';
import Step5 from './components/OnboardingSteps/AfterSetupScreen';
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

function Onboarding({
  setShowAnalyticsResult,
  getActiveProjectDetails
}: OnboardingComponentProps) {
  const [currentStep, setCurrentStep] = useState<number>(1);
  const history = useHistory();
  const routerQuery = useQuery();
  const paramTarget = routerQuery.get('target');
  const paramSetup = routerQuery.get('setup');
  const { active_project } = useSelector((state) => state.global);

  const onboarding_steps: OnboardingStepsConfig = useSelector(
    (state) => state?.global?.currentProjectSettings?.onboarding_steps
  );
  const { projects } = useSelector((state) => state.global);

  const showCloseButton = paramSetup === 'new' || currentStep === 5;

  const incrementStepCount = () => {
    if (currentStep < 5) setCurrentStep(currentStep + 1);
  };

  const handleCloseClick = () => {
    if (paramSetup === 'new') {
      history.goBack();
      return;
    }
    if (currentStep === 5) {
      history.push(PathUrls.ProfileAccounts);
      return;
    }
    // going back to previous project
    if (projects.length) {
      const activeItem = projects?.filter(
        (item) => item.id === localStorage.getItem('prevActiveProject')
      );

      // handling project redirection
      let projectDetails = isEmpty(activeItem) ? projects[0] : activeItem[0];
      // if previous projectId and current projectId are same then changing the project
      if (projectDetails?.id === active_project?.id) {
        if (projects[0]?.id !== active_project?.id) {
          projectDetails = projects[0];
        } else if (
          projects?.length > 1 &&
          projects[1]?.id !== active_project?.id
        ) {
          projectDetails = projects[1];
        }
      }
      localStorage.setItem(
        'prevActiveProject',
        localStorage.getItem('activeProject') || ''
      );
      localStorage.setItem('activeProject', projectDetails?.id);
      getActiveProjectDetails(projectDetails?.id);
      fetchProjectSettings(projectDetails?.id);
      history.push(PathUrls.ProfileAccounts);
    }
  };

  const decrementStepCount = () => {
    if (currentStep > 0) setCurrentStep(currentStep - 1);
  };

  // Effect for hiding the side panel and menu
  useEffect(() => {
    setShowAnalyticsResult(true);

    return () => {
      setShowAnalyticsResult(false);
    };
  }, [setShowAnalyticsResult]);

  // hiding product fruits help center button
  useProductFruitsApi((api) => {
    api.button.hide();
    return () => {
      api.button.show();
      api.button.close();
    };
  }, []);

  useEffect(() => {
    if (paramSetup === 'new') {
      setCurrentStep(1);
      return;
    }
    const step = getCurrentStep(onboarding_steps);
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
    return (
      <Step5
        handleCloseClick={handleCloseClick}
        showCloseButton={showCloseButton}
      />
    );
  }
  return (
    <OnboardingLayout
      currentStep={currentStep}
      stepImage={getIllustrationImage(currentStep)}
      totalSteps={5}
      showCloseButton={showCloseButton}
      handleCloseClick={handleCloseClick}
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
    </OnboardingLayout>
  );
}

const mapDispatchToProps = (dispatch) =>
  bindActionCreators(
    {
      setShowAnalyticsResult,
      getActiveProjectDetails
    },
    dispatch
  );
const connector = connect(null, mapDispatchToProps);
type OnboardingComponentProps = ConnectedProps<typeof connector>;

export default connector(Onboarding);
