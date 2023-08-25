export type CommonStepsProps = {
  incrementStepCount: () => void;
  decrementStepCount: () => void;
};

export const PROJECT_CREATED = 'project_created';
export const SDK_SETUP = 'sdk_setup';
export const VISITOR_IDENTIFICATION_SETUP = 'visitor_identification_setup';
export const MORE_INFO_FORM = 'more_info_form';
export const SETUP_COMPLETED = 'setup_completed';

export interface OnboardingStepsConfig {
  [PROJECT_CREATED]: boolean;
  [SDK_SETUP]: boolean;
  [VISITOR_IDENTIFICATION_SETUP]: boolean;
  [MORE_INFO_FORM]: boolean;
  [SETUP_COMPLETED]: boolean;
}
