import React, { useMemo, memo } from 'react';
import { getValidGranularityOptions } from '../../utils/dataFormatter';
import { Dropdown, Menu, Button } from 'antd';
import { DateBreakdowns } from '../../utils/constants';
import styles from './index.module.scss';
import { SVG } from '../factorsComponents';

function GranularityOptions({ durationObj, onClick, queryType }) {
  const validDateBreakdowns = [...DateBreakdowns];
  
  const options = useMemo(() => {
    const enabledOptions = getValidGranularityOptions(durationObj, queryType);
    return validDateBreakdowns.map((db) => {
      return {
        ...db,
        disabled: enabledOptions.indexOf(db.key) === -1,
      };
    });
  }, [durationObj, queryType, validDateBreakdowns]);

  const currentValue = useMemo(() => {
    if (!durationObj) {
      return validDateBreakdowns[0];
    }
    const { frequency } = durationObj;
    return validDateBreakdowns.find((elem) => elem.key === frequency);
  }, [durationObj, validDateBreakdowns]);

  const menu = (
    <Menu className={styles.dropdownMenu}>
      {options.map((option) => {
        return (
          <Menu.Item
            className={`${styles.dropdownMenuItem} ${
              currentValue.key === option.key ? styles.active : ''
            } ${option.disabled ? styles.disabled : ''}`}
            key={option.key}
            onClick={onClick}
            disabled={option.disabled}
          >
            <div className={'flex items-center'}>
              <span className='mr-3'>{option.title}</span>
              {currentValue.key === option.key ? (
                <SVG name='checkmark' size={17} color='#8692A3' />
              ) : null}
            </div>
          </Menu.Item>
        );
      })}
    </Menu>
  );

  return (
    <Dropdown overlay={menu} placement="bottomLeft">
      <Button className={`ant-dropdown-link flex items-center`}>
        {currentValue.title}
      </Button>
    </Dropdown>
  );
}

export default memo(GranularityOptions);
