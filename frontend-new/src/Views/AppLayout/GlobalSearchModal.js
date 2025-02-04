import React, { useCallback } from 'react';
import { Modal } from 'antd';
import { useDispatch, useSelector } from 'react-redux';
import { TOGGLE_GLOBAL_SEARCH } from 'Reducers/types';
// import GlobalSearch from 'Components/GlobalSearch';
import lazyWithRetry from 'Utils/lazyWithRetry';

const GlobalSearch = lazyWithRetry(
  () =>
    import(
      /* webpackChunkName: "global-search" */ /* webpackPreload: true */ '../../components/GlobalSearch'
    )
);

const GlobalSearchModal = () => {
  const dispatch = useDispatch();

  const isVisibleGlobalSearch = useSelector(
    (state) => state.globalSearch.visible
  );

  const handleCancel = useCallback(() => {
    dispatch({ type: TOGGLE_GLOBAL_SEARCH });
  }, [dispatch]);

  return (
    <Modal
      zIndex={2000}
      keyboard
      visible={isVisibleGlobalSearch}
      footer={null}
      closable={false}
      onCancel={handleCancel}
      bodyStyle={{ padding: 0 }}
      width='40vw'
      className='modal-globalsearch'
      destroyOnClose
    >
      <GlobalSearch />
    </Modal>
  );
};

export default GlobalSearchModal;
