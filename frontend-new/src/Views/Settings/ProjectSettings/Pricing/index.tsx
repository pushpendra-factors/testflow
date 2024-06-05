import { Tabs, notification } from 'antd';
import React, { useState } from 'react';
import { useSelector } from 'react-redux';
import { upgradePlan } from 'Reducers/plansConfig/services';
import { ADDITIONAL_ACCOUNTS_ADDON_ID } from 'Constants/plans.constants';
import logger from 'Utils/logger';
import CommonSettingsHeader from 'Components/GenericComponents/CommonSettingsHeader';
import useTabs from 'hooks/useTabs';
import InvoiceTab from './InvoiceTab';
import UpgradeTab from './UpgradeTab';
import { PRICING_PAGE_TABS, showV2PricingVersion } from './utils';
import BillingTab from './BillingTab';
import styles from './index.module.scss';

function Pricing() {
  const [loading, setIsLoading] = useState<boolean>(false);
  const { activeKey, handleTabChange } = useTabs(PRICING_PAGE_TABS.BILLING);
  const { active_project } = useSelector((state) => state.global);

  const handleBuyAddonClick = async () => {
    try {
      setIsLoading(true);

      const res = await upgradePlan(active_project?.id, '', [
        { addon_id: ADDITIONAL_ACCOUNTS_ADDON_ID, quantity: 1 }
      ]);
      const paymentUrl = res?.data?.url;

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

  return (
    <div>
      <CommonSettingsHeader title='Plans and billing' />
      <div className={`mt-2 ${styles.tab_container}`}>
        <Tabs
          activeKey={activeKey}
          onChange={handleTabChange}
          style={{ overflow: 'none' }}
        >
          <Tabs.TabPane tab='Billing' key={PRICING_PAGE_TABS.BILLING}>
            <BillingTab
              handleBuyAddonClick={handleBuyAddonClick}
              buyAddonLoading={loading}
            />
          </Tabs.TabPane>

          {showV2PricingVersion(active_project) && (
            <>
              <Tabs.TabPane tab='Upgrade' key={PRICING_PAGE_TABS.UPGRADE}>
                <UpgradeTab
                  handleBuyAddonClick={handleBuyAddonClick}
                  buyAddonLoading={loading}
                />
              </Tabs.TabPane>
              <Tabs.TabPane tab='Invoices' key={PRICING_PAGE_TABS.INVOICES}>
                <InvoiceTab />
              </Tabs.TabPane>
            </>
          )}
        </Tabs>
      </div>
    </div>
  );
}

export default Pricing;
