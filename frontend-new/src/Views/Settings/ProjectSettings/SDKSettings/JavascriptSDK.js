import React, { useState, useEffect, useMemo } from 'react';
import {
  Row,
  Col,
  Skeleton,
  Tabs,
  Switch,
  message,
  Table,
  Button,
  Modal,
  Input,
  Spin,
  Divider,
  notification
} from 'antd';
import { Text, SVG } from 'factorsComponents';
import {
  fetchBingAdsIntegration,
  fetchMarketoIntegration,
  fetchProjectSettingsV1,
  udpateProjectSettings
} from 'Reducers/global';
import { connect } from 'react-redux';

import { Link } from 'react-router-dom';
import {
  fetchClickableElements,
  toggleClickableElement
} from '../../../../reducers/settings/middleware';
import MomentTz from '../../../../components/MomentTz';
import ExcludeIp from '../BasicSettings/IpBlocking/excludeIp';

import GtmSteps from './InstructionSteps/gtmSteps';
import ManualSteps from './InstructionSteps/manualSteps';
import { JavascriptHeadDocumentation } from './utils';

const { TabPane } = Tabs;

const JSConfig = ({
  currentProjectSettings,
  activeProject,
  udpateProjectSettings,
  agents,
  currentAgent
}) => {
  const [enableEdit, setEnableEdit] = useState(false);
  const [autoTrack, setAutoTrack] = useState(false);
  const [autoFormCapture, setAutoFormCapture] = useState(false);
  const [autoCaptureFormFills, setAutoCaptureFormFills] = useState(false);
  const [autoTrackSPAPageView, setAutoTrackSPAPageView] = useState(false);
  const [excludeBot, setExcludeBot] = useState(false);
  const [clickCapture, setClickCapture] = useState(false);

  const currentProjectId = activeProject.id;

  useEffect(() => {
    setEnableEdit(false);
    agents &&
      currentAgent &&
      agents.map((agent) => {
        if (agent.uuid === currentAgent.uuid && agent.role === 1)
          setEnableEdit(true);
      });
  }, [activeProject, agents, currentAgent]);

  useEffect(() => {
    if (currentProjectSettings.auto_track) {
      setAutoTrack(true);
    }
    if (currentProjectSettings.auto_track_spa_page_view) {
      setAutoTrackSPAPageView(true);
    }
    if (currentProjectSettings.exclude_bot) {
      setExcludeBot(true);
    }
    if (currentProjectSettings.auto_form_capture) {
      setAutoFormCapture(true);
    }
    if (currentProjectSettings.auto_capture_form_fills) {
      setAutoCaptureFormFills(true);
    }
    if (currentProjectSettings.auto_click_capture) {
      setClickCapture(true);
    }
  }, [currentProjectSettings]);

  const toggleAutoTrack = (checked) => {
    if (!checked) {
      Modal.confirm({
        title: 'Are you sure you want to disable this?',
        content:
          'Doing this will stop Factors from tracking standard events such as page_view, page_load time, page_spent_time and more for each user',
        okText: 'Disable Auto Track',
        cancelText: 'Cancel',
        onOk: () => {
          udpateProjectSettings(currentProjectId, { auto_track: checked })
            .then(() => {
              setAutoTrack(false);
            })
            .catch((err) => {
              console.log('Oops! something went wrong-->', err);
              message.error('Oops! something went wrong.');
            });
        },
        onCancel: () => {
          setAutoTrack(!checked);
        }
      });
    } else {
      udpateProjectSettings(currentProjectId, { auto_track: checked }).catch(
        (err) => {
          console.log('Oops! something went wrong-->', err);
          message.error('Oops! something went wrong.');
        }
      );
    }
  };

  const toggleExcludeBot = (checked) => {
    if (!checked) {
      Modal.confirm({
        title: 'Are you sure you want to disable this?',
        content:
          'Doing this will stop Factors from automatically excluding bot traffic from website traffic using Factor’s proprietary algorithm',
        okText: 'Disable Exclude Bot',
        cancelText: 'Cancel',
        onOk: () => {
          udpateProjectSettings(currentProjectId, { exclude_bot: checked })
            .then(() => {
              setExcludeBot(false);
            })
            .catch((err) => {
              console.log('Oops! something went wrong-->', err);
              message.error('Oops! something went wrong.');
            });
        },
        onCancel: () => {
          setExcludeBot(!checked);
        }
      });
    } else {
      udpateProjectSettings(currentProjectId, { exclude_bot: checked }).catch(
        (err) => {
          console.log('Oops! something went wrong-->', err);
          message.error('Oops! something went wrong.');
        }
      );
    }
  };

  const toggleAutoFormCapture = (checked) => {
    if (!checked) {
      Modal.confirm({
        title: 'Are you sure you want to disable this?',
        content:
          'Doing this will stop Factors from automatically tracking personal identification information such as email and phone number from Form Submissions',
        okText: 'Disable Auto Form Capture',
        cancelText: 'Cancel',
        onOk: () => {
          udpateProjectSettings(currentProjectId, {
            auto_form_capture: checked
          })
            .then(() => {
              setAutoFormCapture(false);
            })
            .catch((err) => {
              console.log('Oops! something went wrong-->', err);
              message.error('Oops! something went wrong.');
            });
        },
        onCancel: () => {
          setAutoFormCapture(!checked);
        }
      });
    } else {
      udpateProjectSettings(currentProjectId, {
        auto_form_capture: checked
      }).catch((err) => {
        console.log('Oops! something went wrong-->', err);
        message.error('Oops! something went wrong.');
      });
    }
  };

  const toggleAutoCaptureFormFills = (checked) => {
    if (!checked) {
      Modal.confirm({
        title: 'Are you sure you want to disable this?',
        content:
          'Doing this will stop Factors from automatically tracking personal identification information such as email and phone number from Form Fills',
        okText: 'Disable Auto Capture Form Fills',
        cancelText: 'Cancel',
        onOk: () => {
          udpateProjectSettings(currentProjectId, {
            auto_capture_form_fills: checked
          })
            .then(() => {
              setAutoCaptureFormFills(false);
            })
            .catch((err) => {
              console.log('Oops! something went wrong-->', err);
              message.error('Oops! something went wrong.');
            });
        },
        onCancel: () => {
          setAutoCaptureFormFills(!checked);
        }
      });
    } else {
      udpateProjectSettings(currentProjectId, {
        auto_capture_form_fills: checked
      }).catch((err) => {
        console.log('Oops! something went wrong-->', err);
        message.error('Oops! something went wrong.');
      });
    }
  };

  const toggleAutoTrackSPAPageView = (checked) => {
    if (!checked) {
      Modal.confirm({
        title: 'Are you sure you want to disable this?',
        content:
          'Doing this will stop Factors from tracking standard events such as page_view, page_load time, page_spent_time and button clicks for each user, on Single Page Applications like React, Angular, Vue, etc,.',
        okText: 'Disable Auto Track SPA',
        cancelText: 'Cancel',
        onOk: () => {
          udpateProjectSettings(currentProjectId, {
            auto_track_spa_page_view: checked
          })
            .then(() => {
              setAutoTrackSPAPageView(false);
            })
            .catch((err) => {
              console.log('Oops! something went wrong-->', err);
              message.error('Oops! something went wrong.');
            });
        },
        onCancel: () => {
          setAutoTrackSPAPageView(!checked);
        }
      });
    } else {
      udpateProjectSettings(currentProjectId, {
        auto_track_spa_page_view: checked
      }).catch((err) => {
        console.log('Oops! something went wrong-->', err);
        message.error('Oops! something went wrong.');
      });
    }
  };

  const toggleClickCapture = (checked) => {
    if (!checked) {
      Modal.confirm({
        title: 'Are you sure you want to disable this?',
        content:
          'Doing this will stop Factors from discovering available buttons and anchors on the website.',
        okText: 'Disable Click Capture',
        cancelText: 'Cancel',
        onOk: () => {
          udpateProjectSettings(currentProjectId, {
            auto_click_capture: checked
          })
            .then(() => {
              setClickCapture(false);
            })
            .catch((err) => {
              console.log('Oops! something went wrong-->', err);
              message.error('Oops! something went wrong.');
            });
        },
        onCancel: () => {
          setClickCapture(!checked);
        }
      });
    } else {
      udpateProjectSettings(currentProjectId, {
        auto_click_capture: checked
      }).catch((err) => {
        console.log('Oops! something went wrong-->', err);
        message.error('Oops! something went wrong.');
      });
    }
  };

  return (
    <Row>
      {enableEdit && (
        <Col span={24}>
          <Text type='title' level={7} color='grey' extraClass='m-0 my-2'>
            *Only Admin(s) can change configurations.
          </Text>
        </Col>
      )}
      <Col span={24}>
        <div span={24} className='flex flex-start items-center mt-2'>
          <span style={{ width: '50px' }}>
            <Switch
              checkedChildren='On'
              disabled={enableEdit}
              unCheckedChildren='OFF'
              onChange={toggleAutoTrack}
              checked={autoTrack}
            />
          </span>
          <Text type='title' level={6} weight='bold' extraClass='m-0 ml-2'>
            Auto-track
          </Text>
        </div>
      </Col>
      <Col span={24} className='flex flex-start items-center'>
        <Text type='paragraph' mini extraClass='m-0 mt-2' color='grey'>
          Automatically track standard events on your website like start of a
          website session, page views and properties associated with these
          events like landing page URL, page load time, page scroll percent etc.
        </Text>
      </Col>
      <Col span={24}>
        <div span={24} className='flex flex-start items-center mt-8'>
          <span style={{ width: '50px' }}>
            <Switch
              checkedChildren='On'
              disabled={enableEdit}
              unCheckedChildren='OFF'
              onChange={toggleAutoTrackSPAPageView}
              checked={autoTrackSPAPageView}
            />
          </span>
          <Text type='title' level={6} weight='bold' extraClass='m-0 ml-2'>
            Auto-track Single Page Application
          </Text>
        </div>
      </Col>
      <Col span={24} className='flex flex-start items-center'>
        <Text type='paragraph' mini extraClass='m-0 mt-2' color='grey'>
          Track standard events on single page applications like React, Angular,
          Vue, etc.
        </Text>
      </Col>
      <Col span={24}>
        <div span={24} className='flex flex-start items-center mt-8'>
          <span style={{ width: '50px' }}>
            <Switch
              checkedChildren='On'
              disabled={enableEdit}
              unCheckedChildren='OFF'
              onChange={toggleExcludeBot}
              checked={excludeBot}
            />
          </span>
          <Text type='title' level={6} weight='bold' extraClass='m-0 ml-2'>
            Exclude Bot
          </Text>
        </div>
      </Col>
      <Col span={24} className='flex flex-start items-center'>
        <Text type='paragraph' mini extraClass='m-0 mt-2' color='grey'>
          Automatically exclude bot traffic from website traffic using Factor’s
          proprietary algorithm
        </Text>
      </Col>
      <Col span={24}>
        <div span={24} className='flex flex-start items-center mt-8'>
          <span style={{ width: '50px' }}>
            <Switch
              checkedChildren='On'
              disabled={enableEdit}
              unCheckedChildren='OFF'
              onChange={toggleAutoFormCapture}
              checked={autoFormCapture}
            />
          </span>
          <Text type='title' level={6} weight='bold' extraClass='m-0 ml-2'>
            Auto Form Capture
          </Text>
        </div>
      </Col>
      <Col span={24} className='flex flex-start items-center'>
        <Text type='paragraph' mini extraClass='m-0 mt-2' color='grey'>
          Automatically record personal identification information such as email
          and phone number whenever someone submits a form on your website
        </Text>
      </Col>
      <Col span={24}>
        <div span={24} className='flex flex-start items-center mt-8'>
          <span style={{ width: '50px' }}>
            <Switch
              checkedChildren='On'
              disabled={
                enableEdit || currentAgent.email === 'solutions@factors.ai'
              }
              unCheckedChildren='OFF'
              onChange={toggleAutoCaptureFormFills}
              checked={autoCaptureFormFills}
            />
          </span>
          <Text type='title' level={6} weight='bold' extraClass='m-0 ml-2'>
            Auto Form Fill Capture
          </Text>
        </div>
      </Col>
      <Col span={24} className='flex flex-start items-center'>
        <Text type='paragraph' mini extraClass='m-0 mt-2' color='grey'>
          Automatically track personal identification information such as email
          and phone number whenever someone fills a form field on your website,
          even if they do not submit the form.
        </Text>
      </Col>
      <Col span={24}>
        <div span={24} className='flex flex-start items-center mt-8'>
          <span style={{ width: '50px' }}>
            <Switch
              checkedChildren='On'
              disabled={enableEdit}
              unCheckedChildren='OFF'
              onChange={toggleClickCapture}
              checked={clickCapture}
            />
          </span>
          <Text type='title' level={6} weight='bold' extraClass='m-0 ml-2'>
            Auto Click Capture
          </Text>
        </div>
      </Col>
      <Col span={24} className='flex flex-start items-center'>
        <Text type='paragraph' mini extraClass='m-0 mt-2' color='grey'>
          Automatically capture button clicks take place on your website. Once
          tracked, they will be available under the Click tracking configuration
          tab and can be turned on to be tracked as events.
        </Text>
      </Col>
      <Col span={24} className='m-0 mt-8'>
        <ExcludeIp />
      </Col>
    </Row>
  );
};

