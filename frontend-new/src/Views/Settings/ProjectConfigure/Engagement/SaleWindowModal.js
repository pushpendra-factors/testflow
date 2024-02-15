import { DownOutlined } from '@ant-design/icons';
import { Button, Dropdown, Input, Menu, Modal, Space } from 'antd';
import { Text } from 'Components/factorsComponents';
import React, { useCallback, useEffect, useState } from 'react';

function SaleWindowModal({ saleWindowValue, visible, onOk, onCancel }) {
  const [valuesState, setValuesState] = useState(saleWindowValue);
  useEffect(() => {
    setValuesState(saleWindowValue);
  }, [saleWindowValue]);
  const onCancelState = () => {
    onCancel();
    setValuesState(saleWindowValue);
  };
  const onSaveState = () => {
    onOk(valuesState);
  };

  const setNumericalValue = (ev) => {
    const userInput = ev.target.value;
    const numericValue = userInput.replace(/\D/g, '');
    setValuesState(numericValue);
  };

  const menu = useCallback(() => {
    let arr = new Array(12).fill(0).map((e, i) => (i + 1) * 15);
    return (
      <Menu
        onClick={(info) => {
          let num = Number(info.key);
          setValuesState(num * 15);
        }}
        style={{ overflow: 'scroll', maxHeight: '185px' }}
      >
        {arr.map((eachValue, eachIndex) => {
          return (
            <Menu.Item style={{ padding: '10px' }} key={eachIndex + 1}>
              {eachValue} Days
            </Menu.Item>
          );
        })}
      </Menu>
    );
  }, []);

  return (
    <Modal
      title={null}
      width={641}
      visible={visible}
      footer={null}
      className={'fa-modal--regular p-4'}
      closable={true}
      maskClosable={true}
      onCancel={onCancelState}
      centered
    >
      <div className='p-6'>
        <div className='pb-4'>
          <Text extraClass='m-0' type='title' level={4} weight='bold'>
            Set engagement window
          </Text>
          <Text extraClass='m-0' type='title' level={7} color='grey'>
            Engagement score is refreshed everyday for all accounts. So, if an
            account shows no new intent, it’s engagement level keeps decreasing
            till it reaches 0 at the end of the engagement window.
          </Text>
        </div>
        <div className='mb-2'>
          <Dropdown overlay={menu} overlayStyle={{ zIndex: 10001 }}>
            <Button
              className='dropdown-btn flex justify-between'
              type='text'
              style={{ width: '142px' }}
            >
              {valuesState ? `${valuesState} Days` : `Select`}

              <DownOutlined />
            </Button>
          </Dropdown>
        </div>
        <div className='flex flex-row-reverse justify-between'>
          <div>
            <Button className='mr-1' type='default' onClick={onCancelState}>
              Cancel
            </Button>
            <Button
              className='ml-1'
              style={{ width: '72px' }}
              type='primary'
              onClick={onSaveState}
            >
              Set
            </Button>
          </div>
        </div>
      </div>
    </Modal>
  );
}
export default SaleWindowModal;
