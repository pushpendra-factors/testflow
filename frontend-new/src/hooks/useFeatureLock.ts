import { useSelector } from 'react-redux';
import { FEATURES } from '../constants/plans.constants';
import { FeatureConfigState } from 'Reducers/featureConfig/types';
import { isFeatureLocked } from 'Reducers/featureConfig/utils';

const useFeatureLock = (
  _featureName: typeof FEATURES[keyof typeof FEATURES]
): { isFeatureLocked: boolean; isLoading: boolean } => {
  const featureConfig = useSelector(
    //@ts-ignore
    (state) => state.featureConfig
  ) as FeatureConfigState;
  if (featureConfig.loading) {
    return { isLoading: true, isFeatureLocked: false };
  }

  return {
    isLoading: false,
    isFeatureLocked: isFeatureLocked(
      _featureName,
      featureConfig.activeFeatures,
      featureConfig.addOns
    )
  };
};

export default useFeatureLock;
