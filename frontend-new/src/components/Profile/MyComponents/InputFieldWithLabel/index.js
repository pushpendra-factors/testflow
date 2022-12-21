import React from 'react';
import { Input } from 'antd';
import { Text } from 'Components/factorsComponents';

const InputFieldWithLabel = ({
  isTextArea = false,
  labelClass,
  inputClass,
  extraClass,
  title,
  placeholer,
  value,
  onChange
}) =>
  isTextArea ? (
    <div className={extraClass}>
      <Text type='title' level={7} extraClass={labelClass}>
        {title}
      </Text>
      <Input.TextArea
        rows={3}
        className={inputClass}
        placeholder={placeholer}
        value={value}
        onChange={onChange}
      ></Input.TextArea>
    </div>
  ) : (
    <div className={extraClass}>
      <Text type='title' level={7} extraClass={labelClass}>
        {title}
      </Text>
      <Input
        className={inputClass}
        placeholder={placeholer}
        value={value}
        onChange={onChange}
      ></Input>
    </div>
  );

export default InputFieldWithLabel;
