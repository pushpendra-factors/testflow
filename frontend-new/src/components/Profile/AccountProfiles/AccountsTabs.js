import React, { useCallback } from 'react';
import { Radio } from 'antd';
import { useDispatch, useSelector } from 'react-redux';
import { selectActiveTab } from 'Reducers/accountProfilesView/selectors';
import { toggleAccountsTab } from 'Reducers/accountProfilesView/actions';
import { SVG as Svg, Text } from 'Components/factorsComponents';
import styles from './index.module.scss';

export default function AccountsTabs() {
  const dispatch = useDispatch();
  const activeTab = useSelector((state) => selectActiveTab(state));

  const handleTabChange = useCallback((e) => {
    dispatch(toggleAccountsTab(e.target.value));
  }, []);

  return (
    <Radio.Group onChange={handleTabChange} value={activeTab}>
      <Radio.Button className={styles['left-tab-button']} value='accounts'>
        <div className='flex gap-x-1 justify-center items-center h-full'>
          <Svg
            size={16}
            name='listUl'
            color={activeTab === 'accounts' ? '#1890ff' : '#000'}
          />
          <Text
            level={7}
            color={activeTab === 'accounts' ? 'brand-color-6' : 'black'}
            type='title'
            extraClass='mb-0'
          >
            List
          </Text>
        </div>
      </Radio.Button>
      <Radio.Button className={styles['right-tab-button']} value='insights'>
        <div className='flex gap-x-1 justify-center items-center h-full'>
          <Svg
            size={16}
            name='eye'
            color={activeTab === 'insights' ? '#1890ff' : '#000'}
          />
          <Text
            level={7}
            color={activeTab === 'insights' ? 'brand-color-6' : 'black'}
            type='title'
            extraClass='mb-0'
          >
            Insights
          </Text>
        </div>
      </Radio.Button>
    </Radio.Group>
  );
}
