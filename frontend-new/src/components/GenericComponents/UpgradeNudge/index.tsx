import { PLANS, PLANS_V0 } from 'Constants/plans.constants';
import React from 'react';
import { useSelector } from 'react-redux';
import {
  showUpgradeNudge,
  showUpgradeNudge
} from 'Views/Settings/ProjectSettings/Pricing/utils';
import AccountLimitNudge from './AccountLimitNudge';
import CarouselNudge from './CarouselNudge';

const UpgradeNudge = ({ showCarousel = false }: UpgradeNudgeProps) => {
  const { sixSignalInfo } = useSelector((state) => state.featureConfig);
  const { currentProjectSettings } = useSelector((state: any) => state.global);
  const { plan } = useSelector((state) => state.featureConfig);
  const amountUsed = sixSignalInfo?.usage || 0;
  const totalLimit = sixSignalInfo?.limit || 0;
  const percentage = Number(((amountUsed / totalLimit) * 100).toFixed(2));
  const isFreePlan =
    plan?.name === PLANS.PLAN_FREE || plan?.name === PLANS_V0.PLAN_FREE;
  const showNudge = showUpgradeNudge(
    amountUsed,
    totalLimit,
    currentProjectSettings
  );
  if (!isFreePlan && showNudge) {
    return (
      <AccountLimitNudge
        percentage={percentage}
        limit={totalLimit}
        usage={amountUsed}
      />
    );
  }

  if (isFreePlan && showCarousel) {
    return (
      <CarouselNudge
        percentage={percentage}
        limit={totalLimit}
        usage={amountUsed}
      />
    );
  }
  return null;
};

interface UpgradeNudgeProps {
  showCarousel?: boolean;
}

export default UpgradeNudge;
