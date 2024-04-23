import React, { useCallback, useEffect } from 'react';
import { Resizable } from 'react-resizable';
import styles from './index.module.scss';
var currentColumn = null;
var currentIndex = -1;
const tableParentId = 'resizing-table-container-div';
var colHead = null;
/*

Use this component only when columns are sticky 
- because sticky antd columns in antd-table, uses 2 native table elements hence we have to check 2 colgroups, and update col-width there.

*/
const ResizableTitle = (props) => {
  const {
    onResize,
    width,
    className,
    children,
    onMouseEnter,
    onMouseLeave,
    onClick,
    ...restProps
  } = props;
  if (!width) {
    return <th {...restProps} />;
  }
  document.onmouseup = () => {
    if (currentColumn) {
      currentColumn = null;
      colHead = null;
      currentIndex = -1;
    }
  };
  const defaultResize = useCallback(()=>{

      const TableParentWidth = document.querySelector('.fa-table--profileslist').querySelector('.ant-table-content').getBoundingClientRect().width - 5 // -5 is added to remove the scrolling for default screen
      const allCols = document.querySelector('.fa-table--profileslist').querySelector('table').querySelector('colgroup').childNodes
      if(TableParentWidth && allCols && allCols.length <= 5){
        allCols.forEach((each)=>{
         
          each.style.width = `${TableParentWidth/allCols.length}px`
        })
      }
  
  },[])
  useEffect(()=>{
    defaultResize();
    window.onresize = defaultResize
    document.onmousemove = (e) => {
      if (currentColumn) {
        let minwidth = Math.max(
          e.clientX -
            currentColumn.parentElement.getBoundingClientRect().left,
          152
        );
        colHead.style.width = minwidth + 'px';
      }
    }
  },[])
  return (
    <th
      {...restProps}
      className={`${className} ${styles['table-custom-th']}`}
      onMouseUp={(e) => {
        if (currentColumn) {
          currentColumn = null;
          colHead = null;
          currentIndex = -1;
        }
      }}
    >
      <span
        className='ant-table-cell-content'
        onClick={onClick}
        onMouseEnter={onMouseEnter}
        onMouseLeave={onMouseLeave}
      >
        {children}
      </span>
      <span
        className={styles['react-resizable-handle']}
        onClick={(e) => {
          e.stopPropagation();
        }}
        onMouseDown={(e) => {
          e.stopPropagation();
          let th = e.currentTarget.parentElement;
          let trs = e.currentTarget.parentElement.parentElement.childNodes;
          let i = 0;
          for (let j = 0; j < trs.length; j++) {
            if (trs[j] === th) {
              i = j;
              break;
            }
          }
          currentIndex = i;

          let colGroups = document
            .getElementById(tableParentId)
            .querySelectorAll('colgroup');

          if (colGroups) {
            colHead = colGroups[0].childNodes[currentIndex];
            
          }
          if (!currentColumn) currentColumn = e.currentTarget;
        }}
      >
        <div
          style={{ height: '28px', width: 0, border: '0.5px solid #dedede' }}
        />
      </span>
    </th>
  );
};
export default ResizableTitle;