const ClickTrackConfiguration = ({
  activeProject,
  agents,
  currentAgent,
  clickableElements,
  toggleClickableElement
}) => {
  const [enableEdit, setEnableEdit] = useState(false);
  const [searchTerm, setSearchTerm] = useState('');
  const [showSearch, setShowSearch] = useState(false);
  const [listData, setListData] = useState([]);

  useEffect(() => {
    setEnableEdit(false);
    agents &&
      currentAgent &&
      agents.forEach((agent) => {
        if (agent.uuid === currentAgent.uuid && agent.role === 1)
          setEnableEdit(true);
      });
  }, [activeProject, agents, currentAgent]);

  const headerClassStr =
    'fai-text fai-text__color--grey-2 fai-text__size--h8 fai-text__weight--bold';

  const columns = [
    {
      title: <span className={headerClassStr}>Name</span>,
      dataIndex: 'displayName',
      key: 'displayName',
      width: 300,
      ellipsis: true
    },
    {
      title: <span className={headerClassStr}>Type</span>,
      dataIndex: 'type',
      key: 'type',
      sorter: (a, b) => (a.type > b.type ? 1 : b.type > a.type ? -1 : 0)
    },
    {
      title: <span className={headerClassStr}>Clicks</span>,
      dataIndex: 'clickCount',
      key: 'clickCount',
      sorter: (a, b) => a.clickCount - b.clickCount
    },
    {
      title: <span className={headerClassStr}>Received At</span>,
      dataIndex: 'createdAt',
      key: 'createdAt',
      render: (item) => MomentTz(item).format('DD MMM YYYY, hh:mm:ss A'),
      defaultSortOrder: 'descend',
      sorter: {
        compare: (a, b) => {
          const aNew = new Date(a.createdAt);
          const bNew = new Date(b.createdAt);
          const aMillisecs = aNew.getTime();
          const bMillisecs = bNew.getTime();
          return aMillisecs - bMillisecs;
        },
        multiple: 1
      }
    },
    {
      title: <span className={headerClassStr}>Tracking</span>,
      dataIndex: 'tracking',
      key: 'tracking',
      render: (e) => (
        <Switch
          value
          checkedChildren='On'
          unCheckedChildren='OFF'
          disabled={enableEdit}
          checked={e.enabled}
          onChange={(checked) =>
            toggleClickableElement(activeProject.id, e.id, checked)
          }
        />
      ),
      defaultSortOrder: 'descend',
      sorter: {
        compare: (a, b) => a.tracking.enabled - b.tracking.enabled,
        multiple: 2
      },
      align: 'right'
    }
  ];

  const dataSource = useMemo(() => {
    const data = clickableElements.map((element) => ({
      index: element.id,
      displayName: element.display_name,
      type: element.element_type,
      clickCount: element.click_count,
      createdAt: element.created_at,
      tracking: { id: element.id, enabled: element.enabled }
    }));
    return data;
  }, [clickableElements]);

  const searchList = (e) => {
    setSearchTerm(e.target.value);
  };

  useEffect(() => {
    const searchResults = dataSource.filter(
      (item) =>
        item?.displayName?.toLowerCase().includes(searchTerm.toLowerCase()) ||
        item?.type?.toLowerCase().includes(searchTerm.toLowerCase())
    );
    setListData(searchResults);
  }, [searchTerm, dataSource]);

  return (
    <Row className='mt-1'>
      <Col span={24}>
        <div className='mb-4 flex justify-between'>
          <Text type='title' level={7} color='grey'>
            *Only Admin(s) can change configurations.
          </Text>
          <div className='flex items-center'>
            {showSearch ? (
              <Input
                onChange={searchList}
                className=''
                placeholder='Search'
                style={{ width: '220px', 'border-radius': '5px' }}
                prefix={<SVG name='search' size={16} color='grey' />}
              />
            ) : null}
            <Button
              type='text'
              ghost
              className='p-2 bg-white'
              onClick={() => {
                setShowSearch(!showSearch);
                if (showSearch) {
                  setSearchTerm('');
                }
              }}
            >
              <SVG
                name={!showSearch ? 'search' : 'close'}
                size={20}
                color='grey'
              />
            </Button>
          </div>
        </div>
      </Col>
      <Col span={24}>
        <Table columns={columns} dataSource={listData} pagination={false} />
      </Col>
    </Row>
  );
};

