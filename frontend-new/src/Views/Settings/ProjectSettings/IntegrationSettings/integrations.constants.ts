import { FEATURES } from 'Constants/plans.constants';
import { IntegrationCategroryType, IntegrationConfig } from './types';
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
// import SixSignalFactors from './SixSignalFactors';
import FactorsAccountIdentification from './FactorsAccountIdentification';
import G2 from './G2';
import SDKSettings from '../SDKSettings';
import { SDKDocumentation } from '../../../../features/onboarding/utils';

export const INTEGRATION_ID = {
  sdk: 'sdk',
  segment: FEATURES.FEATURE_SEGMENT,
  rudderstack: FEATURES.FEATURE_RUDDERSTACK,
  google_ads: FEATURES.FEATURE_GOOGLE_ADS,
  facebook: FEATURES.FEATURE_FACEBOOK,
  linkedIn: FEATURES.FEATURE_LINKEDIN,
  bing_ads: FEATURES.FEATURE_BING_ADS,
  hubspot: FEATURES.FEATURE_HUBSPOT,
  salesforce: FEATURES.FEATURE_SALESFORCE,
  marketo: FEATURES.FEATURE_MARKETO,
  lead_squared: FEATURES.FEATURE_LEADSQUARED,
  clearbit_reveal: FEATURES.FEATURE_CLEARBIT,
  six_signal_by_6_sense: FEATURES.FEATURE_SIX_SIGNAL,
  factors_website_de_anonymization: FEATURES.FEATURE_FACTORS_DEANONYMISATION,
  slack: FEATURES.FEATURE_SLACK,
  microsoft_teams: FEATURES.FEATURE_TEAMS,
  drift: FEATURES.FEATURE_DRIFT,
  g2: FEATURES.FEATURE_G2,
  google_search_console: FEATURES.FEATURE_GOOGLE_ORGANIC
};

export const INTEGRATION_CATEGORY_ID = {
  sdk: 'sdk',
  accountIdentification: 'account_identification',
  adsPlatforms: 'ads',
  crm: 'crm',
  review: 'review',
  cdp: 'cdp',
  communication: 'communication',
  chatbot: 'chatbot',
  organic: 'organic'
};

