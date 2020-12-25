import React from 'react';
import {
  Modal
} from 'antd';
import { Text } from '../factorsComponents';

function ConfirmationModal({
  visible, confirmationText, onOk, onCancel, title, width, cancelText, okText, confirmLoading
}) {
  return (
    <Modal
      centered={true}
      visible={visible}
      width={width || 600}
      title={null}
      onOk={onOk}
      onCancel={onCancel}
      className={'fa-modal--regular p-4 fa-modal--slideInDown'}
      okText={okText}
      cancelText={cancelText}
      closable={false}
      confirmLoading={confirmLoading}
      transitionName=""
      maskTransitionName=""
    >
      <div className="p-6">
        <Text extraClass="m-0" type={'title'} level={3} weight={'bold'}>{title}</Text>
        <Text extraClass={'pt-2'} mini type={'paragraph'}>{confirmationText}</Text>
      </div>
    </Modal>
  );
}

export default ConfirmationModal;
