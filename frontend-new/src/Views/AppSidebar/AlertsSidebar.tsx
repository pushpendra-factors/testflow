import React, { memo, useCallback, useEffect, useState } from 'react';
import cx from 'classnames';
import { Button } from 'antd';
import { SVG, Text } from 'Components/factorsComponents';
import styles from './index.module.scss';
import { isAlertsUrl } from './appSidebar.helpers';
import { useHistory } from 'react-router-dom';
import { PathUrls } from 'Routes/pathUrls';
import useQuery from 'hooks/useQuery';

const AlertsSidebar = () => {
  const history = useHistory();
  const routeQuery = useQuery();

  const [alertType, setAlertType] = useState('realtime');

  useEffect(() => {
    const type = routeQuery.get('type');
    if (type && ['realtime', 'weekly'].includes(type)) {
      setAlertType(type);
    }
  }, [routeQuery]);
  return (
    <div className='flex flex-col row-gap-2'>
      <div
        role='button'
        onClick={() => {
          history.replace(PathUrls.Alerts + '?type=realtime');
        }}
        className='px-4 w-full'
      >
        <div
          className={cx(
            'flex col-gap-1 cursor-pointer py-2 rounded-md items-center w-full px-2',
            styles['draft-title'],
            {
              [styles['item-active']]: alertType === 'realtime'
            }
          )}
        >
          {/* <SVG name='settings' /> */}
          <Text
            color='character-primary'
            type='title'
            level={7}
            extraClass='mb-0'
          >
            Real time alerts
          </Text>
        </div>
      </div>
      <div
        role='button'
        onClick={() => {
          history.replace(PathUrls.Alerts + '?type=weekly');
        }}
        className='px-4 w-full'
      >
        <div
          className={cx(
            'flex col-gap-1 cursor-pointer py-2 rounded-md items-center w-full px-2',
            styles['draft-title'],
            {
              [styles['item-active']]: alertType === 'weekly'
            }
          )}
        >
          {/* <SVG name='settings' /> */}
          <Text
            color='character-primary'
            type='title'
            level={7}
            extraClass='mb-0'
          >
            Weekly updates
          </Text>
        </div>
      </div>
    </div>
  );
};

export default AlertsSidebar;
