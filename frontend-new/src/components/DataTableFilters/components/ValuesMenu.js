import React, { useState } from 'react';
import { Input, Menu } from 'antd';
import cx from 'classnames';
import styles from './valuesMenu.module.scss';
import ControlledComponent from '../../ControlledComponent/ControlledComponent';
import { SVG, Text } from '../../factorsComponents';

const ValuesMenu = ({ options, selectedOptions, onChange }) => {
  const [searchText, setSearchText] = useState('');

  const handleSearchChange = (e) => {
    setSearchText(e.target.value);
  };

  const filteredOptions = options.filter((option) =>
    option.toLowerCase().includes(searchText.toLowerCase())
  );

  const handleItemClick = (option) => {
    onChange(option);
  };

  return (
    <div className="flex flex-col row-gap-1">
      <Input
        type="search"
        className={styles['input-search-box']}
        value={searchText}
        onChange={handleSearchChange}
        placeholder="Search"
      />
      <Menu className={styles['values-menu']}>
        {filteredOptions.map((option) => {
          const isSelected = selectedOptions.indexOf(option) > -1;
          return (
            <Menu.Item
              key={option}
              className={cx(styles['values-menu-item'], {
                [styles['values-menu-item-selected']]: isSelected
              })}
              onClick={handleItemClick.bind(null, option)}
            >
              <Text type="title" color="grey-6" level={7}>
                {option}
              </Text>
              <ControlledComponent controller={isSelected}>
                <SVG name="checkmark" />
              </ControlledComponent>
            </Menu.Item>
          );
        })}
      </Menu>
      <ControlledComponent
        controller={filteredOptions.length === 0 && searchText.length > 0}
      >
        <div className="flex">
          <Text type="title" level={7} color="grey-6">
            No results
          </Text>
        </div>
      </ControlledComponent>
    </div>
  );
};

export default ValuesMenu;
