import React, { useState, useCallback } from 'react';
import {
  Button, Modal, Input, Switch
} from 'antd';
import { SVG, Text } from '../factorsComponents';
import styles from './index.module.scss';
import { saveQuery } from '../../reducers/coreQuery/services';
import { useSelector, useDispatch } from 'react-redux';
import { QUERY_CREATED } from '../../reducers/types';

function SaveQuery({
  requestQuery, setQuerySaved, visible, setVisible
}) {
  const [title, setTitle] = useState('');
  const { active_project } = useSelector(state => state.global);
  const dispatch = useDispatch();

  const handleTitleChange = (e) => {
    setTitle(e.target.value);
  };

  const handleSave = useCallback(async () => {
    if (!title.trim().length) {
      return false;
    }
    try {
      const res = await saveQuery(active_project.id, title, requestQuery);
      dispatch({ type: QUERY_CREATED, payload: res.data });
      setQuerySaved(true);
      setTitle('');
      setVisible(false);
    } catch (err) {
      console.log(err);
    }
  }, [title, active_project.id, requestQuery, dispatch, setQuerySaved, setVisible]);

  return (
    <>
      <Button
        onClick={setVisible.bind(this, true)}
        style={{ display: 'flex' }}
        className="items-center"
        type="primary"
        icon={<SVG extraClass="mr-1" name={'save'} size={24} color="#FFFFFF" />}
      >
        Save
            </Button>

      <Modal
        centered={true}
        visible={visible}
        width={700}
        title={null}
        onOk={handleSave}
        onCancel={setVisible.bind(this, false)}
        className={'fa-modal--regular p-4'}
        okText={'Save'}
        closable={false}
      >
        <div className="p-4">
          <Text extraClass="m-0" type={'title'} level={3} weight={'bold'}>Save this Query</Text>
          <div className="pt-6">
            <Text type={'title'} level={7} extraClass={`m-0 ${styles.inputLabel}`}>Title</Text>
            <Input onChange={handleTitleChange} value={title} className={'fa-input'} size={'large'} />
          </div>
          <div className={`pt-2 ${styles.linkText}`}>Help others to find this query easily?</div>
          <div className={'pt-6 flex items-center'}>
            <Switch className={styles.switchBtn} checkedChildren="On" unCheckedChildren="Off" />
            <Text extraClass="m-0" type="title" level={6} weight="bold">Add to Dashboard</Text>
          </div>
          <Text extraClass={`pt-1 ${styles.noteText}`} mini type={'paragraph'}>Create a dashboard widget for regular monitoring</Text>
        </div>
      </Modal>
    </>
  );
}

export default SaveQuery;
