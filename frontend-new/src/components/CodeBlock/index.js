import { Button, message } from 'antd';
import React from 'react';
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
            type='primary'
            size='small'
            onClick={() => {
              navigator?.clipboard?.writeText(codeContent);
              message.success({
                content: 'copied',
                style: { display: 'flex', justifyContent: 'center', margin: '0 auto' }
              });
            }}
          >
            Copy Code
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
