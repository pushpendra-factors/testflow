import { FEATURES } from 'Constants/plans.constants';
import { IntegrationConfig } from './types';
import Segment from './Segment';
import Rudderstack from './Rudderstack';
import Marketo from './Marketo';
import Slack from './Slack';
import MSTeam from './MSTeam';
import Hubspot from './Hubspot';
import Salesforce from './Salesforce';
import GoogleAdWords from './GoogleAdWords';
import Facebook from './Facebook';
import LinkedIn from './LinkedIn';
import Drift from './Drift';
import GoogleSearchConsole from './GoogleSearchConsole';
import Bing from './Bing';
import Reveal from './Reveal';
import LeadSquared from './LeadSquared';
import SixSignal from './SixSignal';
import SixSignalFactors from './SixSignalFactors';
import G2 from './G2';

export const IntegrationProviderData: IntegrationConfig[] = [
  {
    name: 'Segment',
    desc: 'Segment is a Customer Data Platform (CDP) that simplifies collecting and using data from the users of your digital properties and SaaS applications',
    icon: 'Segment_ads',
    kbLink: 'https://help.factors.ai/en/articles/7261994-segment-integration',
    featureName: FEATURES.INT_SEGMENT,
    Component: Segment
  },
  {
    name: 'Rudderstack',
    desc: 'Rudderstack is a Customer Data Platform (CDP) that simplifies collecting and using data from the users of your digital properties and SaaS applications',
    icon: 'Rudderstack_ads',
    kbLink:
      'https://help.factors.ai/en/articles/7283693-rudderstack-integration',
    featureName: FEATURES.INT_RUDDERSTACK,
    Component: Rudderstack
  },
  {
    name: 'Marketo',
    desc: 'Marketo is a leader in marketing automation. Using our Marketo source, we will ingest your Program, Campaign, Person and List records into Factors',
    icon: 'Marketo',
    featureName: FEATURES.INT_MARKETO,
    Component: Marketo
  },
  {
    name: 'Slack',
    desc: 'Does your team live on Slack? Set up alerts that track KPIs and marketing data. Nudge your team to take the right actions.',
    icon: 'Slack',
    kbLink: 'https://help.factors.ai/en/articles/7283808-slack-integration',
    featureName: FEATURES.INT_SLACK,
    Component: Slack
  },
  {
    name: 'Microsoft Teams',
    desc: 'Does your team live on Teams? Set up alerts that track KPIs and marketing data. Nudge your team to take the right actions.',
    icon: 'MSTeam',
    kbLink:
      'https://help.factors.ai/en/articles/7913152-microsoft-teams-integration',
    featureName: FEATURES.INT_TEAMS,
    Component: MSTeam
  },
  {
    name: 'Hubspot',
    desc: 'Sync your Contact, Company and Deal objects with Factors on a daily basis',
    icon: 'Hubspot_ads',
    kbLink: 'https://help.factors.ai/en/articles/7261985-hubspot-integration',
    featureName: FEATURES.INT_HUBSPOT,
    Component: Hubspot
  },
  {
    name: 'Salesforce',
    desc: 'Sync your Leads, Contact, Account, Opportunity and Campaign objects with Factors on a daily basis',
    icon: 'Salesforce_ads',
    kbLink:
      'https://help.factors.ai/en/articles/7261989-salesforce-integration',
    featureName: FEATURES.INT_SALESFORCE,
    Component: Salesforce
  },
  {
    name: 'Google Ads',
    desc: 'Integrate reporting from Google Search, Youtube and Display Network',
    icon: 'Google_ads',
    kbLink:
      'https://help.factors.ai/en/articles/7283695-google-ads-integration',
    featureName: FEATURES.INT_ADWORDS,
    Component: GoogleAdWords
  },
  {
    name: 'Facebook',
    desc: 'Pull in reports from Facebook, Instagram and Facebook Audience Network',
    icon: 'Facebook_ads',
    kbLink:
      'https://help.factors.ai/en/articles/7283696-facebook-ads-integration',
    featureName: FEATURES.INT_FACEBOOK,
    Component: Facebook
  },
  {
    name: 'LinkedIn',
    desc: 'Sync LinkedIn ads reports with Factors for performance reporting',
    icon: 'Linkedin_ads',
    kbLink:
      'https://help.factors.ai/en/articles/7283729-linkedin-ads-integration',
    featureName: FEATURES.INT_LINKEDIN,
    Component: LinkedIn
  },
  {
    name: 'G2',
    desc: 'Sync G2 intent data with Factors for a complete look at buyer intent',
    icon: 'g2crowd',
    featureName: FEATURES.INT_G2,
    Component: G2
  },
  {
    name: 'Drift',
    desc: 'Track events and conversions from Drift’s chat solution on the website',
    icon: 'DriftLogo',
    featureName: FEATURES.INT_DRIFT,
    Component: Drift
  },
  {
    name: 'Google Search Console',
    desc: 'Track organic search impressions, clicks and position from Google Search',
    icon: 'Google',
    kbLink:
      'https://help.factors.ai/en/articles/7283784-google-search-console-integration',
    featureName: FEATURES.INT_GOOGLE_ORGANIC,
    Component: GoogleSearchConsole
  },
  {
    name: 'Bing Ads',
    desc: 'Sync Bing ads reports with Factors for performance reporting',
    icon: 'Bing',
    kbLink: 'https://help.factors.ai/en/articles/7831204-bing-ads-integration',
    featureName: FEATURES.INT_BING_ADS,
    Component: Bing
  },
  {
    name: 'Clearbit Reveal',
    desc: 'Take action as soon as a target account hits your site',
    icon: 'ClearbitLogo',
    kbLink:
      'https://help.factors.ai/en/articles/7261981-clearbit-reveal-integration',
    featureName: FEATURES.INT_CLEARBIT,
    Component: Reveal
  },
  {
    name: 'LeadSquared',
    desc: 'Leadsquared is a leader in marketing automation. Using our Leadsquared source, we will ingest your Program, Campaign, Person and List records into Factors.',
    icon: 'LeadSquared',
    kbLink:
      'https://help.factors.ai/en/articles/7283684-leadsquared-integration',
    featureName: FEATURES.INT_LEADSQUARED,
    Component: LeadSquared
  },
  {
    name: '6Signal by 6Sense',
    desc: 'Gain insight into who is visiting your website and where they are in the buying journey',
    icon: 'SixSignalLogo',
    kbLink:
      'https://help.factors.ai/en/articles/7261968-6signal-by-6sense-integration',
    featureName: FEATURES.INT_SIX_SIGNAL,
    Component: SixSignal
  },
  {
    name: 'Factors Website De-anonymization',
    desc: 'Gain insight into who is visiting your website and where they are in the buying journey',
    icon: 'Brand',
    featureName: FEATURES.INT_FACTORS_DEANONYMISATION,
    Component: SixSignalFactors
  }
];
