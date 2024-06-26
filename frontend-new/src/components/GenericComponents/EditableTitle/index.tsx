import React, { useState } from 'react';
import { Form, Input } from 'antd';
import { EnterOutlined, EditOutlined } from '@ant-design/icons';
import useAutoFocus from 'hooks/useAutoFocus';
import { EditableTitleProps } from './types';

function EditableTitle({
  title,
  editIcon = <EditOutlined />,
  enterIcon = <EnterOutlined />,
  editable = false,
  handleEdit,
  inputClassName = 'w-56 rounded',
  titleClassName = 'text-lg font-bold'
}: EditableTitleProps) {
  const [editTitle, setEditTitle] = useState(false);
  const inputComponentRef = useAutoFocus(editTitle);

  const handleFinish = (pageTitle: string) => {
    handleEdit(pageTitle);
    setEditTitle(false);
  };
  return editTitle ? (
    <Form
      name='basic'
      labelCol={{ span: 8 }}
      wrapperCol={{ span: 16 }}
      onFinish={({ pageTitle }) => handleFinish(pageTitle)}
      autoComplete='off'
      onBlur={() => setEditTitle(false)}
    >
      <Form.Item name='pageTitle'>
        <Input
          value={title}
          defaultValue={title}
          ref={inputComponentRef}
          suffix={enterIcon}
          className={inputClassName}
        />
      </Form.Item>
    </Form>
  ) : (
    <div className='flex items-center gap-x-1'>
      <div className={titleClassName}>{title}</div>
      {editable && (
        <div className='cursor-pointer' onClick={() => setEditTitle(true)}>
          {editIcon}
        </div>
      )}
    </div>
  );
}

export default EditableTitle;
