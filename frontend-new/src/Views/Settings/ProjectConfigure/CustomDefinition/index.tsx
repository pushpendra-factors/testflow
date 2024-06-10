import CommonSettingsHeader from 'Components/GenericComponents/CommonSettingsHeader';
import { Tabs } from 'antd';

import useTabs from 'hooks/useTabs';
import React from 'react';
import withFeatureLockHOC from 'HOC/withFeatureLock';
import { FEATURES } from 'Constants/plans.constants';
import CommonLockedComponent from 'Components/GenericComponents/CommonLockedComponent';
import Engagement from '../Engagement';
import CustomKPI from '../CustomKPI';
import Events from '../Events';
import PropertyMapping from './PropertyMapping';

const FeatureLockedEngagement = withFeatureLockHOC(Engagement, {
  featureName: FEATURES.FEATURE_ACCOUNT_SCORING,
  LockedComponent: (props) => (
    <CommonLockedComponent
      featureName={FEATURES.FEATURE_ACCOUNT_SCORING}
      variant='tab'
      {...props}
    />
  )
});

const FeatureLockedCustomKPI = withFeatureLockHOC(CustomKPI, {
  featureName: FEATURES.FEATURE_CUSTOM_METRICS,
  LockedComponent: (props) => (
    <CommonLockedComponent
      featureName={FEATURES.FEATURE_CUSTOM_METRICS}
      variant='tab'
      {...props}
    />
  )
});

const FeatureLockedEvents = withFeatureLockHOC(Events, {
  featureName: FEATURES.CONF_CUSTOM_EVENTS,
  LockedComponent: (props) => (
    <CommonLockedComponent
      featureName={FEATURES.CONF_CUSTOM_EVENTS}
      variant='tab'
      {...props}
    />
  )
});

const FeatureLockedPropertyMapping = withFeatureLockHOC(PropertyMapping, {
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
  engagementScoring: 'engagementScoring',
  customKPI: 'customKPI',
  customEvent: 'customEvent',
  propertyMapping: 'propertyMapping'
};

const CustomDefinition = () => {
  const { activeKey, handleTabChange } = useTabs(TabTypes.engagementScoring);
  return (
    <div>
      <CommonSettingsHeader
        hasNoBottomPadding
        title='Custom Definitions'
        description='Define custom metrics, events and groups to analyze all your marketing touchpoints with ease.'
      />
      <div>
        <Tabs activeKey={activeKey} onChange={handleTabChange}>
          <Tabs.TabPane
            tab='Engagement Scoring Rules'
            key={TabTypes.engagementScoring}
          >
            <FeatureLockedEngagement />
          </Tabs.TabPane>
          <Tabs.TabPane tab='Custom KPIs' key={TabTypes.customKPI}>
            <FeatureLockedCustomKPI />
          </Tabs.TabPane>
          <Tabs.TabPane tab='Custom Events' key={TabTypes.customEvent}>
            <FeatureLockedEvents />
          </Tabs.TabPane>
          <Tabs.TabPane tab='Property Mapping' key={TabTypes.propertyMapping}>
            <FeatureLockedPropertyMapping />
          </Tabs.TabPane>
        </Tabs>
      </div>
    </div>
  );
};

export default CustomDefinition;
