import React from 'react';
import { visualizationColors } from '../../../utils/dataFormatter';
import SparkChartWithCount from '../SparkChartWithCount';

export default {
  title: 'Components/SparkLineChartWithCount',
  component: SparkChartWithCount
};

const eventNames = {
  $form_submitted: 'Form Button Click',
  $hubspot_company_created: 'Company Created',
  $hubspot_company_updated: 'Company Updated',
  $hubspot_contact_created: 'Contact Created',
  $hubspot_contact_updated: 'Contact Updated',
  $hubspot_deal_created: 'Deal Created',
  $hubspot_deal_state_changed: 'Deal State Changed',
  $hubspot_deal_updated: 'Deal Updated',
  $hubspot_engagement_call_created: 'Engagement Call Created',
  $hubspot_engagement_call_updated: 'Engagement Call Updated',
  $hubspot_engagement_email: 'Engagement Email',
  $hubspot_engagement_meeting_created: 'Engagement Meeting Created',
  $hubspot_engagement_meeting_updated: 'Engagement Meeting Updated',
  $hubspot_form_submission: 'Hubspot Form Submissions',
  $leadsquared_lead_created: 'Lead Created',
  $leadsquared_lead_updated: 'Lead Updated',
  $offline_touch_point: 'Offline Touchpoint',
  $salesforce_account_created: 'Salesforce Account Created',
  $salesforce_account_updated: 'Salesforce Account Updated',
  $salesforce_opportunity_created: 'Salesforce Opportunity Created',
  $salesforce_opportunity_updated: 'Salesforce Opportunity Updated',
  $session: 'Website Session',
  $sf_account_created: 'Account Created',
  $sf_account_updated: 'Account Updated',
  $sf_campaign_member_created: 'Added To Campaign',
  $sf_campaign_member_updated: 'Interacted With Campaign',
  $sf_contact_created: 'Contact Created',
  $sf_contact_updated: 'Contact Updated',
  $sf_lead_created: 'Lead Created',
  $sf_lead_updated: 'Lead Updated',
  $sf_opportunity_created: 'Opportunity Created',
  $sf_opportunity_updated: 'Opportunity Updated',
  'www.acme.com': 'Session Display Name Test1'
};

const SPARK_CHART_SAMPLE_KPI_DATA = [
  {
    date: new Date('Sat Jul 16 2022 19:30:00 GMT+0530 (India Standard Time)'),
    $form_submitted: 866
  },
  {
    date: new Date('Sun Jul 17 2022 19:30:00 GMT+0530 (India Standard Time)'),
    $form_submitted: 766
  },
  {
    date: new Date('Mon Jul 18 2022 19:30:00 GMT+0530 (India Standard Time)'),
    $form_submitted: 803
  },
  {
    date: new Date('Tue Jul 19 2022 19:30:00 GMT+0530 (India Standard Time)'),
    $form_submitted: 826
  },
  {
    date: new Date('Wed Jul 20 2022 19:30:00 GMT+0530 (India Standard Time)'),
    $form_submitted: 793
  },
  {
    date: new Date('Thu Jul 21 2022 19:30:00 GMT+0530 (India Standard Time)'),
    $form_submitted: 829
  },
  {
    date: new Date('Fri Jul 22 2022 19:30:00 GMT+0530 (India Standard Time)'),
    $form_submitted: 879
  }
];

const SPARK_CHART_SAMPLE_KPI_HOURLY_DATA = [
  {
    date: new Date('Thu Jul 28 2022 19:30:00 GMT+0530 (India Standard Time)'),
    $form_submitted: 26
  },
  {
    date: new Date('Thu Jul 28 2022 20:30:00 GMT+0530 (India Standard Time)'),
    $form_submitted: 42
  },

  {
    date: new Date('Thu Jul 28 2022 21:30:00 GMT+0530 (India Standard Time)'),
    $form_submitted: 36
  },

  {
    date: new Date('Thu Jul 28 2022 22:30:00 GMT+0530 (India Standard Time)'),
    $form_submitted: 26
  },

  {
    date: new Date('Thu Jul 28 2022 23:30:00 GMT+0530 (India Standard Time)'),
    $form_submitted: 45
  },

  {
    date: new Date('Fri Jul 29 2022 00:30:00 GMT+0530 (India Standard Time)'),
    $form_submitted: 29
  },

  {
    date: new Date('Fri Jul 29 2022 01:30:00 GMT+0530 (India Standard Time)'),
    $form_submitted: 27
  },

  {
    date: new Date('Fri Jul 29 2022 02:30:00 GMT+0530 (India Standard Time)'),
    $form_submitted: 40
  },

  {
    date: new Date('Fri Jul 29 2022 03:30:00 GMT+0530 (India Standard Time)'),
    $form_submitted: 23
  },

  {
    date: new Date('Fri Jul 29 2022 04:30:00 GMT+0530 (India Standard Time)'),
    $form_submitted: 22
  },

  {
    date: new Date('Fri Jul 29 2022 05:30:00 GMT+0530 (India Standard Time)'),
    $form_submitted: 32
  },

  {
    date: new Date('Fri Jul 29 2022 06:30:00 GMT+0530 (India Standard Time)'),
    $form_submitted: 25
  },

  {
    date: new Date('Fri Jul 29 2022 07:30:00 GMT+0530 (India Standard Time)'),
    $form_submitted: 36
  },

  {
    date: new Date('Fri Jul 29 2022 08:30:00 GMT+0530 (India Standard Time)'),
    $form_submitted: 31
  },

  {
    date: new Date('Fri Jul 29 2022 09:30:00 GMT+0530 (India Standard Time)'),
    $form_submitted: 42
  },

  {
    date: new Date('Fri Jul 29 2022 10:30:00 GMT+0530 (India Standard Time)'),
    $form_submitted: 26
  },

  {
    date: new Date('Fri Jul 29 2022 11:30:00 GMT+0530 (India Standard Time)'),
    $form_submitted: 25
  },

  {
    date: new Date('Fri Jul 29 2022 12:30:00 GMT+0530 (India Standard Time)'),
    $form_submitted: 23
  },

  {
    date: new Date('Fri Jul 29 2022 13:30:00 GMT+0530 (India Standard Time)'),
    $form_submitted: 29
  },

  {
    date: new Date('Fri Jul 29 2022 14:30:00 GMT+0530 (India Standard Time)'),
    $form_submitted: 34
  },

  {
    date: new Date('Fri Jul 29 2022 15:30:00 GMT+0530 (India Standard Time)'),
    $form_submitted: 30
  },

  {
    date: new Date('Fri Jul 29 2022 16:30:00 GMT+0530 (India Standard Time)'),
    $form_submitted: 15
  },

  {
    date: new Date('Fri Jul 29 2022 17:30:00 GMT+0530 (India Standard Time)'),
    $form_submitted: 37
  },

  {
    date: new Date('Fri Jul 29 2022 18:30:00 GMT+0530 (India Standard Time)'),
    $form_submitted: 34
  }
];

