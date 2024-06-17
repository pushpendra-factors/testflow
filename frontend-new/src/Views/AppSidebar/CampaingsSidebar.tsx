import React, { useEffect, useState } from 'react';
import cx from 'classnames';
import { SVG, Text } from 'Components/factorsComponents';
import { useHistory, useLocation } from 'react-router-dom';
import { PathUrls } from 'Routes/pathUrls';
import useQuery from 'hooks/useQuery';
import { useSelector } from 'react-redux';
import { featureLock } from 'Routes/feature';
import { isAlertsUrl, isCampaignsUrl } from './appSidebar.helpers';
import styles from './index.module.scss';

const CampaignsSidebar = () => {
  const history = useHistory();
  const routeQuery = useQuery();

  const location = useLocation();
  const { pathname } = location;

  const { agent_details } = useSelector((state: any) => state.agent);

  return (
    <div className='flex flex-col gap-y-1 px-2'>
      <div
        role='button'
        onClick={() => {
          history.replace(PathUrls.FreqCap);
        }}
        className={cx(
          'cursor-pointer rounded-md p-2 flex justify-between gap-x-2 items-center',
          styles['draft-title'],
          {
            [styles['item-active']]: isCampaignsUrl(pathname)
          }
        )}
      >
        <div className={cx('flex gap-x-1 items-center w-full')}>
          {/* <SVG name='settings' /> */}
          <Text
            color='brand-color-6'
            type='title'
            level={7}
            weight='bold'
            extraClass='mb-0'
          >
            Frequency Capping
          </Text>
        </div>
      </div>
    </div>
  );
};

export default CampaignsSidebar;
