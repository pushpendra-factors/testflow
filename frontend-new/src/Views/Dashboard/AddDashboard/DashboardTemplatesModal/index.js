import {
  CheckCircleOutlined,
  CopyOutlined,
  LoadingOutlined,
  MinusOutlined,
  PlusOutlined
} from '@ant-design/icons';
import { Alert, Button, Col, Row, Tag } from 'antd';
import { SVG, Text } from 'Components/factorsComponents';
import { ArrowLeftSVG } from 'Components/svgIcons';
import React, { useEffect, useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';

import HorizontalWindow from 'Components/HorizontalWindow';
import { createDashboardFromTemplate } from 'Reducers/dashboard_templates/services';
import { useHistory } from 'react-router-dom';
import { setItemToLocalStorage } from 'Utils/localStorage.helpers';
import { PathUrls } from 'Routes/pathUrls';
import ModalFlow from 'Components/ModalFlow';
import { DASHBOARD_KEYS } from '../../../../constants/localStorage.constants';
import TemplatesThumbnail, {
  FallBackImage,
  Integration_Checks
} from '../../../../constants/templates.constants';
import styles from './index.module.scss';
import { NEW_DASHBOARD_TEMPLATES_MODAL_CLOSE } from '../../../../reducers/types';

const Step2DashboardTemplateModal = ({
  template,
  allTemplates,
  onCancel,
  handleBack,
  handleSelectItem
}) => {
  const [showMore, setShowMore] = useState(true);
  const [copiedState, setCopiedState] = useState(1);
  const history = useHistory();

  // This is created to map Window Title, Image and onClick event
  const [windowTemplates, setWindowTemplates] = useState([]);
  const sdkCheck = useSelector(
    (state) => state?.global?.projectSettingsV1?.int_completed
  );
  const activeProject = useSelector((state) => state.global.active_project);
  const integration = useSelector(
    (state) => state.global.currentProjectSettings
  ); // This is to get All the Integration States, but it doesn't returns SdkCheck
  const [haveRequirements, setHaveRequirements] = useState(false);
  const { bingAds, marketo } = useSelector((state) => state.global);

  const [integrationCheckFailedAt, setIntegrationCheckFailedAt] =
    useState(undefined);
  let integrationChecks = null;

  useEffect(() => {
    if (copiedState === 3) {
      // means dashboard Copied
      // and relad page
      setTimeout(() => {
        window.location.reload();
      }, 500);
    }
  }, [copiedState]);
  // Below useEffect gets called everytime template in Step 2 gets changed
  useEffect(() => {
    integrationChecks = new Integration_Checks(
      sdkCheck,
      integration,
      bingAds,
      marketo
    );
    // let keyname = template.title.toLowerCase().replace(/\s/g, '');
    const integrationResults = integrationChecks.checkRequirements(
      template.required_integrations
    );
    setHaveRequirements(integrationResults.result);
    setIntegrationCheckFailedAt(integrationResults.failedAt);
    setCopiedState(1);
  }, [template]);

  // Below UseEffect gets called once in a lifetime
  useEffect(() => {
    // Below is to check render all the Related Templates
    const temp = [];
    if (allTemplates)
      allTemplates.forEach((element) => {
        temp.push({
          title: element.title,
          image: TemplatesThumbnail.has(
            element.title.toLowerCase().replace(/\s/g, '')
          )
            ? TemplatesThumbnail.get(
                element.title.toLowerCase().replace(/\s/g, '')
              ).image
            : null
        });
      });
    setWindowTemplates(temp);
  }, []);

  const HandleConfirmOkay = async () => {
    try {
      setCopiedState(2);
      const res = await createDashboardFromTemplate(
        activeProject.id,
        template.id
      );

      setItemToLocalStorage(DASHBOARD_KEYS.ACTIVE_DASHBOARD_ID, res.data.id);
      setCopiedState(3);

      if (res) {
        history.push(PathUrls.Dashboard);
      }
    } catch (err) {
      setCopiedState(1);
      console.log(err);
    }
  };

  const onIntegrateNowClick = () => {
    onCancel();
    history.push('/settings/integration');
  };
  return (
    <div>
      <Row className={styles.modalContainerTop}>
        <Col
          span={24}
          style={{ display: 'flex', justifyContent: 'space-between' }}
        >
          <div className='flex items-center'>
            <Button
              size='large'
              type='text'
              icon={<ArrowLeftSVG />}
              onClick={handleBack}
            />
            <Text type='title' level={7} weight='normal' extraClass='m-0 mr-3'>
              {' '}
              Go back to templates
            </Text>
          </div>
          <Button
            size='large'
            type='text'
            icon={<SVG size={20} name='close' />}
            onClick={onCancel}
          />
        </Col>
      </Row>

      <Row className={styles.modalContainerBody}>
        <Col span={24} className={styles.modalTitleHeader}>
          <Text type='title' level={4} weight='bold' extraClass='m-0 mr-3'>
            {template.title}{' '}
          </Text>

          <Row gutter={[8, 8]} className={styles.templateSection}>
            <Col span={12}>
              <Text
                type='title'
                level={7}
                weight='normal'
                extraClass='m-0 mr-3'
              >
                {template.description}
              </Text>

              <Button
                type='primary'
                icon={
                  copiedState === 1 ? (
                    <CopyOutlined />
                  ) : copiedState === 2 ? (
                    <LoadingOutlined />
                  ) : copiedState === 3 ? (
                    <CheckCircleOutlined />
                  ) : (
                    ''
                  )
                }
                className={styles.templateSectionCopyButton}
                disabled={!haveRequirements}
                onClick={HandleConfirmOkay}
                style={{
                  background: copiedState === 3 ? '#5ACA89 ' : '',
                  borderColor: copiedState === 3 ? '#5ACA89 ' : ''
                }}
              >
                {copiedState === 1 ? (
                  <> Copy this dashboard </>
                ) : copiedState === 2 ? (
                  <>Copying</>
                ) : copiedState === 3 ? (
                  <>Copied this Dashboard</>
                ) : (
                  ''
                )}
              </Button>

              <Alert
                style={{ display: haveRequirements ? 'none' : 'block' }}
                message={
                  <div>
                    <Text
                      type='title'
                      level={7}
                      weight='normal'
                      extraClass={`m-0 mr-3 `}
                    >
                      <>
                        Please complete{' '}
                        {integrationCheckFailedAt !== undefined &&
                          integrationCheckFailedAt.map(
                            (eachIntegration, eachIndex) => (
                              <span
                                key={eachIntegration}
                                style={{
                                  fontWeight: '600',
                                  textTransform: 'capitalize'
                                }}
                              >
                                {eachIntegration}
                                {eachIndex ===
                                integrationCheckFailedAt.length - 1
                                  ? ''
                                  : ', '}
                              </span>
                            )
                          )}{' '}
                        integration to use this Dashboard
                      </>
                    </Text>
                    <Button
                      className={styles.templatesSectionAlertBtn}
                      type='link'
                      onClick={onIntegrateNowClick}
                    >
                      Integrate Now
                    </Button>
                  </div>
                }
                className={styles.templatesSectionAlert}
                type='warning'
                showIcon
              />

              <div className={styles.includedReports}>
                <Text
                  type='title'
                  level={6}
                  weight='bold'
                  extraClass={`m-0 mr-3 `}
                >
                  Included Reports
                </Text>
                {showMore
                  ? template.units.length > 4 &&
                    template?.units
                      ?.slice(0, 4)
                      .map((eachReport, eachIndex) => (
                        <Text
                          type='title'
                          level={7}
                          weight='normal'
                          extraClass={`m-0 mr-3 `}
                        >{`${eachIndex + 1}. ${eachReport.title}`}</Text>
                      ))
                  : template?.units?.map((eachReport, eachIndex) => (
                      <Text
                        type='title'
                        level={7}
                        weight='normal'
                        extraClass={`m-0 mr-3 `}
                      >{`${eachIndex + 1}. ${eachReport.title}`}</Text>
                    ))}

                {showMore ? (
                  <Button
                    className={styles.showMoreBtn}
                    type='text'
                    onClick={() => setShowMore(false)}
                    icon={<PlusOutlined />}
                  >
                    Show more
                  </Button>
                ) : (
                  <Button
                    className={styles.showMoreBtn}
                    type='text'
                    onClick={() => setShowMore(true)}
                    icon={<MinusOutlined />}
                  >
                    Show less
                  </Button>
                )}
              </div>
            </Col>
            <Col
              span={12}
              className={styles.templateSectionCol2}
              style={{ padding: '0 10px 0 10px' }}
            >
              <img
                style={{ width: '100%' }}
                alt={template.title}
                src={
                  TemplatesThumbnail.has(
                    template.title.toLowerCase().replace(/\s/g, '')
                  )
                    ? TemplatesThumbnail.get(
                        template.title.toLowerCase().replace(/\s/g, '')
                      ).image
                    : FallBackImage
                }
              />
              <div>
                <Text
                  type='title'
                  level={7}
                  weight='bold'
                  extraClass={`m-0 mr-3 `}
                >
                  Tags :{' '}
                </Text>
                <Tag>{template.tags.value}</Tag>
              </div>
            </Col>
          </Row>
        </Col>
      </Row>

      <Row>
        <Col span={24}>
          <HorizontalWindow
            windowTemplates={windowTemplates}
            onWindowClick={(index) => {
              handleSelectItem(allTemplates[index]);
            }}
          />
        </Col>
      </Row>
    </div>
  );
};

/*
  Main Component Responsible for Rendering of Modal and initial Methods
*/
const DashboardTemplatesModal = () => {
  const dispatch = useDispatch();

  const dashboardTemplates = useSelector(
    (state) => state.dashboardTemplates?.templates
  );

  const dashboard_templates_modal_state = useSelector(
    (state) => state.dashboardTemplatesController
  );
  const [allTemplates, setAllTemplates] = useState([]);
  useEffect(() => {
    if (dashboardTemplates?.data && Array.isArray(dashboardTemplates.data))
      setAllTemplates(
        dashboardTemplates.data.map((each) => ({
          ...each,
          imagePath: TemplatesThumbnail.has(
            each.title.toLowerCase().replace(/\s/g, '')
          )
            ? TemplatesThumbnail.get(
                each.title.toLowerCase().replace(/\s/g, '')
              ).image
            : FallBackImage
        }))
      );
  }, [dashboardTemplates]);
  return (
    <ModalFlow
      data={allTemplates}
      visible={dashboard_templates_modal_state.isNewDashboardTemplateModal}
      onCancel={() => {
        dispatch({ type: NEW_DASHBOARD_TEMPLATES_MODAL_CLOSE });
      }}
      startFreshVisible // this is added to make Start Fresh Option make visible
      FirstScreenIllustration='DashboardTemplateIllustration'
      Step2Screen={Step2DashboardTemplateModal}
      isDashboardTemplatesFlow
    />
  );
};
export default React.memo(DashboardTemplatesModal);

/*
 <Modal
      bodyStyle={{ padding: '0px 0px 24px 0' }}
      title={null}
      visible={dashboard_templates_modal_state.isNewDashboardTemplateModal}
      centered
      zIndex={1005}
      width={1052}
      className='fa-modal--regular p-4 fa-modal--slideInDown '
      confirmLoading={apisCalled}
      closable={false}
      okText={getOkText()}
      transitionName=''
      maskTransitionName=''
      footer={null}
    >
      {' '}
      {step === 1 ? (
        <Step1DashboardTemplateModal
          templates={finalTemplates}
          handleTemplate={handleTemplate}
          searchTemplateHandle={searchTemplateHandle}
          setCategorySelected={setCategorySelected}
          searchValue={searchValue}
          categorySelected={categorySelected}
        />
      ) : (
        <Step2DashboardTemplateModal
          template={selectedTemplate || null}
          setStep={setStep}
          setSelectedTemplate={setSelectedTemplate}
          allTemplates={allTemplates}
        />
      )}
    </Modal>

    */
