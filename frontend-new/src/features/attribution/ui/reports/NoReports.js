import React from 'react';
import { Link, useHistory } from 'react-router-dom';
import { Button } from 'antd';
import { SVG, Text } from 'Components/factorsComponents';
import styles from './index.module.scss';
import {
  ATTRIBUTION_BASICS_LINK,
  ATTRIBUTION_ROUTES
} from 'Attribution/utils/constants';

function NoReports() {
  const history = useHistory();
  const newLocal = 'black';
  return (
    <div className={`flex flex-col items-center ${styles.contentBody}`}>
      <div className='flex w-full justify-between items-center px-8'>
        <Text
          type='title'
          level={6}
          weight='bold'
          color={newLocal}
          extraClass='m-0'
        >
          Attribution Reports
        </Text>
        <Button
          type='primary'
          size='large'
          onClick={() => history.push(ATTRIBUTION_ROUTES.report)}
        >
          <SVG name='plus' color='white' className='w-full' /> Add Report
        </Button>
      </div>
      <div className='flex flex-col justify-center items-center w-2/4 gap-4'>
        <div className='mb-2'>
          <SVG name='attributionReportsBackground' height='190' width='250' />
        </div>
        <div className='flex flex-col items-center gap-1'>
          <Text
            type='title'
            level={6}
            weight='bold'
            color='black'
            extraClass='m-0'
          >
            Lets get started with attribution
          </Text>
          <div className='flex gap-2'>
            <Text
              type='title'
              level={7}
              weight='medium'
              color='grey'
              extraClass='m-0 text-justify'
            >
              Learn and explore more about
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
        <Button
          type='default'
          size='large'
          onClick={() => history.push('/attribution/report')}
        >
          Create an Attribution Report
        </Button>
      </div>
    </div>
  );
}

export default NoReports;
