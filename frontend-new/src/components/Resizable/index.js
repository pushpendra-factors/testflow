import React from 'react';
import { Resizable } from 'react-resizable';
import styles from './index.module.scss';

const ResizableTitle = (props) => {
  const { onResize, width, ...restProps } = props;
  if (!width) {
    return <th {...restProps} />;
  }
  return (
    <Resizable
      width={width}
      height={0}
      handle={
        <span
          className={styles['react-resizable-handle']}
          onClick={(e) => {
            e.stopPropagation();
          }}
        >
          <div
            style={{ height: '28px', width: 0, border: '0.5px solid #dedede' }}
          ></div>
        </span>
      }
      onResize={(e, { size }) => {
        e.stopPropagation();
        e.preventDefault();
        onResize(e, { size });
      }}
      draggableOpts={{
        enableUserSelectHack: false
      }}
    >
      <th {...restProps} />
    </Resizable>
  );
};
export default ResizableTitle;
