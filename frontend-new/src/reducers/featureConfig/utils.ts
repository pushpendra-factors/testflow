import { FEATURES } from 'Constants/plans.constants';
import { FeatureConfig } from './types';

export const getAllActiveFeatures = (
  planFeatures?: FeatureConfig[] | null,
  addons?: FeatureConfig[]
) => {
  let activeFeatures: FeatureConfig[] = [];
  if (Array.isArray(planFeatures)) {
    activeFeatures = [...planFeatures];
  }
  if (Array.isArray(addons)) {
    activeFeatures = [...activeFeatures, ...addons];
  }
  return activeFeatures;
};

export const getFeatureStatusInfo = (
  featureName: (typeof FEATURES)[keyof typeof FEATURES],
  planFeatures?: FeatureConfig[],
  addons?: FeatureConfig[]
): { isFeatureLocked: boolean } => {
  const activeFeatures = getAllActiveFeatures(planFeatures, addons);
  const feature = activeFeatures.find(
    (feature) => feature.name === featureName
  );
  if (!feature) {
    return {
      isFeatureLocked: true
    };
  }
  return {
    isFeatureLocked: !feature?.is_enabled_feature
  };
};
