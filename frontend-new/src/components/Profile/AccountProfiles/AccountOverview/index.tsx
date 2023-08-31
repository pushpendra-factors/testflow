import { Spin } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import { Text } from 'Components/factorsComponents';
import React from 'react';
import { formatDuration } from 'Utils/dataFormatter';
import TableWithHeading from './TableWithHeading';
import TrendsChart from './TrendsChart';
import {
  AccountOverviewProps,
  CustomStyles,
  TopPages,
  TopUsers
} from './types';
import { EngagementTag } from './utils';

const topPageColumns: ColumnsType<TopPages> = [
  {
    title: 'Page URL',
    dataIndex: 'page_url',
    key: 'page_url',
    ellipsis: true,
    width: 224,
    render: (text: string) => (
      <a href={text} target='_blank' rel='noopener noreferrer'>
        {text}
      </a>
    )
  },
  {
    title: '# Views',
    dataIndex: 'views',
    align: 'right',
    width: 96,
    key: 'views'
  },
  {
    title: '# Users',
    dataIndex: 'users_count',
    align: 'right',
    width: 96,
    key: 'users_count'
  },
  {
    title: 'Total Time',
    dataIndex: 'total_time',
    key: 'total_time',
    align: 'right',
    width: 96,
    render: (time: number) => formatDuration(time.toFixed())
  },
  {
    title: 'Avg. Scroll %',
    dataIndex: 'avg_scroll_percent',
    key: 'avg_scroll_percent',
    width: 112,
    align: 'right',
    render: (percent: number) => `${percent?.toFixed(2)}%`
  }
];

const topUserColumns: ColumnsType<TopUsers> = [
  {
    title: 'Name',
    dataIndex: 'name',
    key: 'name',
    width: 264
  },
  {
    title: '# Views',
    align: 'right',
    width: 120,
    dataIndex: 'num_page_views',
    key: 'num_page_views'
  },
  {
    title: 'Active Time',
    align: 'right',
    width: 120,
    dataIndex: 'active_time',
    key: 'active_time',
    render: (time: number) => formatDuration(time.toFixed())
  },
  {
    title: '# Pages',
    align: 'right',
    width: 120,
    dataIndex: 'num_of_pages',
    key: 'num_of_pages'
  }
];

const AccountOverview: React.FC<AccountOverviewProps> = ({
  overview,
  loading
}) => {
  const styles: CustomStyles = {
    '--bg-color': EngagementTag[overview?.engagement]?.bgColor || '#FFF1F0'
  };

  return loading ? (
    <Spin size='large' className='fa-page-loader' />
  ) : (
    <div className='overview'>
      <div className='top-metrics'>
        <div className='metric'>
          <Text type='title' level={7} extraClass='m-0' color='grey'>
            Condition
          </Text>
          {overview?.engagement ? (
            <div
              className='engagement-tag'
              style={styles as React.CSSProperties}
            >
              <img
                src={`../../../assets/icons/${
                  EngagementTag[overview?.engagement]?.icon || 'fire'
                }.svg`}
                alt=''
              />
              <Text type='title' level={7} extraClass='m-0'>
                {overview?.engagement}
              </Text>
            </div>
          ) : (
            <Text
              type='title'
              level={4}
              extraClass='m-0'
              color='red'
              weight='bold'
            >
              NA
            </Text>
          )}
        </div>
        <div className='metric'>
          <Text
            type='title'
            level={7}
            extraClass='m-0 whitespace-no-wrap'
            color='grey'
          >
            Engagement Score
          </Text>
          <Text
            type='title'
            level={4}
            extraClass='m-0'
            color='red'
            weight='bold'
          >
            {overview?.temperature
              ? parseInt(overview?.temperature?.toFixed())
              : 'NA'}
          </Text>
        </div>
        <div className='metric'>
          <Text type='title' level={7} extraClass='m-0' color='grey'>
            #Users
          </Text>
          <Text type='title' level={4} extraClass='m-0' weight='bold'>
            {overview?.users_count > 25 ? '25+' : overview?.users_count || 0}
          </Text>
        </div>
        <div className='metric'>
          <Text
            type='title'
            level={7}
            extraClass='m-0 whitespace-no-wrap'
            color='grey'
          >
            Active Time
          </Text>
          <Text
            type='title'
            level={4}
            extraClass='m-0 whitespace-no-wrap'
            weight='bold'
          >
            {formatDuration(parseInt((overview?.time_active || 0).toFixed()))}
          </Text>
        </div>
      </div>
      <div className='trend'>
        <div className='heading'>
          <Text
            type='title'
            level={7}
            extraClass='m-0 whitespace-no-wrap'
            weight='bold'
            color='grey-2'
          >
            Account Signal Trend
          </Text>
        </div>
        <div className='chart'>
          <TrendsChart data={overview.scores_list} />
        </div>
      </div>
      <div className='top-tables'>
        <TableWithHeading
          heading='Top Pages'
          data={overview.top_pages}
          columns={topPageColumns}
          yScroll={200}
        />
        <TableWithHeading
          heading='Top Users'
          data={overview.top_users}
          columns={topUserColumns}
          yScroll={200}
        />
      </div>
    </div>
  );
};

export default AccountOverview;
