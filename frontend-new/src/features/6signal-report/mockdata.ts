import { ReportApiResponseData } from './types';

export const mockData: ReportApiResponseData = {
  result_group: [
    {
      headers: [
        'Company',
        'Country',
        'Page_Seen',
        'Campaign',
        'Time_Spent',
        'Page_Count',
        'Channel'
      ],
      rows: [
        [
          'Tata Consultancy Services Limited',
          'India',
          'app.factors.ai/#/login',
          '$none',
          '6483.05',
          '56',
          'Direct'
        ],
        [
          'Capital First Limited',
          'India',
          'www.factors.ai/compare-competitors/tapclicks',
          '$none',
          '3.79',
          '1',
          'Organic Search'
        ],
        [
          'Lucidpress',
          'United States',
          'www.factors.ai/lp/clearbit',
          'TD-Search-US-Deanon-Competition',
          '74.19',
          '6',
          'Paid Search'
        ]
      ],
      query: null
    }
  ],
  query: {
    six_signal_query_group: [
      {
        fr: 123,
        to: 123,
        tz: ''
      }
    ]
  }
};
