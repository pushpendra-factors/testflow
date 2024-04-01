import PlanDescriptionCard from 'Components/GenericComponents/PlanDescriptionCard';
// import LastPlanCard from 'Components/GenericComponents/PlanDescriptionCard/LastPlanCard';
import { SVG, Text } from 'Components/factorsComponents';
import {
  ADDITIONAL_ACCOUNTS_ADDON_ID,
  PLANS,
  PLANS_COFIG
} from 'Constants/plans.constants';
import {
  PlansConfigState,
  PlansDetailStateInterface
} from 'Reducers/plansConfig/types';
import logger from 'Utils/logger';
import { Button, Divider, Spin } from 'antd';
import React, { useRef, useState } from 'react';
// import PriceUpgradeModal from './PriceUpgradeModal';
import { useSelector } from 'react-redux';
import PriceUpgradeModal from './PriceUpgradeModal';
import { PRICING_HELP_LINK } from './utils';

function UpgradeTab({ buyAddonLoading, handleBuyAddonClick }: UpgradeTabProps) {
  const [isModalVisible, setIsModalVisible] = useState(false);
  const [isPlansViewCollapsed, setIsPlansViewCollapsed] = useState(true);
  const [modalVariant, setModalVariant] = useState<'plan' | 'only-addon'>(
    'plan'
  );
  const [activePlan, setActivePlan] =
    useState<PlansDetailStateInterface | null>(null);
  const collapseButtonRef = useRef<HTMLButtonElement>(null);
  const { active_project } = useSelector((state: any) => state.global);
  const { uuid: userId } = useSelector(
    (state: any) => state.agent.agent_details
  );
  const isUserBillingAdmin =
    active_project?.billing_admin_agent_uuid === userId;

  const { plansConfig, currentPlanDetail } = useSelector(
    (state: any) => state.plansConfig
  ) as PlansConfigState;
  const { plansDetail } = plansConfig;
  const purchasedAddons = currentPlanDetail?.addons;
  const additionalAccountsAddon = purchasedAddons?.find(
    (addon) => addon.id === ADDITIONAL_ACCOUNTS_ADDON_ID
  );
  const isAdditionalAccountsAddonPurchased = additionalAccountsAddon
    ? additionalAccountsAddon?.quantity > 0
    : false;
  const handleBuyButtonClick = async (
    planName: string,
    isPlanActive: boolean
  ) => {
    try {
      if (!isPlanActive) {
        const activePlan = plansDetail.find((plan) => plan.name === planName);
        if (activePlan) setActivePlan(activePlan);
        setIsModalVisible(true);
        setModalVariant('plan');
      } else {
        // const activePlan = plansDetail.find((plan) => plan.name === planName);
        // if (activePlan) setActivePlan(activePlan);
        // setIsModalVisible(true);
        // setModalVariant('only-addon');
        handleBuyAddonClick();
      }
    } catch (error) {
      logger.error('Error in upgrading plan', error);
    }
  };

  const GrowthPlan = plansDetail?.find(
    (plan) => plan.name === PLANS.PLAN_GROWTH
  );

  const renderPlans = () => (
    <div>
      <div className='flex flex-col gap-5'>
        {plansDetail &&
          plansDetail?.length > 0 &&
          plansDetail
            .filter((plan) => {
              const localPlansConfig = PLANS_COFIG?.[plan.name];
              if (!localPlansConfig) return false;
              return true;
            })
            .sort((a, b) => {
              if (
                b.name === currentPlanDetail?.plan?.externalName ||
                a.name === currentPlanDetail?.plan?.externalName
              ) {
                if (b.name === currentPlanDetail?.plan?.externalName) {
                  return 1;
                }
                if (a.name === currentPlanDetail?.plan?.externalName) {
                  return -1;
                }
              }
              const aPrice =
                a.terms.find((p) => p.period === 'month')?.price || 0;
              const bPrice =
                b.terms.find((p) => p.period === 'month')?.price || 0;
              return aPrice - bPrice;
            })
            .map((plan) => {
              const localPlansConfig = PLANS_COFIG?.[plan.name];
              if (!localPlansConfig) return <></>;

              return (
                <PlanDescriptionCard
                  isPlanActive={
                    currentPlanDetail?.plan?.externalName === plan.name
                  }
                  isRecommendedPlan={localPlansConfig.isRecommendedPlan}
                  plan={plan}
                  planIcon={localPlansConfig.planIcon}
                  planName={localPlansConfig.name}
                  planIconColor={localPlansConfig.planIconColor}
                  planDescription={localPlansConfig.description}
                  planFeatures={localPlansConfig.uniqueFeatures}
                  accountIdentifiedLimit={
                    localPlansConfig.accountIdentifiedLimit
                  }
                  mtuLimit={localPlansConfig.mtuLimit}
                  handleBuyButtonClick={handleBuyButtonClick}
                  isUserBillingAdmin={isUserBillingAdmin}
                  isButtonLoading={buyAddonLoading}
                  isAdditionalAccountsAddonPurchased={
                    isAdditionalAccountsAddonPurchased
                  }
                  seats={localPlansConfig?.seats}
                  icons={localPlansConfig?.icons}
                />
              );
            })}
        {/* <LastPlanCard /> */}
      </div>
    </div>
  );

  if (plansConfig?.loading || currentPlanDetail?.loading) {
    return (
      <div className='w-full h-full flex items-center justify-center'>
        <div className='w-full h-64 flex items-center justify-center'>
          <Spin size='large' />
        </div>
      </div>
    );
  }
  return (
    <div className='py-4'>
      <div className='mb-6'>
        <Text
          type='title'
          level={4}
          weight='bold'
          extraClass='m-0 mb-2'
          color='character-primary'
        >
          Upgrade to get more out of Factors
        </Text>
        <Text
          type='title'
          level={7}
          extraClass='m-0'
          color='character-secondary'
        >
          Check out all our plans and their included features to find the one
          that fits your needs. We are always available for a call if you ever
          need help finding the right one for your organisation.{' '}
          <a href={PRICING_HELP_LINK} target='_blank' rel='noreferrer'>
            Book a call
          </a>{' '}
        </Text>
        <Divider />
      </div>
      {renderPlans()}

      {isModalVisible && (
        <PriceUpgradeModal
          visible={isModalVisible}
          onCancel={() => setIsModalVisible(false)}
          plan={activePlan}
          variant={modalVariant}
        />
      )}
    </div>
  );
}

interface UpgradeTabProps {
  handleBuyAddonClick: () => void;
  buyAddonLoading: boolean;
}

export default UpgradeTab;
