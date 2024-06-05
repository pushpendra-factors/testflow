import { Tabs } from 'antd';
import React from 'react';
import useTabs from 'hooks/useTabs';
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
  const { activeKey, handleTabChange } = useTabs(
    TabTypes.identificationProvider
  );

  return (
    <div>
      <Tabs activeKey={activeKey} onChange={handleTabChange}>
        <Tabs.TabPane
          tab='Identification Provider'
          key={TabTypes.identificationProvider}
        >
          <IndentificationProvider kbLink={kbLink} />
        </Tabs.TabPane>
        <Tabs.TabPane tab='Enrichment Rules' key={TabTypes.enrichmentRules}>
          <EnrichmentRulesTab />
        </Tabs.TabPane>
      </Tabs>
    </div>
  );
};

export default FactorsAccountIdentification;
