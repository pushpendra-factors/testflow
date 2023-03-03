import { Button, Input, notification, Radio } from 'antd';
import AppModal from 'Components/AppModal';
import { Text, SVG } from 'Components/factorsComponents';
import React, { useState } from 'react';
import { ShareData } from '../../../types';
import type { RadioChangeEvent } from 'antd';
import { PlusOutlined } from '@ant-design/icons';

const ShareModal = ({ visible, onCancel, shareData }: ShareModalProps) => {
  const [loading, setIsLoading] = useState<boolean>(false);
  const [emails, setEmails] = useState<string[]>(['']);
  const [subscriptionType, setSubscriptionType] =
    useState<'once' | 'subscribe'>('subscribe');
  const handleRadioChange = (e: RadioChangeEvent) => {
    setSubscriptionType(e.target.value);
  };

  const handleInputChange = (value: string, index: number) => {
    setEmails([...emails.slice(0, index), value, ...emails.slice(index + 1)]);
  };

  const handleAddNewEmailClick = () => {
    setEmails([...emails, '']);
  };

  const renderEmailInputs = () => {
    const inputs = emails?.map((email, index) => (
      <div className={`${index !== 0 ? 'mt-2' : ''}`}>
        <Input
          value={email}
          size='large'
          className='w-100'
          style={{ borderRadius: 6 }}
          onChange={(e) => handleInputChange(e.target.value, index)}
        />
      </div>
    ));
    return inputs;
  };

  const copyToClipboard = async () => {
    try {
      let copied = false;
      const text = shareData?.publicUrl || '';
      if ('clipboard' in navigator) {
        await navigator.clipboard.writeText(text);
        copied = true;
      } else {
        document.execCommand('copy', true, text);
        copied = true;
      }
      if (copied) {
        notification.success({
          message: 'Successfully Copied',
          duration: 3
        });
      }
    } catch (err) {
      console.error('Error in copying data', err);
    }
  };

  const renderFooter = () => (
    <div className='p-2 flex justify-between items-center'>
      {shareData?.publicUrl && document.queryCommandSupported('copy') && (
        <Button
          style={{ color: '#40A9FF' }}
          type='text'
          icon={<SVG name={'link'} color='#40A9FF' />}
          onClick={copyToClipboard}
        >
          Copy link
        </Button>
      )}

      <Button
        onClick={() => {
          console.log('ok clicked---');
          onCancel();
        }}
        size='large'
        type='primary'
      >
        Done
      </Button>
    </div>
  );

  return (
    <div>
      {/* @ts-ignore */}
      <AppModal
        visible={visible}
        footer={renderFooter()}
        onCancel={onCancel}
        isLoading={loading}
      >
        <div className='p-2'>
          <div className='flex items-center justify-between'>
            <Text type={'title'} level={3} weight={'bold'} extraClass='m-0'>
              Share Report
            </Text>
            <Button
              size='middle'
              shape='circle'
              type='text'
              onClick={onCancel}
              icon={<SVG name={'Remove'} color='#8692A3' size={24} />}
            />
          </div>
          <div>
            <Text type={'paragraph'} mini extraClass={'mt-3'} color='grey'>
              By default, everyone in this project will receive weekly and
              monthly updates. Subscribe others to receive the same updates via
              a public link. These users will receive simple, auth-less access.
            </Text>
          </div>
          <div className='mt-4'>
            <Radio.Group onChange={handleRadioChange} value={subscriptionType}>
              <Radio value={'once'}>Send once</Radio>
              <Radio value={'subscribe'}>Subscribe</Radio>
            </Radio.Group>
          </div>
          <div className='mt-4'>
            <Text type={'title'} level={7} weight={'bold'} extraClass={'m-0'}>
              Recipients
            </Text>
            <div className='mt-2'>{renderEmailInputs()}</div>
          </div>
          <div className='mt-3'>
            <Button
              onClick={handleAddNewEmailClick}
              type='text'
              icon={<PlusOutlined color='#8692A3' />}
            >
              Add Email
            </Button>
          </div>
        </div>
      </AppModal>
    </div>
  );
};

type ShareModalProps = {
  visible: boolean;
  onCancel: () => void;
  shareData?: ShareData;
};

export default ShareModal;