export const IntegrationProviderData: IntegrationConfig[] = [
  // SDK
  {
    id: INTEGRATION_ID.sdk,
    categoryId: INTEGRATION_CATEGORY_ID.sdk,
    name: 'Javascript SDK',
    desc: 'Place Factors SDK on your website to identify accounts visiting your website and track their activity',
    icon: 'Brand',
    featureName: 'sdk',
    Component: SDKSettings,
    kbLink: SDKDocumentation,
    showInstructionMenu: false
  },

  // Account Identification
  // {
  //   id: INTEGRATION_ID.clearbit_reveal,
  //   categoryId: INTEGRATION_CATEGORY_ID.accountIdentification,
  //   name: 'Clearbit Reveal',
  //   desc: 'Take action as soon as a target account hits your site',
  //   icon: 'ClearbitLogo',
  //   kbLink:
  //     'https://help.factors.ai/en/articles/7261981-clearbit-reveal-integration',
  //   featureName: FEATURES.FEATURE_CLEARBIT,
  //   Component: Reveal
  // },
  // {
  //   id: INTEGRATION_ID.six_signal_by_6_sense,
  //   categoryId: INTEGRATION_CATEGORY_ID.accountIdentification,
  //   name: '6Signal by 6Sense',
  //   desc: 'Gain insight into who is visiting your website and where they are in the buying journey',
  //   icon: 'SixSignalLogo',
  //   kbLink:
  //     'https://help.factors.ai/en/articles/7261968-6signal-by-6sense-integration',
  //   featureName: FEATURES.FEATURE_SIX_SIGNAL,
  //   Component: SixSignal
  // },
  {
    id: INTEGRATION_ID.factors_website_de_anonymization,
    categoryId: INTEGRATION_CATEGORY_ID.accountIdentification,
    name: 'Factors Account Identification',
    desc: 'Gain insight into who is visiting your website and where they are in the buying journey',
    icon: 'Brand',
    featureName: FEATURES.FEATURE_FACTORS_DEANONYMISATION,
    Component: FactorsAccountIdentification,
    showInstructionMenu: false
  },

  // ads
  {
    id: INTEGRATION_ID.google_ads,
    categoryId: INTEGRATION_CATEGORY_ID.adsPlatforms,
    name: 'Google Ads',
    desc: 'Integrate reporting from Google Search, Youtube and Display Network',
    icon: 'Google_ads',
    kbLink:
      'https://help.factors.ai/en/articles/7283695-google-ads-integration',
    featureName: FEATURES.FEATURE_GOOGLE_ADS,
    Component: GoogleAdWords,
    showInstructionMenu: true,
    instructionTitle: 'Integration Details',
    instructionDescription:
      'Click Enable using Google, you will be redirected to authorise the connection between Factors and your Google account. Once you have authorised the connection, you will be asked to select the ad account that you wish to bring data from. Data will only be pulled once an account has been selected. '
  },
  {
    id: INTEGRATION_ID.google_search_console,
    categoryId: INTEGRATION_CATEGORY_ID.adsPlatforms,
    name: 'Google Search Console',
    desc: 'Track organic search impressions, clicks and position from Google Search',
    icon: 'Google',
    kbLink:
      'https://help.factors.ai/en/articles/7283784-google-search-console-integration',
    featureName: FEATURES.FEATURE_GOOGLE_ORGANIC,
    Component: GoogleSearchConsole,

    showInstructionMenu: true,
    instructionTitle: 'Integration Details',
    instructionDescription:
      'Click Enable using Google, you will be redirected to authorise the connection between Factors and your Search Console account. Once you have authorised the connection, you will be asked to select the URL(s) that you wish to bring data from. Data will only be pulled once a URL has been selected. '
  },
  {
    id: INTEGRATION_ID.facebook,
    categoryId: INTEGRATION_CATEGORY_ID.adsPlatforms,
    name: 'Facebook',
    desc: 'Pull in reports from Facebook, Instagram and Facebook Audience Network',
    icon: 'Facebook_ads',
    kbLink:
      'https://help.factors.ai/en/articles/7283696-facebook-ads-integration',
    featureName: FEATURES.FEATURE_FACEBOOK,
    Component: Facebook,
    showInstructionMenu: true,
    instructionTitle: 'Integration Details',
    instructionDescription:
      'Click Login with Facebook, you will be redirected to authorise the connection between Factors and your Facebook account. Once you have authorised the connection, you will be asked to select the ad account that you wish to bring data from. Data will only be pulled once an account has been selected. '
  },
  {
    id: INTEGRATION_ID.linkedIn,
    categoryId: INTEGRATION_CATEGORY_ID.adsPlatforms,
    name: 'LinkedIn',
    desc: 'Sync LinkedIn ads reports with Factors for performance reporting',
    icon: 'Linkedin_ads',
    kbLink:
      'https://help.factors.ai/en/articles/7283729-linkedin-ads-integration',
    featureName: FEATURES.FEATURE_LINKEDIN,
    Component: LinkedIn,
    showInstructionMenu: true,
    instructionTitle: 'Integration Details',
    instructionDescription:
      'Click Connect Now, you will be redirected to authorise the connection between Factors and your LinkedIn account. Once you have authorised the connection, you will be asked to select the ad account that you wish to bring data from. Data will only be pulled once an account has been selected. '
  },
  {
    id: INTEGRATION_ID.bing_ads,
    categoryId: INTEGRATION_CATEGORY_ID.adsPlatforms,
    name: 'Bing Ads',
    desc: 'Sync Bing ads reports with Factors for performance reporting',
    icon: 'Bing',
    kbLink: 'https://help.factors.ai/en/articles/7831204-bing-ads-integration',
    featureName: FEATURES.FEATURE_BING_ADS,
    Component: Bing,
    showInstructionMenu: true,
    instructionTitle: 'Integration Details',
    instructionDescription:
      'Click Connect Now, you will be redirected to authorise the connection between Factors and your Microsoft Advertising account. Once you have authorised the connection, you will be asked to select the Bing ad account that you wish to bring data from. Data will only be pulled once an account has been selected.'
  },

  // crm
  {
    id: INTEGRATION_ID.hubspot,
    categoryId: 'crm',
    name: 'Hubspot',
    desc: 'Sync your Contact, Company and Deal objects with Factors on a daily basis',
    icon: 'Hubspot_ads',
    kbLink: 'https://help.factors.ai/en/articles/7261985-hubspot-integration',
    featureName: FEATURES.FEATURE_HUBSPOT,
    Component: Hubspot,
    showInstructionMenu: true,
    instructionTitle: 'Integration Details',
    instructionDescription:
      'Click Enable using Hubspot, you will be redirected to authorise the connection between Factors and HubSpot.'
  },
  {
    id: INTEGRATION_ID.salesforce,
    categoryId: INTEGRATION_CATEGORY_ID.crm,
    name: 'Salesforce',
    desc: 'Sync your Leads, Contact, Account, Opportunity and Campaign objects with Factors on a daily basis',
    icon: 'Salesforce_ads',
    kbLink:
      'https://help.factors.ai/en/articles/7261989-salesforce-integration',
    featureName: FEATURES.FEATURE_SALESFORCE,
    Component: Salesforce,
    showInstructionMenu: true,
    instructionTitle: 'Integration Details',
    instructionDescription:
      'Click Connect Salesforce, you will be redirected to authorise the connection between Factors and Salesforce.'
  },
  {
    id: INTEGRATION_ID.marketo,
    categoryId: INTEGRATION_CATEGORY_ID.crm,
    name: 'Marketo',
    desc: 'Marketo is a leader in marketing automation. Using our Marketo source, we will ingest your Program, Campaign, Person and List records into Factors',
    icon: 'Marketo',
    featureName: FEATURES.FEATURE_MARKETO,
    Component: Marketo,
    showInstructionMenu: true,
    instructionTitle: 'Integration Details',
    instructionDescription:
      'Click Connect Marketo, you will be redirected to authorise the connection between Factors and Marketo. Simple follow the instructions that are displayed in the next step to add details about the API endpoint, client ID and client secret to establish the data connection between Factors and Marketo.'
  },
  {
    id: INTEGRATION_ID.lead_squared,
    categoryId: INTEGRATION_CATEGORY_ID.crm,
    name: 'LeadSquared',
    desc: 'Leadsquared is a leader in marketing automation. Using our Leadsquared source, we will ingest your Program, Campaign, Person and List records into Factors.',
    icon: 'LeadSquared',
    kbLink:
      'https://help.factors.ai/en/articles/7283684-leadsquared-integration',
    featureName: FEATURES.FEATURE_LEADSQUARED,
    Component: LeadSquared,
    showInstructionMenu: true,
    instructionTitle: 'Integration Details',
    instructionDescription:
      'Enter your LeadSquared access key, secret key and host and then click Connect Now to connect your LeadSquared account to Factors.'
  },

  // Review platform
  {
    id: INTEGRATION_ID.g2,
    categoryId: INTEGRATION_CATEGORY_ID.review,
    name: 'G2',
    desc: 'Sync G2 intent data with Factors for a complete look at buyer intent',
    icon: 'g2crowd',
    featureName: FEATURES.FEATURE_G2,
    Component: G2,
    showInstructionMenu: true,
    instructionTitle: 'Integration Details',
    instructionDescription:
      'Enter your G2 API key to connect Factors to G2 and bring in intent data. You can find your G2 API key by going inside “Integrations” in your G2 admin portal and then creating a new API token. Once created, copy the API token from G2 and enter it here to authorise the connection. '
  },
  // communication
  {
    id: INTEGRATION_ID.slack,
    categoryId: INTEGRATION_CATEGORY_ID.communication,
    name: 'Slack',
    desc: 'Does your team live on Slack? Set up alerts that track KPIs and marketing data. Nudge your team to take the right actions.',
    icon: 'Slack',
    kbLink: 'https://help.factors.ai/en/articles/7283808-slack-integration',
    featureName: FEATURES.FEATURE_SLACK,
    Component: Slack,
    showInstructionMenu: true,
    instructionTitle: 'Integration Details',
    instructionDescription:
      'Click Connect Now, you will be redirected to authorise the connection between Factors and Slack.'
  },
  {
    id: INTEGRATION_ID.microsoft_teams,
    categoryId: INTEGRATION_CATEGORY_ID.communication,
    name: 'Microsoft Teams',
    desc: 'Does your team live on Teams? Set up alerts that track KPIs and marketing data. Nudge your team to take the right actions.',
    icon: 'MSTeam',
    kbLink:
      'https://help.factors.ai/en/articles/7913152-microsoft-teams-integration',
    featureName: FEATURES.FEATURE_TEAMS,
    Component: MSTeam,
    showInstructionMenu: true,
    instructionTitle: 'Integration Details',
    instructionDescription:
      'Click Connect Now, you will be redirected to authorise the connection between Factors and Microsoft Teams.'
  },

  // cdp
  {
    id: INTEGRATION_ID.segment,
    categoryId: INTEGRATION_CATEGORY_ID.cdp,
    name: 'Segment',
    desc: 'Segment is a Customer Data Platform (CDP) that simplifies collecting and using data from the users of your digital properties and SaaS applications',
    icon: 'Segment_ads',
    kbLink: 'https://help.factors.ai/en/articles/7261994-segment-integration',
    featureName: FEATURES.FEATURE_SEGMENT,
    Component: Segment,
    showInstructionMenu: true,
    instructionTitle: 'Integration Details',
    instructionDescription:
      "First, take your API key and configure Factors as a destination in your Segment Workspace. Once done, enable all the data sources inside Segment that you would like to send to Factors. We start bringing in data only once you've completed these steps."
  },
  {
    id: INTEGRATION_ID.rudderstack,
    categoryId: INTEGRATION_CATEGORY_ID.cdp,
    name: 'Rudderstack',
    desc: 'Rudderstack is a Customer Data Platform (CDP) that simplifies collecting and using data from the users of your digital properties and SaaS applications',
    icon: 'Rudderstack_ads',
    kbLink:
      'https://help.factors.ai/en/articles/7283693-rudderstack-integration',
    featureName: FEATURES.FEATURE_RUDDERSTACK,
    Component: Rudderstack,
    showInstructionMenu: true,
    instructionTitle: 'Integration Details',
    instructionDescription:
      "First, take your API key and configure Factors as a destination in your Rudderstack Workspace. Once done, enable all the data sources inside Rudderstack that you would like to send to Factors. We start bringing in data only once you've completed these steps."
  },

  // chatbot
  {
    id: INTEGRATION_ID.drift,
    categoryId: INTEGRATION_CATEGORY_ID.chatbot,
    name: 'Drift',
    desc: 'Track events and conversions from Drift’s chat solution on the website',
    icon: 'DriftLogo',
    featureName: FEATURES.FEATURE_DRIFT,
    Component: Drift,
    showInstructionMenu: true,
    instructionTitle: 'Integration Details',
    instructionDescription:
      'Click Enable Now and Factors will start reading data from your Drift SDK.'
  }
];

