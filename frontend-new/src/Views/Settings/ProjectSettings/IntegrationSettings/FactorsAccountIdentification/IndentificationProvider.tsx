import CollapsibleContainer from 'Components/GenericComponents/CollapsibleContainer';
import Header from 'Components/GenericComponents/CollapsibleContainer/CollasibleHeader';
import React from 'react';
import { SVG } from 'Components/factorsComponents';
import SixSignalFactors from '../SixSignalFactors';
import ExternalProvider from './ExternalProvider';

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
    <div className='flex flex-col gap-4'>
      <CollapsibleContainer
        showBorder={false}
        key='factors'
        HeaderComponent={
          <Header
            title='Factor Deanonymization'
            description='Use Factors deanonymization to identify accounts visiting your website. Monthly quota is calculated based on your plan.'
          />
        }
        BodyComponent={<SixSignalFactors kbLink={kbLink} />}
        expandIcon={collapseIcon}
      />
      <CollapsibleContainer
        showBorder={false}
        expandIcon={collapseIcon}
        key='third-party'
        HeaderComponent={
          <Header
            title='Third party Integrations'
            description='Use your existing API key to identify accounts visiting your website. Your usage is metered by your API provider directly.'
          />
        }
        BodyComponent={<ExternalProvider />}
      />
    </div>
  );
};

export default IndentificationProvider;
