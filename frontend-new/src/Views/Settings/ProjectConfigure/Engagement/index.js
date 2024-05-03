import {
  Button,
  Col,
  Modal,
  notification,
  Popover,
  Row,
  Table,
  Tooltip
} from 'antd';
import { SVG, Text } from 'Components/factorsComponents';
import {
  findKeyByValue,
  transformPayloadForWeightConfig,
  transformWeightConfigForQuery
} from 'Components/Profile/utils';
import React, { useMemo, useState, useEffect } from 'react';
import { connect, useSelector } from 'react-redux';
import _ from 'lodash';
import { bindActionCreators } from 'redux';
import { updateAccountScores } from 'Reducers/timelines';
import { fetchProjectSettings } from 'Reducers/global';
import { fetchEventNames, getGroups } from 'Reducers/coreQuery/middleware';
import { headerClassStr } from 'Components/Profile/constants';
import { InfoCircleFilled } from '@ant-design/icons';
import SaleWindowModal from './SaleWindowModal';
import EngagementModal from './EngagementModal';

import EngagementCategoryModal from './EngagementCategoryModal';
import styles from './index.module.scss';

const filterConfigRuleCheck = (existingConfig, newConfig) => {
  try {
    let result = true;
    if (Array.isArray(existingConfig) && Array.isArray(newConfig)) {
      existingConfig?.forEach((eachrule, eachIndex) => {
        result &&=
          _.isEqual(eachrule?.value, newConfig[eachIndex]?.value) &&
          eachrule?.operator === newConfig[eachIndex]?.operator &&
          eachrule?.property_type === newConfig[eachIndex]?.property_type &&
          eachrule?.value_type === newConfig[eachIndex]?.value_type &&
          eachrule?.lower_bound === newConfig[eachIndex]?.lower_bound;
      });
    } else if (existingConfig === null && newConfig === null) {
      result &&= true;
    } else {
      result = false;
    }
    return result;
  } catch (err) {
    return false;
  }
};

