import {
  OnboardingStepsConfig,
  PROJECT_CREATED,
  SDK_SETUP,
  VISITOR_IDENTIFICATION_SETUP,
  MORE_INFO_FORM
} from '../ui/types';

export const generateCopyText = (script: string) => {
  return `
        ${CopyTitle}

        ${CopySdkTitle}

        ${script}

        Note: ${CopyNote}

        ${CopyOption}

        ${CopyOption1}
        ${CopyOption1Title}

        ${SDK_FLOW.GTM.step1}
        ${SDK_FLOW.GTM.step2}
        ${SDK_FLOW.GTM.step3}
        ${SDK_FLOW.GTM.step4}
        ${SDK_FLOW.GTM.step5}
        ${SDK_FLOW.GTM.step6}

        
        ${CopyOption2}
        ${CopyOption2Title}
        ${CopyOption2Desc}
            
    `;
};

export const SDK_FLOW = {
  GTM: {
    step1:
      '1. Sign in to Google Tag Manager, select “Workspace”, and “Add a new tag”',
    step2: '2. Name it “Factors tag”. Select Edit on Tag Configuration',
    step3: '3. Under custom, select custom HTML',
    step4:
      '4. Copy the above tracking script and paste it on the HTML field, Select Save',
    step5: '5. In the Triggers popup, select Add Trigger and select All Pages',
    step6:
      '6. The trigger has been added. Click on Publish at the top of your GTM window!'
  }
};

export const CopyTitle = 'Implementation instructions for Factors SDK';
export const CopySdkTitle = 'Factors SDK';
export const CopyNote =
  "This SDK is specific to your project, please don't share it with anyone outside your orgnisation.";
export const CopyOption =
  'You can implement this SDK with either of these methods';
export const CopyOption1 = '1. GTM Setup';
export const CopyOption1Title =
  'Add Factors SDK quickly using Google Tag Manager without any engineering effort';
export const CopyOption2 = '2. Manual setup';
export const CopyOption2Title =
  'Add Factors SDK manually in the head section for all pages you wish to get data for';
export const CopyOption2Desc =
  'Add the above javascript code on every page between the <head> and </head> tags.';

const isOnboardingStepCompleted = (config: any, steps: string[]) => {
  let isStepOnboarded = true;
  steps.forEach((step) => {
    if (!config[step]) {
      isStepOnboarded = false;
    }
  });
  return isStepOnboarded;
};

export const getCurrentStep = (onboardingConfig: OnboardingStepsConfig) => {
  if (!onboardingConfig) return 1;
  if (
    isOnboardingStepCompleted(onboardingConfig, [
      PROJECT_CREATED,
      SDK_SETUP,
      VISITOR_IDENTIFICATION_SETUP,
      MORE_INFO_FORM
    ])
  )
    return 5;
  if (
    isOnboardingStepCompleted(onboardingConfig, [
      PROJECT_CREATED,
      SDK_SETUP,
      VISITOR_IDENTIFICATION_SETUP
    ])
  )
    return 4;
  if (isOnboardingStepCompleted(onboardingConfig, [PROJECT_CREATED, SDK_SETUP]))
    return 3;
  if (isOnboardingStepCompleted(onboardingConfig, [PROJECT_CREATED])) return 2;
  return 1;
};
