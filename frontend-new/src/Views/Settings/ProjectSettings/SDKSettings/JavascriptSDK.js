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
  Spin
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

const { TabPane } = Tabs;

const ViewSetup = ({ activeProject }) => {
  const projectToken = activeProject.token;
  // eslint-disable-next-line
  const assetURL = BUILD_CONFIG.sdk_asset_url;

  return (
    <Row>
      <Col span={24}>
        <Text
          type={'title'}
          level={5}
          weight={'bold'}
          color={'grey'}
          extraClass={'m-0 mt-2'}
        >
          Setup 1
        </Text>
        <Text type={'paragraph'} extraClass={'m-0'}>
          Add the below javascript code on every page between the {'<head>'} and{' '}
          {'</head>'} tags.
        </Text>
      </Col>
      <Col span={24}>
        <pre className={'fa-code-block my-4'}>
          <code>
            {`<script>
window.factors=window.factors||function(){this.q=[];var i=new CustomEvent("FACTORS_QUEUED_EVENT"),n=function(t,e){this.q.push({k:t,a:e}),window.dispatchEvent(i)};return this.track=function(t,e,i){n("track",arguments)},this.init=function(t,e,i){this.TOKEN=t,this.INIT_PARAMS=e,this.INIT_CALLBACK=i,window.dispatchEvent(new CustomEvent("FACTORS_INIT_EVENT"))},this.reset=function(){n("reset",arguments)},this.page=function(t,e){n("page",arguments)},this.updateEventProperties=function(t,e){n("updateEventProperties",arguments)},this.identify=function(t,e){n("identify",arguments)},this.addUserProperties=function(t){n("addUserProperties",arguments)},this.getUserId=function(){n("getUserId",arguments)},this.call=function(){var t={k:"",a:[]};if(arguments&&1<=arguments.length){for(var e=1;e<arguments.length;e++)t.a.push(arguments[e]);t.k=arguments[0]}this.q.push(t),window.dispatchEvent(i)},this.init("${projectToken}"),this}(),function(){var t=document.createElement("script");t.type="text/javascript",t.src="${assetURL}",t.async=!0,d=document.getElementsByTagName("script")[0],d.parentNode.insertBefore(t,d)}(); 
</script>`}
          </code>
        </pre>
      </Col>
      <Col span={24}>
        <Text type={'paragraph'} extraClass={'m-0 mt-2 mb-2'}>
          For detailed help or instructions to setup via GTM (Google Tag
          Manager), please refer to our{' '}
          <a
            className={'fa-anchor'}
            href='https://help.factors.ai/en/articles/5754974-placing-factors-s-javascript-sdk-on-your-website'
            target='_blank'
          >
            JavaScript developer documentation.
          </a>
        </Text>
      </Col>
      <Col span={24}>
        <Text
          type={'title'}
          level={5}
          weight={'bold'}
          color={'grey'}
          extraClass={'m-0 mt-4'}
        >
          Setup 2 (Optional)
        </Text>
        <Text type={'paragraph'} extraClass={'m-0'}>
          Send us an event (Enable Auto-track for capturing user visits
          automatically).
        </Text>
      </Col>
      <Col span={24}>
        <pre className={'fa-code-block my-4'}>
          <code>{'factors.track("YOUR_EVENT");'}</code>
        </pre>
      </Col>
    </Row>
  );
};

