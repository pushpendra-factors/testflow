import React, {
  useState,
  useEffect,
  useCallback,
  useRef,
  useMemo
} from 'react';
import { connect, useDispatch, useSelector } from 'react-redux';
import {
  Dropdown,
  Button,
  Input,
  Tag,
  Collapse,
  Select,
  Form,
  message,
  notification
} from 'antd';
import { Text, SVG } from 'factorsComponents';
import { paragon } from '@useparagon/connect/dist/src/index';
import isEmpty from 'lodash/isEmpty';
import MapComponent from '../MapComponent';
import logger from 'Utils/logger';

const FactorsSalesforceCompany = ({
  propertyMapMandatory,
  setPropertyMapMandatory,
  filterOptions,
  dropdownOptions,
  propertyMapAdditional,
  setPropertyMapAdditional,
  user,
  saveWorkflowFn,
  selectedTemp,
  isTemplate
}) => {
  const { Panel } = Collapse;

  const [SFCompanyProps, SetSFCompanyProps] = useState([]);

  const isSFIntEnabled = user?.integrations?.salesforce?.enabled;

  const isSFInt = () => {
    if (isSFIntEnabled) {
      return <SVG name='Check_circle' size={22} color='green' />;
    }
    return null;
  };

  const fetchSFCompanies = () => {
    if (user) {
      paragon
        .request('salesforce', '/sobjects/Account/describe', {
          method: 'GET'
        })
        .then((response) => {
          const SFdropdownOptions = response?.fields?.map((item) => ({
            label: item.label,
            value: item.name,
            type: item.type
          }));
          SetSFCompanyProps(SFdropdownOptions || []);
        })
        .catch((err) => {
          logger.log('fetchSFCompanies error===>>>>>', err);
        });
    }
  };

  useEffect(() => {
    fetchSFCompanies();
  }, [isSFIntEnabled]);

  useEffect(() => {
    if (selectedTemp && !isTemplate) {
      setPropertyMapMandatory(
        selectedTemp?.message_properties?.mandatory_properties
      );
      setPropertyMapAdditional(
        selectedTemp?.message_properties?.additional_properties_company
      );
    }
  }, selectedTemp);

  try {
    return (
      <Collapse
        accordion
        bordered={false}
        defaultActiveKey={[isSFIntEnabled ? '2' : '1']}
      >
        <Panel
          header='Integrate Salesforce'
          className='bg-white'
          key='1'
          extra={isSFInt()}
        >
          <div className='flex flex-col p-4'>
            <Text type='title' level={7} color='grey' extraClass='m-0 mb-2'>
            {`Your credentials are encrypted & can be removed at any time. You can manage all of your connected accounts `}
            <a target="_blank" href='https://app.factors.ai/settings/integration'>here.</a></Text>
            <div className=''>
              <Button
                // disabled={isSFIntEnabled}
                icon={
                  isSFIntEnabled ? (
                    <SVG name='Check_circle' size={16} color='green' />
                  ) : (
                    ''
                  )
                }
                onClick={() => paragon.installIntegration('salesforce')}
              >
                {isSFIntEnabled ? 'Salesforce Connected' : 'Connect Salesforce'}
              </Button>
            </div>
          </div>
        </Panel>
        <Panel
          header='Configurations'
          key='2'
          className='bg-white'
          disabled={!isSFIntEnabled}
        >
          <div className='flex p-4'>
            <div className='flex flex-col'>
              <Text
                type='title'
                weight='bold'
                level={7}
                color='black'
                extraClass='m-0'
              >
                Mandatory fields
              </Text>
              <div className='flex justify-between items-center mt-4'>
                <div className=''>
                  <Text type='title' level={8} color='black' extraClass='m-0'>
                    Factors Properties
                  </Text>
                </div>
                <div className='mr-2 ml-2' />
                <div className=''>
                  <Text type='title' level={8} color='black' extraClass='m-0'>
                    Salesforce Properties
                  </Text>
                </div>
              </div>
              <MapComponent
                dropdownOptions1={dropdownOptions}
                dropdownOptions2={SFCompanyProps}
                propertyMap={propertyMapMandatory}
                setPropertyMap={setPropertyMapMandatory}
                limit={2}
                isTemplate={isTemplate}
              />
              <div className='mt-6'>
                <Text
                  type='title'
                  weight='bold'
                  level={7}
                  color='black'
                  extraClass='m-0'
                >
                  Additional fields (Optional)
                </Text>
                <MapComponent
                  dropdownOptions1={dropdownOptions}
                  dropdownOptions2={SFCompanyProps}
                  propertyMap={propertyMapAdditional}
                  setPropertyMap={setPropertyMapAdditional}
                  isTemplate={isTemplate}
                />
              </div>
            </div>
          </div>
          <div className='border-top--thin-2 p-4 mt-4 flex items-center justify-end'>
            <Button
              type='primary'
              className='mt-2'
              onClick={() => saveWorkflowFn()}
            >
              Save and Publish
            </Button>
          </div>
        </Panel>
      </Collapse>
    );
  } catch (err) {
    logger.log('error inside FactorsSalesforceCompany', err);
    return null;
  }
};

export default FactorsSalesforceCompany;
