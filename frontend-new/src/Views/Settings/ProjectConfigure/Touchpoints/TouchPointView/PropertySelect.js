import React, { useState } from 'react';
import { Button, Tooltip } from 'antd';
import { SVG } from 'factorsComponents';
import styles from './index.module.scss';
import FaSelect from 'Components/GenericComponents/FaSelect';

export const PropertySelect = ({
  title,
  setPropValue,
  renderOptions,
  allowSearch
}) => {
  const [propSelectorOpen, setPropSelectorOpen] = useState(false);

  const handleOptionClick = (option) => {
    setPropSelectorOpen(false);
    setPropValue(option);
  };
  return (
    <div className={`flex flex-col relative items-center ${styles.dropdown}`}>
      <Tooltip title={title}>
        <Button
          className={`${styles.dropdownbtn}`}
          type='text'
          onClick={() => setPropSelectorOpen(true)}
        >
          <div className={styles.dropdownbtntext + '  text-sm'}>{title}</div>
          <div className={styles.dropdownbtnicon}>
            <SVG name='caretDown' size={18} />
          </div>
        </Button>
      </Tooltip>
      {propSelectorOpen && (
        <FaSelect
          allowSearch={allowSearch}
          options={renderOptions()}
          optionClickCallback={handleOptionClick}
          onClickOutside={() => setPropSelectorOpen(false)}
          extraClass={`${styles.dropdownSelect}`}
        ></FaSelect>
      )}
    </div>
  );
};