const GTMSetup = ({ activeProject }) => {
  const projectToken = activeProject.token;
  // eslint-disable-next-line
  const assetURL = BUILD_CONFIG.sdk_asset_url;

  return (
    <Row>
      <Col span={24}>
        <Text
          type={'title'}
          level={5}
          weight={'bold'}
          color={'grey'}
          extraClass={'m-0 mt-2 mb-1'}
        >
          Setup 1
        </Text>
        <Text type={'paragraph'} extraClass={'m-0'}>
          1. Sign in to{' '}
          <span className={'underline'}>
            <a href='https://tagmanager.google.com/' target='_blank'>
              Google Tag Manager
            </a>
          </span>
          , select “Workspace”, and “Add a new tag”
        </Text>
        <Text type={'paragraph'} extraClass={'m-0'}>
          2. Name it “Factors tag”. Select{' '}
          <span className={'italic'}>Edit</span> on Tag Configuration
        </Text>
        <Text type={'paragraph'} extraClass={'m-0'}>
          3. Under custom, select <span className={'italic'}>custom HTML</span>
        </Text>
        <Text type={'paragraph'} extraClass={'m-0'}>
          4. Copy the below tracking script and{' '}
          <span className={'italic'}>paste</span> it on the HTML field, Select{' '}
          <span className={'font-extrabold'}>Save</span>
        </Text>
      </Col>
      <Col span={24}>
        <pre className={'fa-code-block my-4'}>
          <code>
            {`<script>
window.factors=window.factors||function(){this.q=[];var i=new CustomEvent("FACTORS_QUEUED_EVENT"),n=function(t,e){this.q.push({k:t,a:e}),window.dispatchEvent(i)};return this.track=function(t,e,i){n("track",arguments)},this.init=function(t,e,i){this.TOKEN=t,this.INIT_PARAMS=e,this.INIT_CALLBACK=i,window.dispatchEvent(new CustomEvent("FACTORS_INIT_EVENT"))},this.reset=function(){n("reset",arguments)},this.page=function(t,e){n("page",arguments)},this.updateEventProperties=function(t,e){n("updateEventProperties",arguments)},this.identify=function(t,e){n("identify",arguments)},this.addUserProperties=function(t){n("addUserProperties",arguments)},this.getUserId=function(){n("getUserId",arguments)},this.call=function(){var t={k:"",a:[]};if(arguments&&1<=arguments.length){for(var e=1;e<arguments.length;e++)t.a.push(arguments[e]);t.k=arguments[0]}this.q.push(t),window.dispatchEvent(i)},this.init("${projectToken}"),this}(),function(){var t=document.createElement("script");t.type="text/javascript",t.src="${assetURL}",t.async=!0,d=document.getElementsByTagName("script")[0],d.parentNode.insertBefore(t,d)}(); 
</script>`}
          </code>
        </pre>
      </Col>
      <Col span={24}>
        <Text type={'paragraph'} extraClass={'m-0'}>
          5. In the <span className={'italic'}>Triggers</span> popup, select{' '}
          <span className={'italic'}>Add Trigger</span> and select{' '}
          <span className={'italic'}>All Pages</span>
        </Text>
        <Text type={'paragraph'} extraClass={'m-0'}>
          6. The trigger has been added. Click on{' '}
          <span className={'font-extrabold'}>Publish</span> at the top of your
          GTM window!
        </Text>
      </Col>
      <Col span={24}>
        <Text type={'paragraph'} extraClass={'m-0 mt-4 mb-2'}>
          For detailed help or instructions to setup via GTM (Google Tag
          Manager), please refer to our{' '}
          <a
            className={'fa-anchor'}
            href='https://help.factors.ai/en/articles/5754974-placing-factors-s-javascript-sdk-on-your-website'
            target='_blank'
          >
            JavaScript developer documentation.
          </a>
        </Text>
      </Col>
      <Col span={24}>
        <Text
          type={'title'}
          level={5}
          weight={'bold'}
          color={'grey'}
          extraClass={'m-0 mt-4'}
        >
          Setup 2 (Optional)
        </Text>
        <Text type={'paragraph'} extraClass={'m-0'}>
          Send us custom events that you define using GTM’s triggers (Enable
          Auto-track for capturing user visits automatically).
        </Text>
      </Col>
      <Col span={24}>
        <pre className={'fa-code-block my-4'}>
          <code>{'factors.track("YOUR_EVENT");'}</code>
        </pre>
      </Col>
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
          </span>{' '}
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
          </span>{' '}
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
          </span>{' '}
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
          </span>{' '}
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
              disabled={enableEdit}
              unCheckedChildren='OFF'
              onChange={toggleAutoCaptureFormFills}
              checked={autoCaptureFormFills}
            />
          </span>{' '}
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
          </span>{' '}
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
      agents.map((agent) => {
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
  const [sdkCheck, setSdkCheck] = useState();
  const [loading, setloading] = useState(false);
  const fetchProjectsV1 = async () => {
    console.log('FETCH');
    try {
      fetchProjectSettingsV1(activeProject.id).then((res) => {
        setSdkCheck(res.data.int_completed);
        console.log(res);
      });
    } catch (error) {
      console.error(error);
    }
  };
  useEffect(() => {
    fetchProjectsV1();
  }, [sdkCheck]);

  const onSDKcheck = () => {
    setloading(true);
    setTimeout(() => {
      setloading(false);
    }, 2000);
    setSdkCheck(!sdkCheck);
  };

  return (
    <React.Fragment>
      {loading ? (
        <div className='flex justify-center items-center w-full'>
          <Spin />
        </div>
      ) : (
        <>
          {sdkCheck && sdkCheck !== '' && sdkCheck !== true ? (
            <Row justify={'space-between'}>
              <Col span={20}>
                <SVG name={'CheckCircle'} extraClass={'inline'} />
                <Text
                  type={'title'}
                  level={6}
                  color={'grey-2'}
                  extraClass={'m-0 ml-2 inline'}
                >
                  Verified. Your script is up and running.
                </Text>
              </Col>
              <Col>
                <Button
                  type={'text'}
                  size={'small'}
                  style={{ color: '#1890FF' }}
                  onClick={onSDKcheck}
                >
                  Verify again
                </Button>
              </Col>
            </Row>
          ) : (
            <div>
              <Text type={'title'} level={6} extraClass={'m-0 ml-2 mr-1'}>
                SDK not detected yet. Have you added the code?{' '}
                <Button type={'default'} size={'small'} onClick={onSDKcheck}>
                  Verify it now
                </Button>
              </Text>
            </div>
          )}
        </>
      )}
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
  fetchProjectSettingsV1
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

  return (
    <>
      <div className={'mb-4 pl-4'}>
        <Row>
          <Col span={12}>
            <Text type={'title'} level={3} weight={'bold'} extraClass={'m-0'}>
              Javascript SDK
            </Text>
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
              Your website data will be visible on the platform from the time
              the your javascript SDK is placed on your site. Hence, no
              historical data prior to the setup would be available on the
              platform.
            </Text>
            <Text type={'title'} level={6} color={'grey-2'} extraClass={'m-0'}>
              The website data you see in Factors is real-time.
            </Text>
          </Col>
        </Row>
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
                      {renderTabs()}{' '}
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
