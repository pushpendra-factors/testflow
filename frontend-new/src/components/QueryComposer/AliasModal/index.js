import React, { useState, useEffect } from 'react';
import { Modal, Input, Button } from 'antd';
import { SVG, Text } from 'factorsComponents';
import { connect } from 'react-redux';
import styles from './index.module.scss';

function AliasModal({ visible, onOk, onCancel, alias, event }) {
  const [aliasName, setAliasName] = useState('');
  const handleUserInput = (e) => {
    setAliasName(e.target.value);
  };

  useEffect(() => {
    setAliasName(alias);
  }, [alias]);

  const onCancelState = () => {
    onCancel();
  };

  const resetInputField = () => {
    setAliasName('');
    onOk();
  };

  return (
    <Modal
      title={null}
      width={750}
      visible={visible}
      footer={null}
      className={'fa-modal--regular p-6'}
      closable={false}
    >
      <div className='p-6'>
        <Text extraClass='m-0' type={'title'} level={3} weight={'bold'}>
          Alias for...
        </Text>
        <Text extraClass={'pt-0'} type={'paragraph'} weight={'bold'}>
          {event}
        </Text>
      </div>
      <div className='px-6'>
        <Text extraClass={'pb-2'} mini type={'paragraph'}>
          Use this alias to easily reference this event with its filters on your
          report.
        </Text>
        <Input
          className='fa-input'
          placeholder='Alias title'
          value={aliasName}
          onChange={handleUserInput}
        ></Input>
      </div>
      <div className={`p-6 flex flex-row-reverse justify-between`}>
        <div>
          <Button className='mr-1' type='default' onClick={onCancelState}>
            Cancel
          </Button>
          <Button
            disabled={!aliasName?.length}
            className='ml-1'
            type='primary'
            onClick={() => onOk(aliasName)}
          >
            {' '}
            {alias?.length ? 'Update' : 'Create'}
          </Button>
        </div>
        {alias?.length ? (
          <Button
            type='text'
            onClick={resetInputField}
            icon={<SVG size={16} name='trash' color={'grey'} />}
          >
            Delete Alias
          </Button>
        ) : null}
      </div>
    </Modal>
  );
}

const mapStateToProps = (state) => ({
  eventNames: state.coreQuery.eventNames,
});

export default connect(mapStateToProps, {})(AliasModal);
