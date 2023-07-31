import { fetchProjectSettingsV1 } from 'Reducers/global';

const TEMPLATES_HOSTCDN =
  'https://s3.amazonaws.com/www.factors.ai/assets/img/product/templates/';
export const IntegrationKeyNames = {
  adwords: 'adwords',
  sdk: 'website_sdk',
  bingads: 'bingads',
  googlesearchconsole: 'googlesearchconsole',
  hubspot: 'hubspot',
  linkedin: 'linkedin',
  facebook: 'facebook',
  segment: 'segment'
};
export class Integration_Checks {
  // These are for Templates
  website_sdk = undefined;
  adwords = undefined;
  bingads = undefined;
  googlesearchconsole = undefined;
  hubspot = undefined;
  linkedin = undefined;
  facebook = undefined;

  // Other Integrations
  segment = undefined;

  // Integration = current prokect Settings Object
  constructor(
    sdk,
    integration,
    bingAds,
    marketo,
    isFactorsDeanonymizationConnected
  ) {
    this.website_sdk = sdk;
    // Now Checking Other Integrations
    this.adwords = !!integration.int_adwords_enabled_agent_uuid;
    this.bingads = bingAds.accounts;
    this.google_search_console =
      integration.int_google_organic_url_prefixes &&
      integration.int_google_organic_url_prefixes !== '';
    this.hubspot = integration.int_hubspot;
    this.linkedin = integration.int_linkedin_agent_uuid;
    this.facebook = integration.int_facebook_user_id;
    this.marketo = marketo.status;
    this['6signal'] =
      integration?.int_client_six_signal_key ||
      isFactorsDeanonymizationConnected;
    // Other Integrations
    this.segment = integration.int_segment;
  }

  // This Function Accepts
  // 1. Requirements = Array<{mandatory, name, keyname}>
  checkRequirements = (requirements = []) => {
    let result = undefined;
    let failed = [];
    try {
      requirements.forEach((element) => {
        if (result === undefined) {
          result = !!this[element];
        } else {
          result = result && this[element];
        }
        if (!this[element]) failed.push(element);
      });
    } catch (error) {
      console.log(error);
    }
    return { result, failedAt: failed };
  };
}
const ThumbnailAssetsWithName = [
  {
    name: 'allpaidmarketing',
    image: TEMPLATES_HOSTCDN + 'Thumbnail_AllPaidMarketing.png'
  },
  {
    name: 'googleadwords',
    image: TEMPLATES_HOSTCDN + 'Thumbnail_GoogleAdwords.png'
  },
  {
    name: 'googlesearchconsole',
    image: TEMPLATES_HOSTCDN + 'Thumbnail_GoogleSearchConsole.png'
  },
  {
    name: 'hubspotcontactsattribution',
    image: TEMPLATES_HOSTCDN + 'Thumbnail_HubspotContactsAttribution.png'
  },
  {
    name: 'hubspotinsights',
    image: TEMPLATES_HOSTCDN + 'Thumbnail_HubspotInsights.png'
  },
  {
    name: 'organicperformance',
    image: TEMPLATES_HOSTCDN + 'Thumbnail_OrganicPerformance.png'
  },
  {
    name: 'overallreporting',
    image: TEMPLATES_HOSTCDN + 'Thumbnail_OverallReporting.png'
  },
  {
    name: 'paidsearchmarketing',
    image: TEMPLATES_HOSTCDN + 'Thumbnail_PaidSearchMarketing.png'
  },
  {
    name: 'paidsocialmarketing',
    image: TEMPLATES_HOSTCDN + 'Thumbnail_PaidSocialMarketing.png'
  },
  {
    name: 'webanalytics',
    image: TEMPLATES_HOSTCDN + 'Thumbnail_WebAnalytics.png'
  },
  {
    name: 'webkpisandoverview',
    image: TEMPLATES_HOSTCDN + 'Thumbnail_WebKPIsAndOverview.png'
  },
  {
    name: 'landingpageengagement',
    image: TEMPLATES_HOSTCDN + 'Thumbnail_LandingPageEngagement.png'
  },
  {
    name: 'websitevisitoridentification',
    image: TEMPLATES_HOSTCDN + 'Thumbnail_WebsiteVisitorIdentification.png'
  }
];

const TemplatesThumbnail = new Map();

ThumbnailAssetsWithName.forEach((element) => {
  TemplatesThumbnail.set(element.name, element);
});

export const FallBackImage = TEMPLATES_HOSTCDN + 'FallBack.png';
export const StartFreshImage = TEMPLATES_HOSTCDN + 'StartFresh.png';

export default TemplatesThumbnail;
