import { Button, message } from 'antd';
import { SVG } from 'Components/factorsComponents';
import React from 'react';
import styles from './index.module.scss';
const CodeBlock = ({
  codeContent,
  preClassName = 'my-4 fa-code-block',
  codeClassName = 'fa-code-code-block',
  preProps = {},
  codeProps = {},
  pureTextCode = ``
}) => {
  return (
    <div>
      <pre className={preClassName} {...preProps}>
        <div style={{ position: 'absolute', right: '8px' }}>
          <Button
            className={styles['btn']}
            onClick={() => {
              navigator?.clipboard
                ?.writeText(pureTextCode)
                .then(() => {
                  message.success({
                    content: 'copied',
                    style: {
                      display: 'flex',
                      justifyContent: 'center',
                      margin: '0 auto'
                    }
                  });
                })
                .catch(() => {
                  console.log('ERROR');
                  message.error({
                    content: 'copying failed',
                    style: {
                      display: 'flex',
                      justifyContent: 'center',
                      margin: '0 auto'
                    }
                  });
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
