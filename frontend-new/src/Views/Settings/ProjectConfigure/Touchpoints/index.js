import React from 'react';

import { Tabs } from 'antd';

import CommonSettingsHeader from 'Components/GenericComponents/CommonSettingsHeader';
import useTabs from 'hooks/useTabs';
import { FEATURES } from 'Constants/plans.constants';
import CommonLockedComponent from 'Components/GenericComponents/CommonLockedComponent';
import withFeatureLockHOC from 'HOC/withFeatureLock';
import MarketingInteractions from '../MarketingInteractions';

import EmailClicks from './EmailClicks/index';
import CampaignGroup from './CampaignGroup';
import ChannelGroup from './ChannelGroup';
import ContentGroups from '../ContentGroups';
import HubspotSalesforceTouchpoint from './HubspotSalesforceTouchpoint';

const { TabPane } = Tabs;

const FeatureLockedMarketingInteraction = withFeatureLockHOC(
  MarketingInteractions,
  {
    featureName: FEATURES.FEATURE_OFFLINE_TOUCHPOINTS,
    LockedComponent: (props) => (
      <CommonLockedComponent
        featureName={FEATURES.FEATURE_OFFLINE_TOUCHPOINTS}
        variant='tab'
        {...props}
      />
    )
  }
);

const FeatureLockedEmailClicks = withFeatureLockHOC(EmailClicks, {
  featureName: FEATURES.FEATURE_OFFLINE_TOUCHPOINTS,
  LockedComponent: (props) => (
    <CommonLockedComponent
      featureName={FEATURES.FEATURE_OFFLINE_TOUCHPOINTS}
      variant='tab'
      {...props}
    />
  )
});

const FeatureLockedHubspotSalesforceTouchpoint = withFeatureLockHOC(
  HubspotSalesforceTouchpoint,
  {
    featureName: FEATURES.FEATURE_OFFLINE_TOUCHPOINTS,
    LockedComponent: (props) => (
      <CommonLockedComponent
        featureName={FEATURES.FEATURE_OFFLINE_TOUCHPOINTS}
        variant='tab'
        {...props}
      />
    )
  }
);

const FeatureLockedCampaignGroup = withFeatureLockHOC(CampaignGroup, {
  featureName: FEATURES.CONF_CUSTOM_PROPERTIES,
  LockedComponent: (props) => (
    <CommonLockedComponent
      featureName={FEATURES.CONF_CUSTOM_PROPERTIES}
      variant='tab'
      {...props}
    />
  )
});

const FeatureLockedChannelGroup = withFeatureLockHOC(ChannelGroup, {
  featureName: FEATURES.CONF_CUSTOM_PROPERTIES,
  LockedComponent: (props) => (
    <CommonLockedComponent
      featureName={FEATURES.CONF_CUSTOM_PROPERTIES}
      variant='tab'
      {...props}
    />
  )
});

const FeatureLockedContentGroup = withFeatureLockHOC(ContentGroups, {
  featureName: FEATURES.FEATURE_CONTENT_GROUPS,
  LockedComponent: (props) => (
    <CommonLockedComponent
      featureName={FEATURES.FEATURE_CONTENT_GROUPS}
      variant='tab'
      {...props}
    />
  )
});

const TabTypes = {
  utmParameters: 'utmParameters',
  campaignGroup: 'campaignGroup',
  channelGroup: 'channelGroup',
  contentGroup: 'contentGroup',
  emailTracking: 'emailTracking',
  hubspotTouchpoint: 'hubspotTouchpoint',
  salesforceTouchpoint: 'salesforceTouchpoint'
};

const Touchpoints = () => {
  const { activeKey, handleTabChange } = useTabs(TabTypes.utmParameters);

  return (
    <div>
      <CommonSettingsHeader
        hasNoBottomPadding
        title='Touchpoint Definitions'
        description='Effortlessly map and standardize all your digital touchpoints and use them for powerful analytics'
      />
      <div>
        <Tabs activeKey={activeKey} onChange={handleTabChange}>
          <TabPane tab='UTM Parameters' key={TabTypes.utmParameters}>
            <FeatureLockedMarketingInteraction />
          </TabPane>
          <TabPane tab='Campaign Groups' key={TabTypes.campaignGroup}>
            <FeatureLockedCampaignGroup />
          </TabPane>
          <TabPane tab='Channel Groups' key={TabTypes.channelGroup}>
            <FeatureLockedChannelGroup />
          </TabPane>
          <TabPane tab='Content Groups' key={TabTypes.contentGroup}>
            <FeatureLockedContentGroup />
          </TabPane>
          <TabPane tab='Email tracking' key={TabTypes.emailTracking}>
            <FeatureLockedEmailClicks />
          </TabPane>

          <TabPane tab='Hubspot' key={TabTypes.hubspotTouchpoint}>
            <FeatureLockedHubspotSalesforceTouchpoint
              type='hubspot'
              descriptionText='Capture offline interactions from HubSpot to attribute them precisely to leads, opportunities, and pipeline stages. Utilize custom rules to define your touchpoints accurately.'
            />
          </TabPane>

          <TabPane tab='Salesforce' key={TabTypes.salesforceTouchpoint}>
            <FeatureLockedHubspotSalesforceTouchpoint
              type='salesforce'
              descriptionText='Capture offline interactions from Salesforce to attribute them precisely to leads, opportunities, and pipeline stages. Utilize custom rules to define your touchpoints accurately.'
            />
          </TabPane>
        </Tabs>
      </div>
    </div>
  );
};

export default Touchpoints;
