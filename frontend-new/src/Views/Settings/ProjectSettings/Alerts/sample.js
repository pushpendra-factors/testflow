export const ALERTS_DATA = [
  {
    id: 1,
    title: 'One of your accounts just visited the pricing page',
    description:
      'Get alerts whenever one of your accounts just out your pricing page and reach out in real time.',
    categories: ['Account Executives'],
    required_integrations: [
      ['hubspot', 'website_sdk'],
      ['salesforce', 'website_sdk']
    ],
    alert_name: 'An account visited pricing page',
    alert_message: 'One your accounts just visited the pricing page',
    questions: [],
    payload_props: {
      'hubspot,website_sdk': [
        {
          prop_category: 'user',
          eventName: '$page_view',
          property: '$hubspot_company_name',
          prop_type: 'categorical',
          eventIndex: 1,
          overAllIndex: 0
        },
        {
          prop_category: 'user',
          eventName: '$page_view',
          property: '$hubspot_company_lifecyclestage',
          prop_type: 'categorical',
          eventIndex: 1,
          overAllIndex: 1
        },
        {
          prop_category: 'user',
          eventName: '$page_view',
          property: '$hubspot_company_hubspot_owner_id',
          prop_type: 'categorical',
          eventIndex: 1,
          overAllIndex: 2
        },
        {
          prop_category: 'event',
          eventName: '$page_view',
          property: '$page_url',
          prop_type: 'categorical',
          eventIndex: 1,
          overAllIndex: 3
        },
        {
          prop_category: 'event',
          eventName: '$page_view',
          property: '$page_title',
          prop_type: 'categorical',
          eventIndex: 1,
          overAllIndex: 4
        },
        {
          prop_category: 'event',
          eventName: '$page_view',
          property: '$page_spent_time',
          prop_type: 'numerical',
          eventIndex: 1,
          overAllIndex: 5
        },
        {
          prop_category: 'event',
          eventName: '$page_view',
          property: '$page_scroll_percent',
          prop_type: 'numerical',
          eventIndex: 1,
          overAllIndex: 6
        },
        {
          prop_category: 'event',
          eventName: '$page_view',
          property: '$referrer_url',
          prop_type: 'categorical',
          eventIndex: 1,
          overAllIndex: 7
        }
      ],
      'salesforce,website_sdk': [
        {
          prop_category: 'user',
          eventName: '$page_view',
          property: '$salesforce_company_name',
          prop_type: 'categorical',
          eventIndex: 1,
          overAllIndex: 0
        },
        {
          prop_category: 'user',
          eventName: '$page_view',
          property: '$salesforce_account_ownerid',
          prop_type: 'categorical',
          eventIndex: 1,
          groupName: '$hubspot_company',
          overAllIndex: 1
        },
        {
          prop_category: 'event',
          eventName: '$page_view',
          property: '$page_url',
          prop_type: 'categorical',
          eventIndex: 1,
          overAllIndex: 2
        },
        {
          prop_category: 'event',
          eventName: '$page_view',
          property: '$page_title',
          prop_type: 'categorical',
          eventIndex: 1,
          overAllIndex: 3
        },
        {
          prop_category: 'event',
          eventName: '$page_view',
          property: '$page_spent_time',
          prop_type: 'numerical',
          eventIndex: 1,
          overAllIndex: 4
        },
        {
          prop_category: 'event',
          eventName: '$page_view',
          property: '$page_scroll_percent',
          prop_type: 'numerical',
          eventIndex: 1,
          overAllIndex: 5
        },
        {
          prop_category: 'event',
          eventName: '$page_view',
          property: '$referrer_url',
          prop_type: 'categorical',
          eventIndex: 1,
          overAllIndex: 6
        }
      ]
    },
    icon: 'UserTie',
    backgroundColor: '#FFF7E6',
    color: '#D46B08',
    question: 'What is the URL of your pricing page?',
    prepopulate: {
      'hubspot,website_sdk': {
        event: { label: '$page_view', group: 'Website activity' },
        filterBy: [
          {
            operator: 'contains',
            props: ['event', '$page_url', 'categorical', 'event'],
            values: ['/pricing'],
            ref: 1
          }
        ]
      },
      'salesforce,website_sdk': {
        event: { label: '$page_view', group: 'Website activity' },
        filterBy: [
          {
            operator: 'contains',
            props: ['event', '$page_url', 'categorical', 'event'],
            values: ['/pricing'],
            ref: 1
          }
        ]
      }
    }
  },
  {
    id: 2,
    title: 'One of your opportunities is on the website',
    description:
      'Stay on top activity from your deals/opportunities and reach out to them at the right time.',
    categories: ['Account Executives'],
    required_integrations: [
      ['hubspot', 'website_sdk'],
      ['salesforce', 'website_sdk']
    ],
    alert_name: 'One of your opportunities is on the website',
    alert_message:
      'An existing opportunity just visited our website. Time to reach out!',
    questions: [],
    payload_props: {
      'hubspot,website_sdk': [
        {
          prop_category: 'user',
          eventName: '$session',
          property: '$hubspot_deal_dealname',
          prop_type: 'categorical',
          eventIndex: 1,
          overAllIndex: 0
        },
        {
          prop_category: 'user',
          eventName: '$session',
          property: '$hubspot_deal_amount',
          prop_type: 'numerical',
          eventIndex: 1,
          groupName: '$hubspot_deal',
          gbty: 'raw_values',
          overAllIndex: 1
        },
        {
          prop_category: 'user',
          eventName: '$session',
          property: '$hubspot_deal_hubspot_owner_id',
          prop_type: 'categorical',
          eventIndex: 1,
          groupName: '$hubspot_deal',
          overAllIndex: 2
        },
        {
          prop_category: 'user',
          eventName: '$session',
          property: '$6Signal_domain',
          prop_type: 'categorical',
          eventIndex: 1,
          groupName: '$6signal',
          overAllIndex: 3
        },
        {
          prop_category: 'user',
          eventName: '$session',
          property: '$6Signal_name',
          prop_type: 'categorical',
          eventIndex: 1,
          groupName: '$6signal',
          overAllIndex: 4
        },
        {
          prop_category: 'event',
          eventName: '$session',
          property: '$initial_page_url',
          prop_type: 'categorical',
          eventIndex: 1,
          groupName: 'Session properties',
          overAllIndex: 5
        },
        {
          prop_category: 'event',
          eventName: '$session',
          property: '$campaign',
          prop_type: 'categorical',
          eventIndex: 1,
          groupName: 'Session properties',
          overAllIndex: 6
        },
        {
          prop_category: 'event',
          eventName: '$session',
          property: '$term',
          prop_type: 'categorical',
          eventIndex: 1,
          groupName: 'Session properties',
          overAllIndex: 7
        },
        {
          prop_category: 'event',
          eventName: '$session',
          property: '$city',
          prop_type: 'categorical',
          eventIndex: 1,
          groupName: 'Session properties',
          overAllIndex: 8
        },
        {
          prop_category: 'event',
          eventName: '$session',
          property: '$session_spent_time',
          prop_type: 'numerical',
          eventIndex: 1,
          groupName: 'Session properties',
          gbty: 'raw_values',
          overAllIndex: 9
        },
        {
          prop_category: 'event',
          eventName: '$session',
          property: '$page_count',
          prop_type: 'numerical',
          eventIndex: 1,
          groupName: 'Session properties',
          gbty: 'raw_values',
          overAllIndex: 10
        },
        {
          prop_category: 'event',
          eventName: '$session',
          property: '$session_latest_page_url',
          prop_type: 'categorical',
          eventIndex: 1,
          groupName: 'Session properties',
          overAllIndex: 11
        }
      ],
      'salesforce,website_sdk': [
        {
          prop_category: 'user',
          eventName: '$session',
          property: '$salesforce_opportunity_name',
          prop_type: 'categorical',
          eventIndex: 1,
          overAllIndex: 0
        },
        {
          prop_category: 'user',
          eventName: '$session',
          property: '$salesforce_opportunity_amount',
          prop_type: 'numerical',
          eventIndex: 1,
          groupName: '$hubspot_deal',
          gbty: 'raw_values',
          overAllIndex: 1
        },
        {
          prop_category: 'user',
          eventName: '$session',
          property: '$salesforce_opportunity_id',
          prop_type: 'categorical',
          eventIndex: 1,
          groupName: '$hubspot_deal',
          overAllIndex: 2
        },
        {
          prop_category: 'user',
          eventName: '$session',
          property: '$6Signal_domain',
          prop_type: 'categorical',
          eventIndex: 1,
          groupName: '$6signal',
          overAllIndex: 3
        },
        {
          prop_category: 'user',
          eventName: '$session',
          property: '$6Signal_name',
          prop_type: 'categorical',
          eventIndex: 1,
          groupName: '$6signal',
          overAllIndex: 4
        },
        {
          prop_category: 'event',
          eventName: '$session',
          property: '$initial_page_url',
          prop_type: 'categorical',
          eventIndex: 1,
          groupName: 'Session properties',
          overAllIndex: 5
        },
        {
          prop_category: 'event',
          eventName: '$session',
          property: '$campaign',
          prop_type: 'categorical',
          eventIndex: 1,
          groupName: 'Session properties',
          overAllIndex: 6
        },
        {
          prop_category: 'event',
          eventName: '$session',
          property: '$term',
          prop_type: 'categorical',
          eventIndex: 1,
          groupName: 'Session properties',
          overAllIndex: 7
        },
        {
          prop_category: 'event',
          eventName: '$session',
          property: '$city',
          prop_type: 'categorical',
          eventIndex: 1,
          groupName: 'Session properties',
          overAllIndex: 8
        },
        {
          prop_category: 'event',
          eventName: '$session',
          property: '$session_spent_time',
          prop_type: 'numerical',
          eventIndex: 1,
          groupName: 'Session properties',
          gbty: 'raw_values',
          overAllIndex: 9
        },
        {
          prop_category: 'event',
          eventName: '$session',
          property: '$page_count',
          prop_type: 'numerical',
          eventIndex: 1,
          groupName: 'Session properties',
          gbty: 'raw_values',
          overAllIndex: 10
        },
        {
          prop_category: 'event',
          eventName: '$session',
          property: '$session_latest_page_url',
          prop_type: 'categorical',
          eventIndex: 1,
          groupName: 'Session properties',
          overAllIndex: 11
        }
      ]
    },
    icon: 'UserTie',
    backgroundColor: '#FFF7E6',
    color: '#D46B08',
    question: 'How do you define an opportunity in CRM?',
    prepopulate: {
      'hubspot,website_sdk': {
        event: { label: '$session', group: 'Website activity' },
        filterBy: [
          {
            operator: 'is known',
            props: ['user', '$hubspot_contact_email', 'categorical', 'user'],
            values: ['$none'],
            ref: 1
          }
        ]
      },
      'salesforce,website_sdk': {
        event: { label: '$session', group: 'Website activity' },
        filterBy: [
          {
            operator: 'is known',
            props: [
              'user',
              '$salesforce_opportunity_name',
              'categorical',
              'user'
            ],
            values: ['$none'],
            ref: 1
          }
        ]
      }
    }
  },
  {
    id: 3,
    title: 'One of your churned accounts just came back to the website',
    description:
      'Known when a churned account revisits your website and the activities they perform.',
    categories: ['Account Executives'],
    required_integrations: [
      ['hubspot', 'website_sdk'],
      ['salesforce', 'website_sdk']
    ],
    alert_name: 'Churned account just visited the website',
    alert_message: 'This churned account just revisited our website',
    questions: [],
    payload_props: {
      'hubspot,website_sdk': [
        {
          prop_category: 'user',
          eventName: '$session',
          property: '$hubspot_deal_dealname',
          prop_type: 'categorical',
          eventIndex: 1,
          overAllIndex: 0
        },
        {
          prop_category: 'user',
          eventName: '$session',
          property: '$hubspot_deal_amount',
          prop_type: 'numerical',
          eventIndex: 1,
          groupName: '$hubspot_deal',
          gbty: 'raw_values',
          overAllIndex: 1
        },
        {
          prop_category: 'user',
          eventName: '$session',
          property: '$hubspot_deal_hubspot_owner_id',
          prop_type: 'categorical',
          eventIndex: 1,
          groupName: '$hubspot_deal',
          overAllIndex: 2
        },
        {
          prop_category: 'event',
          eventName: '$session',
          property: '$initial_page_url',
          prop_type: 'categorical',
          eventIndex: 1,
          groupName: 'Session properties',
          overAllIndex: 3
        },
        {
          prop_category: 'event',
          eventName: '$session',
          property: '$campaign',
          prop_type: 'categorical',
          eventIndex: 1,
          groupName: 'Session properties',
          overAllIndex: 4
        },
        {
          prop_category: 'event',
          eventName: '$session',
          property: '$term',
          prop_type: 'categorical',
          eventIndex: 1,
          groupName: 'Session properties',
          overAllIndex: 5
        },
        {
          prop_category: 'event',
          eventName: '$session',
          property: '$city',
          prop_type: 'categorical',
          eventIndex: 1,
          groupName: 'Session properties',
          overAllIndex: 6
        },
        {
          prop_category: 'event',
          eventName: '$session',
          property: '$session_spent_time',
          prop_type: 'numerical',
          eventIndex: 1,
          groupName: 'Session properties',
          gbty: 'raw_values',
          overAllIndex: 7
        },
        {
          prop_category: 'event',
          eventName: '$session',
          property: '$page_count',
          prop_type: 'numerical',
          eventIndex: 1,
          groupName: 'Session properties',
          gbty: 'raw_values',
          overAllIndex: 8
        },
        {
          prop_category: 'event',
          eventName: '$session',
          property: '$session_latest_page_url',
          prop_type: 'categorical',
          eventIndex: 1,
          groupName: 'Session properties',
          overAllIndex: 9
        }
      ],
      'salesforce,website_sdk': [
        {
          prop_category: 'user',
          eventName: '$session',
          property: '$salesforce_opportunity_name',
          prop_type: 'categorical',
          eventIndex: 1,
          overAllIndex: 0
        },
        {
          prop_category: 'user',
          eventName: '$session',
          property: '$salesforce_opportunity_amount',
          prop_type: 'numerical',
          eventIndex: 1,
          groupName: '$hubspot_deal',
          gbty: 'raw_values',
          overAllIndex: 1
        },
        {
          prop_category: 'user',
          eventName: '$session',
          property: '$salesforce_opportunity_ownerid',
          prop_type: 'categorical',
          eventIndex: 1,
          groupName: '$hubspot_deal',
          overAllIndex: 2
        },
        {
          prop_category: 'event',
          eventName: '$session',
          property: '$initial_page_url',
          prop_type: 'categorical',
          eventIndex: 1,
          groupName: 'Session properties',
          overAllIndex: 3
        },
        {
          prop_category: 'event',
          eventName: '$session',
          property: '$campaign',
          prop_type: 'categorical',
          eventIndex: 1,
          groupName: 'Session properties',
          overAllIndex: 4
        },
        {
          prop_category: 'event',
          eventName: '$session',
          property: '$term',
          prop_type: 'categorical',
          eventIndex: 1,
          groupName: 'Session properties',
          overAllIndex: 5
        },
        {
          prop_category: 'event',
          eventName: '$session',
          property: '$city',
          prop_type: 'categorical',
          eventIndex: 1,
          groupName: 'Session properties',
          overAllIndex: 6
        },
        {
          prop_category: 'event',
          eventName: '$session',
          property: '$session_spent_time',
          prop_type: 'numerical',
          eventIndex: 1,
          groupName: 'Session properties',
          gbty: 'raw_values',
          overAllIndex: 7
        },
        {
          prop_category: 'event',
          eventName: '$session',
          property: '$page_count',
          prop_type: 'numerical',
          eventIndex: 1,
          groupName: 'Session properties',
          gbty: 'raw_values',
          overAllIndex: 8
        },
        {
          prop_category: 'event',
          eventName: '$session',
          property: '$session_latest_page_url',
          prop_type: 'categorical',
          eventIndex: 1,
          groupName: 'Session properties',
          overAllIndex: 9
        }
      ]
    },
    icon: 'UserTie',
    backgroundColor: '#FFF7E6',
    color: '#D46B08',
    question: 'How do you define a churned account in your CRM?',
    prepopulate: {
      'hubspot,website_sdk': {
        event: { label: '$session', group: 'Website activity' },
        filterBy: [
          {
            operator: 'equals',
            props: ['user', '$hubspot_deal_dealstage', 'categorical', 'user'],
            values: ['closed lost'],
            ref: 1
          }
        ]
      },
      'salesforce,website_sdk': {
        event: { label: '$session', group: 'Website activity' },
        filterBy: [
          {
            operator: 'equals',
            props: [
              'user',
              '$salesforce_opportunity_stagename',
              'categorical',
              'user'
            ],
            values: ['closed lost'],
            ref: 1
          }
        ]
      }
    }
  },
  {
    id: 4,
    title: 'An ICP account just visited the website',
    description:
      'Get alerted when one of your target accounts visits your website.',
    categories: ['SDRs'],
    required_integrations: [['website_sdk']],
    alert_name: 'ICP account visited website',
    alert_message: 'An account that matches our ICP just visited the website',
    questions: [],
    payload_props: {'website_sdk': [{"prop_category":"user","eventName":"$session","property":"$6Signal_domain","prop_type":"categorical","eventIndex":1,"overAllIndex":0},{"prop_category":"user","eventName":"$session","property":"$6Signal_name","prop_type":"categorical","eventIndex":1,"groupName":"$6signal","overAllIndex":1},{"prop_category":"user","eventName":"$session","property":"$6Signal_industry","prop_type":"categorical","eventIndex":1,"groupName":"$6signal","overAllIndex":2},{"prop_category":"user","eventName":"$session","property":"$6Signal_revenue_range","prop_type":"categorical","eventIndex":1,"groupName":"$6signal","overAllIndex":3},{"prop_category":"user","eventName":"$session","property":"$6Signal_employee_range","prop_type":"categorical","eventIndex":1,"groupName":"$6signal","overAllIndex":4},{"prop_category":"event","eventName":"$session","property":"$initial_page_url","prop_type":"categorical","eventIndex":1,"groupName":"Session properties","overAllIndex":5},{"prop_category":"event","eventName":"$session","property":"$city","prop_type":"categorical","eventIndex":1,"groupName":"Session properties","overAllIndex":6},{"prop_category":"event","eventName":"$session","property":"$session_spent_time","prop_type":"numerical","eventIndex":1,"groupName":"Session properties","gbty":"raw_values","overAllIndex":7},{"prop_category":"event","eventName":"$session","property":"$page_count","prop_type":"numerical","eventIndex":1,"groupName":"Session properties","gbty":"raw_values","overAllIndex":8},{"prop_category":"event","eventName":"$session","property":"$session_latest_page_url","prop_type":"categorical","eventIndex":1,"groupName":"Session properties","overAllIndex":9}]},
    icon: 'Headset',
    backgroundColor: '#E6FFFB',
    color: '#08979C',
    question: 'How do you define an ICP Account?',
    prepopulate: {
      'website_sdk': {
        event: { label: '$session', group: 'Website activity' },
        filterBy: [
          {
            operator: 'equals',
            props: ['$6signal', '$6Signal_industry', 'categorical', 'group'],
            values: ['Business Services', 'Software and Technology'],
            ref: 1
          },
          {
            operator: 'equals',
            props: [
              '$6signal',
              '$6Signal_revenue_range',
              'categorical',
              'group'
            ],
            values: ['$1M - $5M', '$5M - $10M'],
            ref: 2
          },
          {
            operator: 'equals',
            props: ['$6signal', '$6Signal_country', 'categorical', 'group'],
            values: ['Australia', 'United Kingdom', 'United States'],
            ref: 3
          }
        ]
      }
    }
  },
  {
    id: 5,
    title: 'An account just read one of your blogs',
    description:
      'Get notified if an account of interest spends time on your blog content pages.',
    categories: ['SDRs'],
    required_integrations: [['website_sdk']],
    alert_name: 'An account just visited a blog page',
    alert_message: 'This account just spent 30 seconds on a blog page',
    questions: [],
    payload_props: {'website_sdk': [{"prop_category":"user","eventName":"$page_view","property":"$6Signal_domain","prop_type":"categorical","eventIndex":1,"overAllIndex":0},{"prop_category":"user","eventName":"$page_view","property":"$6Signal_name","prop_type":"categorical","eventIndex":1,"overAllIndex":1},{"prop_category":"user","eventName":"$page_view","property":"$6Signal_industry","prop_type":"categorical","eventIndex":1,"overAllIndex":2},{"prop_category":"user","eventName":"$page_view","property":"$6Signal_revenue_range","prop_type":"categorical","eventIndex":1,"groupName":"$6signal","overAllIndex":3},{"prop_category":"user","eventName":"$page_view","property":"$6Signal_employee_range","prop_type":"categorical","eventIndex":1,"groupName":"$6signal","overAllIndex":4},{"prop_category":"event","eventName":"$page_view","property":"$page_url","prop_type":"categorical","eventIndex":1,"groupName":"Page properties","overAllIndex":5},{"prop_category":"event","eventName":"$page_view","property":"$page_title","prop_type":"categorical","eventIndex":1,"groupName":"Page properties","overAllIndex":6},{"prop_category":"event","eventName":"$page_view","property":"$page_spent_time","prop_type":"numerical","eventIndex":1,"groupName":"Page properties","gbty":"raw_values","overAllIndex":7},{"prop_category":"event","eventName":"$page_view","property":"$page_scroll_percent","prop_type":"numerical","eventIndex":1,"groupName":"Page properties","gbty":"raw_values","overAllIndex":8},{"prop_category":"event","eventName":"$page_view","property":"$referrer_url","prop_type":"categorical","eventIndex":1,"groupName":"Traffic source","overAllIndex":9}]},
    icon: 'Headset',
    backgroundColor: '#E6FFFB',
    color: '#08979C',
    question: 'How do you define your blog pages URL?',
    prepopulate: {
      'website_sdk': {
        event: { label: '$page_view', group: 'Website activity' },
        filterBy: [
          {
            operator: 'contains',
            props: ['event', '$page_url', 'categorical', 'event'],
            values: ['/blog'],
            ref: 1
          },
          {
            operator: '>',
            props: ['event', '$page_spent_time', 'numerical', 'event'],
            values: ['30'],
            ref: 2
          }
        ]
      }
    }
  },
  {
    id: 6,
    title: 'An account came to your website via paid search keywords',
    description:
      'Know when an account comes to your website through a paid search keyword.',
    categories: ['SDRs'],
    required_integrations: [['website_sdk']],
    alert_name: 'Account came to website through paid search',
    alert_message:
      'This account just came to the website through a paid search keyword',
    questions: [],
    payload_props: {'website_sdk': [{"prop_category":"user","eventName":"$session","property":"$6Signal_domain","prop_type":"categorical","eventIndex":1,"overAllIndex":0},{"prop_category":"user","eventName":"$session","property":"$6Signal_name","prop_type":"categorical","eventIndex":1,"overAllIndex":1},{"prop_category":"user","eventName":"$session","property":"$6Signal_industry","prop_type":"categorical","eventIndex":1,"overAllIndex":2},{"prop_category":"user","eventName":"$session","property":"$6Signal_revenue_range","prop_type":"categorical","eventIndex":1,"groupName":"$6signal","overAllIndex":3},{"prop_category":"user","eventName":"$session","property":"$6Signal_employee_range","prop_type":"categorical","eventIndex":1,"groupName":"$6signal","overAllIndex":4},{"prop_category":"event","eventName":"$session","property":"$initial_page_url","prop_type":"categorical","eventIndex":1,"groupName":"Session properties","overAllIndex":5},{"prop_category":"event","eventName":"$session","property":"$term","prop_type":"categorical","eventIndex":1,"groupName":"Session properties","overAllIndex":6},{"prop_category":"event","eventName":"$session","property":"$campaign","prop_type":"categorical","eventIndex":1,"groupName":"Session properties","overAllIndex":7},{"prop_category":"event","eventName":"$session","property":"$session_spent_time","prop_type":"numerical","eventIndex":1,"groupName":"Session properties","gbty":"raw_values","overAllIndex":8},{"prop_category":"event","eventName":"$session","property":"$session_latest_page_url","prop_type":"categorical","eventIndex":1,"groupName":"Session properties","overAllIndex":9}]},
    icon: 'Headset',
    backgroundColor: '#E6FFFB',
    color: '#08979C',
    prepopulate: {
      'website_sdk': {
        event: { label: '$session', group: 'Website activity' },
        filterBy: [
          {
            operator: 'equals',
            props: ['event', '$channel', 'categorical', 'event'],
            values: ['Paid Search'],
            ref: 1
          },
        ]
      }
    }
  },
  {
    id: 7,
    title: 'An account just visited a competitor comparison page',
    description:
      'Get notified whenever an account visits a competitor comparison page on your webiste.',
    categories: ['SDRs'],
    required_integrations: [['website_sdk']],
    alert_name: 'Account visited competitor comparison page',
    alert_message:
      'This account just visited a comparison page for one of our competitors',
    questions: [],
    payload_props: {'website_sdk': [{"prop_category":"user","eventName":"$page_view","property":"$6Signal_domain","prop_type":"categorical","eventIndex":1,"overAllIndex":0},{"prop_category":"user","eventName":"$page_view","property":"$6Signal_name","prop_type":"categorical","eventIndex":1,"groupName":"$6signal","overAllIndex":1},{"prop_category":"user","eventName":"$page_view","property":"$6Signal_industry","prop_type":"categorical","eventIndex":1,"groupName":"$6signal","overAllIndex":2},{"prop_category":"user","eventName":"$page_view","property":"$6Signal_revenue_range","prop_type":"categorical","eventIndex":1,"groupName":"$6signal","overAllIndex":3},{"prop_category":"user","eventName":"$page_view","property":"$6Signal_employee_range","prop_type":"categorical","eventIndex":1,"groupName":"$6signal","overAllIndex":4},{"prop_category":"event","eventName":"$page_view","property":"$page_url","prop_type":"categorical","eventIndex":1,"groupName":"Page properties","overAllIndex":5},{"prop_category":"event","eventName":"$page_view","property":"$page_title","prop_type":"categorical","eventIndex":1,"groupName":"Page properties","overAllIndex":6},{"prop_category":"event","eventName":"$page_view","property":"$page_spent_time","prop_type":"numerical","eventIndex":1,"groupName":"Page properties","gbty":"raw_values","overAllIndex":7},{"prop_category":"event","eventName":"$page_view","property":"$page_scroll_percent","prop_type":"numerical","eventIndex":1,"groupName":"Page properties","gbty":"raw_values","overAllIndex":8},{"prop_category":"event","eventName":"$page_view","property":"$referrer_url","prop_type":"categorical","eventIndex":1,"groupName":"Traffic source","overAllIndex":9}]},
    icon: 'Headset',
    backgroundColor: '#E6FFFB',
    color: '#08979C',
    question: 'How would you define your competitor pages?',
    prepopulate: {
      'website_sdk': {
        event: { label: '$page_view', group: 'Website activity' },
        filterBy: [
          {
            operator: 'contains',
            props: ['event', '$page_url', 'categorical', 'event'],
            values: ['/versus'],
            ref: 1
          }
        ]
      }
    }
  },
  {
    id: 8,
    title: 'A new lead just submitted a form on your website',
    description:
      'Get instant alerts for all form submissions that take place on your website.',
    categories: ['Marketing'],
    required_integrations: [['website_sdk']],
    alert_name: 'New form submission on website',
    alert_message: 'A form was just submitted on the website',
    questions: [],
    payload_props: {'website_sdk': [{"prop_category":"user","eventName":"$form_submitted","property":"$email","prop_type":"categorical","eventIndex":1,"overAllIndex":0},{"prop_category":"user","eventName":"$form_submitted","property":"$first_name","prop_type":"categorical","eventIndex":1,"groupName":"User identification","overAllIndex":1},{"prop_category":"user","eventName":"$form_submitted","property":"$last_name","prop_type":"categorical","eventIndex":1,"groupName":"User identification","overAllIndex":2},{"prop_category":"event","eventName":"$form_submitted","property":"$page_url","prop_type":"categorical","eventIndex":1,"groupName":"Page properties","overAllIndex":3},{"prop_category":"event","eventName":"$form_submitted","property":"$page_title","prop_type":"categorical","eventIndex":1,"groupName":"Page properties","overAllIndex":4},{"prop_category":"user","eventName":"$form_submitted","property":"$initial_channel","prop_type":"categorical","eventIndex":1,"groupName":"Traffic source","overAllIndex":5},{"prop_category":"user","eventName":"$form_submitted","property":"$latest_channel","prop_type":"categorical","eventIndex":1,"groupName":"Traffic source","overAllIndex":6}]},
    icon: 'SponsorShip',
    backgroundColor: '#F9F0FF',
    color: '#722ED1',
    prepopulate: {
      'website_sdk': {
        event: {label: '$form_submitted', group: 'Website activity'}
      }
    }
  },
  {
    id: 9,
    title: 'A company clicked on your LinkedIn ad',
    description: 'Get to know about companies clicking on your LinkedIn ads.',
    categories: ['Marketing'],
    required_integrations: [['linkedin', 'website_sdk']],
    alert_name: 'A LinkedIn ad was clicked',
    alert_message:
      'Someone from this company clicked on one of our LinkedIn ads',
    questions: [],
    payload_props: {'website_sdk': [{"prop_category":"user","eventName":"$linkedin_clicked_ad","property":"$li_localized_name","prop_type":"categorical","eventIndex":1,"overAllIndex":0},{"prop_category":"user","eventName":"$linkedin_clicked_ad","property":"$li_domain","prop_type":"categorical","eventIndex":1,"groupName":"$linkedin_company","overAllIndex":1},{"prop_category":"user","eventName":"$linkedin_clicked_ad","property":"$li_preferred_country","prop_type":"categorical","eventIndex":1,"groupName":"$linkedin_company","overAllIndex":2},{"prop_category":"event","eventName":"$linkedin_clicked_ad","property":"$campaign","prop_type":"categorical","eventIndex":1,"groupName":"Traffic source","overAllIndex":3},{"prop_category":"event","eventName":"$linkedin_clicked_ad","property":"$campaign_id","prop_type":"categorical","eventIndex":1,"groupName":"Traffic source","overAllIndex":4},{"prop_category":"user","eventName":"$linkedin_clicked_ad","property":"$li_total_ad_click_count","prop_type":"numerical","eventIndex":1,"groupName":"$linkedin_company","gbty":"raw_values","overAllIndex":5}]},
    icon: 'SponsorShip',
    backgroundColor: '#F9F0FF',
    color: '#722ED1',
    question: 'Which of the campaigns do you want to get alerted about?',
    prepopulate: {
      'website_sdk': {
        event: {
          group: 'Linkedin Company Engagements',
          label: '$linkedin_clicked_ad'
        },
        filterBy: [
          {
            operator: 'is known',
            props: ['user', '$campaign', 'categorical', 'user'],
            values: ['$none'],
            ref: 1
          }
        ]
      }
    }
  },
  {
    id: 10,
    title: 'A company is researching about you on G2',
    description: 'Know which companies are researching about you on G2.',
    categories: ['Marketing'],
    required_integrations: [['g2', 'website_sdk']],
    alert_name: 'A company research our product on G2',
    alert_message: 'This company researched about us on G2',
    questions: [],
    payload_props: {
      'g2,website_sdk': [{"prop_category":"event","eventName":"$g2_all","property":"$page_url","prop_type":"categorical","eventIndex":1,"overAllIndex":0,"groupName":"Page properties"},{"prop_category":"event","eventName":"$g2_all","property":"$page_title","prop_type":"categorical","eventIndex":1,"groupName":"Page properties","overAllIndex":1},{"prop_category":"user","eventName":"$g2_all","property":"$g2_domain","prop_type":"categorical","eventIndex":1,"groupName":"$g2","overAllIndex":2},{"prop_category":"user","eventName":"$g2_all","property":"$g2_name","prop_type":"categorical","eventIndex":1,"groupName":"$g2","overAllIndex":3},{"prop_category":"user","eventName":"$g2_all","property":"$g2_employees_range","prop_type":"categorical","eventIndex":1,"groupName":"$g2","overAllIndex":4},{"prop_category":"event","eventName":"$g2_all","property":"$g2_visitor_state","prop_type":"categorical","eventIndex":1,"groupName":"G2 Properties","overAllIndex":5},{"prop_category":"event","eventName":"$g2_all","property":"$g2_visitor_country","prop_type":"categorical","eventIndex":1,"groupName":"G2 Properties","overAllIndex":6}]
    },
    icon: 'SponsorShip',
    backgroundColor: '#F9F0FF',
    color: '#722ED1',
    question: 'What G2 activity do you want to get alerted about?',
    prepopulate: {
      'g2,website_sdk': {
        event: {
          group: 'G2 Engagements',
          label: '$g2_all'
        },
        filterBy: []
      }
    }
  },
  {
    id: 11,
    title: 'One of your customers is checking out your pricing page',
    description:
      'Get notified when a customer checks out your pricing page and take action for a potential upsell.',
    categories: ['Customer Success'],
    required_integrations: [
      ['hubspot', 'website_sdk'],
      ['salesforce', 'website_sdk']
    ],
    alert_name: 'A customer just checked out our pricing page',
    alert_message:
      'This customer just checked out our pricing page, looks like a potential upsell',
    questions: [],
    payload_props: {
      'hubspot,website_sdk': [
        {"prop_category":"user","eventName":"$page_view","property":"$hubspot_company_name","prop_type":"categorical","eventIndex":1,"overAllIndex":0},
        {"prop_category":"user","eventName":"$page_view","property":"$hubspot_company_hubspot_owner_id","prop_type":"categorical","eventIndex":1,"groupName":"$hubspot_company","overAllIndex":1},
        {"prop_category":"user","eventName":"$page_view","property":"$hubspot_deal_amount","prop_type":"numerical","eventIndex":1,"groupName":"$hubspot_deal","gbty":"raw_values","overAllIndex":2},
        {"prop_category":"user","eventName":"$page_view","property":"$email","prop_type":"categorical","eventIndex":1,"groupName":"User identification","overAllIndex":3},{"prop_category":"event","eventName":"$page_view","property":"$page_url","prop_type":"categorical","eventIndex":1,"groupName":"Page properties","overAllIndex":4},{"prop_category":"event","eventName":"$page_view","property":"$page_title","prop_type":"categorical","eventIndex":1,"groupName":"Page properties","overAllIndex":5},{"prop_category":"event","eventName":"$page_view","property":"$page_spent_time","prop_type":"numerical","eventIndex":1,"groupName":"Page properties","gbty":"raw_values","overAllIndex":6}],
      'salesforce,website_sdk': [{
        prop_category: 'user',
        eventName: '$session',
        property: '$salesforce_account_name',
        prop_type: 'categorical',
        eventIndex: 1,
        overAllIndex: 0
      },
      {
        prop_category: 'user',
        eventName: '$session',
        property: '$salesforce_account_ownerid',
        prop_type: 'numerical',
        eventIndex: 1,
        groupName: '$hubspot_deal',
        gbty: 'raw_values',
        overAllIndex: 1
      },
      {
        prop_category: 'user',
        eventName: '$session',
        property: '$salesforce_opportunity_amount',
        prop_type: 'categorical',
        eventIndex: 1,
        groupName: '$hubspot_deal',
        overAllIndex: 2
      },{"prop_category":"user","eventName":"$page_view","property":"$email","prop_type":"categorical","eventIndex":1,"groupName":"User identification","overAllIndex":3},{"prop_category":"event","eventName":"$page_view","property":"$page_url","prop_type":"categorical","eventIndex":1,"groupName":"Page properties","overAllIndex":4},{"prop_category":"event","eventName":"$page_view","property":"$page_title","prop_type":"categorical","eventIndex":1,"groupName":"Page properties","overAllIndex":5},{"prop_category":"event","eventName":"$page_view","property":"$page_spent_time","prop_type":"numerical","eventIndex":1,"groupName":"Page properties","gbty":"raw_values","overAllIndex":6}]
    },
    icon: 'Handshake',
    backgroundColor: '#F0F5FF',
    color: '#2F54EB',
    question:
      'What is your pricing page URL? How do you define a customer in your CRM?',
    prepopulate: {
      'hubspot,website_sdk': {
        event: { label: '$page_view', group: 'Website activity' },
        filterBy: [
          {
            operator: 'contains',
            props: ['event', '$page_url', 'categorical', 'event'],
            values: ['/pricing'],
            ref: 1
          },
          {
            operator: 'equals',
            props: [
              'user',
              '$hubspot_company_lifecyclestage',
              'categorical',
              'user'
            ],
            values: ['Customer'],
            ref: 2
          }
        ]
      },
      'salesforce,website_sdk': {
        event: { label: '$page_view', group: 'Website activity' },
        filterBy: [
          {
            operator: 'contains',
            props: ['event', '$page_url', 'categorical', 'event'],
            values: ['/pricing'],
            ref: 1
          },
          {
            operator: 'equals',
            props: [
              'user',
              '$salesforce_opportunity_stagename',
              'categorical',
              'user'
            ],
            values: ['Closed won'],
            ref: 2
          }
        ]
      }
    }
  },
  {
    id: 12,
    title:
      'One of your customers just checked out your help docsOne of your customers just checked out your help docs',
    description:
      'Get notified when a customer checks out your pricing page and offer right help exactly when they need it.',
    categories: ['Customer Success'],
    required_integrations: [
      ['hubspot', 'website_sdk'],
      ['salesforce', 'website_sdk']
    ],
    alert_name: 'A customer just checked out help docs',
    alert_message:
      'This customer just checked out a help doc, they might need help',
    questions: [],
    payload_props: {
      'hubspot,website_sdk': [{"prop_category":"user","eventName":"$page_view","property":"$6Signal_domain","prop_type":"categorical","eventIndex":1,"overAllIndex":0},{"prop_category":"user","eventName":"$page_view","property":"$6Signal_name","prop_type":"categorical","eventIndex":1,"groupName":"$6signal","overAllIndex":1},{"prop_category":"user","eventName":"$page_view","property":"$email","prop_type":"categorical","eventIndex":1,"groupName":"User identification","overAllIndex":2},{"prop_category":"event","eventName":"$page_view","property":"$page_url","prop_type":"categorical","eventIndex":1,"groupName":"Page properties","overAllIndex":3},{"prop_category":"event","eventName":"$page_view","property":"$page_title","prop_type":"categorical","eventIndex":1,"groupName":"Page properties","overAllIndex":4},{"prop_category":"event","eventName":"$page_view","property":"$page_spent_time","prop_type":"numerical","eventIndex":1,"groupName":"Page properties","gbty":"raw_values","overAllIndex":5}],
      'salesforce,website_sdk': [{"prop_category":"user","eventName":"$page_view","property":"$6Signal_domain","prop_type":"categorical","eventIndex":1,"overAllIndex":0},{"prop_category":"user","eventName":"$page_view","property":"$6Signal_name","prop_type":"categorical","eventIndex":1,"groupName":"$6signal","overAllIndex":1},{"prop_category":"user","eventName":"$page_view","property":"$email","prop_type":"categorical","eventIndex":1,"groupName":"User identification","overAllIndex":2},{"prop_category":"event","eventName":"$page_view","property":"$page_url","prop_type":"categorical","eventIndex":1,"groupName":"Page properties","overAllIndex":3},{"prop_category":"event","eventName":"$page_view","property":"$page_title","prop_type":"categorical","eventIndex":1,"groupName":"Page properties","overAllIndex":4},{"prop_category":"event","eventName":"$page_view","property":"$page_spent_time","prop_type":"numerical","eventIndex":1,"groupName":"Page properties","gbty":"raw_values","overAllIndex":5}]
    },
    icon: 'Handshake',
    backgroundColor: '#F0F5FF',
    color: '#2F54EB',
    question:
      'What is URL of your help doc pages? How do you define a customer in your CRM?',
    prepopulate: {
      'hubspot,website_sdk': {
        event: { label: '$page_view', group: 'Website activity' },
        filterBy: [
          {
            operator: 'contains',
            props: ['event', '$page_url', 'categorical', 'event'],
            values: ['help'],
            ref: 1
          },
          {
            operator: 'equals',
            props: [
              'user',
              '$hubspot_company_lifecyclestage',
              'categorical',
              'user'
            ],
            values: ['Customer'],
            ref: 2
          }
        ]
      },
      'salesforce,website_sdk': {
        event: { label: '$page_view', group: 'Website activity' },
        filterBy: [
          {
            operator: 'contains',
            props: ['event', '$page_url', 'categorical', 'event'],
            values: ['help'],
            ref: 1
          },
          {
            operator: 'equals',
            props: [
              'user',
              '$salesforce_opportunity_stagename',
              'categorical',
              'user'
            ],
            values: ['Closed won'],
            ref: 2
          }
        ]
      }
    }
  }
];