function JavascriptSDK({
  activeProject,
  currentProjectSettings,
  udpateProjectSettings,
  agents,
  currentAgent,
  fetchClickableElements,
  toggleClickableElement,
  clickableElements,
  kbLink
}) {
  const [dataLoading, setDataLoading] = useState(true);
  const projectToken = activeProject.token;
  const assetURL = currentProjectSettings.sdk_asset_url;
  const apiURL = currentProjectSettings.sdk_api_url;
  useEffect(() => {
    fetchClickableElements(activeProject.id).then(() => {
      setDataLoading(false);
    });
  }, [activeProject]);

  const callback = (key) => {
    console.log(key);
  };

  currentProjectSettings =
    currentProjectSettings?.project_settings || currentProjectSettings;

  const renderTabs = () => {
    const tabs = [
      <TabPane tab='Setup' key='1'>
        <div className='flex justify-between items-center'>
          <Text
            type='title'
            level={4}
            weight='bold'
            extraClass='m-0'
            color='character-primary'
          >
            Integration Details
          </Text>
          {kbLink && (
            <Link
              className='inline-block ml-1'
              target='_blank'
              to={{
                pathname: kbLink
              }}
            >
              <div className='flex items-center gap-2'>
                <Text type='paragraph' mini weight='bold' color='brand-color-6'>
                  View Documentation
                </Text>
                <SVG name='ArrowUpRightSquare' size={14} color='#1890ff' />
              </div>
            </Link>
          )}
        </div>
        <div>
          <Text
            type='title'
            level={7}
            extraClass='m-0 mt-2'
            color='character-secondary'
          >
            Your website data will be visible on the platform from the time the
            your javascript SDK is placed on your site. Hence, no historical
            data prior to the setup would be available on the platform. The
            website data you see in Factors is real-time.
          </Text>
          <div>
            <Text
              type='title'
              level={7}
              extraClass='m-0 mt-2'
              color='character-secondary'
            >
              Add javascript below on every page you want to track between the{' '}
              {'<head>'} and {'</head>'} tags{' '}
              <Link
                className='inline-block ml-1'
                target='_blank'
                to={{
                  pathname: JavascriptHeadDocumentation
                }}
              >
                <div className='flex items-center gap-2'>
                  <Text
                    type='paragraph'
                    mini
                    weight='bold'
                    color='brand-color-6'
                  >
                    Here's why
                  </Text>
                  <SVG name='ArrowUpRightSquare' size={14} color='#1890ff' />
                </div>
              </Link>
            </Text>
          </div>
        </div>
        <div className='mt-6'>
          <GtmSteps
            apiURL={apiURL}
            assetURL={assetURL}
            projectToken={projectToken}
            isOnboardingFlow={false}
          />
        </div>
        <div className='mt-6'>
          <ManualSteps
            apiURL={apiURL}
            assetURL={assetURL}
            projectToken={projectToken}
            isOnboardingFlow={false}
          />
        </div>
      </TabPane>,

      <TabPane tab='General Configuration' key='2'>
        <>
          <Col span={24} className='mb-4'>
            <JSConfig
              udpateProjectSettings={udpateProjectSettings}
              currentProjectSettings={currentProjectSettings}
              activeProject={activeProject}
              agents={agents}
              currentAgent={currentAgent}
            />
          </Col>
        </>
      </TabPane>
    ];

    if (currentProjectSettings.auto_click_capture)
      tabs.push(
        <TabPane tab='Click Tracking Configuration' key='3'>
          <ClickTrackConfiguration
            activeProject={activeProject}
            agents={agents}
            currentAgent={currentAgent}
            currentProjectSettings={currentProjectSettings}
            clickableElements={clickableElements}
            toggleClickableElement={toggleClickableElement}
          />
        </TabPane>
      );

    return tabs;
  };
  return (
    <div className='mb-4'>
      <Row>
        <Col span={24}>
          {dataLoading ? (
            <Skeleton active paragraph={{ rows: 4 }} />
          ) : (
            <Tabs defaultActiveKey='1' onChange={callback}>
              {renderTabs()}
            </Tabs>
          )}
        </Col>
      </Row>
      {/* <Row
        style={{
          margin: '10px 15px',
          display: 'flex'
        }}
      >
        <Text type='paragraph' color='mono-6' extraClass='m-0 inline'>
          For detailed instructions on how to install and initialize the
          JavaScript SDK please refer to our
        </Text>
        <Text type='paragraph' color='mono-6' extraClass='m-0 inline'>
          <a href={SDKDocumentation} target='_blank' rel='noreferrer'>
            JavaScript developer documentation &#8594;
          </a>
        </Text>
      </Row> */}
    </div>
  );
}

const mapStateToProps = (state) => ({
  currentProjectSettings: state.global.currentProjectSettings,
  activeProject: state.global.active_project,
  agents: state.agent.agents,
  currentAgent: state.agent.agent_details,
  clickableElements: state.settings.clickableElements
});
export default connect(mapStateToProps, {
  udpateProjectSettings,
  fetchClickableElements,
  toggleClickableElement,
  fetchBingAdsIntegration,
  fetchMarketoIntegration,
  fetchProjectSettingsV1
})(JavascriptSDK);
