import { SVG } from 'Components/factorsComponents';
import { Breadcrumb, Tabs, notification } from 'antd';
import React, { useEffect, useState } from 'react';
import { useHistory } from 'react-router-dom';
import BillingTab from './BillingTab';
import EnrichmentRulesTab from './EnrichmentRulesTab';
import { PRICING_PAGE_TABS, showV2PricingVersion } from './utils';
import { PathUrls } from 'Routes/pathUrls';
import useQuery from 'hooks/useQuery';
import UpgradeTab from './UpgradeTab';
import InvoiceTab from './InvoiceTab';
import { useSelector } from 'react-redux';
import { startCase } from 'lodash';
import { upgradePlan } from 'Reducers/plansConfig/services';
import { ADDITIONAL_ACCOUNTS_ADDON_ID } from 'Constants/plans.constants';
import logger from 'Utils/logger';

function Pricing() {
  const [activeKey, setActiveKey] = useState(PRICING_PAGE_TABS.BILLING);
  const [loading, setIsLoading] = useState<boolean>(false);
  const history = useHistory();
  const routerQuery = useQuery();
  const { active_project } = useSelector((state) => state.global);
  const paramActiveTab = routerQuery.get('activeTab');

  const handleTabChange = (activeKey: string) => {
    setActiveKey(activeKey);
    history.replace(`${PathUrls.SettingsPricing}?activeTab=${activeKey}`);
  };

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

  useEffect(() => {
    if (!paramActiveTab) {
      history.replace(`${PathUrls.SettingsPricing}?activeTab=${activeKey}`);
    }
    if (paramActiveTab && activeKey !== paramActiveTab) {
      setActiveKey(paramActiveTab);
    }
  }, [paramActiveTab, activeKey]);

  return (
    <div>
      <div className='flex gap-3 items-center'>
        <div className='cursor-pointer' onClick={() => history.goBack()}>
          <SVG name='ArrowLeft' size='16' />
        </div>
        <div>
          <Breadcrumb>
            <Breadcrumb.Item>Settings</Breadcrumb.Item>
            <Breadcrumb.Item>Pricing</Breadcrumb.Item>
            <Breadcrumb.Item>{startCase(activeKey)}</Breadcrumb.Item>
          </Breadcrumb>
        </div>
      </div>
      <div className='mt-6'>
        <Tabs activeKey={activeKey} onChange={handleTabChange}>
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
          <Tabs.TabPane
            tab='Enrichment Rules'
            key={PRICING_PAGE_TABS.ENRICHMENT_RULES}
          >
            <EnrichmentRulesTab />
          </Tabs.TabPane>
        </Tabs>
      </div>
    </div>
  );
}

export default Pricing;
