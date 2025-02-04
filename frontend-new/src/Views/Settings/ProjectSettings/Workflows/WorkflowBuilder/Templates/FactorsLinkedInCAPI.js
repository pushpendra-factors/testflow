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
  Tooltip
} from 'antd';
import { Text, SVG } from 'factorsComponents';
import MapComponent from '../MapComponent';
import isEmpty from 'lodash/isEmpty';
import logger from 'Utils/logger';
import { fetchConversionAPIData, fetchProjectSettings } from 'Reducers/global';

const FactorsLinkedInCAPI = ({
  propertyMapMandatory,
  setPropertyMapMandatory,
  user,
  saveWorkflowFn,
  selectedTemp,
  isTemplate,
  fetchProjectSettings,
  fetchConversionAPIData
}) => {
  const { Panel } = Collapse;

  const [DropdownProps, SetDropdownProps] = useState([]);
  const [conversionData, SetConversionData] = useState([]);
  const [selectedProps, SetSelectedProps] = useState('');

  const { active_project: activeProject, currentProjectSettings } = useSelector(
    (state) => state.global
  );

  useEffect(() => {
    if (activeProject?.id) {
      fetchProjectSettings(activeProject.id);
    }
  }, [activeProject?.id]);

  useEffect(() => {
    if (selectedTemp && !isTemplate) {
      setPropertyMapMandatory(selectedTemp?.addtional_configuration);
      SetSelectedProps(
        selectedTemp?.addtional_configuration?.conversions?.elements?.[0]?.name
      );
    }
  }, [selectedTemp]);

  const fetchDropdownData = () => {
    fetchConversionAPIData(activeProject?.id)
      .then((res) => {
        const dropdownOptions = res?.data?.elements?.map((item) => {
          return {
            value: item.id,
            label: item.name
          };
        });
        SetDropdownProps(dropdownOptions);
        SetConversionData(res?.data?.elements);
      })
      .catch((err) => logger.log('fetch conversion API data error=>', err));
  };

  useEffect(() => {
    if (currentProjectSettings?.int_linkedin_access_token) {
      fetchDropdownData();
    }
  }, [currentProjectSettings]);

  const handleChange = (id) => {
    SetSelectedProps(id);
    const data = conversionData.filter((val) => val?.id === id);
    setPropertyMapMandatory(data);
  };

  const renderLinkedinLogin = () => {
    if (!currentProjectSettings?.int_linkedin_access_token) {
      const { hostname } = window.location;
      const { protocol } = window.location;
      const { port } = window.location;
      let redirect_uri = `${protocol}//${hostname}:${port}`;
      if (port === undefined || port === '') {
        redirect_uri = `${protocol}//${hostname}`;
      }

      const href = `https://www.linkedin.com/oauth/v2/authorization?response_type=code&client_id=${BUILD_CONFIG.linkedin_client_id}&redirect_uri=${redirect_uri}&state=factors&scope=r_basicprofile%20r_liteprofile%20r_ads_reporting%20rw_ads%20rw_conversions%20rw_dmp_segments`;
      return (
        <div
          className='flex justify-center items-center mt-4'
          style={{
            backgroundColor: '#FAFAFA',
            width: '800px',
            height: '200px'
          }}
        >
          <div>
            <Text
              type={'title'}
              level={7}
              color={'grey'}
              extraClass={'m-0 italic'}
            >
              Please connect to LinkedIn to set this up
            </Text>
            <div className='ml-8'>
              <Button
                className='m-0 mr-2'
                icon={<SVG name='Linkedin_ads'></SVG>}
                href={href}
                target='_blank'
              >
                Connect
              </Button>
              <Button
                className='m-0'
                icon={<SVG name='SyncAlt'></SVG>}
                onClick={() => fetchProjectSettings(activeProject.id)}
              >
                Refresh
              </Button>
            </div>
          </div>
        </div>
      );
    }
  };

  try {
    return (
      <>
        <Collapse accordion bordered={false} defaultActiveKey={['1']}>
          <Panel header='LinkedIn configuration' key='1' className='bg-white'>
            <div className='flex p-4 m-2'>
              <div className='flex flex-col'>
                <Text
                  type={'title'}
                  level={7}
                  color={'gray'}
                  extraClass={'m-0'}
                >
                  Select the Conversion from linked to which you want to push
                  the accounts
                  <Tooltip
                    placement='top'
                    title={`If you don’t find a conversion here, please create a new one with data source set as ’Direct API’ inside your LinkedIn account manager `}
                  >
                    <div className='inline ml-1'>
                      <SVG
                        name='InfoCircle'
                        size={16}
                        color='#8C8C8C'
                        extraClass='inline'
                      />
                    </div>
                  </Tooltip>
                </Text>
                {renderLinkedinLogin()}
                {currentProjectSettings?.int_linkedin_access_token && (
                  <div className='mt-4'>
                    <Select
                      options={DropdownProps}
                      onChange={handleChange}
                      style={{ width: 800 }}
                      showSearch
                      placeholder='Select property'
                      optionFilterProp='label'
                      className='fa-select ml-4'
                      value={selectedProps}
                    />
                  </div>
                )}
                <Text
                  type={'title'}
                  level={7}
                  color={'grey'}
                  extraClass={'m-0 mt-2 mb-2'}
                >
                  {`Your credentials are encrypted & can be removed at any time. You can manage all of your connected accounts `}
                  <a
                    target='_blank'
                    href='https://app.factors.ai/settings/integration'
                  >
                    here.
                  </a>
                </Text>
              </div>
            </div>
            <div className='p-4 mt-4 flex items-center justify-end'>
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
    logger.log('error inside FactorsLinkedInCAPI', err);
    return null;
  }
};

const mapStateToProps = (state) => ({
  currentAgent: state.agent.agent_details
});

export default connect(mapStateToProps, {
  fetchProjectSettings,
  fetchConversionAPIData
})(FactorsLinkedInCAPI);
