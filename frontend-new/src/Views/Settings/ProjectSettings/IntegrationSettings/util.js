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
  } else {
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
  }
};
