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
    return (
      <Menu
        onClick={(info) => {
          let num = Number(info.key);
          setValuesState(num * 15);
        }}
        style={{ overflow: 'scroll', maxHeight: '185px' }}
      >
        <Menu.Item key={1}>15 Days</Menu.Item>
        <Menu.Item key={2}>30 Days</Menu.Item>
        <Menu.Item key={3}>45 Days</Menu.Item>
        <Menu.Item key={4}>60 Days</Menu.Item>
        <Menu.Item key={5}>75 Days</Menu.Item>
        <Menu.Item key={6}>90 Days</Menu.Item>
        <Menu.Item key={7}>105 Days</Menu.Item>
        <Menu.Item key={8}>120 Days</Menu.Item>
        <Menu.Item key={9}>135 Days</Menu.Item>
        <Menu.Item key={10}>150 Days</Menu.Item>
        <Menu.Item key={11}>165 Days</Menu.Item>
        <Menu.Item key={12}>180 Days</Menu.Item>
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
            <Button className='dropdown-btn' type='text'>
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
