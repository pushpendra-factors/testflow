import { Input } from 'antd';
import TextArea from 'antd/lib/input/TextArea';
import { getBackendHost } from '../IntegrationSettings/util';

export type SettingsLoginAndSecurityCardProps = {
  title: string | JSX.Element;
  description: string | JSX.Element;
  loginMethodState: typeof InitialLoginMethodMap;
  loading?: boolean;
  documentationLink?: null | string;
  methodName: LoginMethodTypes;
  onMethodChange: (state: boolean, type: LoginMethodTypes) => void;
  methodProperties: typeof LoginMethodProperties | undefined;
  isAdmin: boolean;
  updateMethodProperties?: (payload) => void;
};

export type LoginMethodTypes = 'google' | 'saml' | '';

export const InitialLoginMethodMap = {
  google: false,
  saml: true,
  '': false
};
const LoginMethodProperties = [
  {
    id: 1,
    label: 'ACS URL',
    type: 'input',
    value: '',
    placeholder: ``,
    hint: 'You can use this for getting your Login URL',
    required: false,
    copy: true,
    name: 'acs_url'
  },
  {
    id: 2,
    label: 'Login URL',
    type: 'input',
    value: '',
    placeholder: 'http://localhost:11880/settings/sso/login...',
    hint: 'Team members will be redirected to this URL to login',
    required: true,
    copy: false,
    name: 'login_url'
  },
  {
    id: 3,
    label: 'SAML Certificate',
    type: 'textarea',
    value: '',
    placeholder:
      '<ds:X509Certificate>MIIDpDCCAoygAWIBAgIGAVL66FjyMA0GCSqGSIb3DQEBBQUAMIGSMQ3WC...',
    hint: 'It is required to authenticate your identity provider',
    required: true,
    copy: false,
    name: 'certificate'
  },
  {
    id: 4,
    label: 'Remote Logout URL',
    type: 'input',
    value: '',
    placeholder: 'http://localhost:11880/settings/sso/logout',
    hint: 'It is required to authenticate your identity provider',
    required: false,
    copy: false,
    name: 'logout_url'
  }
];
export const InitialLoginMethodProperties = {
  google: [],
  saml: LoginMethodProperties
};