function EngagementConfig({
  fetchProjectSettings,
  getGroups,
  fetchEventNames
}) {
  const [editIndex, setEditIndex] = useState(undefined);
  const [showCategoryModal, setShowCategoryModal] = useState(false);
  const [renderCategoryModal, setRenderCategoryModal] = useState(false);
  const [showModal, setShowModal] = useState(false);
  const [renderModal, setRenderModal] = useState(false);
  const [saleWindowValue, setSaleWindowValue] = useState();
  const [showSaleWindowModal, setShowSaleWindowModal] = useState(false);
  const [activeEvent, setActiveEvent] = useState({});
  const activeProject = useSelector((state) => state.global.active_project);
  const currentProjectSettings = useSelector(
    (state) => state.global.currentProjectSettings
  );
  const { eventNamesSpecial, eventNamesMapSpecial } = useSelector(
    (state) => state.coreQuery
  );

  useEffect(() => {
    fetchEventNames(activeProject?.id, true);
  }, [activeProject]);

  useEffect(() => {
    getGroups(activeProject?.id);
  }, [activeProject?.id]);

  const columns = [
    {
      title: <div className={headerClassStr}>Engagement Signals</div>,
      dataIndex: 'label',
      key: 'label',
      ellipsis: true
    },
    {
      title: <div className={headerClassStr}>Weight assigned</div>,
      width: 250,
      dataIndex: 'weight',
      key: 'weight'
    }
  ];

  useEffect(() => {
    const initialSaleWindow =
      currentProjectSettings?.acc_score_weights?.salewindow;
    setSaleWindowValue(initialSaleWindow);
  }, [currentProjectSettings?.acc_score_weights?.salewindow]);

  const weightsConfig = useMemo(
    () => currentProjectSettings?.acc_score_weights?.WeightConfig || [],
    [currentProjectSettings]
  );

  const showSuccessMessage = ({ title, description }) => {
    notification.success({
      message: title,
      description,
      duration: 3
    });
  };

  const showErrorMessage = ({ description, title = 'Error' }) => {
    notification.error({
      message: title,
      description,
      duration: 3,
      style: {
        zIndex: 6000
      }
    });
  };

  const handleOk = (config, editMode) => {
    const weightConf = [...weightsConfig];
    const newConfig = transformPayloadForWeightConfig(config);

    if (editMode) {
      const noChangesMade = weightConf.find(
        (existingConfig) =>
          existingConfig.weight === newConfig.weight &&
          existingConfig.fname === newConfig.fname
      );

      if (noChangesMade) {
        showErrorMessage({ description: 'No changes to save.' });
        return;
      }
      const configExistsIndex = weightConf.findIndex(
        (existingConfig) =>
          existingConfig.event_name === newConfig.event_name &&
          existingConfig.wid === newConfig.wid
      );

      weightConf.splice(configExistsIndex, 1, newConfig);
    } else {
      if (!config.weight || config.weight === '' || config.weight === 0) {
        showErrorMessage({ description: 'Please add a score for this rule.' });
        return;
      }
      const configExistsIndex = weightConf.findIndex(
        (existingConfig) =>
          existingConfig.event_name === newConfig.event_name &&
          filterConfigRuleCheck(existingConfig?.rule, newConfig?.rule)
      );
      if (configExistsIndex !== -1) {
        const configExists = weightConf[configExistsIndex];

        if (configExists.is_deleted) {
          configExists.is_deleted = false;
          newConfig.wid = configExists.wid;
          if (!newConfig.wid) delete newConfig.wid;
          weightConf.splice(configExistsIndex, 1, newConfig);
        } else {
          showErrorMessage({
            description:
              'Another signal with the same event condition already exists.',
            title: 'Signal already exists'
          });
          return;
        }
      } else {
        delete newConfig.wid;
        weightConf.push(newConfig);
      }
    }
    updateAccountScores(activeProject.id, {
      WeightConfig: weightConf,
      salewindow: parseInt(saleWindowValue)
    })
      .then(() => fetchProjectSettings(activeProject.id))
      .then(() =>
        showSuccessMessage({
          title: `${
            editMode ? 'Signal updated' : 'New signal added'
          } successfully`,
          description: `${
            editMode
              ? 'Signal has been saved with changes. Updates will reflect across accounts within 24 hours.'
              : 'Signal saved. Monitoring for this signal has now begun across accounts.'
          }`
        })
      )
      .catch((err) => {
        console.log(err);
        showErrorMessage({
          description: `Error ${editMode ? 'updating' : 'adding'} score.`
        });
      });
    setShowModal(false);
    const timeoutHandle = setTimeout(() => {
      setRenderModal(false);
      clearTimeout(timeoutHandle);
    }, 500);
  };
  const handleCategoryModal = {
    onCancel: () => {
      setShowCategoryModal(false);
      const timeoutHandle = setTimeout(() => {
        setRenderCategoryModal(false);
        clearTimeout(timeoutHandle);
      }, 500);
    },
    onOk: () => {
      setShowCategoryModal(false);
      const timeoutHandle = setTimeout(() => {
        setRenderCategoryModal(false);
        clearTimeout(timeoutHandle);
      }, 500);
    }
  };
  const handleSaleWindowOk = (value) => {
    setSaleWindowValue(value);
    setShowSaleWindowModal(false);
    updateAccountScores(activeProject.id, {
      WeightConfig: [...weightsConfig],
      salewindow: parseInt(value)
    }).then(() => {
      fetchProjectSettings(activeProject.id);
      showSuccessMessage({
        title: 'Engagement window updated',
        description:
          'The engagement window time has been updated and changes will reflect within 24 hours.'
      });
    });
  };

  const onDelete = (event, index) => {
    const updatedWeightConfig = [...weightsConfig];

    updatedWeightConfig[index].is_deleted = true;

    updateAccountScores(activeProject.id, {
      WeightConfig: updatedWeightConfig,
      salewindow: parseInt(saleWindowValue)
    })
      .then(() => fetchProjectSettings(activeProject.id))
      .then(() =>
        showSuccessMessage({
          title: `Signal removed successfully`,
          description:
            'Signal removed from monitoring. Changes will reflect across accounts within 24 hours.'
        })
      )
      .catch((err) => {
        console.log(err);
        showErrorMessage({ description: `Error removing score.` });
      });
  };

  const renderDeleteModal = (event, index) => {
    Modal.confirm({
      title: 'Do you want to remove this score?',
      okText: 'Yes',
      cancelText: 'Cancel',
      closable: true,
      centered: true,
      onOk: () => {
        onDelete(event, index);
      },
      onCancel: () => {}
    });
  };

  const handleCancel = () => {
    setShowModal(false);
    setEditIndex(undefined);
    const timeoutHandle = setTimeout(() => {
      setRenderModal(false);
      clearTimeout(timeoutHandle);
    }, 500);
  };

  const handleCancelSaleWindow = () => {
    setShowSaleWindowModal(false);
  };

  const setEdit = (event, index) => {
    setActiveEvent(event);
    setShowModal(true);
    setRenderModal(true);
    setEditIndex(index);
  };

  const setAddNewScore = () => {
    setActiveEvent({});
    setShowModal(true);
    setRenderModal(true);
  };

  const tableData = useMemo(
    () =>
      weightsConfig
        ?.map((q, index) => {
          const event = transformWeightConfigForQuery(q);
          event.group = findKeyByValue(eventNamesMapSpecial, event.label);
          return {
            ...event,
            is_deleted: q.is_deleted,
            label: event.fname || event.label,
            weight: (
              <div className='flex justify-between items-center'>
                <div>{event.weight}</div>
                <div className='flex justify-between items-center'>
                  <Tooltip title='Edit Signal'>
                    <Button
                      onClick={() => setEdit(event, index)}
                      type='text'
                      icon={<SVG name='edit' />}
                    />
                  </Tooltip>
                  <Tooltip title='Delete Signal'>
                    <Button
                      onClick={() => renderDeleteModal(event, index)}
                      type='text'
                      icon={<SVG name='delete' />}
                    />
                  </Tooltip>
                </div>
              </div>
            )
          };
        })
        .filter((item) => item.is_deleted === false),
    [eventNamesSpecial, weightsConfig]
  );

  return (
    <div className='fa-container'>
      <Row gutter={[24, 24]} justify='center'>
        <Col span={22}>
          <Row>
            <Col span={18}>
              <Text
                type='title'
                level={4}
                weight='bold'
                id='fa-at-text--page-title'
              >
                Engagement Scoring
              </Text>
            </Col>
          </Row>
          <Row>
            <Col span={12}>
              <Text type='title' level={7}>
                Define signals of engagement that matter to your organisation
                and assign them weights to accurately score the engagement level
                of your accounts.
              </Text>
            </Col>
            <Col span={12}>
              <div className='flex justify-end' style={{ gap: '10px' }}>
                <Popover
                  placement='bottom'
                  overlayClassName={styles.engagementpopover}
                  trigger='hover'
                  style={{ margin: 0 }}
                  overlayInnerStyle={{
                    borderRadius: '8px',
                    margin: 0,
                    padding: 0
                  }}
                  content={<SVG name='EngagementCategoryPillsPopover' />}
                >
                  <Button
                    className='dropdown-btn'
                    onClick={() => {
                      setShowCategoryModal(true);
                      setRenderCategoryModal(true);
                    }}
                  >
                    Engagement Category
                  </Button>
                </Popover>
                <Button
                  type='primary'
                  icon={<SVG name='plus' color='white' />}
                  onClick={setAddNewScore}
                >
                  Add signal
                </Button>
              </div>
            </Col>
          </Row>
          <Row className='my-6'>
            <Col span={24}>
              <div className='flex items-center'>
                <div className='mr-2'>Set engagement window</div>
                {Number(saleWindowValue) <= 0 ||
                saleWindowValue === undefined ||
                saleWindowValue === null ? (
                  <Button
                    className='fa-button--truncate filter-buttons-radius filter-buttons-margin mx-1'
                    type='link'
                    onClick={() => setShowSaleWindowModal(true)}
                  >
                    Sale Window
                  </Button>
                ) : (
                  <Button
                    className='dropdown-btn'
                    type='text'
                    onClick={() => setShowSaleWindowModal(true)}
                  >
                    {saleWindowValue} Days
                    <SVG size={16} name='edit' color='black' />
                  </Button>
                )}
                <Tooltip title='This helps keep engagement levels up to date for all your accounts.'>
                  <InfoCircleFilled className='ml-2' />
                </Tooltip>
              </div>
            </Col>
          </Row>
          <Row className='my-10'>
            <Col span={24}>
              {weightsConfig.filter((item) => !item?.is_deleted)?.length ? (
                <Table columns={columns} dataSource={tableData} />
              ) : (
                <div className='grid h-full place-items-center'>
                  <img src='../../../../assets/icons/empty_file.svg' alt='' />
                  <Text type='title' level={6} weight='bold' extraClass='m-4'>
                    Looks like there aren&apos;t any rules here yet
                  </Text>
                  <Button
                    type='primary'
                    icon={<SVG name='plus' color='white' />}
                    onClick={setAddNewScore}
                  >
                    Add new signal
                  </Button>
                </div>
              )}
            </Col>
          </Row>
        </Col>
      </Row>

      {renderCategoryModal && (
        <EngagementCategoryModal
          visible={showCategoryModal}
          {...handleCategoryModal}
        />
      )}
      {renderModal && (
        <EngagementModal
          event={activeEvent}
          visible={showModal}
          onOk={handleOk}
          onCancel={handleCancel}
          editMode={Object.entries(activeEvent).length}
        />
      )}
      <SaleWindowModal
        saleWindowValue={saleWindowValue}
        visible={showSaleWindowModal}
        onOk={handleSaleWindowOk}
        onCancel={handleCancelSaleWindow}
      />
    </div>
  );
}
const mapDispatchToProps = (dispatch) =>
  bindActionCreators(
    {
      fetchProjectSettings,
      getGroups,
      fetchEventNames
    },
    dispatch
  );

export default connect(null, mapDispatchToProps)(EngagementConfig);
