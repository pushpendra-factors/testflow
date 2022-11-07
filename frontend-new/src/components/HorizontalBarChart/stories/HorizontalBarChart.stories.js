import React from 'react';
import BarChart from '../HorizontalBarChart';

export default {
  title: 'Components/BarChart',
  component: BarChart
};

export function DefaultChart() {
  return (
    <BarChart
      categories={[
        '(Not Set)',
        'Traditional_media',
        'Seasonal_push',
        'Context_marketing',
        'Email_marketing',
        'Brand_awareness',
        'Product_launch',
        'Rebranding_campaign',
        'Brand_launch'
      ]}
      series={[
        {
          name: 'OG',
          data: [
            {
              y: 6907,
              color: '#82AEE0'
            },
            {
              y: 947,
              color: '#7CBAD9'
            },
            {
              y: 767,
              color: '#83D2D2'
            },
            {
              y: 751,
              color: '#86D3A3'
            },
            {
              y: 481,
              color: '#E5DD8C'
            },
            {
              y: 472,
              color: '#F9C06E'
            },
            {
              y: 218,
              color: '#E89E7B'
            },
            {
              y: 200,
              color: '#D4787D'
            },
            {
              y: 61,
              color: '#B87B7E'
            }
          ]
        }
      ]}
    />
  );
}

export function WithComparison() {
  return (
    <BarChart
      categories={[
        '(Not Set)',
        'Traditional_media',
        'Seasonal_push',
        'Context_marketing',
        'Email_marketing',
        'Brand_awareness',
        'Product_launch',
        'Rebranding_campaign',
        'Brand_launch'
      ]}
      comparisonApplied
      series={[
        {
          name: 'OG',
          data: [
            {
              y: 6907,
              color: '#82AEE0'
            },
            {
              y: 947,
              color: '#7CBAD9'
            },
            {
              y: 767,
              color: '#83D2D2'
            },
            {
              y: 751,
              color: '#86D3A3'
            },
            {
              y: 481,
              color: '#E5DD8C'
            },
            {
              y: 472,
              color: '#F9C06E'
            },
            {
              y: 218,
              color: '#E89E7B'
            },
            {
              y: 200,
              color: '#D4787D'
            },
            {
              y: 61,
              color: '#B87B7E'
            }
          ]
        },
        {
          name: 'compare',
          data: [
            {
              y: 5400,
              color: '#82AEE0'
            },
            {
              y: 1850,
              color: '#7CBAD9'
            },
            {
              y: 1777,
              color: '#83D2D2'
            },
            {
              y: 221,
              color: '#86D3A3'
            },
            {
              y: 581,
              color: '#E5DD8C'
            },
            {
              y: 842,
              color: '#F9C06E'
            },
            {
              y: 828,
              color: '#E89E7B'
            },
            {
              y: 350,
              color: '#D4787D'
            },
            {
              y: 251,
              color: '#B87B7E'
            }
          ]
        }
      ]}
    />
  );
}
