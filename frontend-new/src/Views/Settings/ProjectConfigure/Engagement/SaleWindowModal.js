import { Button, Input, Modal } from 'antd';
import { Text } from 'Components/factorsComponents';
import React, { useState } from 'react';

function SaleWindowModal({ visible, onOk, onCancel }) {
  const [valuesState, setValuesState] = useState('');

  const onCancelState = () => {
    onCancel();
  };
  const onSaveState = () => {
    onOk(valuesState);
  };

  const setNumericalValue = (ev) => {
    const userInput = ev.target.value;
    const numericValue = userInput.replace(/\D/g, '');
    setValuesState(numericValue);
  };

  return (
    <Modal
      title={null}
      width={500}
      visible={visible}
      footer={null}
      className={'fa-modal--regular p-4'}
      closable={false}
      centered
    >
      <div className='p-6'>
        <div className='pb-4'>
          <Text extraClass='m-0' type='title' level={4} weight='bold'>
            Set sale window
          </Text>
          <Text extraClass='m-0' type='title' level={7} color='grey'>
            How long is your average sales cycle
          </Text>
        </div>
        <div className='mb-2'>
          <Input
            type='text'
            value={valuesState}
            className={`input-value`}
            style={{ width: '76px' }}
            autoFocus={true}
            onPressEnter={() => setNumericalValue}
            onChange={setNumericalValue}
          ></Input>
          <Button> Days</Button>
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
