import anchorme from 'anchorme';
import {
  OnboardingStepsConfig,
  PROJECT_CREATED,
  SDK_SETUP,
  VISITOR_IDENTIFICATION_SETUP,
  MORE_INFO_FORM
} from '../ui/types';

export const SDK_FLOW = {
  GTM: {
    step1:
      '1. Sign in to Google Tag Manager(https://tagmanager.google.com/), select “Workspace”, and “Add a new tag”',
    step2: '2. Name it “Factors tag”. Select Edit on Tag Configuration',
    step3: '3. Under custom, select custom HTML',
    step4:
      '4. Copy the above tracking script and paste it on the HTML field, Select Save',
    step5: '5. In the Triggers popup, select Add Trigger and select All Pages',
    step6:
      '6. The trigger has been added. Click on Publish at the top of your GTM window!'
  },
  helpText:
    'SDK still says "Not verified"? Check these steps(https://help.factors.ai/en/articles/7260638-connecting-factors-to-your-website)'
};

export const OnboardingSupportLink =
  'https://factors.schedulehero.io/meet/naveena/integration-call';

export const CopyTitle = 'Connecting Factors to your website';
export const CopySdkTitle = 'Factors SDK snippet';
export const CopyNote =
  "This SDK is specific to your project, please don't share it with anyone outside your orgnisation.";
export const CopyOption =
  'You can implement this SDK with either of these methods';
export const CopyOption1 = '1. GTM Setup';
export const CopyOption2 = '2. Manual setup';
export const CopyOption2Desc =
  'Add the above javascript code on every page between the <head> and </head> tags.';

export const generateCopyText = (script: string) => `
  ${CopyTitle}

  ${CopySdkTitle}

  ${script}

  Note: ${CopyNote}

  ${CopyOption}

  ${CopyOption1}

    ${SDK_FLOW.GTM.step1}
    ${SDK_FLOW.GTM.step2}
    ${SDK_FLOW.GTM.step3}
    ${SDK_FLOW.GTM.step4}
    ${SDK_FLOW.GTM.step5}
    ${SDK_FLOW.GTM.step6}

  
  ${CopyOption2}
    1.${CopyOption2Desc}
      


  If you have any questions or issues, please reach out to our support team(${OnboardingSupportLink}) for assistance. 
`;

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

export const getCompanyDomainFromEmail = (email: string) => {
  if (!email) return '';
  const domain = email?.split('@')[1];
  if (!domain) return '';
  return domain;
};

export const SDKDocumentation =
  'https://help.factors.ai/en/articles/8999789-setting-up-factors-sdk-on-your-site';

export const ClearbitTermsOfUseLink =
  'https://www.factors.ai/customer-agreement';

export function extractDomainFromUrl(url: string): string | null {
  // validating url
  if (!anchorme.validate.url(url)) {
    return null;
  }
  const regex = /^(?:https?:\/\/)?(?:www\.)?([^\/\?]+)/i;
  const match = url.match(regex);

  return (match && match?.[1]) || null;
}
