import {
  TEMPLATE_CONSTANTS,
  Integration_Checks
} from 'Constants/templates.constants';
import { createDashboardFromTemplate } from 'Reducers/dashboard_templates/services';
import logger from 'Utils/logger';

export const INTEGRATION_HOME_PAGE = '/settings/integration';
export const ADWORDS_INTERNAL_REDIRECT_URI = '?googleAds=manageAccounts';
export const ADWORDS_REDIRECT_URI_NEW = '/adwords/v1/auth/redirect';

export const getDefaultTimelineConfigForSixSignal = (config) => {
  const defaultProps = [
    '$6Signal_name',
    '$6Signal_industry',
    '$6Signal_employee_range',
    '$6Signal_revenue_range'
  ];

  const mergeProps = (existingProps, additionalProps) => [
    ...new Set(existingProps.concat(additionalProps))
  ];

  const defaultTimelineConfig = {
    account_config: {
      table_props: mergeProps(defaultProps, [])
    },
    user_config: {
      table_props: mergeProps(defaultProps, ['$session_spent_time'])
    }
  };

  if (!config) {
    return defaultTimelineConfig;
  }
  const { account_config, user_config } = config.timelines_config || {};
  const prevAccountTableProps = account_config?.table_props || [];
  const prevUserTableProps = user_config?.table_props || [];

  return {
    ...config.timelines_config,
    account_config: {
      ...account_config,
      table_props: mergeProps(
        prevAccountTableProps,
        defaultTimelineConfig.account_config.table_props
      )
    },
    user_config: {
      ...user_config,
      table_props: mergeProps(
        prevUserTableProps,
        defaultTimelineConfig.user_config.table_props
      )
    }
  };
};

const checkIfDashboardIsAlreadyCreated = (
  dashboards: any[],
  templateId: string
) => {
  const templateDashboard = dashboards?.find(
    (dashboard: any) => dashboard?.template_type === templateId
  );
  return !!templateDashboard;
};

const getTemplateFromTemplateConstant = (
  templates: any[],
  templateConstant: string
) => {
  const template = templates?.find((t) => t?.type === templateConstant);
  if (template) {
    return template;
  }
  logger.warn('Template not found', templateConstant);
  return null;
};

export const createDashboardsFromTemplatesForRequiredIntegration = async (
  projectId: string,
  dashboards: any,
  templates: any[],
  sdkCheck: boolean,
  currentProjectSettings: any,
  bingAds: any,
  marketo: any
) => {
  try {
    const possibleTemplates = [
      TEMPLATE_CONSTANTS.GOOGLE_ADWORDS,
      TEMPLATE_CONSTANTS.G2_INFLLUENCE_SALESFORCE,
      TEMPLATE_CONSTANTS.G2_INFLUENCE_HUBSPOT,
      TEMPLATE_CONSTANTS.LINKEDIN_INFLUENCE_HUBSPOT,
      TEMPLATE_CONSTANTS.LINKEDIN_INFLUENCE_SALESFORCE
    ];
    const IntegrationChecks = new Integration_Checks(
      sdkCheck,
      currentProjectSettings,
      bingAds,
      marketo
    );

    let dashboardAddedFlag = false;

    // looping through each possible template
    possibleTemplates
      .filter(
        // filtering only templates for which dashboards are not there
        (templateConstant) =>
          !checkIfDashboardIsAlreadyCreated(dashboards, templateConstant)
      )
      .map((templateConstant) => {
        // mapping only template constants which have template id
        const template = getTemplateFromTemplateConstant(
          templates,
          templateConstant
        );
        if (!template || !template?.id) return false;
        return {
          id: template.id,
          requiredIntegrations: template?.required_integrations
        };
      })
      .filter((t) => !!t)
      .forEach(async ({ id, requiredIntegrations }) => {
        if (IntegrationChecks.checkRequirements(requiredIntegrations)?.result) {
          try {
            dashboardAddedFlag = true;
            await createDashboardFromTemplate(projectId, id);
          } catch (error) {
            logger.error('Error in template', error);
          }
        }
      });
    return dashboardAddedFlag;
  } catch (error) {
    logger.error('Error in creating dashboard', error);
    return false;
  }
};
