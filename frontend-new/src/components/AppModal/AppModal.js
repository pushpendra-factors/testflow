import React from 'react';
import PropTypes from 'prop-types';
import noop from 'lodash/noop';
import { Modal } from 'antd';

const AppModal = (props) => {
  const {
    visible,
    width,
    title,
    onOk,
    onCancel,
    okText,
    cancelText,
    closable,
    isLoading,
    className,
    children,
    ...rest
  } = props;

  return (
    <Modal
      centered={true}
      visible={visible}
      width={width}
      title={title}
      onOk={onOk}
      onCancel={onCancel}
      className={className}
      okText={okText}
      closable={closable}
      confirmLoading={isLoading}
      cancelText={cancelText}
      {...rest}
    >
      {children}
    </Modal>
  );
};

export default AppModal;

AppModal.propTypes = {
  children: PropTypes.element,
  visible: PropTypes.bool,
  width: PropTypes.number,
  title: PropTypes.oneOfType([PropTypes.string, PropTypes.instanceOf(null)]),
  onOk: PropTypes.func,
  onCancel: PropTypes.func,
  okText: PropTypes.string,
  cancelText: PropTypes.string,
  closable: PropTypes.bool,
  isLoading: PropTypes.bool,
  className: PropTypes.string
};

AppModal.defaultProps = {
  visible: false,
  width: 700,
  title: null,
  onOk: noop,
  onCancel: noop,
  okText: 'Ok',
  cancelText: 'Cancel',
  closable: false,
  isLoading: false,
  className: 'fa-modal--regular'
};
