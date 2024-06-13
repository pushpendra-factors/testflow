import React, {
  useState,
  useEffect,
  useCallback,
  useRef,
  useMemo
} from 'react';
import { connect, useDispatch, useSelector } from 'react-redux';
import { Dropdown, Button, Input, Tag, Collapse, Select, Form } from 'antd';
import { Text, SVG } from 'factorsComponents';
import MapComponent from '../MapComponent';
import { paragon } from '@useparagon/connect/dist/src/index';
import isEmpty from 'lodash/isEmpty';
import logger from 'Utils/logger';

const FactorsApolloHubspotContacts = ({
  propertyMapMandatory,
  setPropertyMapMandatory,
  filterOptions,
  dropdownOptions,
  propertyMapAdditional,
  setPropertyMapAdditional,
  user,
  saveWorkflowFn,
  selectedTemp,
  isTemplate,
  setPropertyMapAdditional2,
  propertyMapAdditional2,
  apolloFormDetails,
  setApolloFormDetails
}) => {
  const { Panel } = Collapse;
  const [form] = Form.useForm();

  const [HSCompanyProps, SetHSCompanyProps] = useState([]);
  const [HSContactsProps, SetHSContactsProps] = useState([]);

  const isHubspotIntEnabled = user?.integrations?.hubspot?.enabled;

  const isHubspotInt = () => {
    if (isHubspotIntEnabled) {
      return <SVG name={'Check_circle'} size={22} color={'green'} />;
    } else return null;
  };

  const fetchHubspotCompanies = () => {
    if (user) {
      paragon
        .request('hubspot', '/crm/v3/properties/companies', {
          method: 'GET'
        })
        .then((response) => {
          let finalCompList = response.results?.filter((val) => !val?.hidden);

          let HSdropdownOptions = finalCompList?.map((item) => {
            return {
              label: item.label,
              value: item.name
            };
          });
          SetHSCompanyProps(HSdropdownOptions || []);
        })
        .catch((err) => {
          logger.log('fetchHubspotCompanies error', err);
        });
    }
  };

  const fetchHubspotContacts = () => {
    if (user) {
      paragon
        .request('hubspot', '/crm/v3/properties/contacts', {
          method: 'GET'
        })
        .then((response) => {
          let finalCompList = response.results?.filter((val) => !val?.hidden);

          let HSdropdownOptions = finalCompList?.map((item) => {
            return {
              label: item.label,
              value: item.name
            };
          });
          SetHSContactsProps(HSdropdownOptions || []);
        })
        .catch((err) => {
          logger.log('fetchHubspotCompanies error', err);
        });
    }
  };

  useEffect(() => {
    fetchHubspotCompanies();
    fetchHubspotContacts();
  }, [isHubspotIntEnabled]);

  useEffect(() => {
    if (selectedTemp && !isTemplate) {
      setPropertyMapMandatory(
        selectedTemp?.message_properties?.mandatory_properties
      );
      setPropertyMapAdditional(
        selectedTemp?.message_properties?.additional_properties_company
      );
      setPropertyMapAdditional2(
        selectedTemp?.message_properties?.additional_properties_contact
      );
      setApolloFormDetails(selectedTemp?.addtional_configuration?.[0]);
    }
  }, [selectedTemp]);

  const saveFormValidateApollo = () => {
    form.validateFields().then((value) => {
      setApolloFormDetails(value);
      saveWorkflowFn(value);
    });
  };

  try {
    return (
      <>
        <Collapse
          accordion
          bordered={false}
          defaultActiveKey={[isHubspotIntEnabled ? '2' : '1']}
        >
          <Panel
            header='Integrate Hubspot'
            className='bg-white'
            key='1'
            extra={isHubspotInt()}
          >
            <div className='flex flex-col p-4'>
              <Text
                type={'title'}
                level={7}
                color={'grey'}
                extraClass={'m-0 mb-2'}
              >{`Factors is a secure partner with Zapier. Your credentials are encrypted & can be removed at any time. You can manage all of your connected accounts here.`}</Text>
              <div className=''>
                <Button
                  // disabled={isHubspotIntEnabled}
                  icon={
                    isHubspotIntEnabled ? (
                      <SVG name={'Check_circle'} size={16} color={'green'} />
                    ) : (
                      ''
                    )
                  }
                  onClick={() => paragon.installIntegration('hubspot')}
                >
                  {isHubspotIntEnabled
                    ? 'Hubspot Connected'
                    : 'Connect Hubspot'}
                </Button>
              </div>
            </div>
          </Panel>

          <Panel
            header='Configurations'
            key='2'
            className='bg-white'
            disabled={!isHubspotIntEnabled}
          >
            <div className='flex p-4'>
              <div className='flex flex-col'>
                <Text
                  type={'title'}
                  weight={'bold'}
                  level={7}
                  color={'black'}
                  extraClass={'m-0'}
                >{`Mandatory fields`}</Text>
                <div className='flex justify-between items-center mt-4'>
                  <div className=''>
                    <Text
                      type={'title'}
                      level={8}
                      color={'black'}
                      extraClass={'m-0'}
                    >{`Factors Properties`}</Text>
                  </div>
                  <div className='mr-2 ml-2'></div>
                  <div className=''>
                    <Text
                      type={'title'}
                      level={8}
                      color={'black'}
                      extraClass={'m-0'}
                    >{`Hubspot Properties`}</Text>
                  </div>
                </div>
                <MapComponent
                  dropdownOptions1={dropdownOptions}
                  dropdownOptions2={HSCompanyProps}
                  propertyMap={propertyMapMandatory}
                  setPropertyMap={setPropertyMapMandatory}
                  limit={2}
                  isTemplate={isTemplate}
                />

                <div className='mt-6'>
                  <Text
                    type={'title'}
                    weight={'bold'}
                    level={7}
                    color={'black'}
                    extraClass={'m-0'}
                  >{`Additional fields (for Company)`}</Text>
                  <MapComponent
                    dropdownOptions1={dropdownOptions}
                    dropdownOptions2={HSCompanyProps}
                    propertyMap={propertyMapAdditional}
                    setPropertyMap={setPropertyMapAdditional}
                    isTemplate={isTemplate}
                  />
                </div>

                <div className='mt-6'>
                  <Text
                    type={'title'}
                    weight={'bold'}
                    level={7}
                    color={'black'}
                    extraClass={'m-0'}
                  >{`Apollo Configuration`}</Text>
                  <Form
                    form={form}
                    name='apollo'
                    className='w-full'
                    initialValues={apolloFormDetails}
                    onFinish={(val) => saveFormValidateApollo(val)}
                  >
                    <div className='mt-4'>
                      <Text
                        type={'title'}
                        weight={'thin'}
                        level={8}
                        extraClass={'m-0'}
                      >{`Apollo API key`}</Text>
                      <Form.Item
                        label={null}
                        name='ApiKey'
                        className='w-full'
                        rules={[
                          {
                            required: true,
                            message: 'Please enter API key'
                          }
                        ]}
                      >
                        <Input
                          onChange={(e) =>
                            setApolloFormDetails({
                              ...apolloFormDetails,
                              ApiKey: e.target.value
                            })
                          }
                          value={apolloFormDetails?.ApiKey}
                          className='fa-input w-full'
                          placeholder='API key'
                        />
                      </Form.Item>
                    </div>
                    <div className='mt-4'>
                      <Text
                        type={'title'}
                        weight={'thin'}
                        level={8}
                        extraClass={'m-0'}
                      >{`Job title list`}</Text>
                      <Form.Item
                        label={null}
                        name='PersonTitles'
                        className='w-full'
                      >
                        <Input
                          onChange={(e) =>
                            setApolloFormDetails({
                              ...apolloFormDetails,
                              PersonTitles: e.target.value
                            })
                          }
                          value={apolloFormDetails?.PersonTitles}
                          className='fa-input w-full'
                          placeholder={`Marketing,CEO,Founder`}
                        />
                      </Form.Item>
                    </div>
                    <div className='mt-4'>
                      <Text
                        type={'title'}
                        weight={'thin'}
                        level={8}
                        extraClass={'m-0'}
                      >{`Seniorities to include`}</Text>
                      <Form.Item
                        label={null}
                        name='PersonSeniorities'
                        className='w-full'
                      >
                        <Input
                          onChange={(e) =>
                            setApolloFormDetails({
                              ...apolloFormDetails,
                              PersonSeniorities: e.target.value
                            })
                          }
                          value={apolloFormDetails?.PersonSeniorities}
                          className='fa-input w-full'
                          placeholder={`manager,vp,c_suite,director`}
                        />
                      </Form.Item>
                    </div>
                    <div className='mt-4'>
                      <Text
                        type={'title'}
                        weight={'thin'}
                        level={8}
                        extraClass={'m-0'}
                      >{`Maximum number of contacts to enrich for a company`}</Text>
                      <Form.Item
                        label={null}
                        name='MaxContacts'
                        className='w-full'
                      >
                        <Input
                          onChange={(e) =>
                            setApolloFormDetails({
                              ...apolloFormDetails,
                              MaxContacts: e.target.value
                            })
                          }
                          value={apolloFormDetails?.MaxContacts}
                          className='fa-input w-full'
                          placeholder={`10`}
                        />
                      </Form.Item>
                    </div>
                  </Form>
                </div>
                <div className='mt-6'>
                  <Text
                    type={'title'}
                    weight={'bold'}
                    level={7}
                    color={'black'}
                    extraClass={'m-0'}
                  >{`Additional fields (for Contact)`}</Text>
                  <MapComponent
                    dropdownOptions1={dropdownOptions}
                    dropdownOptions2={HSContactsProps}
                    propertyMap={propertyMapAdditional2}
                    setPropertyMap={setPropertyMapAdditional2}
                    isTemplate={isTemplate}
                  />
                </div>
              </div>
            </div>
            <div className='border-top--thin-2 p-4 mt-4 flex items-center justify-end'>
              <Button
                type={'primary'}
                className='mt-2'
                onClick={() => saveWorkflowFn()}
              >
                Save and Publish
              </Button>
            </div>
          </Panel>
        </Collapse>
      </>
    );
  } catch (err) {
    logger.log('error inside FactorsApolloHubspotContacts', err);
    return null;
  }
};

export default FactorsApolloHubspotContacts;
