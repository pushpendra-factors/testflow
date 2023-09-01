import { useSelector } from 'react-redux';
import { FEATURES } from '../constants/plans.constants';
import { FeatureConfigState } from 'Reducers/featureConfig/types';
import { getFeatureStatusInfo } from 'Reducers/featureConfig/utils';

const useFeatureLock = (
  _featureName: typeof FEATURES[keyof typeof FEATURES]
): {
  isFeatureLocked: boolean;
  isLoading: boolean;
} => {
  const featureConfig = useSelector(
    //@ts-ignore
    (state) => state.featureConfig
  ) as FeatureConfigState;
  if (featureConfig.loading) {
    return {
      isLoading: true,
      isFeatureLocked: false
    };
  }
  const featureStatus = getFeatureStatusInfo(
    _featureName,
    featureConfig.activeFeatures,
    featureConfig.addOns
  );

  return {
    isLoading: false,
    isFeatureLocked: featureStatus.isFeatureLocked
  };
};

export default useFeatureLock;