const SPARK_CHART_SAMPLE_KPI_DATA_WITH_COMPARISON_DATA = [
  {
    date: new Date('Sat Jul 16 2022 19:30:00 GMT+0530 (India Standard Time)'),
    compareDate: new Date(
      'Sat Jul 16 2022 19:30:00 GMT+0530 (India Standard Time)'
    ),
    $form_submitted: 866,
    compareValue: 800
  },
  {
    date: new Date('Sun Jul 17 2022 19:30:00 GMT+0530 (India Standard Time)'),
    $form_submitted: 766,
    compareDate: new Date(
      'Sun Jul 17 2022 19:30:00 GMT+0530 (India Standard Time)'
    ),
    compareValue: 800
  },
  {
    date: new Date('Mon Jul 18 2022 19:30:00 GMT+0530 (India Standard Time)'),
    compareDate: new Date(
      'Mon Jul 18 2022 19:30:00 GMT+0530 (India Standard Time)'
    ),
    $form_submitted: 803,
    compareValue: 18000
  },
  {
    date: new Date('Tue Jul 19 2022 19:30:00 GMT+0530 (India Standard Time)'),
    compareDate: new Date(
      'Tue Jul 19 2022 19:30:00 GMT+0530 (India Standard Time)'
    ),
    $form_submitted: 826,
    compareValue: 800
  },
  {
    date: new Date('Wed Jul 20 2022 19:30:00 GMT+0530 (India Standard Time)'),
    compareDate: new Date(
      'Wed Jul 20 2022 19:30:00 GMT+0530 (India Standard Time)'
    ),
    $form_submitted: 793,
    compareValue: 800
  },
  {
    date: new Date('Thu Jul 21 2022 19:30:00 GMT+0530 (India Standard Time)'),
    compareDate: new Date(
      'Thu Jul 21 2022 19:30:00 GMT+0530 (India Standard Time)'
    ),
    $form_submitted: 829,
    compareValue: 800
  },
  {
    date: new Date('Fri Jul 22 2022 19:30:00 GMT+0530 (India Standard Time)'),
    compareDate: new Date(
      'Fri Jul 22 2022 19:30:00 GMT+0530 (India Standard Time)'
    ),
    $form_submitted: 879,
    compareValue: 1000
  }
];

export const DefaultChart = () => {
  return (
    // event prop should match the key present in the chartData array. In this case, event is $form_submitted
    <SparkChartWithCount
      chartData={SPARK_CHART_SAMPLE_KPI_DATA}
      event="$form_submitted"
      total={SPARK_CHART_SAMPLE_KPI_DATA.reduce(
        (total, elem) => total + elem.$form_submitted,
        0
      )}
      headerTitle="Form Submitted"
    />
  );
};

export const WithComparisonEnabled = () => {
  return (
    // event prop should match the key present in the chartData array. In this case, event is $form_submitted
    <SparkChartWithCount
      chartColor={visualizationColors[7]}
      chartData={SPARK_CHART_SAMPLE_KPI_DATA_WITH_COMPARISON_DATA}
      event="$form_submitted"
      comparisonEnabled={true}
      total={SPARK_CHART_SAMPLE_KPI_DATA.reduce(
        (total, elem) => total + elem.$form_submitted,
        0
      )}
      compareTotal={SPARK_CHART_SAMPLE_KPI_DATA_WITH_COMPARISON_DATA.reduce(
        (total, elem) => total + elem.compareValue,
        0
      )}
      eventNames={eventNames}
      smallFont={true}
      headerTitle="Form Submitted"
    />
  );
};

export const WithVerticalAlignment = () => {
  return (
    // event prop should match the key present in the chartData array. In this case, event is $form_submitted
    <SparkChartWithCount
      chartColor={visualizationColors[4]}
      chartData={SPARK_CHART_SAMPLE_KPI_DATA_WITH_COMPARISON_DATA}
      event="$form_submitted"
      comparisonEnabled={true}
      total={SPARK_CHART_SAMPLE_KPI_DATA.reduce(
        (total, elem) => total + elem.$form_submitted,
        0
      )}
      compareTotal={SPARK_CHART_SAMPLE_KPI_DATA_WITH_COMPARISON_DATA.reduce(
        (total, elem) => total + elem.compareValue,
        0
      )}
      eventNames={eventNames}
      smallFont={false}
      alignment="vertical"
      headerTitle="Form Submitted"
    />
  );
};
