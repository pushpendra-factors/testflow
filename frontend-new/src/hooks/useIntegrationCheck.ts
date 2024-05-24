import { useSelector } from 'react-redux';
import { INTEGRATION_ID } from 'Views/Settings/ProjectSettings/IntegrationSettings/integrations.constants';

const useIntegrationCheck = () => {
  const currentProjectSetting = useSelector(
    (state) => state.global.currentProjectSettings
  );
  const integrationV1 = useSelector((state) => state.global.projectSettingsV1);
  const { bingAds, marketo } = useSelector((state) => state.global);
  return {
    integrationInfo: {
      [INTEGRATION_ID.segment]: Boolean(currentProjectSetting?.int_segment),
      [INTEGRATION_ID.rudderstack]: Boolean(
        currentProjectSetting?.int_rudderstack
      ),
      [INTEGRATION_ID.google_ads]: Boolean(
        currentProjectSetting?.int_adwords_enabled_agent_uuid
      ),
      [INTEGRATION_ID.facebook]: Boolean(
        currentProjectSetting?.int_facebook_user_id
      ),
      [INTEGRATION_ID.linkedIn]: Boolean(
        currentProjectSetting?.int_linkedin_agent_uuid
      ),
      [INTEGRATION_ID.bing_ads]: Boolean(bingAds?.accounts),
      [INTEGRATION_ID.hubspot]: Boolean(currentProjectSetting?.int_hubspot),
      [INTEGRATION_ID.salesforce]: Boolean(
        currentProjectSetting?.int_salesforce_enabled_agent_uuid
      ),
      [INTEGRATION_ID.marketo]: Boolean(marketo?.status),
      [INTEGRATION_ID.lead_squared]: Boolean(
        currentProjectSetting?.lead_squared_config
      ),
      [INTEGRATION_ID.clearbit_reveal]: Boolean(
        currentProjectSetting?.int_clear_bit
      ),
      [INTEGRATION_ID.six_signal_by_6_sense]: Boolean(
        currentProjectSetting?.int_client_six_signal_key
      ),
      [INTEGRATION_ID.factors_website_de_anonymization]: Boolean(
        currentProjectSetting?.int_factors_six_signal_key
      ),
      [INTEGRATION_ID.slack]: Boolean(integrationV1?.int_slack),
      [INTEGRATION_ID.microsoft_teams]: Boolean(integrationV1?.int_teams), // check
      [INTEGRATION_ID.drift]: Boolean(currentProjectSetting?.int_drift),
      [INTEGRATION_ID.g2]: Boolean(currentProjectSetting?.int_g2),
      [INTEGRATION_ID.google_search_console]: Boolean(
        currentProjectSetting?.int_google_organic_enabled_agent_uuid
      ),
      [INTEGRATION_ID.sdk]: Boolean(integrationV1?.int_completed)
    }
  };
};

export interface IntegrationInfoInterface {
  [key: string]: boolean;
}

export default useIntegrationCheck;
