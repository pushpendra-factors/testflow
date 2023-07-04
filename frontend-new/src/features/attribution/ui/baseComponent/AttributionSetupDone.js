import React from 'react';
import { Link } from 'react-router-dom';
import { Button } from 'antd';
import { SVG, Text } from 'Components/factorsComponents';
import styles from './index.module.scss';
import { ATTRIBUTION_BASICS_LINK } from 'Attribution/utils/constants';
import { PathUrls } from '../../../../routes/pathUrls';

function AttributionSetupDone() {
  return (
    <div className={`flex flex-col items-center ${styles.contentBody}`}>
      <div className='flex w-full justify-between items-center px-8'>
        <Text
          type='title'
          level={6}
          weight='bold'
          color='black'
          extraClass='m-0'
        >
          Attribution Reports
        </Text>
        <div className='flex items-center gap-2'>
          <Button
            type='link'
            size='large'
            onClick={() => history.push(PathUrls.SettingsAttribution)}
          >
            Configuration
          </Button>
          <Button type='primary' disabled size='large'>
            <SVG name='plus' color='white' className='w-full' /> Add Report
          </Button>
        </div>
      </div>
      <div className='flex flex-col justify-center items-center w-2/4 gap-4'>
        <div className='mb-2'>
          <SVG name='attributionHomeBackground' height='190' width='250' />
        </div>
        <div className='flex flex-col items-center gap-1'>
          <Text
            type='title'
            level={6}
            weight='bold'
            color='black'
            extraClass='m-0'
          >
            Pre-computing the attribution engine...
          </Text>
          <Text
            type='title'
            level={7}
            weight='medium'
            color='grey'
            extraClass='m-0 text-justify'
          >
            Come back here after a day to create your attribution reporting
          </Text>
        </div>
        <div className='flex flex-col items-center gap-1'>
          <Text
            type='title'
            level={7}
            weight='medium'
            color='grey'
            extraClass='m-0 text-justify'
          >
            Learn more about Multitouch Attribution Reporting
          </Text>

          <Link
            className='flex items-center font-semibold gap-2'
            style={{ color: `#1d89ff` }}
            target='_blank'
            to={{
              pathname: ATTRIBUTION_BASICS_LINK
            }}
          >
            Attribution Basics{' '}
            <SVG size={20} name='Arrowright' color='#1d89ff' />
          </Link>
        </div>
      </div>
    </div>
  );
}

export default AttributionSetupDone;
