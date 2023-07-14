import { FEATURES } from 'Constants/plans.constants';
import { FeatureConfig } from './types';

export const isFeatureLocked = (
  featureName: typeof FEATURES[keyof typeof FEATURES],
  planFeatures?: FeatureConfig[],
  addons?: FeatureConfig[]
) => {
  const activeFeatures = getAllActiveFeatures(planFeatures, addons);
  const unlockedFeatures = activeFeatures?.map((feature) => feature.name) || [];
  const isFeatureLocked = unlockedFeatures?.includes(featureName)
    ? false
    : true;
  return isFeatureLocked;
};

export const getAllActiveFeatures = (
  planFeatures?: FeatureConfig[],
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
