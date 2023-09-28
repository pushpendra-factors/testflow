import { uniq } from 'lodash';
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
  if (!config) {
    return {
      account_config: {
        table_props: [...defaultProps],
        leftpane_props: [...defaultProps]
      },
      user_config: {
        table_props: [...defaultProps, '$session_spent_time'],
        leftpane_props: [...defaultProps, '$session_spent_time']
      }
    };
  } else {
    const prevAccountTableProps = config?.account_config?.table_props || [];
    const prevUsertTableProps = config?.user_config?.table_props || [];
    return {
      ...config?.timelines_config,
      account_config: {
        ...config?.timelines_config?.account_config,
        table_props: [...uniq(prevAccountTableProps.concat(defaultProps))]
      },
      user_config: {
        ...config?.timelines_config?.user_config,
        table_props: [...uniq(prevUsertTableProps.concat(defaultProps))]
      }
    };
  }
};
