import CollapsibleContainer from 'Components/GenericComponents/CollapsibleContainer';
import Header from 'Components/GenericComponents/CollapsibleContainer/CollasibleHeader';
import React from 'react';
import { SVG } from 'Components/factorsComponents';
import { Divider } from 'antd';
import SixSignalFactors from '../SixSignalFactors';
import ExternalProvider from './ExternalProvider';
import styles from './index.module.scss';

interface IndentificationProviderProps {
  kbLink: string;
}

const IndentificationProvider = ({ kbLink }: IndentificationProviderProps) => {
  const collapseIcon = (panelProps: string) => {
    if (panelProps?.isActive) {
      return (
        <span className='anticon anticon-right ant-collapse-arrow'>
          <SVG name='CaretUp' size='24' />
        </span>
      );
    }
    return (
      <span className='anticon anticon-right ant-collapse-arrow'>
        <SVG name='CaretDown' size='24' />
      </span>
    );
  };
  return (
    <div className={`${styles.collapse} flex flex-col`}>
      <CollapsibleContainer
        showBorder={false}
        keyName='factors'
        openByDefault
        HeaderComponent={
          <Header
            title='Factors Deanonymization'
            description='Use Factors Deanonymization to identify accounts visiting your website. Monthly quota is calculated based on your plan.'
          />
        }
        BodyComponent={<SixSignalFactors kbLink={kbLink} />}
        expandIcon={collapseIcon}
      />
      <Divider />
      <CollapsibleContainer
        showBorder={false}
        expandIcon={collapseIcon}
        keyName='third-party'
        HeaderComponent={
          <Header
            title='Third-Party Integrations'
            description='Use your existing API key to identify accounts visiting your website. Your usage is metered by your API provider directly.'
          />
        }
        BodyComponent={<ExternalProvider />}
      />
    </div>
  );
};

export default IndentificationProvider;
