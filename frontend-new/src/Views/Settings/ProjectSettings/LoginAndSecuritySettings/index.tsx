import CommonSettingsHeader from 'Components/GenericComponents/CommonSettingsHeader';
import { Text } from 'Components/factorsComponents';
import {
  Button,
  Card,
  Divider,
  Form,
  Input,
  Modal,
  Switch,
  Tooltip,
  message
} from 'antd';
import React, { useCallback, useEffect, useMemo, useState } from 'react';
import {
  CopyOutlined,
  EditOutlined,
  ExclamationCircleOutlined
} from '@ant-design/icons';
import { CopyTextToClipboard } from 'Utils/global';
import { ConnectedProps, connect, useSelector } from 'react-redux';
import { udpateProjectSettings, fetchProjectSettings } from 'Reducers/global';
import { bindActionCreators } from 'redux';
import {
  InitialLoginMethodMap,
  InitialLoginMethodProperties,
  LoginMethodTypes,
  SettingsLoginAndSecurityCardProps
} from './types';
import { getBackendHost } from '../IntegrationSettings/util';
import logger from 'Utils/logger';

const host = getBackendHost();

function SecurityCard(props: SettingsLoginAndSecurityCardProps) {
  const {
    title,
    description,
    loginMethodState = InitialLoginMethodMap,
    loading = false,
    documentationLink = null,
    methodName,
    onMethodChange,
    methodProperties,
    isAdmin = false,
    updateMethodProperties
  } = props;

  const [enable, setEnable] = useState(false);
  const [mode, setMode] = useState<'edit' | 'view'>('view');
  const active_project = useSelector(
    (state: any) => state?.global?.active_project
  );

  const onSwitchChange = useCallback(
    (state) => {
      onMethodChange(state, methodName);
    },
    [onMethodChange]
  );

  useEffect(() => {
    if (methodName in loginMethodState) {
      setEnable(loginMethodState[methodName]);
      if (loginMethodState[methodName] === false) {
        setMode('view');
      }
    }
  }, [loginMethodState]);

  const handleEditHandle = useCallback(() => {
    setMode('edit');
  }, []);

  const handleFormSubmit = useCallback((values) => {
    Modal.confirm({
      title: 'Do you Want to update fields?',
      icon: <ExclamationCircleOutlined />,
      content: "You'll get logged out once you enable this login-method",
      okButtonProps: { style: { padding: '0 12px' } },
      className: 'fa-modal--regular',
      onOk() {
        if (!values?.login_url || !values?.certificate) {
          message.error('Please give us the Login URL and Certificate');
          return;
        }
        if (updateMethodProperties)
          updateMethodProperties(values, true, methodName);
        setMode('view');
      }
    });
  }, []);

  const handleCopyButton = useCallback((text) => {
    CopyTextToClipboard(text);
  }, []);
  return (
    <div className='fa-login-security-card'>
      <div>
        <div>
          <Text type='title' level={5} extraClass='mb-1'>
            {title}
          </Text>
          <Text color='grey' type='title' level={7}>
            {description}{' '}
            {documentationLink && (
              <a href={documentationLink}>View Documentation</a>
            )}
          </Text>
        </div>
        <div>
          <Text type='title' extraClass='mb-0'>
            Enable
          </Text>{' '}
          <Tooltip
            placement='left'
            title={
              isAdmin === false
                ? `Only project admins can make a change to project's login settings.`
                : ''
            }
          >
            {' '}
            <Switch
              disabled={isAdmin === false}
              checked={enable}
              onChange={onSwitchChange}
            />
          </Tooltip>
        </div>
      </div>

      {isAdmin && methodProperties && methodProperties.length > 0 && enable && (
        <div className='fa-container'>
          <Form layout='vertical' onFinish={handleFormSubmit}>
            {methodProperties &&
              methodProperties.map(
                (eachProperty: {
                  id: React.Key | null | undefined;
                  label: string;
                  required: boolean | undefined;
                  type: string;
                  placeholder: string | undefined;
                  copy: boolean;
                  hint: string;
                  name: string;
                  value: string;
                }) => (
                  <Input.Group className='w-full' key={eachProperty.id}>
                    <Form.Item
                      label={eachProperty.label}
                      required={eachProperty.required}
                      style={{ padding: '8px 0' }}
                      className='w-2/5'
                      name={eachProperty?.name}
                      initialValue={
                        eachProperty.name === 'acs_url'
                          ? `${host}/project/${active_project?.id}/saml/acs`
                          : eachProperty?.value
                      }
                      extra={eachProperty.hint}
                    >
                      {eachProperty.type === 'input' ? (
                        <Input
                          className='fa-input'
                          type='text'
                          style={{ borderRadius: '8px' }}
                          placeholder={
                            eachProperty.name === 'acs_url'
                              ? `${host}/project/${active_project?.id}/saml/acs`
                              : eachProperty.placeholder
                          }
                          disabled={eachProperty.copy || mode === 'view'}
                          suffix={
                            eachProperty.copy && (
                              <Tooltip title={`Copy ${eachProperty.label}`}>
                                <CopyOutlined
                                  style={{ cursor: 'pointer' }}
                                  onClick={() =>
                                    handleCopyButton(eachProperty.placeholder)
                                  }
                                />
                              </Tooltip>
                            )
                          }
                        />
                      ) : (
                        <Input.TextArea
                          placeholder={eachProperty.placeholder}
                          style={{ borderRadius: '8px' }}
                          disabled={eachProperty.copy || mode === 'view'}
                          className='fa-input'
                        />
                      )}
                    </Form.Item>
                  </Input.Group>
                )
              )}
            <Divider style={{ margin: '8px 0' }} />
            <div className='flex justify-end p-1 px-3'>
              {mode === 'edit' ? (
                <Form.Item>
                  <Button
                    size='large'
                    icon={<EditOutlined />}
                    htmlType='submit'
                  >
                    Save
                  </Button>
                </Form.Item>
              ) : (
                ''
              )}
            </div>
          </Form>
          {mode === 'view' && (
            <div className='flex justify-end p-1 px-3'>
              <Button
                size='large'
                icon={<EditOutlined />}
                onClick={handleEditHandle}
              >
                Edit
              </Button>
            </div>
          )}
        </div>
      )}
    </div>
  );
}

