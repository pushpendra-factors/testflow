import { FEATURES } from 'Constants/plans.constants';
import AccountScoringIllustration from '../../../assets/images/illustrations/AccountScoringIllustration.png';
import AutomationIllustration from '../../../assets/images/illustrations/AutomationIllustration.png';
import G2Illustration from '../../../assets/images/illustrations/G2Illustration.png';
import LinkedinIllustration from '../../../assets/images/illustrations/LinkedinIllustration.png';

export interface UpgradeNudgeConfig {
  name: string;
  title: string;
  description: string;
  image: string;
  backgroundColor: string;
  featureName: string;
  videoId?: string;
}

export const UpgradeNudges: UpgradeNudgeConfig[] = [
  {
    name: 'Account Scoring',
    title: 'Unlock Success with Advanced Account Scoring!',
    description:
      'Harness the Power of Account Scoring to Optimize Client Relationships and Maximize Revenue in Real Time.',
    image: AccountScoringIllustration,
    backgroundColor: '#120338',
    featureName: FEATURES.FEATURE_ACCOUNT_SCORING,
    videoId: 'sbgrCYaAnwQ'
  },
  {
    name: 'Linkedin',
    title: 'Identify companies engaging with your ads',
    description:
      'Find out which companies are seeing or clicking your LinkedIn ads. Use this intent signal to go after companies engaging with your brand.',
    image: LinkedinIllustration,
    backgroundColor: '#0050B3',
    featureName: FEATURES.FEATURE_LINKEDIN
  },
  {
    name: 'G2',
    title: 'Identify companies that are researching about you on G2',
    description:
      'Find out which companies are checking out your G2 page or comparing you against competitors.',
    image: G2Illustration,
    backgroundColor: '#FF7A45',
    featureName: FEATURES.FEATURE_G2,
    videoId: 'D04fvnLc1xg'
  },
  {
    name: 'Workflows',
    title: 'Setup data syncs and automations',
    description:
      'Sync Factors data with your internal tools like your CRM or setup automations like auto enrolling into sales sequences and building a dynamic audience on LinkedIn',
    image: AutomationIllustration,
    backgroundColor: '#262626',
    featureName: FEATURES.FEATURE_WORKFLOWS
  }
];
