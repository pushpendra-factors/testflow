import AppModal from 'Components/AppModal';
import { Number, SVG, Text } from 'Components/factorsComponents';
import React, { useEffect, useState } from 'react';
import style from './index.module.scss';
import {
  Button,
  Divider,
  InputNumber,
  Radio,
  RadioChangeEvent,
  notification
} from 'antd';
import { MinusCircleOutlined } from '@ant-design/icons';
import {
  ADDITIONAL_ACCOUNTS_ADDON_LIMIT,
  PLANS_COFIG,
  ADDITIONAL_ACCOUNTS_ADDON_ID,
  PLANS
} from 'Constants/plans.constants';
import {
  PlanTerm,
  PlansConfigState,
  PlansDetailStateInterface
} from 'Reducers/plansConfig/types';
import logger from 'Utils/logger';
import { upgradePlan } from 'Reducers/plansConfig/services';
import { useSelector } from 'react-redux';
import { startCase, toLower } from 'lodash';

function PriceUpgradeModal({
  visible,
  onCancel,
  plan,
  variant
}: UpgradeModalProps) {
  const [loading, setIsLoading] = useState<boolean>(false);
  const [addonVisible, setAddonVisible] = useState<boolean>(false);
  const [addonCount, setAddonCount] = useState<number>(
    variant === 'only-addon' ? 1 : 0
  );
  const planConfig = PLANS_COFIG[plan?.name || ''] || {};
  const [selectedPlanTerm, setSelectedPlanTerm] = useState<PlanTerm | null>(
    null
  );

  const { active_project } = useSelector((state: any) => state.global);
  const { plansConfig, differentialPricing } = useSelector(
    (state: any) => state.plansConfig
  ) as PlansConfigState;
  const additionalAccountsAddon =
    plansConfig?.addOnsDetail?.find(
      (addon) => addon.id === ADDITIONAL_ACCOUNTS_ADDON_ID
    ) || false;

  const planPrice = selectedPlanTerm?.price || 0;

  const {
    planIcon,
    planIconColor,
    description: planDescription,
    accountIdentifiedLimit,
    mtuLimit
  } = planConfig;

  const handleBuyAddonClick = () => {
    if (addonVisible) return;
    setAddonVisible(true);
  };

  const handleAddonDecrement = () => {
    if (addonCount === 0) return;
    setAddonCount(addonCount - 1);
  };

  const handleAddonClose = () => {
    setAddonVisible(false);
    setAddonCount(0);
  };

  const getTermOptions = () => {
    return (
      plan?.terms?.map((plan) => {
        return {
          value: plan.id,
          label: startCase(toLower(plan.period + 'ly'))
        };
      }) || []
    );
  };

  const handleTermOptionChange = ({ target: { value } }: RadioChangeEvent) => {
    const selectedPlanTerm = plan?.terms.find((plan) => plan.id === value);
    if (selectedPlanTerm) setSelectedPlanTerm(selectedPlanTerm);
  };

  const handleContinueClick = async () => {
    try {
      setIsLoading(true);
      let paymentUrl = '';
      if (variant === 'only-addon') {
        const res = await upgradePlan(active_project?.id, '', [
          { addon_id: ADDITIONAL_ACCOUNTS_ADDON_ID, quantity: addonCount }
        ]);
        paymentUrl = res?.data?.url;
      } else if (addonVisible && addonCount) {
        const res = await upgradePlan(
          active_project?.id,
          selectedPlanTerm?.id,
          [{ addon_id: ADDITIONAL_ACCOUNTS_ADDON_ID, quantity: addonCount }]
        );
        paymentUrl = res?.data?.url;
      } else {
        const res = await upgradePlan(active_project?.id, selectedPlanTerm?.id);
        paymentUrl = res?.data?.url;
      }
      if (!paymentUrl) {
        notification.error({
          message: 'Failed!',
          description: 'Payment URL not found!',
          duration: 3
        });
      } else {
        window.open(paymentUrl, '_self');
      }
      setIsLoading(false);
    } catch (error) {
      logger.error('Error in upgrading plan', error);
      notification.error({
        message: 'Failed!',
        description: 'Something went wrong!',
        duration: 3
      });
      setIsLoading(false);
    }
  };

  useEffect(() => {
    if (plan && plan.terms?.length > 0) {
      const yearlyTerm = plan.terms.find((p) => p.period === 'year');
      if (yearlyTerm) setSelectedPlanTerm(yearlyTerm);
      else setSelectedPlanTerm(plan.terms[0]);
    }
  }, [plan]);

  const differentialPriceForPlan = differentialPricing?.data?.find(
    (data) =>
      data.parent_item_id === plan?.name &&
      data.item_price_id === ADDITIONAL_ACCOUNTS_ADDON_ID
  );

  const addonAmount = differentialPriceForPlan
    ? differentialPriceForPlan.price
    : additionalAccountsAddon
      ? additionalAccountsAddon?.price
      : 0;

  const renderPlanVaraint = () => (
    <>
      <div className='flex gap-2 items-center '>
        <SVG name={planIcon} size='28' color={planIconColor} />
        {plan?.name && (
          <Text
            type={'title'}
            level={3}
            weight={'bold'}
            color='character-primary'
            extraClass={'m-0 '}
          >
            {plan.name}
          </Text>
        )}
      </div>
      <Text
        type={'title'}
        level={6}
        color='character-primary'
        extraClass={'m-0 mt-2'}
      >
        {planDescription}
      </Text>
      <Divider />

      <div className='mt-2 flex gap-2 items-center justify-between'>
        <Text
          type={'title'}
          level={6}
          color='character-primary'
          extraClass={'m-0 '}
        >
          Plan Details
        </Text>
        <div>
          <Radio.Group
            options={getTermOptions()}
            onChange={handleTermOptionChange}
            value={selectedPlanTerm?.id}
            optionType='button'
          />
        </div>
      </div>

      <div className={style.planDetailContainer}>
        <div className='flex items-center justify-between'>
          <div className='flex flex-col gap-2'>
            <div className='flex gap-2 items-center '>
              <SVG name={'Buildings'} size='20' color={'#BFBFBF'} />
              <Text
                type={'title'}
                level={7}
                color='character-primary'
                extraClass={'m-0 '}
              >
                <span style={{ fontWeight: 600 }}>
                  <Number number={accountIdentifiedLimit} />{' '}
                </span>
                Accounts Identified/month
              </Text>
            </div>
            <div className='flex gap-2 items-center '>
              <SVG name={'UserEvent'} size='20' color={'#BFBFBF'} />
              <Text
                type={'title'}
                level={7}
                color='character-primary'
                extraClass={'m-0 '}
              >
                <span style={{ fontWeight: 600 }}>
                  <Number number={mtuLimit} />{' '}
                </span>
                Monthly tracked users
              </Text>
            </div>
          </div>
          <div>
            <Text
              type={'title'}
              level={6}
              color='character-primary'
              extraClass={'m-0 '}
              weight={'bold'}
            >
              {selectedPlanTerm?.period === 'month' && (
                <>
                  $<Number number={planPrice} />
                  /Month
                </>
              )}
              {selectedPlanTerm?.period === 'year' && (
                <>
                  $<Number number={planPrice / 12} />
                  /Month
                </>
              )}
            </Text>
          </div>
        </div>
      </div>
      {additionalAccountsAddon && plan?.name !== PLANS.PLAN_FREE && (
        <>
          <Divider />

          <div className='mt-6'>
            <Text
              type={'title'}
              level={6}
              color='character-primary'
              extraClass={'m-0 '}
              weight={'bold'}
            >
              Need to identify more accounts?{' '}
              <a
                onClick={handleBuyAddonClick}
                style={{
                  color: addonVisible ? 'inherit' : '#1890FF',
                  cursor: addonVisible ? 'auto' : 'pointer'
                }}
              >
                Buy an Add on
              </a>
            </Text>
          </div>
        </>
      )}

      {addonVisible && renderAddonContainer()}

      <div className='mt-10'>
        <div className='flex flex-row-reverse '>
          <div className='flex flex-col gap-6  items-end'>
            <div className='flex gap-2 items-end'>
              <Text
                type={'title'}
                level={5}
                color='character-secondary'
                extraClass={'m-0 '}
                weight={'bold'}
              >
                Total
              </Text>
              <Text
                type={'title'}
                level={2}
                weight={'bold'}
                color='character-primary'
                extraClass={style.amountText}
              >
                {' '}
                $<Number number={planPrice + addonCount * addonAmount} />
              </Text>
            </div>
          </div>
        </div>
      </div>
      <Divider />
      <div className='flex gap-3 flex-row-reverse '>
        <Button
          type='primary'
          style={{ width: 120 }}
          onClick={handleContinueClick}
          loading={loading}
        >
          Continue
        </Button>
        <Button
          className={style.outlineButton}
          type='text'
          onClick={onCancel}
          style={{ width: 120 }}
        >
          Cancel
        </Button>
      </div>
    </>
  );

  const renderAddonContainer = () => (
    <div className={style.planDetailContainer}>
      <div className='flex items-center justify-between'>
        <div className='flex flex-col gap-4'>
          <Text
            type={'title'}
            level={7}
            color='character-primary'
            extraClass={'m-0 '}
          >
            <span style={{ fontWeight: 600 }}>
              Buy add-on Extra {ADDITIONAL_ACCOUNTS_ADDON_LIMIT}{' '}
            </span>
            (${addonAmount}/{ADDITIONAL_ACCOUNTS_ADDON_LIMIT} accounts)
          </Text>
          <Text
            type={'paragraph'}
            mini
            extraClass={'m-0'}
            color='character-secondary'
          >
            You will get extra 500 account identification every month based on
            your plan. ${addonAmount}/{ADDITIONAL_ACCOUNTS_ADDON_LIMIT} Accounts
          </Text>
        </div>
        <div>
          <Text
            type={'title'}
            level={6}
            color='character-primary'
            extraClass={'m-0 '}
            weight={'bold'}
          >
            +${addonAmount}
          </Text>
        </div>
      </div>
      <Divider />
      <div className='flex items-center justify-between'>
        <div className='flex items-center'>
          <InputNumber
            className={style.InputNumber}
            addonBefore={
              <Button
                type='text'
                className={style.inputNumberButtonLeft}
                onClick={handleAddonDecrement}
              >
                <SVG name='Minus' />
              </Button>
            }
            addonAfter={
              <Button
                type='text'
                className={style.inputNumberButtonRight}
                onClick={() => setAddonCount((count) => count + 1)}
              >
                <SVG name='Plus' />
              </Button>
            }
            defaultValue={0}
            min={0}
            readOnly
            controls={false}
            formatter={(value) => `${value} Qty`}
            value={addonCount}
          />
        </div>
        <div>
          <Text
            type={'title'}
            level={6}
            color='character-primary'
            extraClass={'m-0 '}
            weight={'bold'}
          >
            +${addonAmount * addonCount}
          </Text>
        </div>
      </div>
      {variant === 'plan' && (
        <div className={style.addonCancelButton}>
          <Button
            type='text'
            onClick={handleAddonClose}
            icon={<MinusCircleOutlined color='#8C8C8C' size={20} />}
          />
        </div>
      )}
    </div>
  );

  const renderAddonVariant = () => (
    <>
      <div className='flex gap-2 items-center '>
        <SVG name='Bookmark' size='28' color='#1890FF' />
        {plan?.name && (
          <Text
            type={'title'}
            level={3}
            weight={'bold'}
            color='character-primary'
            extraClass={'m-0 '}
          >
            Add-on
          </Text>
        )}
      </div>
      <Text
        type={'title'}
        level={6}
        color='character-primary'
        extraClass={'m-0 mt-2'}
      >
        Get additional accounts without switching your plan.
      </Text>
      <Divider />
      {renderAddonContainer()}
      <div className='mt-10'>
        <div className='flex flex-row-reverse '>
          <div className='flex flex-col gap-6  items-end'>
            <div className='flex gap-2 items-end'>
              <Text
                type={'title'}
                level={5}
                color='character-secondary'
                extraClass={'m-0 '}
                weight={'bold'}
              >
                Total
              </Text>
              <Text
                type={'title'}
                level={2}
                weight={'bold'}
                color='character-primary'
                extraClass={style.amountText}
              >
                {' '}
                ${addonCount * addonAmount}
              </Text>
            </div>
          </div>
        </div>
      </div>
      <Divider />
      <div className='flex gap-3 flex-row-reverse '>
        <Button
          type='primary'
          style={{ width: 120 }}
          onClick={handleContinueClick}
          disabled={addonCount < 1}
          loading={loading}
        >
          Continue
        </Button>
        <Button
          className={style.outlineButton}
          type='text'
          onClick={onCancel}
          style={{ width: 120 }}
        >
          Cancel
        </Button>
      </div>
    </>
  );
  return (
    <div>
      <AppModal
        visible={visible}
        footer={<></>}
        onCancel={onCancel}
        isLoading={loading}
        className={style.priceUpgradeModal}
      >
        {variant === 'plan' ? renderPlanVaraint() : renderAddonVariant()}
      </AppModal>
    </div>
  );
}

interface UpgradeModalProps {
  visible: boolean;
  onCancel: () => void;
  plan: PlansDetailStateInterface | null;
  variant: 'plan' | 'only-addon';
}

export default PriceUpgradeModal;
