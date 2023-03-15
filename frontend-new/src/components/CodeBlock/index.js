import { Button, message } from 'antd';
import { SVG } from 'Components/factorsComponents';
import React from 'react';
import styles from './index.module.scss';
const CodeBlock = ({
  codeContent,
  preClassName = '',
  codeClassName = '',
  preProps = {},
  codeProps = {}
}) => {
  return (
    <div>
      <pre className={preClassName} {...preProps}>
        <div style={{ position: 'absolute', right: '8px' }}>
          <Button
            className={styles['btn']}
            onClick={() => {
              navigator?.clipboard?.writeText(codeContent);
              message.success({
                content: 'copied',
                style: {
                  display: 'flex',
                  justifyContent: 'center',
                  margin: '0 auto'
                }
              });
            }}
          >
            <SVG name='copycode' />
          </Button>
        </div>
        <code className={codeClassName} {...codeProps}>
          {codeContent}
        </code>
      </pre>
    </div>
  );
};
export default CodeBlock;
