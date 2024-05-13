import React from 'react';
import { Text } from 'Components/factorsComponents';
import CodeBlockV2 from 'Components/CodeBlock/CodeBlockV2';
import CollapsibleContainer from 'Components/GenericComponents/CollapsibleContainer';
import Header from 'Components/GenericComponents/CollapsibleContainer/CollasibleHeader';
import ScriptHtml from '../ScriptHtml';
import { generateSdkScriptCode } from '../utils';
import SdkVerificationFooter from '../SdkVerificationFooter';

interface ManualStepsProps {
  projectToken: string;
  assetURL: string;
  apiURL: string;
  isOnboardingFlow: boolean;
}

interface ManualStepsBodyProps {
  projectToken: string;
  assetURL: string;
  apiURL: string;
  showFooter: boolean;
}

const ManualStepsBody = ({
  projectToken,
  assetURL,
  apiURL,
  showFooter
}: ManualStepsBodyProps) => (
  <div className='flex flex-col gap-1.5 px-4'>
    <Text type='paragraph' color='mono-6' extraClass='m-0'>
      Add the below javascript code on every page between the &lt;head&gt; and
      &lt;/head&gt; tags.
    </Text>
    <div className='py-4'>
      <CodeBlockV2
        collapsedViewText={
          <>
            <span style={{ color: '#2F80ED' }}>{`<script>`}</span>
            {`(function(c)d.appendCh.....func("`}
            <span style={{ color: '#EB5757' }}>{`${projectToken}`}</span>
            {`")`}
            <span style={{ color: '#2F80ED' }}>{`</script>`}</span>
          </>
        }
        fullViewText={
          <ScriptHtml
            projectToken={projectToken}
            assetURL={assetURL}
            apiURL={apiURL}
          />
        }
        textToCopy={generateSdkScriptCode(assetURL, projectToken, apiURL)}
      />
    </div>
    {showFooter && <SdkVerificationFooter type='manual' />}
  </div>
);

const ManualSteps = ({
  projectToken,
  apiURL,
  assetURL,
  isOnboardingFlow = false
}: ManualStepsProps) => (
  <CollapsibleContainer
    showBorder
    key='manual'
    BodyComponent={
      <ManualStepsBody
        projectToken={projectToken}
        apiURL={apiURL}
        assetURL={assetURL}
        showFooter={isOnboardingFlow}
      />
    }
    HeaderComponent={
      <Header
        title='Manual Setup'
        description='Add Factors SDK manually in the head section for all pages you wish to
      get data for'
      />
    }
  />
);

export default ManualSteps;
