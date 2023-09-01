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
  fetchProjectSettings,
  fetchProjectSettingsV1,
  udpateProjectSettings
} from 'Reducers/global';
import {
  fetchClickableElements,
  toggleClickableElement
} from '../../../../reducers/settings/middleware';
import { connect, useSelector } from 'react-redux';
import MomentTz from '../../../../components/MomentTz';
import DemoSDK from './DemoSDK';
import CodeBlock from 'Components/CodeBlock';
import styles from './index.module.scss';
import { UserAddOutlined } from '@ant-design/icons';
import InviteUsers from 'Views/Settings/ProjectSettings/UserSettings/InviteUsers';
import ExcludeIp from '../BasicSettings/IpBlocking/excludeIp';
import { generateSdkScriptCode } from './utils';
import ScriptHtml from './ScriptHtml';
import CodeBlockV2 from 'Components/CodeBlock/CodeBlockV2';
import logger from 'Utils/logger';

const { TabPane } = Tabs;

const ViewSetup = ({ currentProjectSettings, activeProject }) => {
  const projectToken = activeProject.token;
  const assetURL = currentProjectSettings.sdk_asset_url;
  const apiURL = currentProjectSettings.sdk_api_url;

  return (
    <Row>
      <Col span={24}>
        <Text
          type={'title'}
          level={5}
          weight={'bold'}
          extraClass={'m-0 mt-2'}
        >
          Setup manually
        </Text>
        <Text
          type={'title'}
          level={6}
          color={'grey'}
          extraClass={'m-0 mb-3'}
        >
          Add Factors SDK manually in the head section for all pages you wish to
          get data for
        </Text>
        <div className='ml-3'>
          <Text
            type={'title'}
            level={6}
            weight={'bold'}
            color={'grey'}
            extraClass={'m-0 mt-2 mb-2'}
          >
            Setup 1
          </Text>
          <Text type='paragraph' color='mono-6' extraClass={'m-0'}>
            Add the javascript code below on every page between the {'<head>'}{' '}
            and {'</head>'} tags.
          </Text>
          <div className='py-4'>
            <CodeBlockV2
              collapsedViewText={
                <>
                  <span style={{ color: '#2F80ED' }}>{`<script>`}</span>
                  {`(function(c)d.appendCh.....func("`}
                  <span style={{ color: '#EB5757' }}>{`${projectToken}`}</span>
                  {`")`}
                  <span style={{ color: '#2F80ED' }}>{`</script>`}</span>
                </>
              }
              fullViewText={
                <ScriptHtml
                  projectToken={projectToken}
                  assetURL={assetURL}
                  apiURL={apiURL}
                />
              }
              textToCopy={generateSdkScriptCode(assetURL, projectToken, apiURL)}
            />
          </div>
        </div>
      </Col>
      <div className='ml-3'>
        <Col span={24}>
          <Text
            type={'title'}
            level={6}
            weight={'bold'}
            color={'grey'}
            extraClass={'m-0 mt-4'}
          >
            Setup 2 (Optional)
          </Text>
          <Text type='paragraph' color='mono-6' extraClass={'m-0'}>
            Send us an event (Enable Auto-track under{' '}
            <span className='italic'>General Configuration tab</span> above for
            capturing user visits automatically).
          </Text>
        </Col>
        <Col span={24}>
          <CodeBlock
            codeContent={
              <>
                <span style={{ color: '#2F80ED' }}>{'faitracker.call'}</span>
                {'("track", "'}
                <span style={{ color: '#EB5757' }}>{'YOUR_EVENT'}</span>
                {'");'}
              </>
            }
            pureTextCode={`faitracker.call("track", "YOUR_EVENT");`}
            hideCopyBtn={true}
          ></CodeBlock>
        </Col>
      </div>
    </Row>
  );
};

