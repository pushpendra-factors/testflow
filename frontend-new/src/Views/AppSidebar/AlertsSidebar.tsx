import React, { memo, useCallback, useEffect, useState } from 'react';
import cx from 'classnames';
import { Button } from 'antd';
import { SVG, Text } from 'Components/factorsComponents';
import styles from './index.module.scss';
import { isAlertsUrl } from './appSidebar.helpers';
import { useHistory } from 'react-router-dom';
import { PathUrls } from 'Routes/pathUrls';

const AlertsSidebar = () => {
  const history = useHistory();
  return (
    <div className='flex flex-col row-gap-2'>
      <div
        role='button'
        onClick={() => {
          history.replace(PathUrls.Alerts);
        }}
        className='px-4 w-full'
      >
        <div
          className={cx(
            'flex col-gap-1 cursor-pointer py-2 rounded-md items-center w-full px-2',
            styles['draft-title'],
            {
              [styles['item-active']]: true
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
            Alerts
          </Text>
        </div>
      </div>
    </div>
  );
};

export default AlertsSidebar;
