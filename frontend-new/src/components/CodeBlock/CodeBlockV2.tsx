import React, { ReactNode, useState } from 'react';
import style from './index.module.scss';
import { Button, notification } from 'antd';
import { SVG } from 'Components/factorsComponents';
const CodeBlockV2 = ({
  textToCopy,
  collapsedViewText,
  fullViewText
}: CodeBlockV2Props) => {
  const [collapsed, setCollapsed] = useState(true);
  const copyCode = (e) => {
    e.stopPropagation();
    navigator?.clipboard
      ?.writeText(textToCopy)
      .then(() => {
        notification.success({
          message: 'Success',
          description: 'Successfully copied!',
          duration: 3
        });
      })
      .catch(() => {
        notification.error({
          message: 'Failed!',
          description: 'Failed to copy!',
          duration: 3
        });
      });
  };
  const handleCollapse = (e) => {
    e.stopPropagation();
    setCollapsed((state) => !state);
  };
  return (
    <div className={style.codeBlockV2} onClick={copyCode}>
      <pre>
        <code className='fa-code-code-block'>
          {collapsed && <>{collapsedViewText}</>}
          {!collapsed && <>{fullViewText}</>}
        </code>
      </pre>
      <div className='flex gap-2 '>
        <Button onClick={handleCollapse} type='text' className={style.btnV2}>
          <SVG
            name={collapsed ? 'Expand' : 'Collapse'}
            size='16'
            color='#8C8C8C'
          />
        </Button>
        <Button className={style.btnV2} onClick={copyCode} type='text'>
          <SVG name='TextCopy' size='16' color='#8C8C8C' />
        </Button>
      </div>
    </div>
  );
};

interface CodeBlockV2Props {
  textToCopy: string;
  collapsedViewText: ReactNode;
  fullViewText: ReactNode;
}

export default CodeBlockV2;
