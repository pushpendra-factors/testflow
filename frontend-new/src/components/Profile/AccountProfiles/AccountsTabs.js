import React, { useCallback } from 'react';
import { Button, Dropdown, Menu, Radio } from 'antd';
import { useDispatch, useSelector } from 'react-redux';
import { selectActiveTab } from 'Reducers/accountProfilesView/selectors';
import { toggleAccountsTab } from 'Reducers/accountProfilesView/actions';
import { SVG as Svg, Text } from 'Components/factorsComponents';

export default function AccountsTabs() {
  const dispatch = useDispatch();
  const activeTab = useSelector((state) => selectActiveTab(state));

  const handleTabChange = useCallback((e) => {
    dispatch(toggleAccountsTab(e.target.value));
  }, []);

  const renderAutomationMenu = () => {
    const items = [
      { label: 'Set Alert', icon: 'alarmPlus' },
      { label: 'Set Workflow', icon: 'path' }
    ];
    return (
      <Menu items={items}>
        {items.map((item) => (
          <Menu.Item key={item.label}>
            <div className='inline-flex gap-x-2'>
              <Svg name={item.icon} />
              {item.label}
            </div>
          </Menu.Item>
        ))}
      </Menu>
    );
  };

  return (
    <>
      <Radio.Group onChange={handleTabChange} value={activeTab}>
        <Radio.Button value='accounts'>
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
        <Radio.Button value='insights'>
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
      <Dropdown overlay={renderAutomationMenu()}>
        <Button className='button-shadow'>
          <Svg name='AlarmPlus' /> Automation
        </Button>
      </Dropdown>
    </>
  );
}
