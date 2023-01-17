import { render, screen, fireEvent } from '@testing-library/react';
import React from 'react';
import SearchBar from './index';
import * as reactRedux from 'react-redux';

jest.mock('react-redux', () => ({
  useSelector: jest.fn()
}));

describe('Testing SearchBar Component', () => {
  beforeEach(() => {
    useSelectorMock.mockImplementation((selector) => selector(mockStore));
  });
  afterEach(() => {
    useSelectorMock.mockClear();
  });

  const useSelectorMock = reactRedux.useSelector;

  const mockStore = {
    queries: {
      loading: false,
      error: false,
      data: [
        {
          id: '27000049',
          project_id: 51,
          title: 'Custom KPI',
          query: {
            cl: 'kpi',
            gFil: [],
            gGBy: [
              {
                dpNa: 'Initial Referrer URL',
                en: 'user',
                gr: '',
                objTy: '',
                prDaTy: 'categorical',
                prNa: '$initial_referrer_url'
              }
            ],
            qG: [
              {
                ca: 'profiles',
                dc: 'hubspot_contacts',
                fil: [],
                fr: 1673096400,
                gBy: [],
                gbt: '',
                me: ['testcustomkpi'],
                pgUrl: '',
                qt: 'custom',
                to: 1673528399,
                tz: 'Australia/Sydney'
              },
              {
                ca: 'profiles',
                dc: 'hubspot_contacts',
                fil: [],
                fr: 1673096400,
                gBy: [],
                gbt: 'date',
                me: ['testcustomkpi'],
                pgUrl: '',
                qt: 'custom',
                to: 1673528399,
                tz: 'Australia/Sydney'
              },
              {
                ca: 'profiles',
                dc: 'hubspot_contacts',
                fil: [],
                fr: 1673096400,
                gBy: [],
                gbt: '',
                me: ['Default KPI with >2 KPIs'],
                pgUrl: '',
                qt: 'custom',
                to: 1673528399,
                tz: 'Australia/Sydney'
              },
              {
                ca: 'profiles',
                dc: 'hubspot_contacts',
                fil: [],
                fr: 1673096400,
                gBy: [],
                gbt: 'date',
                me: ['Default KPI with >2 KPIs'],
                pgUrl: '',
                qt: 'custom',
                to: 1673528399,
                tz: 'Australia/Sydney'
              }
            ]
          },
          type: 2,
          is_deleted: false,
          created_by: '8b629994-e660-4365-9154-1367653ecdef',
          created_by_name: ' ',
          created_by_email: 'solutions@factors.ai',
          created_at: '2023-01-13T11:22:53.31117Z',
          updated_at: '2023-01-13T11:22:53.31117Z',
          settings: {
            chart: 'pb',
            dateSorter: [
              {
                key: 'testcustomkpi - 0',
                order: 'descend',
                subtype: null,
                type: 'numerical'
              }
            ],
            pivotConfig:
              '{"rows":[],"cols":[],"vals":[],"aggregatorName":"Integer Sum","rowOrder":"value_a_to_z","configLoaded":false}',
            sorter: [
              {
                key: 'testcustomkpi - 0',
                order: 'descend',
                subtype: null,
                type: 'numerical'
              }
            ]
          },
          id_text:
            'rE5g8LRzRy8BSCn420Ch2oudiUbsWpYLTaDEsyF2sX1673608973IyY5jeIK',
          Converted: false,
          is_dashboard_query: false
        },
        {
          id: '27000048',
          project_id: 51,
          title: 'Saved for kartheek',
          query: {
            query_group: [
              {
                cl: 'events',
                ec: 'each_given_event',
                ewp: [
                  {
                    an: '',
                    grpa: 'Most Recent',
                    na: '$session',
                    pr: []
                  },
                  {
                    an: '',
                    grpa: 'Most Recent',
                    na: 'www.acme.com/pricing',
                    pr: []
                  },
                  {
                    an: '',
                    grpa: 'Most Recent',
                    na: 'Schedule A Demo Form',
                    pr: []
                  }
                ],
                fr: 1673096400,
                gbp: [
                  {
                    en: 'event',
                    ena: '$session',
                    eni: 1,
                    pr: '$channel',
                    pty: 'categorical'
                  },
                  {
                    en: 'event',
                    ena: 'www.acme.com/pricing',
                    eni: 2,
                    pr: 'page_url',
                    pty: 'categorical'
                  },
                  {
                    en: 'event',
                    ena: 'Schedule A Demo Form',
                    eni: 3,
                    pr: 'Source-Medium',
                    pty: 'categorical'
                  },
                  {
                    en: 'user',
                    ena: '$present',
                    pr: '$initial_campaign',
                    pty: 'categorical'
                  }
                ],
                gbt: 'date',
                gup: [],
                to: 1673528399,
                ty: 'unique_users',
                tz: 'Australia/Sydney'
              },
              {
                cl: 'events',
                ec: 'each_given_event',
                ewp: [
                  {
                    an: '',
                    grpa: 'Most Recent',
                    na: '$session',
                    pr: []
                  },
                  {
                    an: '',
                    grpa: 'Most Recent',
                    na: 'www.acme.com/pricing',
                    pr: []
                  },
                  {
                    an: '',
                    grpa: 'Most Recent',
                    na: 'Schedule A Demo Form',
                    pr: []
                  }
                ],
                fr: 1673096400,
                gbp: [
                  {
                    en: 'event',
                    ena: '$session',
                    eni: 1,
                    pr: '$channel',
                    pty: 'categorical'
                  },
                  {
                    en: 'event',
                    ena: 'www.acme.com/pricing',
                    eni: 2,
                    pr: 'page_url',
                    pty: 'categorical'
                  },
                  {
                    en: 'event',
                    ena: 'Schedule A Demo Form',
                    eni: 3,
                    pr: 'Source-Medium',
                    pty: 'categorical'
                  },
                  {
                    en: 'user',
                    ena: '$present',
                    pr: '$initial_campaign',
                    pty: 'categorical'
                  }
                ],
                gbt: '',
                gup: [],
                to: 1673528399,
                ty: 'unique_users',
                tz: 'Australia/Sydney'
              }
            ]
          },
          type: 2,
          is_deleted: false,
          created_by: '8b629994-e660-4365-9154-1367653ecdef',
          created_by_name: ' ',
          created_by_email: 'solutions@factors.ai',
          created_at: '2023-01-13T09:11:08.41192Z',
          updated_at: '2023-01-13T09:11:08.41192Z',
          settings: {
            chart: 'pb',
            pivotConfig:
              '{"rows":[],"cols":[],"vals":[],"aggregatorName":"Integer Sum","rowOrder":"value_a_to_z","configLoaded":false}'
          },
          id_text:
            'KwjU7COO33fR8V1673601068MfTFDmscR7mJjKoINl6dQRkr3QxUQksvFbBZ',
          Converted: false,
          is_dashboard_query: false
        },
        {
          id: '27000047',
          project_id: 51,
          title: 'Google Ads: Performance By Campaign, Ad Group, Keyword',
          query: {
            cl: 'kpi',
            gFil: [],
            gGBy: [
              {
                en: '',
                gr: '',
                objTy: 'campaign',
                prDaTy: 'categorical',
                prNa: 'campaign_name'
              },
              {
                en: '',
                gr: '',
                objTy: 'ad_group',
                prDaTy: 'categorical',
                prNa: 'ad_group_name'
              },
              {
                en: '',
                gr: '',
                objTy: 'keyword',
                prDaTy: 'categorical',
                prNa: 'keyword_name'
              }
            ],
            qG: [
              {
                ca: 'channels',
                dc: 'google_ads_metrics',
                fil: [],
                fr: 1660415400,
                gBy: [],
                gbt: '',
                me: ['impressions'],
                pgUrl: '',
                qt: 'static',
                to: 1661020199,
                tz: 'Asia/Kolkata'
              },
              {
                ca: 'channels',
                dc: 'google_ads_metrics',
                fil: [],
                fr: 1660415400,
                gBy: [],
                gbt: 'date',
                me: ['impressions'],
                pgUrl: '',
                qt: 'static',
                to: 1661020199,
                tz: 'Asia/Kolkata'
              },
              {
                ca: 'channels',
                dc: 'google_ads_metrics',
                fil: [],
                fr: 1660415400,
                gBy: [],
                gbt: '',
                me: ['clicks'],
                pgUrl: '',
                qt: 'static',
                to: 1661020199,
                tz: 'Asia/Kolkata'
              },
              {
                ca: 'channels',
                dc: 'google_ads_metrics',
                fil: [],
                fr: 1660415400,
                gBy: [],
                gbt: 'date',
                me: ['clicks'],
                pgUrl: '',
                qt: 'static',
                to: 1661020199,
                tz: 'Asia/Kolkata'
              },
              {
                ca: 'channels',
                dc: 'google_ads_metrics',
                fil: [],
                fr: 1660415400,
                gBy: [],
                gbt: '',
                me: ['spend'],
                pgUrl: '',
                qt: 'static',
                to: 1661020199,
                tz: 'Asia/Kolkata'
              },
              {
                ca: 'channels',
                dc: 'google_ads_metrics',
                fil: [],
                fr: 1660415400,
                gBy: [],
                gbt: 'date',
                me: ['spend'],
                pgUrl: '',
                qt: 'static',
                to: 1661020199,
                tz: 'Asia/Kolkata'
              },
              {
                ca: 'channels',
                dc: 'google_ads_metrics',
                fil: [],
                fr: 1660415400,
                gBy: [],
                gbt: '',
                me: ['conversion'],
                pgUrl: '',
                qt: 'static',
                to: 1661020199,
                tz: 'Asia/Kolkata'
              },
              {
                ca: 'channels',
                dc: 'google_ads_metrics',
                fil: [],
                fr: 1660415400,
                gBy: [],
                gbt: 'date',
                me: ['conversion'],
                pgUrl: '',
                qt: 'static',
                to: 1661020199,
                tz: 'Asia/Kolkata'
              },
              {
                ca: 'channels',
                dc: 'google_ads_metrics',
                fil: [],
                fr: 1660415400,
                gBy: [],
                gbt: '',
                me: ['cost_per_conversion'],
                pgUrl: '',
                qt: 'static',
                to: 1661020199,
                tz: 'Asia/Kolkata'
              },
              {
                ca: 'channels',
                dc: 'google_ads_metrics',
                fil: [],
                fr: 1660415400,
                gBy: [],
                gbt: 'date',
                me: ['cost_per_conversion'],
                pgUrl: '',
                qt: 'static',
                to: 1661020199,
                tz: 'Asia/Kolkata'
              },
              {
                ca: 'channels',
                dc: 'google_ads_metrics',
                fil: [],
                fr: 1660415400,
                gBy: [],
                gbt: '',
                me: ['search_impression_share'],
                pgUrl: '',
                qt: 'static',
                to: 1661020199,
                tz: 'Asia/Kolkata'
              },
              {
                ca: 'channels',
                dc: 'google_ads_metrics',
                fil: [],
                fr: 1660415400,
                gBy: [],
                gbt: 'date',
                me: ['search_impression_share'],
                pgUrl: '',
                qt: 'static',
                to: 1661020199,
                tz: 'Asia/Kolkata'
              }
            ]
          },
          type: 2,
          is_deleted: false,
          created_by: '4e0a4c43-a921-4f5c-aabc-a93de27e96ef',
          created_by_name: 'Pushpendra Vishwakarma',
          created_by_email: 'pushpendra@factors.ai',
          created_at: '2023-01-13T08:32:54.667478Z',
          updated_at: '2023-01-13T08:32:54.667478Z',
          settings: {},
          id_text:
            'mry2hKvPYJ2Ox4123Sa5vhuZciFGAwlnBDpq3Ta1673598774rEwuuoFT4fT',
          Converted: false,
          is_dashboard_query: true
        }
      ]
    }
  };

  it('Should render search bar', () => {
    render(<SearchBar />);
    const inputElement = screen.getByPlaceholderText(/Search Reports/i);
    expect(inputElement).toBeInTheDocument();
  });

  it('Should render search bar with data when clicked on search bar', async () => {
    render(<SearchBar />);
    const inputElement = screen.getByPlaceholderText(/Search Reports/i);
    inputElement.focus();
    const searchModalElement = await screen.findByTestId('search-modal');
    expect(searchModalElement).toBeInTheDocument();
  });

  it('Should render search bar with autofocus', async () => {
    render(<SearchBar />);
    const inputElement = screen.getByPlaceholderText(/Search Reports/i);
    inputElement.focus();
    const searchModalElement = await screen.findByTestId('search-modal-input');
    expect(searchModalElement).toHaveFocus();
  });

  it('Should render 1 query search input', async () => {
    render(<SearchBar />);
    const inputElement = screen.getByPlaceholderText(/Search Reports/i);
    inputElement.focus();
    const searchModalElement = screen.getByTestId('search-modal-input');
    expect(searchModalElement).toBeInTheDocument();
    fireEvent.change(searchModalElement, { target: { value: 'Google' } });
    const searchElements = screen.getAllByTestId(/search-element-/i);
    expect(searchElements.length).toBe(1);
  });

  it('Should render no query search input does not match anything', async () => {
    render(<SearchBar />);
    const inputElement = screen.getByPlaceholderText(/Search Reports/i);
    inputElement.focus();
    const searchModalElement = screen.getByTestId('search-modal-input');
    fireEvent.change(searchModalElement, { target: { value: 'tttesstiing' } });
    const searchElements = screen.queryAllByTestId(/search-element-/i);
    expect(searchElements.length).toBe(0);
    const noDataElement = await screen.findByTestId('no-data');
    expect(noDataElement).toBeInTheDocument();
  });
});
