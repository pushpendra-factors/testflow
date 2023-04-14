import { Button, Form, Input, notification, Radio } from 'antd';
import AppModal from 'Components/AppModal';
import { Text, SVG } from 'Components/factorsComponents';
import React, { useState } from 'react';
import { ShareData } from '../../../types';
import { PlusOutlined, MinusCircleOutlined } from '@ant-design/icons';
import style from './index.module.scss';
import logger from 'Utils/logger';
import {
  shareSixSignalReportToEmails,
  subscribeToVistorIdentificationEmails
} from '../../../state/services';

const ShareModal = ({ visible, onCancel, shareData }: ShareModalProps) => {
  const [loading, setIsLoading] = useState<boolean>(false);
  const [form] = Form.useForm();

  const handleFinish = async (values: {
    emails: string[];
    subscriptionType: string;
  }) => {
    const { from, timezone, to, domain, publicUrl, projectId } =
      shareData || {};
    const isSubscriptionType = values.subscriptionType === 'subscribe';
    try {
      if (
        !from ||
        !to ||
        !publicUrl ||
        !timezone ||
        !values?.emails ||
        !projectId
      ) {
        notification.error({
          message: 'Fields are missing!',
          duration: 3
        });
        return;
      }
      setIsLoading(true);
      if (!isSubscriptionType) {
        await shareSixSignalReportToEmails(
          values.emails,
          publicUrl,
          domain || '',
          from,
          to,
          timezone,
          projectId
        );
      } else {
        await subscribeToVistorIdentificationEmails(values.emails, projectId);
      }

      notification.success({
        message: `${
          isSubscriptionType
            ? 'Emails added to subscription list'
            : 'Emails successfully sent'
        }`,
        duration: 3
      });
      onCancel();
    } catch (error) {
      logger.error('Error in sending report over emails', error);
      notification.error({
        message: 'Error',
        description: error?.data?.error || 'Something went wrong',
        duration: 5
      });
    }
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
      logger.error('Error in copying data', err);
    }
  };

  return (
    <div>
      {/* @ts-ignore */}
      <AppModal
        visible={visible}
        footer={<></>}
        onCancel={onCancel}
        isLoading={loading}
        className={style.shareModal}
      >
        <div>
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
              Share a public link that allows others in your team to view this.
              These users will have access to this report without having to sign
              up or login.
            </Text>
          </div>
          <Form
            name='share-modal-form'
            onFinish={handleFinish}
            initialValues={{ subscriptionType: 'subscribe', emails: [''] }}
            form={form}
          >
            <div className='mt-4'>
              <Form.Item
                name='subscriptionType'
                rules={[
                  {
                    required: true,
                    message: 'Please select subscription type',
                    type: 'string'
                  }
                ]}
              >
                <Radio.Group>
                  <Radio value={'once'}>Send once</Radio>
                  <Radio value={'subscribe'}>Subscribe</Radio>
                </Radio.Group>
              </Form.Item>
            </div>

            <div className='mt-4'>
              <Text type={'title'} level={7} weight={'bold'} extraClass={'m-0'}>
                Recipients
              </Text>
              <div className='mt-2'>
                <Form.List
                  name='emails'
                  rules={[
                    {
                      validator: async (_, emails) => {
                        if (!emails || emails.length < 1) {
                          return Promise.reject(
                            new Error('Enter at least one email')
                          );
                        }
                      }
                    }
                  ]}
                >
                  {(fields, { add, remove }, { errors }) => (
                    <>
                      {fields.map((field, index) => (
                        <Form.Item required={true} key={field.key}>
                          <div
                            className={`flex w-100 items-center gap-2 ${
                              index !== 0 ? 'mt-3' : ''
                            }`}
                          >
                            <Form.Item
                              {...field}
                              validateTrigger={['onBlur']}
                              rules={[
                                {
                                  required: true,
                                  whitespace: true,
                                  type: 'email',
                                  message: 'Please Enter a valid Email'
                                }
                              ]}
                              noStyle
                            >
                              <Input
                                size='large'
                                className='w-100'
                                style={{ borderRadius: 6 }}
                              />
                            </Form.Item>
                            {fields.length > 1 ? (
                              <Button
                                size='middle'
                                shape='circle'
                                type='text'
                                onClick={() => remove(field.name)}
                                icon={
                                  <MinusCircleOutlined
                                    style={{ color: '#8692A3' }}
                                  />
                                }
                              />
                            ) : null}
                          </div>
                        </Form.Item>
                      ))}
                      {fields?.length < 5 && (
                        <Form.Item>
                          <Button
                            onClick={() => add()}
                            type='text'
                            icon={<PlusOutlined color='#8692A3' />}
                            className='mt-3'
                          >
                            Add Email
                          </Button>
                          <Form.ErrorList errors={errors} />
                        </Form.Item>
                      )}
                    </>
                  )}
                </Form.List>
              </div>
            </div>

            <div className='flex justify-between items-center w-100 mt-6'>
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
              <Form.Item>
                <Button
                  htmlType='submit'
                  size='large'
                  type='primary'
                  loading={loading}
                >
                  Done
                </Button>
              </Form.Item>
            </div>
          </Form>
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
