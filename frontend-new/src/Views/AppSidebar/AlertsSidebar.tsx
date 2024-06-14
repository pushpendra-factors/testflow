import React, { useEffect, useState } from 'react';
import cx from 'classnames';
import { SVG, Text } from 'Components/factorsComponents';
import styles from './index.module.scss';
import { isAlertsUrl } from './appSidebar.helpers';
import { useHistory, useLocation } from 'react-router-dom';
import { PathUrls } from 'Routes/pathUrls';
import useQuery from 'hooks/useQuery'; 
import useFeatureLock from 'hooks/useFeatureLock';
import { FEATURES } from 'Constants/plans.constants';

const AlertsSidebar = () => {
  const history = useHistory();
  const routeQuery = useQuery();

  const location = useLocation();
  const { pathname } = location;

  const [alertType, setAlertType] = useState('realtime');

  const featureFlag = useFeatureLock(FEATURES.FEATURE_WORKFLOWS);

  const featureLocked =
    featureFlag.isLoading === false && featureFlag.isFeatureLocked === true;


  console.log("isWorkflowsFeatureLocked===>>",featureLocked)
  
  useEffect(() => {
    const type = routeQuery.get('type');
    if (type && ['realtime', 'weekly'].includes(type)) {
      setAlertType(type);
    } else {
      setAlertType('workflows');
    }
  }, [routeQuery]);
  return (
    <div className='flex flex-col gap-y-1 px-2'>
      <div
        role='button'
        onClick={() => {
          history.replace(PathUrls.Alerts + '?type=realtime');
        }}
        className={cx(
          'cursor-pointer rounded-md p-2 flex justify-between gap-x-2 items-center',
          styles['draft-title'],
          {
            [styles['item-active']]:
              isAlertsUrl(pathname) && alertType === 'realtime'
          }
        )}
      >
        <div className={cx('flex gap-x-1 items-center w-full')}>
          {/* <SVG name='settings' /> */}
          <Text
            color={
              alertType === 'realtime' ? 'brand-color-6' : 'character-primary'
            }
            type='title'
            level={7}
            weight={alertType === 'realtime' && 'bold'}
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
        className={cx(
          'cursor-pointer rounded-md p-2 flex justify-between gap-x-2 items-center',
          styles['draft-title'],
          {
            [styles['item-active']]:
              isAlertsUrl(pathname) && alertType === 'weekly'
          }
        )}
      >
        <div className={cx('flex gap-x-1 items-center w-full')}>
          {/* <SVG name='settings' /> */}
          <Text
            color={
              alertType === 'weekly' ? 'brand-color-6' : 'character-primary'
            }
            type='title'
            level={7}
            weight={alertType === 'weekly' && 'bold'}
            extraClass='mb-0'
          >
            Weekly updates
          </Text>
        </div>
      </div>

      {!featureLocked && (
        <div
          role='button'
          onClick={() => {
            history.replace(PathUrls.Workflows);
            setAlertType('workflows');
          }}
          className={cx(
            'cursor-pointer rounded-md p-2 flex justify-between gap-x-2 items-center',
            styles['draft-title'],
            {
              [styles['item-active']]:
                isAlertsUrl(pathname) && alertType === 'workflows'
            }
          )}
        >
          <div className={cx('flex gap-x-1 items-center w-full')}>
            {/* <SVG name='settings' /> */}
            <Text
              color={
                alertType === 'workflows'
                  ? 'brand-color-6'
                  : 'character-primary'
              }
              type='title'
              level={7}
              weight={alertType === 'workflows' && 'bold'}
              extraClass='mb-0'
            >
              Workflows
            </Text>
          </div>
        </div>
      )}
    </div>
  );
};

export default AlertsSidebar;
