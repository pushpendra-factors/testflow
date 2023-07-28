import { Spin } from 'antd';
import { Text } from 'Components/factorsComponents';
import React from 'react';
import { formatDurationIntoString } from 'Utils/dataFormatter';
import TrendsChart from './TrendsChart';
import { EngagementTag } from './utils';

export interface DataMap {
  [key: string]: number;
}

export type Overview = {
  temperature: string;
  engagement: string;
  users_count: number;
  time_active: number;
  scores_list: DataMap;
};

interface AccountOverviewProps {
  overview: Overview;
  loading: boolean;
}

interface CustomStyles {
  '--bg-color': string;
}

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
              level={3}
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
            level={3}
            extraClass='m-0'
            color='red'
            weight='bold'
          >
            {overview?.temperature || 'NA'}
          </Text>
        </div>
        <div className='metric'>
          <Text
            type='title'
            level={7}
            extraClass='m-0 whitespace-no-wrap'
            color='grey'
          >
            #Users
          </Text>
          <Text type='title' level={3} extraClass='m-0' weight='bold'>
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
          <Text type='title' level={3} extraClass='m-0' weight='bold'>
            {formatDurationIntoString(overview?.time_active || 0)}
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
    </div>
  );
};

export default AccountOverview;
