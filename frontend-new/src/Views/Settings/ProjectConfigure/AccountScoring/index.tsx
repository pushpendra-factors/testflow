import CommonSettingsHeader from 'Components/GenericComponents/CommonSettingsHeader';
import { Tabs } from 'antd';

import useTabs from 'hooks/useTabs';
import React from 'react';
import withFeatureLockHOC from 'HOC/withFeatureLock';
import { FEATURES } from 'Constants/plans.constants';
import CommonLockedComponent from 'Components/GenericComponents/CommonLockedComponent';
import Engagement from '../Engagement';

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

const TabTypes = {
  engagementScoring: 'engagementScoring'
};

const CustomDefinition = () => {
  const { activeKey, handleTabChange } = useTabs(TabTypes.engagementScoring);
  return (
    <div>
      <CommonSettingsHeader
        hasNoBottomPadding
        title='Account Scoring'
        description='Setup custom rules to categorize, score and prioritize your target accounts with ease.'
      />
      <div>
        <Tabs activeKey={activeKey} onChange={handleTabChange}>
          <Tabs.TabPane
            tab='Engagement Scoring'
            key={TabTypes.engagementScoring}
          >
            <FeatureLockedEngagement />
          </Tabs.TabPane>
        </Tabs>
      </div>
    </div>
  );
};

export default CustomDefinition;
