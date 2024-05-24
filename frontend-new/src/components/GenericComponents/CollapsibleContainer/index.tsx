import { Collapse } from 'antd';
import React from 'react';
import style from './index.module.scss';

const { Panel } = Collapse;

interface CollapsibleContainerProps {
  keyName: string;
  BodyComponent: React.ReactNode;
  HeaderComponent: React.ReactNode;
  expandIcon?: (panelProps: any) => React.ReactNode;
  showBorder: boolean;
  openByDefault?: boolean;
}

const CollapsibleContainer = ({
  keyName,
  BodyComponent,
  HeaderComponent,
  expandIcon,
  showBorder = true,
  openByDefault = false
}: CollapsibleContainerProps) => (
  <Collapse
    bordered={showBorder}
    defaultActiveKey={openByDefault ? [keyName] : undefined}
    expandIconPosition='right'
    className={`${style.collapse} ${showBorder ? style.border : ''}`}
    expandIcon={expandIcon}
  >
    <Panel key={keyName} header={HeaderComponent}>
      {BodyComponent}
    </Panel>
  </Collapse>
);

export default CollapsibleContainer;
