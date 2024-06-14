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

const FactorsHubspotCompany = ({
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

  const [HSCompanyProps, SetHSCompanyProps] = useState([]);

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

  useEffect(() => {
    fetchHubspotCompanies();
  }, [isHubspotIntEnabled]);

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
              >{`Your credentials are encrypted & can be removed at any time. You can manage all of your connected accounts `}
              <a target="_blank" href='https://app.factors.ai/settings/integration'>here.</a></Text>
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
                  >{`Additional fields (Optional)`}</Text>
                  <MapComponent
                    dropdownOptions1={dropdownOptions}
                    dropdownOptions2={HSCompanyProps}
                    propertyMap={propertyMapAdditional}
                    setPropertyMap={setPropertyMapAdditional}
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
    logger.log('error inside FactorsHubspotCompany', err);
    return null;
  }
};

export default FactorsHubspotCompany;
