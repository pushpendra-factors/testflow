import React from 'react';
import { Modal, Spin } from 'antd';
import styles from './index.module.scss';
import ActiveUnitContent from './ActiveUnitContent';

function ExpandableView({ widgetModal, setwidgetModal, loading }) {
  let content = null;

  if (loading) {
    content = (
            <div className="flex justify-center items-center w-full min-h-screen">
                <Spin size="small" />
            </div>
    );
  } else if (widgetModal.data) {
    const { unit, data } = widgetModal;
    content = (
            <ActiveUnitContent
                unit={unit}
                unitData={data}
            />
    );
  }

  return (
        <Modal
            title={null}
            visible={widgetModal}
            footer={null}
            centered={false}
            zIndex={1015}
            mask={false}
            onCancel={() => setwidgetModal(false)}
            className={`w-full inset-0 ${styles.fullModal}`}
        >
            {content}
        </Modal>
  );
}

export default ExpandableView;
