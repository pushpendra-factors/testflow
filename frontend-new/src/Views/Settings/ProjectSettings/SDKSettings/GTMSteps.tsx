import CodeBlockV2 from 'Components/CodeBlock/CodeBlockV2';
import { Text } from 'Components/factorsComponents';
import React from 'react';
import ScriptHtml from './ScriptHtml';
import { generateSdkScriptCode } from './utils';

const GTMSteps = ({
  projectToken,
  assetURL,
  apiURL
}: {
  projectToken: string;
  assetURL: string;
  apiURL: string;
}) => {
  return (
    <>
      <Text type='paragraph' color='mono-6' extraClass={'m-0'}>
        1. Sign in to&nbsp;
        <span>
          <a
            href='https://tagmanager.google.com/'
            target='_blank'
            rel='noreferrer'
          >
            Google Tag Manager
          </a>
        </span>
        &nbsp;and select “Workspace”.
      </Text>
      <Text type='paragraph' color='mono-6' extraClass={'m-0'}>
        2. Click on “Add a new tag” and name it “Factors tag”.
      </Text>
      <Text type='paragraph' color='mono-6' extraClass={'m-0'}>
        3. Click <span className='italic'>Edit</span> on Tag Configuration and
        under custom, select <span className='italic'>Custom HTML</span>
      </Text>
      <Text type='paragraph' color='mono-6' extraClass={'m-0'}>
        4. Copy the tracking script below and paste it on the HTML field. Hit{' '}
        <span className='italic'>Save</span>.
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

      <Text type='paragraph' color='mono-6' extraClass={'m-0'}>
        5. In the Triggers popup, click{' '}
        <span className='italic'>Add Trigger</span> and select All Pages.
      </Text>
      <Text type='paragraph' color='mono-6' extraClass={'m-0'}>
        6. Once the trigger has been added, click on Publish at the top of your
        GTM window and that’s it!
      </Text>
    </>
  );
};

export default GTMSteps;
