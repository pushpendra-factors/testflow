import factorsai from 'factorsai';
import useAgentInfo from './useAgentInfo';
import { useSelector } from 'react-redux';
import { useHistory } from 'react-router-dom';
import { PathUrls } from 'Routes/pathUrls';
import {
  PRICING_PAGE_TABS,
  showV2PricingVersion
} from 'Views/Settings/ProjectSettings/Pricing/utils';
import { FeatureConfigState } from 'Reducers/featureConfig/types';

const usePlanUpgrade = (): {
  handlePlanUpgradeClick: (featureName: string, type?: ButtonType) => void;
} => {
  const { email, firstName, lastName } = useAgentInfo();
  const { active_project } = useSelector((state) => state.global);
  const { sixSignalInfo } = useSelector(
    (state: any) => state.featureConfig
  ) as FeatureConfigState;
  const showV2PricingVersionFlag = showV2PricingVersion(active_project);
  const history = useHistory();
  const handlePlanUpgradeClick = (
    featureName: string,
    type?: ButtonType = 'upgradeClick'
  ) => {
    // triggering the Upgrade click event
    factorsai.track('UPGRADE_CLICK', {
      feature_name: featureName,
      first_name: firstName || '',
      email,
      last_name: lastName || '',
      project_name: active_project?.name,
      project_id: active_project?.id,
      account_identification_limit: sixSignalInfo?.limit,
      account_identification_usage: sixSignalInfo?.usage
    });

    if (showV2PricingVersionFlag) {
      history.push(
        `${PathUrls.SettingsPricing}?activeTab=${PRICING_PAGE_TABS.UPGRADE}`
      );
    } else {
      history.push(PathUrls.SettingsPricing);
    }
  };

  return { handlePlanUpgradeClick };
};

type ButtonType = 'addonClick' | 'upgradeClick';

export default usePlanUpgrade;
