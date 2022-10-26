import React, { memo } from 'react';
import { Dropdown, Button, Menu, Tooltip } from 'antd';
import { SVG, Text } from '../factorsComponents';
import styles from './index.module.scss';

import { TOOLTIP_CONSTANTS } from '../../constants/tooltips.constans';
import { CHART_TYPES_DROPDOWN_CONSTANTS } from '../../constants/chartTypesDropDown.constants';

function ChartTypeDropdown({ menuItems, onClick, chartType }) {
  // console.log("CHART_TYPE",  chartType)
  const menu = (
    <div
      className={`flex shadow-md rounded items-center flex-wrap bg-white p-4 text-center ${
        styles.dropdownMenu
      } ${menuItems.length < 3 ? styles.smallMenu : ''}`}
    >
      {menuItems.map((item) => {
        console.log("CHART_TYPE", item)
        return (
          <Tooltip 
            title={CHART_TYPES_DROPDOWN_CONSTANTS[item.key]}
            color={TOOLTIP_CONSTANTS.DARK}
            >
              <div
              className={`${
                styles.item
              } flex flex-col items-center justify-center p-3 cursor-pointer ${
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
          </Tooltip>
        );
      })}
    </div>
  );

  const activeItem = menuItems.find((item) => item.key === chartType);

  return (
    <Dropdown 
      overlay={menu}                 
      placement="bottomRight"
    >
      <Button
        size={'large'}
        className={`ant-dropdown-link flex items-center ${styles.dropdownBtn}`}
      >
        {chartType ? (
          <>
            <SVG name={chartType} size={25} color='#0E2647' />
            {activeItem ? activeItem.name : ''}
          </>
        ) : null}
        <SVG name={'dropdown'} size={25} color='#3E516C' />
      </Button>
    </Dropdown>
  );
}

export default memo(ChartTypeDropdown);
