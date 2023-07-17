import { Button, Col, Modal, notification, Row, Table } from 'antd';
import { SVG, Text } from 'Components/factorsComponents';
import {
  transformPayloadForWeightConfig,
  transformWeightConfigForQuery
} from 'Components/Profile/utils';
import React, { useMemo, useState, useEffect } from 'react';
import { connect, useSelector } from 'react-redux';
import EngagementModal from './EngagementModal';
import _ from 'lodash';
import { bindActionCreators } from 'redux';
import { updateAccountScores } from 'Reducers/timelines';
import { fetchProjectSettings } from 'Reducers/global';
import SaleWindowModal from './SaleWindowModal';

function EngagementConfig({ fetchProjectSettings }) {
  const [showModal, setShowModal] = useState(false);
  const [saleWindowValue, setSaleWindowValue] = useState();
  const [showSaleWindowModal, setShowSaleWindowModal] = useState(false);
  const [activeEvent, setActiveEvent] = useState({});
  const activeProject = useSelector((state) => state.global.active_project);
  const currentProjectSettings = useSelector(
    (state) => state.global.currentProjectSettings
  );
  const { eventNames } = useSelector((state) => state.coreQuery);

  const columns = [
    {
      title: 'Event',
      dataIndex: 'label',
      key: 'label',
      ellipsis: true
    },
    {
      title: 'Engagement Score',
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

  const weightsConfig = useMemo(() => {
    return currentProjectSettings?.acc_score_weights?.WeightConfig || [];
  }, [currentProjectSettings]);

  const handleOk = (config, editMode) => {
    const weightConf = [...weightsConfig];
    const newConfig = transformPayloadForWeightConfig(config);

    if (editMode) {
      const noChangesMade = weightConf.find(
        (existingConfig) =>
          existingConfig.event_name === newConfig.event_name &&
          existingConfig.wid === newConfig.wid &&
          _.isEqual(newConfig.rule, existingConfig.rule) &&
          existingConfig.weight === newConfig.weight
      );

      if (noChangesMade) {
        showErrorMessage('No changes to save.');
        return;
      } else {
        const configExistsIndex = weightConf.findIndex(
          (existingConfig) =>
            existingConfig.event_name === newConfig.event_name &&
            existingConfig.wid === newConfig.wid
        );

        weightConf.splice(configExistsIndex, 1, newConfig);
      }
    } else {
      if (!config.weight || config.weight === '' || config.weight === 0) {
        showErrorMessage('Please add a score for this rule.');
        return;
      }

      const configExistsIndex = weightConf.findIndex(
        (existingConfig) =>
          existingConfig.event_name === newConfig.event_name &&
          _.isEqual(newConfig.rule, existingConfig.rule)
      );

      if (configExistsIndex !== -1) {
        const configExists = weightConf[configExistsIndex];

        if (configExists.is_deleted) {
          configExists.is_deleted = false;
          newConfig.wid = configExists.wid;
          if (!newConfig.wid) delete newConfig.wid;
          weightConf.splice(configExistsIndex, 1, newConfig);
        } else {
          showErrorMessage('Rule already exists.');
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
        showSuccessMessage(
          `Score ${editMode ? 'updated' : 'added'} successfully`,
          `The ${editMode ? '' : 'new'} score has been ${
            editMode ? 'updated' : 'added'
          }. This will start reflecting in Accounts shortly.`
        )
      )
      .catch((err) => {
        console.log(err);
        showErrorMessage(`Error ${editMode ? 'updating' : 'adding'} score.`);
      });
    setShowModal(false);
  };

  const handleSaleWindowOk = (value) => {
    setSaleWindowValue(value);
    setShowSaleWindowModal(false);
    updateAccountScores(activeProject.id, {
      WeightConfig: [...weightsConfig],
      salewindow: parseInt(value)
    }).then(() => fetchProjectSettings(activeProject.id));
    return;
  };

  const showErrorMessage = (description) => {
    notification.error({
      message: 'Error',
      description: description,
      duration: 3
    });
  };

  const showSuccessMessage = (message, description) => {
    notification.success({
      message: message,
      description: description,
      duration: 3
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

  const onDelete = (event, index) => {
    const updatedWeightConfig = [...weightsConfig];

    updatedWeightConfig[index].is_deleted = true;
    updateAccountScores(activeProject.id, {
      WeightConfig: updatedWeightConfig,
      salewindow: parseInt(saleWindowValue)
    })
      .then(() => fetchProjectSettings(activeProject.id))
      .then(() => showSuccessMessage(`Score removed successfully`))
      .catch((err) => {
        console.log(err);
        showErrorMessage(`Error removing score.`);
      });
  };

  const handleCancel = () => {
    setShowModal(false);
  };

  const handleCancelSaleWindow = () => {
    setShowSaleWindowModal(false);
  };

  const setEdit = (event) => {
    setActiveEvent(event);
    setShowModal(true);
  };

  const setAddNewScore = () => {
    setActiveEvent({});
    setShowModal(true);
  };

  const tableData = useMemo(() => {
    return weightsConfig
      ?.map((q, index) => {
        const event = transformWeightConfigForQuery(q);
        return {
          ...event,
          is_deleted: q.is_deleted,
          label: eventNames[event.label]
            ? eventNames[event.label]
            : event.label,
          weight: (
            <div className='flex justify-between items-center'>
              <div>{event.weight}</div>
              <div className='flex justify-between items-center'>
                <Button
                  onClick={() => setEdit(event)}
                  type='text'
                  icon={<SVG name='edit' />}
                />
                <Button
                  onClick={() => renderDeleteModal(event, index)}
                  type='text'
                  icon={<SVG name='delete' />}
                />
              </div>
            </div>
          )
        };
      })
      .filter((item) => item.is_deleted === false);
  }, [eventNames, weightsConfig]);

  return (
    <div className='fa-container'>
      <Row gutter={[24, 24]} justify='center'>
        <Col span={18}>
          <Row>
            <Col span={18}>
              <Text type='title' level={4} weight='bold'>
                Engagement Scoring
              </Text>
            </Col>
          </Row>
          <Row>
            <Col span={18}>
              <Text type='title' level={7}>
                Some events matter more than others, and are better indicators
                of buying intent. Configure scores for them, tag them as intent
                signals, and more.
              </Text>
            </Col>
            <Col span={6}>
              <div className='flex justify-end'>
                <Button
                  type='primary'
                  icon={<SVG name='plus' color='white' />}
                  onClick={setAddNewScore}
                >
                  Add a rule
                </Button>
              </div>
            </Col>
          </Row>
          <Row className='my-6'>
            <Col span={24}>
              <div className='flex items-center'>
                <div className='mr-2'>How long is your average sales cycle</div>
                {Number(saleWindowValue) <= 0 ||
                saleWindowValue == undefined ||
                saleWindowValue == null ? (
                  <Button
                    className={`fa-button--truncate filter-buttons-radius filter-buttons-margin mx-1`}
                    type='link'
                    onClick={() => setShowSaleWindowModal(true)}
                  >
                    Sale Window
                  </Button>
                ) : (
                  <Button
                    className={`dropdown-btn`}
                    type='text'
                    onClick={() => setShowSaleWindowModal(true)}
                  >
                    {saleWindowValue} Days
                    <SVG size={16} name='edit' color={'black'} />
                  </Button>
                )}
              </div>
            </Col>
          </Row>
          <Row className='my-10'>
            <Col span={24}>
              {weightsConfig.filter((item) => !item?.is_deleted)?.length ? (
                <Table columns={columns} dataSource={tableData} />
              ) : (
                <div className='grid h-full place-items-center'>
                  <img
                    src='../../../../assets/icons/empty_file.svg'
                    alt=''
                  ></img>
                  <Text type='title' level={6} weight='bold' extraClass='m-4'>
                    Looks like there aren't any rules here yet
                  </Text>
                  <Button
                    type='primary'
                    icon={<SVG name='plus' color='white' />}
                    onClick={setAddNewScore}
                  >
                    Add a rule
                  </Button>
                </div>
              )}
            </Col>
          </Row>
        </Col>
      </Row>
      <EngagementModal
        event={activeEvent}
        visible={showModal}
        onOk={handleOk}
        onCancel={handleCancel}
        editMode={Object.entries(activeEvent).length}
      />
      <SaleWindowModal
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
      fetchProjectSettings
    },
    dispatch
  );

export default connect(null, mapDispatchToProps)(EngagementConfig);