function LoginAndSecuritySettings({
  udpateProjectSettings,
  fetchProjectSettings
}) {
  const agentState = useSelector((state: any) => state.agent);
  const activeAgent = agentState?.agent_details?.email;
  const isAdmin = useMemo(
    () => agentState?.agents.find((e) => e?.email === activeAgent)?.role === 2,
    [agentState]
  );

  const project_settings = useSelector(
    (state: any) => state?.global?.currentProjectSettings
  );

  const active_project = useSelector(
    (state: any) => state?.global?.active_project
  );

  const [loginMethodState, setLoginMethodState] = useState(
    InitialLoginMethodMap
  );

  const [methodProperties, setMethodProperties] = useState(
    InitialLoginMethodProperties
  );

  const updateMethodProperties = async (
    updatedPayload: { certificate: any; login_url: any } | null,
    enabled_state: boolean,
    methodName: string
  ) => {
    if (
      methodName === 'saml' &&
      enabled_state === true &&
      updatedPayload === null
    ) {
      if (!updatedPayload?.login_url || !updatedPayload?.certificate)
        message.error('Please fill the required fields');
      return;
    }
    const messageLoadingHandle = message.loading('Updating SAML Settings', 0);
    try {
      const tmpPayload = {};
      const methodNameState =
        methodName === 'saml' ? 2 : methodName === 'google' ? 3 : 0;

      if (enabled_state === true && methodNameState === 2) {
        tmpPayload.saml_configuration = {
          ...updatedPayload,
          created_by: `${agentState?.agent_details?.first_name} ${agentState?.agent_details?.last_name}`
        };
        delete tmpPayload.saml_configuration.acs_url;
      }

      tmpPayload.sso_state = enabled_state === true ? methodNameState : 1;

      await udpateProjectSettings(active_project?.id, tmpPayload);
      await fetchProjectSettings(active_project?.id);

      message.success('SAML Settings updated');
    } catch (error) {
      logger.error('error', error);
      message.error(error?.data?.error || 'SAML Settings Failed to update');
    } finally {
      messageLoadingHandle();
    }
  };

  const loginMethodChangeFn = (
    changeState: boolean,
    methodName: LoginMethodTypes
  ) => {
    loginMethodState[methodName] = changeState;
    Object.keys(loginMethodState).forEach((eachMethod) => {
      if (eachMethod !== methodName) loginMethodState[eachMethod] = false;
    });

    updateMethodProperties(null, changeState, methodName);
    setLoginMethodState({ ...loginMethodState });
  };

  const onLoginMethodChange = (
    changeState: boolean,
    methodName: LoginMethodTypes
  ) => {
    Modal.confirm({
      title: `Are you sure about changing login method to ${methodName} ?`,
      icon: <ExclamationCircleOutlined />,
      content:
        changeState &&
        methodName === 'google' &&
        "You'll get logged out once you enable this login-method",
      okButtonProps: { style: { padding: '0 12px' } },
      className: 'fa-modal--regular',
      onOk() {
        loginMethodChangeFn(changeState, methodName);
      }
    });
  };

  useEffect(() => {
    if (project_settings) {
      setLoginMethodState((prev) => {
        const tmp = { ...prev };
        const ssoName: any =
          project_settings.sso_state == 2
            ? 'saml'
            : project_settings.sso_state == 3
              ? 'google'
              : '';

        Object.keys(prev).forEach((eachKey) => {
          if (ssoName === eachKey) tmp[eachKey] = true;
          else tmp[eachKey] = false;
        });
        return tmp;
      });

      setMethodProperties((prevMap) => {
        const tmp = { ...prevMap };
        Object.keys(tmp).forEach((eachKey) => {
          if (eachKey === 'saml') {
            tmp[eachKey].forEach((eachProperty, ei) => {
              tmp[eachKey][ei].value =
                project_settings.saml_configuration?.[eachProperty.name];
            });
          }
        });

        return tmp;
      });
    }
  }, [project_settings]);

  return (
    <div className='fa-container'>
      <CommonSettingsHeader
        title='Login and Security'
        description='Configure project wide login settings for users. Choose compulsory login through one of these methods.'
      />
      <SecurityCard
        title='Mandatory login with Google SSO'
        description='Require all project members to login through Google SSO to.'
        loginMethodState={loginMethodState}
        methodName='google'
        onMethodChange={onLoginMethodChange}
        methodProperties={methodProperties['google']}
        isAdmin={isAdmin}
        updateMethodProperties={updateMethodProperties}
      />
      <SecurityCard
        title='Mandatory login with SAML credentials'
        description='Require all project members to login through their SAML identity
            provider credentials.'
        documentationLink='https://api.factors.ai'
        loginMethodState={loginMethodState}
        methodName='saml'
        onMethodChange={onLoginMethodChange}
        methodProperties={methodProperties['saml']}
        isAdmin={isAdmin}
        updateMethodProperties={updateMethodProperties}
      />
    </div>
  );
}

const mapDispatchToProps = (dispatch: any) =>
  bindActionCreators(
    {
      udpateProjectSettings,
      fetchProjectSettings
    },
    dispatch
  );
const connector = connect(null, mapDispatchToProps);
type ThirdPartyStepsBodyProps = ConnectedProps<typeof connector>;

export default connector(LoginAndSecuritySettings);