const GTMSetup = ({ currentProjectSettings, activeProject }) => {
  const projectToken = activeProject.token;
  const assetURL = currentProjectSettings.sdk_asset_url;
  const apiURL = currentProjectSettings.sdk_api_url;

  return (
    <Row>
      <Col span={24}>
        <Text
          type={'title'}
          level={5}
          weight={'bold'}
          extraClass={'m-0 mt-2'}
        >
          Setup using GTM
        </Text>
        <Text
          type={'title'}
          level={6}
          color={'grey'}
          extraClass={'m-0 mb-3'}
        >
          Add Factors SDK quickly using Google Tag Manager without any
          engineering effort
        </Text>
        <div className='ml-3'>
          <Text
            type={'title'}
            level={6}
            weight={'bold'}
            color={'grey'}
            extraClass={'m-0 mt-2 mb-2'}
          >
            Setup 1
          </Text>
          <Text type='paragraph' color='mono-6' extraClass={'m-0'}>
            1. Sign in to&nbsp;
            <span>
              <a href='https://tagmanager.google.com/' target='_blank'>
                Google Tag Manager
              </a>
            </span>
            &nbsp;and select “Workspace”.
          </Text>
          <Text type='paragraph' color='mono-6' extraClass={'m-0'}>
            2. Click on “Add a new tag” and name it “Factors tag”.
          </Text>
          <Text type='paragraph' color='mono-6' extraClass={'m-0'}>
            3. Click <span className='italic'>Edit</span> on Tag Configuration
            and under custom, select <span className='italic'>Custom HTML</span>
          </Text>
          <Text type='paragraph' color='mono-6' extraClass={'m-0'}>
            4. Copy the tracking script below and paste it on the HTML field.
            Hit <span className='italic'>Save</span>.
          </Text>
          <div className='py-4'>
            <CodeBlockV2
              collapsedViewText={
                <>
                  <span style={{ color: '#2F80ED' }}>{`<script>`}</span>
                  {`(function(c)d.appendCh.....func("`}
                  <span style={{ color: '#EB5757' }}>{`${projectToken}`}</span>
                  {`")`}
                  <span style={{ color: '#2F80ED' }}>{`</script>`}</span>
                </>
              }
              fullViewText={
                <ScriptHtml
                  projectToken={projectToken}
                  assetURL={assetURL}
                  apiURL={apiURL}
                />
              }
              textToCopy={generateSdkScriptCode(assetURL, projectToken, apiURL)}
            />
          </div>

          <Text type='paragraph' color='mono-6' extraClass={'m-0'}>
            5. In the Triggers popup, click{' '}
            <span className='italic'>Add Trigger</span> and select All Pages.
          </Text>
          <Text type='paragraph' color='mono-6' extraClass={'m-0'}>
            6. Once the trigger has been added, click on Publish at the top of
            your GTM window and that’s it!
          </Text>
        </div>
      </Col>

      <div className='ml-3'>
        <Col span={24}>
          <Text
            type={'title'}
            level={6}
            weight={'bold'}
            color={'grey'}
            extraClass={'m-0 mt-4'}
          >
            Setup 2 (Optional)
          </Text>
          <Text type='paragraph' color='mono-6' extraClass={'m-0'}>
            Send us an event (Enable Auto-track under{' '}
            <span className='italic'>General Configuration tab</span> above for
            capturing user visits automatically).
          </Text>
        </Col>
        <Col span={24}>
          <CodeBlock
            codeContent={
              <>
                <span style={{ color: '#2F80ED' }}>{'faitracker.call'}</span>
                {'("track", "'}
                <span style={{ color: '#EB5757' }}>{'YOUR_EVENT'}</span>
                {'");'}
              </>
            }
            pureTextCode={`faitracker.call("track", "YOUR_EVENT");`}
            hideCopyBtn={true}
          ></CodeBlock>
        </Col>
      </div>
    </Row>
  );
};

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
          <Text type={'title'} level={7} color={'grey'} extraClass={'m-0 my-2'}>
            *Only Admin(s) can change configurations.
          </Text>
        </Col>
      )}
      <Col span={24}>
        <div span={24} className={'flex flex-start items-center mt-2'}>
          <span style={{ width: '50px' }}>
            <Switch
              checkedChildren='On'
              disabled={enableEdit}
              unCheckedChildren='OFF'
              onChange={toggleAutoTrack}
              checked={autoTrack}
            />
          </span>
          <Text
            type={'title'}
            level={6}
            weight={'bold'}
            extraClass={'m-0 ml-2'}
          >
            Auto-track
          </Text>
        </div>
      </Col>
      <Col span={24} className={'flex flex-start items-center'}>
        <Text type={'paragraph'} mini extraClass={'m-0 mt-2'} color={'grey'}>
          Track standard events such as page_view, page_load time,
          page_spent_time and more for each user
        </Text>
      </Col>
      <Col span={24}>
        <div span={24} className={'flex flex-start items-center mt-8'}>
          <span style={{ width: '50px' }}>
            <Switch
              checkedChildren='On'
              disabled={enableEdit}
              unCheckedChildren='OFF'
              onChange={toggleAutoTrackSPAPageView}
              checked={autoTrackSPAPageView}
            />
          </span>
          <Text
            type={'title'}
            level={6}
            weight={'bold'}
            extraClass={'m-0 ml-2'}
          >
            Auto-track Single Page Application
          </Text>
        </div>
      </Col>
      <Col span={24} className={'flex flex-start items-center'}>
        <Text type={'paragraph'} mini extraClass={'m-0 mt-2'} color={'grey'}>
          Track standard events such as page_view, page_load time,
          page_spent_time and button clicks for each user, on Single Page
          Applications like React, Angular, Vue, etc,.
        </Text>
      </Col>
      <Col span={24}>
        <div span={24} className={'flex flex-start items-center mt-8'}>
          <span style={{ width: '50px' }}>
            <Switch
              checkedChildren='On'
              disabled={enableEdit}
              unCheckedChildren='OFF'
              onChange={toggleExcludeBot}
              checked={excludeBot}
            />
          </span>
          <Text
            type={'title'}
            level={6}
            weight={'bold'}
            extraClass={'m-0 ml-2'}
          >
            Exclude Bot
          </Text>
        </div>
      </Col>
      <Col span={24} className={'flex flex-start items-center'}>
        <Text type={'paragraph'} mini extraClass={'m-0 mt-2'} color={'grey'}>
          Automatically exclude bot traffic from website traffic using Factor’s
          proprietary algorithm
        </Text>
      </Col>
      <Col span={24}>
        <div span={24} className={'flex flex-start items-center mt-8'}>
          <span style={{ width: '50px' }}>
            <Switch
              checkedChildren='On'
              disabled={enableEdit}
              unCheckedChildren='OFF'
              onChange={toggleAutoFormCapture}
              checked={autoFormCapture}
            />
          </span>
          <Text
            type={'title'}
            level={6}
            weight={'bold'}
            extraClass={'m-0 ml-2'}
          >
            Auto Form Capture
          </Text>
        </div>
      </Col>
      <Col span={24} className={'flex flex-start items-center'}>
        <Text type={'paragraph'} mini extraClass={'m-0 mt-2'} color={'grey'}>
          Automatically track personal identification information such as email
          and phone number from Form Submissions
        </Text>
      </Col>
      <Col span={24}>
        <div span={24} className={'flex flex-start items-center mt-8'}>
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
          <Text
            type={'title'}
            level={6}
            weight={'bold'}
            extraClass={'m-0 ml-2'}
          >
            Auto Form Fill Capture
          </Text>
        </div>
      </Col>
      <Col span={24} className={'flex flex-start items-center'}>
        <Text type={'paragraph'} mini extraClass={'m-0 mt-2'} color={'grey'}>
          Automatically track personal identification information such as email
          and phone number from Form Filled Values
        </Text>
      </Col>
      <Col span={24}>
        <div span={24} className={'flex flex-start items-center mt-8'}>
          <span style={{ width: '50px' }}>
            <Switch
              checkedChildren='On'
              disabled={enableEdit}
              unCheckedChildren='OFF'
              onChange={toggleClickCapture}
              checked={clickCapture}
            />
          </span>
          <Text
            type={'title'}
            level={6}
            weight={'bold'}
            extraClass={'m-0 ml-2'}
          >
            Auto Click Capture
          </Text>
        </div>
      </Col>
      <Col span={24} className={'flex flex-start items-center'}>
        <Text type={'paragraph'} mini extraClass={'m-0 mt-2'} color={'grey'}>
          Starts discovering available buttons and anchors on the website. After
          discovery, it will be listed under Click Tracking Configurations and
          can be enabled for tracking as events.
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

  var columns = [
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
    const data = clickableElements.map((element) => {
      return {
        index: element.id,
        displayName: element.display_name,
        type: element.element_type,
        clickCount: element.click_count,
        createdAt: element.created_at,
        tracking: { id: element.id, enabled: element.enabled }
      };
    });
    return data;
  }, [clickableElements]);

  const searchList = (e) => {
    setSearchTerm(e.target.value);
  };

  useEffect(() => {
    const searchResults = dataSource.filter((item) => {
      return (
        item?.displayName?.toLowerCase().includes(searchTerm.toLowerCase()) ||
        item?.type?.toLowerCase().includes(searchTerm.toLowerCase())
      );
    });
    setListData(searchResults);
  }, [searchTerm, dataSource]);

  return (
    <Row className={'mt-1'}>
      <Col span={24}>
        <div className={'mb-4 flex justify-between'}>
          <Text type={'title'} level={7} color={'grey'}>
            *Only Admin(s) can change configurations.
          </Text>
          <div className={'flex items-center'}>
            {showSearch ? (
              <Input
                onChange={searchList}
                className={''}
                placeholder={'Search'}
                style={{ width: '220px', 'border-radius': '5px' }}
                prefix={<SVG name='search' size={16} color={'grey'} />}
              />
            ) : null}
            <Button
              type='text'
              ghost={true}
              className={'p-2 bg-white'}
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
                color={'grey'}
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
const VerifySdkCheck = ({
  activeProject,
  setDataLoading,

  fetchBingAdsIntegration,
  fetchMarketoIntegration,
  fetchProjectSettingsV1
}) => {
  const int_completed = useSelector(
    (state) => state?.global?.projectSettingsV1?.int_completed
  );
  const [sdkVerified, setSdkVerified] = useState(int_completed ? true : false);
  const [loading, setLoading] = useState(false);
  const [errorState, setErrorState] = useState(false);

  const handleSdkVerification = async () => {
    try {
      setLoading(true);
      setErrorState(false);
      const res = await fetchProjectSettingsV1(activeProject.id);

      if (res?.data?.int_completed) {
        setSdkVerified(true);
        notification.success({
          message: 'Success',
          description: 'SDK Verified!',
          duration: 3
        });
      } else {
        notification.error({
          message: 'Error',
          description: 'SDK not Verified!',
          duration: 3
        });
        setErrorState(true);
      }

      setLoading(false);
    } catch (error) {
      logger.error(error);
      setErrorState(true);
      setLoading(false);
    }
  };

  return (
    <React.Fragment>
      <div className='mt-2 ml-2'>
        <Divider />
        {sdkVerified && (
          <div className='flex justify-between items-center'>
            <div>
              <SVG name={'CheckCircle'} extraClass={'inline'} />
              <Text
                type={'title'}
                level={6}
                color={'character-primary'}
                extraClass={'m-0 ml-2 inline'}
              >
                {'Verified. Your script is up and running.'}
              </Text>
            </div>
            <Button
              type={'text'}
              size={'small'}
              style={{ color: '#1890FF' }}
              onClick={() => handleSdkVerification()}
              loading={loading}
            >
              {'Verify again'}
            </Button>
          </div>
        )}
        {!int_completed && !errorState && (
          <div className='flex gap-2 items-center'>
            <Text type='paragraph' color='mono-6' extraClass='m-0'>
              {'Have you already added the code?'}
            </Text>
            <Button onClick={() => handleSdkVerification()}>
              {'Verify it now'}
            </Button>
          </div>
        )}
        {errorState && (
          <div className='flex items-center'>
            <SVG name={'CloseCircle'} extraClass={'inline'} color='#F5222D' />
            <Text
              type={'title'}
              level={6}
              color={'character-primary'}
              extraClass={'m-0 ml-2 inline'}
            >
              {'Couldn’t detect SDK.'}
            </Text>
            <Button
              type={'text'}
              size={'small'}
              style={{ color: '#1890FF', padding: 0 }}
              onClick={() => handleSdkVerification()}
              loading={loading}
            >
              Verify again
            </Button>
            <Text
              type={'title'}
              level={6}
              color={'character-primary'}
              extraClass={'m-0 ml-1 inline'}
            >
              or
            </Text>
            <Button
              type={'text'}
              size={'small'}
              style={{ color: '#1890FF', padding: 0, marginLeft: 4 }}
              onClick={() =>
                window.open('https://calendly.com/aravindhvetri', '_blank')
              }
            >
              book a call
            </Button>
          </div>
        )}
      </div>
    </React.Fragment>
  );
};
function JavascriptSDK({
  activeProject,
  fetchProjectSettings,
  currentProjectSettings,
  udpateProjectSettings,
  agents,
  currentAgent,
  fetchClickableElements,
  toggleClickableElement,
  clickableElements,

  fetchBingAdsIntegration,
  fetchMarketoIntegration,
  fetchProjectSettingsV1,
  isOnBoardFlow
}) {
  const [dataLoading, setDataLoading] = useState(true);
  const [isDemo, setIsDemo] = useState(null);

  const projectId = useSelector((state) => state?.global?.active_project?.id);
  useEffect(() => {
    console.log(projectId);
    if (projectId == '519') {
      setIsDemo(true);
    }
  }, []);
  useEffect(() => {
    fetchProjectSettings(activeProject.id).then(() => {
      setDataLoading(false);
    });

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
    let tabs = [
      <TabPane tab='GTM Setup' key='1'>
        <GTMSetup
          currentProjectSettings={currentProjectSettings}
          activeProject={activeProject}
        />
      </TabPane>,
      <TabPane tab='Manual Setup' key='2'>
        <ViewSetup
          currentProjectSettings={currentProjectSettings}
          activeProject={activeProject}
          fetchBingAdsIntegration={fetchBingAdsIntegration}
          fetchMarketoIntegration={fetchMarketoIntegration}
          fetchProjectSettingsV1={fetchProjectSettingsV1}
        />
      </TabPane>,
      <TabPane tab='General Configuration' key='3'>
        <React.Fragment>
          <Col span={24} className={'mb-4'}>
            <JSConfig
              udpateProjectSettings={udpateProjectSettings}
              currentProjectSettings={currentProjectSettings}
              activeProject={activeProject}
              agents={agents}
              currentAgent={currentAgent}
            />
          </Col>
        </React.Fragment>
      </TabPane>
    ];

    if (currentProjectSettings.auto_click_capture)
      tabs.push(
        <TabPane tab='Click Tracking Configuration' key='4'>
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
  const [inviteModal, setInviteModal] = useState(false);
  const handleOk = () => {};
  const confirmLoading = () => {};
  return (
    <>
      <div className={'mb-4 pl-4'}>
        <Row style={{ width: '100%', justifyContent: 'space-between' }}>
          <Col span={12}>
            <Text type={'title'} level={3} weight={'bold'} extraClass={'m-0'}>
              {isOnBoardFlow === true ? 'Add our ' : ''} Javascript SDK
            </Text>
          </Col>
          <Col>
            {isOnBoardFlow === true ? (
              <Row>
                <div
                  style={{
                    display: 'flex',
                    alignItems: 'center',
                    margin: '0 10px',
                    fontWeight: '600'
                  }}
                >
                  Need help?
                </div>
                <Button
                  size='large'
                  icon={<UserAddOutlined />}
                  className={styles['btn']}
                  onClick={() => {
                    setInviteModal((prev) => !prev);
                  }}
                >
                  Invite Team
                </Button>
              </Row>
            ) : (
              ''
            )}
          </Col>
        </Row>
        <Row>
          <Col span={24}>
            <Text
              type={'title'}
              level={6}
              color={'grey-2'}
              extraClass={'m-0 my-1'}
            >
              {isOnBoardFlow === true
                ? 'Track Data Natively With a Factors SDK (Coded Tracking)'
                : 'Your website data will be visible on the platform from the time the your javascript SDK is placed on your site. Hence, no historical data prior to the setup would be available on the platform.'}
            </Text>
            {isOnBoardFlow === true ? (
              <></>
            ) : (
              <Text
                type={'title'}
                level={6}
                color={'grey-2'}
                extraClass={'m-0 my-1'}
              >
                The website data you see in Factors is real-time.
              </Text>
            )}
          </Col>
        </Row>
        <InviteUsers
          visible={inviteModal}
          onCancel={() => setInviteModal(false)}
          onOk={() => handleOk()}
          confirmLoading={confirmLoading}
        />
        <Row className={'mt-2'}>
          <Col span={24}>
            {isDemo === true ? (
              <DemoSDK />
            ) : (
              <>
                {dataLoading ? (
                  <Skeleton active paragraph={{ rows: 4 }} />
                ) : (
                  <>
                    <Tabs defaultActiveKey='1' onChange={callback}>
                      {renderTabs()}
                    </Tabs>

                    <VerifySdkCheck
                      activeProject={activeProject}
                      setDataLoading={setDataLoading}
                      fetchBingAdsIntegration={fetchBingAdsIntegration}
                      fetchMarketoIntegration={fetchMarketoIntegration}
                      fetchProjectSettingsV1={fetchProjectSettingsV1}
                    />
                  </>
                )}
              </>
            )}
          </Col>
        </Row>
        <Row
          style={{
            margin: '10px 15px',
            display: 'flex'
          }}
        >
          <Text type='paragraph' color='mono-6' extraClass={'m-0 inline'}>
            For detailed instructions on how to install and initialize the
            JavaScript SDK please refer to our
          </Text>
          <Text type='paragraph' color='mono-6' extraClass={'m-0 inline'}>
            <a
              href='https://help.factors.ai/en/articles/7260638-placing-factors-sdk'
              target='_blank'
              rel='noreferrer'
            >
              JavaScript developer documentation &#8594;
            </a>
          </Text>
        </Row>
      </div>
    </>
  );
}

const mapStateToProps = (state) => {
  return {
    currentProjectSettings: state.global.currentProjectSettings,
    activeProject: state.global.active_project,
    agents: state.agent.agents,
    currentAgent: state.agent.agent_details,
    clickableElements: state.settings.clickableElements
  };
};
export default connect(mapStateToProps, {
  fetchProjectSettings,
  udpateProjectSettings,
  fetchClickableElements,
  toggleClickableElement,
  fetchBingAdsIntegration,
  fetchMarketoIntegration,
  fetchProjectSettingsV1
})(JavascriptSDK);
