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
        title='Touchpoint Customisation'
        description='Unlock productivity with our robust ecosystem of seamless software integrations.'
      />
      <div>
        <Tabs activeKey={activeKey} onChange={handleTabChange}>
          <TabPane tab='UTM Parameters' key={TabTypes.utmParameters}>
            <FeatureLockedMarketingInteraction />
          </TabPane>
          <TabPane tab='Campaign Grouping' key={TabTypes.campaignGroup}>
            <FeatureLockedCampaignGroup />
          </TabPane>
          <TabPane tab='Channel Grouping' key={TabTypes.channelGroup}>
            <FeatureLockedChannelGroup />
          </TabPane>
          <TabPane tab='Content Grouping' key={TabTypes.contentGroup}>
            <ContentGroups />
          </TabPane>
          <TabPane tab='Email tracking' key={TabTypes.emailTracking}>
            <FeatureLockedEmailClicks />
          </TabPane>

          <TabPane tab='HS Offline Touchpoint' key={TabTypes.hubspotTouchpoint}>
            <FeatureLockedHubspotSalesforceTouchpoint type='hubspot' />
          </TabPane>

          <TabPane
            tab='Salesforce Offline Touchpoint'
            key={TabTypes.salesforceTouchpoint}
          >
            <FeatureLockedHubspotSalesforceTouchpoint type='salesforce' />
          </TabPane>
        </Tabs>
      </div>
    </div>
  );
};

export default Touchpoints;
