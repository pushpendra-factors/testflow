import React from 'react';
import { Link, useHistory } from 'react-router-dom';
import { Button } from 'antd';
import { SVG, Text } from 'Components/factorsComponents';
import styles from './index.module.scss';
import { ABOUT_ATTRIBUTION_LINK } from 'Attribution/utils/constants';

function AttributionSetupPending() {
  const history = useHistory();
  return (
    <div
      className={`flex flex-col justify-center items-center m-auto ${styles.contentBody}`}
    >
      <div className='flex flex-col justify-center w-2/4 gap-4'>
        <div className='mb-2'>
          <SVG name='attributionHomeBackground' height='190' width='250' />
        </div>
        <Text
          type='title'
          level={6}
          weight='bold'
          color='black'
          extraClass='m-0'
        >
          Get Started with Attribution
        </Text>
        <Text
          type='title'
          level={7}
          weight='medium'
          color='grey'
          extraClass='m-0 text-justify'
        >
          Lorem ipsum dolor sit amet, consectetur adipiscing elit. Aenean
          placerat eros id porttitor dignissim. In order to set up conversion
          goals for attribution and attribution window, click on the "Setup now"
          button below.
        </Text>
        <div className='flex gap-8'>
          <Button
            type='primary'
            size='large'
            onClick={() => history.push('/settings/attribution')}
          >
            Setup Now
          </Button>
          <Link
            className='flex items-center font-semibold gap-2'
            style={{ color: `#1d89ff` }}
            target='_blank'
            to={{
              pathname: ABOUT_ATTRIBUTION_LINK
            }}
          >
            Learn More <SVG size={20} name='Arrowright' color='#1d89ff' />
          </Link>
        </div>
      </div>
    </div>
  );
}

export default AttributionSetupPending;
