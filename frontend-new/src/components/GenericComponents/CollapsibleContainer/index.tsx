import { Collapse } from 'antd';
import React from 'react';
import style from './index.module.scss';

const { Panel } = Collapse;

interface GTMStepsContainerProps {
  key: string;
  BodyComponent: React.ReactNode;
  HeaderComponent: React.ReactNode;
  expandIcon?: (panelProps: any) => React.ReactNode;
  showBorder: boolean;
}

const CollapsibleContainer = ({
  key,
  BodyComponent,
  HeaderComponent,
  expandIcon,
  showBorder = true
}: GTMStepsContainerProps) => (
  <Collapse
    bordered={showBorder}
    key={key}
    expandIconPosition='right'
    className={`${style.collapse} ${showBorder ? style.border : ''}`}
    expandIcon={expandIcon}
  >
    <Panel key={key} header={HeaderComponent}>
      {BodyComponent}
    </Panel>
  </Collapse>
);

export default CollapsibleContainer;
