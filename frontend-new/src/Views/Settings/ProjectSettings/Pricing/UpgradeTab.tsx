import PlanDescriptionCard from 'Components/GenericComponents/PlanDescriptionCard';
import LastPlanCard from 'Components/GenericComponents/PlanDescriptionCard/LastPlanCard';
import { SVG, Text } from 'Components/factorsComponents';
import { PLANS, PLANS_COFIG } from 'Constants/plans.constants';
import {
  PlansConfigState,
  PlansDetailStateInterface
} from 'Reducers/plansConfig/types';
import logger from 'Utils/logger';
import { Button, Divider, Spin } from 'antd';
import React, { useState } from 'react';
// import PriceUpgradeModal from './PriceUpgradeModal';
import { useSelector } from 'react-redux';
import PriceUpgradeModal from './PriceUpgradeModal';

const UpgradeTab = () => {
  const [isModalVisible, setIsModalVisible] = useState(false);
  const [isPlansViewCollapsed, setIsPlansViewCollapsed] = useState(true);
  const [modalVariant, setModalVariant] =
    useState<'plan' | 'only-addon'>('plan');
  const [activePlan, setActivePlan] =
    useState<PlansDetailStateInterface | null>(null);
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
        const activePlan = plansDetail.find((plan) => plan.name === planName);
        if (activePlan) setActivePlan(activePlan);
        setIsModalVisible(true);
        setModalVariant('only-addon');
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
      {isPlansViewCollapsed && GrowthPlan && (
        <PlanDescriptionCard
          isPlanActive={
            currentPlanDetail?.plan?.externalName === GrowthPlan.name
          }
          isRecommendedPlan={PLANS_COFIG[GrowthPlan.name].isRecommendedPlan}
          plan={GrowthPlan}
          planIcon={PLANS_COFIG[GrowthPlan.name].planIcon}
          planName={PLANS_COFIG[GrowthPlan.name].name}
          planIconColor={PLANS_COFIG[GrowthPlan.name].planIconColor}
          planDescription={PLANS_COFIG[GrowthPlan.name].description}
          planFeatures={PLANS_COFIG[GrowthPlan.name].uniqueFeatures}
          accountIdentifiedLimit={
            PLANS_COFIG[GrowthPlan.name].accountIdentifiedLimit
          }
          mtuLimit={PLANS_COFIG[GrowthPlan.name].mtuLimit}
          handleBuyButtonClick={handleBuyButtonClick}
          isUserBillingAdmin={isUserBillingAdmin}
        />
      )}
      {!isPlansViewCollapsed && (
        <div className='flex flex-col gap-5'>
          {plansDetail &&
            plansDetail?.length > 0 &&
            plansDetail
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
                  />
                );
              })}
          <LastPlanCard />
        </div>
      )}
      <Button
        className='mt-4'
        type='text'
        onClick={() => setIsPlansViewCollapsed((value) => !value)}
      >
        {isPlansViewCollapsed && (
          <div className='flex gap-2 items-center'>
            <Text
              type={'title'}
              level={6}
              extraClass='m-0'
              color='brand-color-6'
            >
              Show All Plans
            </Text>
            <SVG name={'CaretDown'} size={20} color='#1890FF' />
          </div>
        )}
        {!isPlansViewCollapsed && (
          <div className='flex gap-2 items-center'>
            <Text
              type={'title'}
              level={6}
              extraClass='m-0'
              color='brand-color-6'
            >
              Hide All Plans
            </Text>
            <SVG name={'CaretUp'} size={20} color={'#1890FF'} />
          </div>
        )}
      </Button>
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
          type={'title'}
          level={4}
          weight={'bold'}
          extraClass={'m-0 mb-2'}
          color='character-primary'
        >
          Upgrade to get the most out of Factors
        </Text>
        <Text
          type={'title'}
          level={7}
          extraClass={'m-0'}
          color='character-secondary'
        >
          Familairize yourself with the payment plans below. See for yourself
          that the basic service for your business is not as expensive as it
          might seem
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
};

export default UpgradeTab;
