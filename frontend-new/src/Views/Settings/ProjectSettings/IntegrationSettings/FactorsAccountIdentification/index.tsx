import { Tabs } from 'antd';
import React, { useEffect, useState } from 'react';
import { useHistory, useLocation } from 'react-router-dom';
import useQuery from 'hooks/useQuery';
import EnrichmentRulesTab from '../../Pricing/EnrichmentRulesTab';
import IndentificationProvider from './IndentificationProvider';

const TabTypes = {
  identificationProvider: 'identificationProvider',
  enrichmentRules: 'enrichmentRules'
};
interface FactorsAccountIdentificationProps {
  kbLink: string;
}
const FactorsAccountIdentification = ({
  kbLink
}: FactorsAccountIdentificationProps) => {
  const location = useLocation();
  const history = useHistory();
  const [activeKey, setActiveKey] = useState(TabTypes.identificationProvider);
  const routerQuery = useQuery();
  const paramActiveTab = routerQuery.get('activeTab');

  const handleTabChange = (key: string) => {
    setActiveKey(key);
    history.replace(`${location.pathname}?activeTab=${key}`);
  };

  useEffect(() => {
    if (!paramActiveTab) {
      history.replace(`${location.pathname}?activeTab=${activeKey}`);
    }
    if (paramActiveTab && activeKey !== paramActiveTab) {
      setActiveKey(paramActiveTab);
    }
  }, [paramActiveTab, activeKey]);

  return (
    <div>
      <Tabs activeKey={activeKey} onChange={handleTabChange}>
        <Tabs.TabPane
          tab='Identification Provider'
          key={TabTypes.identificationProvider}
        >
          <IndentificationProvider kbLink={kbLink} />
        </Tabs.TabPane>
        <Tabs.TabPane tab='EnrichmentRules' key={TabTypes.enrichmentRules}>
          <EnrichmentRulesTab />
        </Tabs.TabPane>
      </Tabs>
    </div>
  );
};

export default FactorsAccountIdentification;
