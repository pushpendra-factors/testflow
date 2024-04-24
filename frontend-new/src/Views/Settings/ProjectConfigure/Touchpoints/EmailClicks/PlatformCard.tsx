import React from 'react';
import { Row, Col, Divider, Alert } from 'antd';
import { Text } from 'factorsComponents';
import CodeBlock from 'Components/CodeBlock';
import unplugImg from '../../../../../assets/images/unplug.png';
import styles from './index.module.scss';

const tagsMap = {
  hubspot: '{{contact.email}}',
  salesforceOutreach: '{{{Recipient.Email}}}',
  salesforceEmailStudio: '%%emailaddr%%',
  apollo: '{{email}}',
  outreach: '{{email}}'
};

interface Props {
  tags: string[];
  selectedPlatform: object;
}

const PlatformCard = ({ tags, selectedPlatform }: Props) => {
  const selectedTags: string = tagsMap[selectedPlatform?.value];

  const renderEmptyBox = () => (
    <div className='flex flex-col items-center mt-16'>
      <Row className=''>
        <Col span={24}>
          <img
            src={unplugImg}
            alt='unplugImg'
            className={`${styles.unplugImg}`}
          />
        </Col>
      </Row>
      <Row className='mt-3'>
        <Col span={24}>
          <Text
            type='title'
            level={6}
            weight='bold'
            extraClass='m-0 flex justify-center'
          >
            Please Select any Platform
          </Text>
          <Text
            type='title'
            level={7}
            color='grey'
            extraClass={`${styles.textEmptyBox}`}
          >
            Choose the platform that you use to send emails to your prospects
          </Text>
        </Col>
      </Row>
    </div>
  );

  const renderData = () => (
    <div>
      <Row className='m-0'>
        <Col span={24}>
          <Text type='title' level={6} weight='bold' extraClass='m-0'>
            {selectedPlatform?.label} UTM Parameter
          </Text>
          <Text type='title' level={7} color='grey' extraClass='m-0'>
            Next to a link in your email, add the UTM tag for email followed by{' '}
            {selectedTags}
          </Text>
        </Col>
      </Row>
      {tags &&
        tags?.map((item, index) => (
          <>
            <Row>
              <Col span={24}>
                <CodeBlock
                  codeContent={
                    <span>
                      ?{item}={selectedTags}
                    </span>
                  }
                  pureTextCode={`?${item}=${selectedTags}`}
                />
              </Col>
            </Row>
            {index !== tags.length - 1 && (
              <Divider className={`${styles.dividerOR}`}>or</Divider>
            )}
          </>
        ))}
      <Alert
        message='If you have other UTM tags added in your link, separate these with ‘&’'
        type='info'
        showIcon
      />
      <Row className='m-0 mt-4'>
        <Col span={24}>
          <Text type='title' level={6} weight='bold' extraClass='m-0'>
            Example Links
          </Text>
          <Text type='title' level={7} color='grey-2' extraClass='m-0 mt-2'>
            If your link to your website is www.acme.com/pricing, add the UTM
            tag like so:
          </Text>
          <Text type='title' level={7} color='grey' extraClass='m-0 mt-2'>
            www.acme.com/pricing
            <span className={`${styles.textColor}`}>
              ?{tags?.[0]}={selectedTags}
            </span>
          </Text>
          <Text type='title' level={7} color='grey-2' extraClass='m-0 mt-2'>
            OR, if you have multiple UTM parameters
          </Text>
          <Text type='title' level={7} color='grey' extraClass='m-0 mt-2'>
            www.acme.com/pricing
            <span className={`${styles.textColor}`}>
              ?utm_source=email&{tags?.[0]}={selectedTags}
            </span>
          </Text>
        </Col>
      </Row>
    </div>
  );

  return (
    <div className={`${styles.container}`}>
      {selectedPlatform?.label ? renderData() : renderEmptyBox()}
    </div>
  );
};

export default PlatformCard;
