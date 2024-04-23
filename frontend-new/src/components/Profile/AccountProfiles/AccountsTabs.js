import React, { useCallback } from 'react';
import { Tabs } from 'antd';
import { useDispatch, useSelector } from 'react-redux';
import { selectActiveTab } from 'Reducers/accountProfilesView/selectors';
import { toggleAccountsTab } from 'Reducers/accountProfilesView/actions';
import { Text } from 'Components/factorsComponents';
import styles from './index.module.scss';

export default function AccountsTabs() {
  const dispatch = useDispatch();
  const activeTab = useSelector((state) => selectActiveTab(state));

  const handleTabChange = useCallback((newActiveTab) => {
    dispatch(toggleAccountsTab(newActiveTab));
  }, []);

  return (
    <Tabs
      className={styles.accountsTabs}
      activeKey={activeTab}
      type='card'
      size='small'
      onChange={handleTabChange}
    >
      <Tabs.TabPane
        tab={
          <Text
            level={7}
            color={
              activeTab === 'accounts' ? 'brand-color-6' : 'neutral-gray-8'
            }
            type='title'
            extraClass='mb-0'
          >
            Accounts
          </Text>
        }
        key='accounts'
      />
      <Tabs.TabPane
        tab={
          <Text
            level={7}
            mini
            color={
              activeTab === 'insights' ? 'brand-color-6' : 'neutral-gray-8'
            }
            type='title'
            extraClass='mb-0'
          >
            Insights
          </Text>
        }
        key='insights'
      />
    </Tabs>
  );
}