export const AccountIdentificationProviderData: IntegrationConfig[] = [
  // Account Identification
  {
    id: INTEGRATION_ID.six_signal_by_6_sense,
    categoryId: INTEGRATION_CATEGORY_ID.accountIdentification,
    name: '6Signal by 6Sense',
    desc: 'Use 6Signal by 6Sense to identify accounts. Your usage will be billed by 6Signal directly.',
    icon: 'SixSignalLogo',
    kbLink:
      'https://help.factors.ai/en/articles/7261968-6signal-by-6sense-integration',
    featureName: FEATURES.FEATURE_SIX_SIGNAL,
    Component: SixSignal,
    showInstructionMenu: false
  },
  {
    id: INTEGRATION_ID.clearbit_reveal,
    categoryId: INTEGRATION_CATEGORY_ID.accountIdentification,
    name: 'Clearbit Reveal',
    desc: 'Use Clearbit Reveal to identify accounts. Your usage will be billed by Clearbit directly.',
    icon: 'ClearbitLogo',
    kbLink:
      'https://help.factors.ai/en/articles/7261981-clearbit-reveal-integration',
    featureName: FEATURES.FEATURE_CLEARBIT,
    Component: Reveal,
    showInstructionMenu: false
  }
];

export const IntegrationPageCategories: IntegrationCategroryType[] = [
  {
    name: 'SDK',
    id: INTEGRATION_CATEGORY_ID.sdk,
    sortOrder: 1
  },
  {
    name: 'Account Identification',
    id: INTEGRATION_CATEGORY_ID.accountIdentification,
    sortOrder: 2
  },
  {
    name: 'CRMs & MAPs',
    id: INTEGRATION_CATEGORY_ID.crm,
    sortOrder: 3
  },
  {
    name: 'Ad Platforms',
    id: INTEGRATION_CATEGORY_ID.adsPlatforms,
    sortOrder: 4
  },
  {
    name: 'Review Platforms',
    id: INTEGRATION_CATEGORY_ID.review,
    sortOrder: 5
  },
  {
    name: 'Communication Apps',
    id: INTEGRATION_CATEGORY_ID.communication,
    sortOrder: 6
  },
  {
    name: 'Customer Data Platforms (CDP)',
    id: INTEGRATION_CATEGORY_ID.cdp,
    sortOrder: 7
  },
  {
    name: 'Chatbot',
    id: INTEGRATION_CATEGORY_ID.chatbot,
    sortOrder: 8
  }
];
