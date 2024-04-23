import { Button, message, notification } from 'antd';
import { SVG } from 'Components/factorsComponents';
import React from 'react';
import styles from './index.module.scss';

const CodeBlock = ({
  codeContent,
  preClassName = 'my-4 fa-code-block',
  codeClassName = 'fa-code-code-block',
  preProps = {},
  codeProps = {},
  pureTextCode = ``,
  hideCopyBtn = false
}) => (
  <div>
    <pre className={preClassName} {...preProps}>
      {!hideCopyBtn && (
        <div style={{ position: 'absolute', right: '8px' }}>
          <Button
            className={styles.btnV2}
            style={{ marginTop: '-6px' }}
            type='text'
            onClick={(e) => {
              e.stopPropagation();
              navigator?.clipboard
                ?.writeText(pureTextCode)
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
            }}
          >
            <SVG name='TextCopy' size='16' color='#8C8C8C' />
          </Button>
        </div>
      )}
      <code className={codeClassName} {...codeProps}>
        {codeContent}
      </code>
    </pre>
  </div>
);
export default CodeBlock;
