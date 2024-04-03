import React, { useCallback, useEffect, useMemo, useState } from 'react';
import cx from 'classnames';
import AppModal from 'Components/AppModal/AppModal';
import { Text, SVG } from 'Components/factorsComponents';
import { Button, Input, Checkbox, notification } from 'antd';
import styles from './index.module.scss';

const ALLOWED_COLUMNS_COUNT = 10;

const DownloadCSVModal = ({
  visible,
  onCancel,
  onSubmit,
  options,
  displayTableProps,
  isLoading
}) => {
  const [searchText, setSearchText] = useState('');
  const [selectedColumns, setSelectedColumns] = useState([]);

  const handleOk = () => {
    if (selectedColumns.length === 0) {
      notification.error({
        message: 'Error',
        description: 'Please choose at least 1 property to export',
        duration: 2
      });
      return;
    }
    onSubmit(selectedColumns);
  };

  const handleSearchChange = (e) => {
    setSearchText(e.target.value);
  };

  const handleOptionChange = useCallback((option) => {
    setSelectedColumns((curr) => {
      if (curr.indexOf(option.prop_name) === -1) {
        if (curr.length === 10) {
          notification.error({
            message: 'Error',
            description: 'Maximum 10 properties can be chosen',
            duration: 2
          });
          return curr;
        }
        return [...curr, option.prop_name];
      }
      return curr.filter((elem) => elem !== option.prop_name);
    });
  }, []);

  useEffect(() => {
    setSelectedColumns(
      displayTableProps.filter(
        (p) => options.findIndex((option) => option.prop_name === p) > -1
      )
    );
  }, [displayTableProps, options]);

  const filteredOptions = useMemo(() => {
    if (!options || !searchText) {
      return options;
    }

    return options.filter((option) =>
      option && option.display_name
        ? option.display_name.toLowerCase().includes(searchText.toLowerCase())
        : false
    );
  }, [options, searchText]);

  return (
    <AppModal
      okText='Export CSV'
      visible={visible}
      onOk={handleOk}
      onCancel={onCancel}
      width={635}
      isLoading={isLoading}
    >
      <div className='flex flex-col gap-y-4'>
        <Text
          type='title'
          level={5}
          color='character-primary'
          extraClass='mb-0'
          weight='bold'
        >
          Selects which columns to include
        </Text>
        <Input
          prefix={<SVG color='#bfbfbf' name='search' />}
          value={searchText}
          onChange={handleSearchChange}
          placeholder='Search properties'
          className={styles['download-csv-search-input-box']}
        />
        <div className='flex justify-between items-center'>
          <Text
            color='character-disabled-placeholder'
            type='title'
            extraClass='mb-0'
            weight='medium'
          >
            {selectedColumns.length}/{ALLOWED_COLUMNS_COUNT}
          </Text>
          <Button
            className={styles['download-csv-clear-all-button']}
            type='text'
            onClick={() => setSelectedColumns([])}
          >
            <Text
              type='title'
              extraClass='mb-0'
              color='brand-color-6'
              weight='medium'
            >
              Clear All
            </Text>
          </Button>
        </div>
        <div
          className={cx(
            styles['download-csv-modal-options-container'],
            'border p-4 flex flex-col gap-y-4'
          )}
        >
          {filteredOptions.map((option) => {
            return (
              <Checkbox
                key={option.prop_name}
                value={option.prop_name}
                onChange={() => handleOptionChange(option)}
                checked={selectedColumns.indexOf(option.prop_name) > -1}
                className={styles['download-csv-modal-checkbox']}
              >
                <Text type='title' extraClass='mb-0' color='character-primary'>
                  {option.display_name}
                </Text>
              </Checkbox>
            );
          })}
        </div>
      </div>
    </AppModal>
  );
};

export default DownloadCSVModal;
