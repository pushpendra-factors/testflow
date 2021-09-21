import React, { memo } from 'react';
import { Dropdown, Button, Menu } from 'antd';
import { SVG, Text } from '../factorsComponents';
import styles from './index.module.scss';

function ChartTypeDropdown({ menuItems, onClick, chartType }) {
  const menu = (
    <Menu className={styles.dropdownMenu}>
      {menuItems.map((item) => {
        return (
          <Menu.Item
            key={item.key}
            onClick={onClick}
            className={`${styles.dropdownMenuItem} ${
              chartType === item.key ? styles.active : ''
            }`}
          >
            <div className={'flex items-center'}>
              <SVG
                extraClass='mr-1'
                name={item.key}
                size={25}
                color={chartType === item.key ? '#8692A3' : '#3E516C'}
              />
              <span className='mr-3'>{item.name}</span>
              {chartType === item.key ? (
                <SVG name='checkmark' size={17} color='#8692A3' />
              ) : null}
            </div>
          </Menu.Item>
        );
      })}
    </Menu>
  );

  const menu1 = (
    <div
      className={`flex shadow-md rounded items-center flex-wrap justify-between bg-white p-4 text-center ${
        styles.dropdownMenu
      } ${menuItems.length < 3 ? styles.smallMenu : ''}`}
    >
      {menuItems.map((item) => {
        return (
          <div
            className={`${
              styles.item
            } flex flex-col items-center justify-center p-2 cursor-pointer ${
              chartType === item.key ? styles.selectedItem : ''
            }`}
            key={item.key}
            onClick={onClick.bind(null, { key: item.key })}
          >
            <SVG
              name={item.key}
              size={25}
              color={chartType === item.key ? '#5949BC' : '#3E516C'}
            />
            <Text extraClass={`mb-0 ${styles.chartName}`} type='title'>
              {item.name}
            </Text>
          </div>
        );
      })}
    </div>
  );

  const activeItem = menuItems.find((item) => item.key === chartType);

  return (
    <Dropdown overlay={menu1}>
      <Button
        size={'large'}
        className={`ant-dropdown-link flex items-center ${styles.dropdownBtn}`}
      >
        {chartType ? (
          <>
            <SVG name={chartType} size={25} color='#0E2647' />{' '}
            {activeItem ? activeItem.name : ''}
          </>
        ) : null}
        <SVG name={'dropdown'} size={25} color='#3E516C' />
      </Button>
    </Dropdown>
  );
}

export default memo(ChartTypeDropdown);
