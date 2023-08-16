import {
  CheckCircleOutlined,
  CopyOutlined,
  LoadingOutlined,
  MinusOutlined,
  PlusOutlined,
  SearchOutlined
} from '@ant-design/icons';
import {
  Alert,
  Button,
  Col,
  Divider,
  Input,
  Modal,
  Row,
  Spin,
  Tag
} from 'antd';
import { SVG, Text } from 'Components/factorsComponents';
import { ArrowLeftSVG } from 'Components/svgIcons';
import React, { useEffect, useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { CATEGORY_TYPES } from './../../../../constants/categories.constants';

import {
  ADD_DASHBOARD_MODAL_OPEN,
  NEW_DASHBOARD_TEMPLATES_MODAL_CLOSE,
  NEW_DASHBOARD_TEMPLATES_MODAL_OPEN,
  UPDATE_PICKED_FIRST_DASHBOARD_TEMPLATE
} from '../../../../reducers/types';
import styles from './index.module.scss';
import HorizontalWindow from 'Components/HorizontalWindow';
import TemplatesThumbnail, {
  FallBackImage,
  Integration_Checks
} from '../../../../constants/templates.constants';
import { createDashboardFromTemplate } from 'Reducers/dashboard_templates/services';
import { fetchDashboards } from 'Reducers/dashboard/services';
import {
  fetchAttrContentGroups,
  fetchGroups,
  fetchQueries,
  fetchSmartPropertyRules
} from 'Reducers/coreQuery/services';
import { fetchKPIConfig, fetchPageUrls } from 'Reducers/kpi';
import {
  fetchEventNames,
  getGroupProperties,
  getUserPropertiesV2
} from 'Reducers/coreQuery/middleware';
import { fetchWeeklyIngishtsMetaData } from 'Reducers/insights';
import { useHistory } from 'react-router-dom';
import { setItemToLocalStorage } from 'Utils/localStorage.helpers';
import { DASHBOARD_KEYS } from './../../../../constants/localStorage.constants';
import TemplateThumbnailImage from './TemplateThumbnailImage';
import { PathUrls } from 'Routes/pathUrls';

const CATEGORY_SELECTED_STYLES = {
  background: '#f5f6f8',
  color: '#1890ff'
};

let Step1DashboardTemplateModal = ({
  templates,
  handleTemplate,
  searchTemplateHandle,
  setCategorySelected,
  categorySelected,
  searchValue
}) => {
  let dispatch = useDispatch();

  let dashboardTemplates = useSelector((state) => state.dashboardTemplates);

  let handleCategoryFunction = (eachCategory) => {
    if (!eachCategory) {
      setCategorySelected(null);
      return;
    } else {
      setCategorySelected(eachCategory);
    }
  };
  return (
    <>
      <Row className={styles.modalContainerTop}>
        <Col span={24} style={{ display: 'flex', justifyContent: 'end' }}>
          <Button
            size='large'
            type='text'
            icon={<SVG size={20} name='close' />}
            onClick={() => {
              dispatch({ type: NEW_DASHBOARD_TEMPLATES_MODAL_CLOSE });
            }}
          />
        </Col>
        <Row className='w-full' style={{ padding: '10px 0px' }}>
          <Col span={12} className={styles.modalTitleHeader}>
            <Text
              type={'title'}
              level={3}
              weight={'bold'}
              extraClass={`m-0 mr-3`}
            >
              What are you tracking today?
            </Text>
          </Col>
          <Col
            span={12}
            style={{
              display: 'flex',
              justifyContent: 'end',
              paddingRight: '40px'
            }}
          >
            <Input
              value={searchValue}
              onChange={searchTemplateHandle}
              style={{ height: '100%', width: '50%' }}
              placeholder='Search Templates'
              prefix={<SearchOutlined />}
            />
          </Col>
        </Row>
      </Row>
      <Row className={styles.modalContainerBody}>
        <Col span={6} style={{ borderRight: '1px solid #dedede' }}>
          <div className={styles.categoryLists}>
            <div className={styles.categoryLeftItemTitle}>
              <Text
                type={'title'}
                level={5}
                weight={'bold'}
                extraClass={`m-0 mr-3`}
              >
                Categories
              </Text>
            </div>
            <div
              className={`${styles.categoryLeftItem} ${
                categorySelected == null ? styles.categoryItemSelectedArray : ''
              }`}
              onClick={() => handleCategoryFunction()}
              style={{
                color:
                  categorySelected == null
                    ? CATEGORY_SELECTED_STYLES.color
                    : '',
                background:
                  categorySelected == null
                    ? CATEGORY_SELECTED_STYLES.background
                    : ''
              }}
            >
              All Categories
            </div>
            {CATEGORY_TYPES.map((eachCategory, eachIndex) => {
              return (
                <div
                  key={eachIndex}
                  className={`${styles.categoryLeftItem} ${
                    eachCategory == categorySelected
                      ? styles.categoryItemSelectedArray
                      : ''
                  }`}
                  onClick={() => handleCategoryFunction(eachCategory)}
                  style={{
                    color:
                      eachCategory == categorySelected
                        ? CATEGORY_SELECTED_STYLES.color
                        : '',
                    background:
                      eachCategory == categorySelected
                        ? CATEGORY_SELECTED_STYLES.background
                        : ''
                  }}
                >
                  {' '}
                  {eachCategory}{' '}
                </div>
              );
            })}
          </div>
        </Col>
        <Col span={18} className={styles.templatesShowSection + ' text-center'}>
          <Row
            gutter={[8, 8]}
            style={{ cursor: 'pointer' }}
            onClick={() => dispatch({ type: ADD_DASHBOARD_MODAL_OPEN })}
          >
            <Col style={{ width: '50%' }}>
              <TemplateThumbnailImage isStartFresh={true} />
            </Col>
            <Col
              style={{ width: '50%', display: 'grid', placeContent: 'center' }}
            >
              <Text
                type={'title'}
                level={6}
                weight={'bold'}
                extraClass={`m-0 mr-3`}
              >
                Start Fresh
              </Text>
              <Text
                type={'title'}
                level={7}
                weight={'normal'}
                extraClass={`m-0 mr-3 ${styles.startFreshDescription}`}
              >
                Create an empty dashboard, run queries and add widgets to start
                monitoring.{' '}
              </Text>
            </Col>
          </Row>
          <Divider />
          {dashboardTemplates.templates.loading ? (
            <Spin style={{ margin: '0 auto' }} />
          ) : (
            <Row gutter={[8, 8]} style={{ height: '50%' }}>
              {templates?.length > 0 ? (
                <>
                  {templates?.map((eachState, eachIndex) => {
                    return (
                      <Col
                        key={eachIndex + '-' + eachState.title}
                        span={12}
                        style={{
                          padding: '0 20px 0 20px',
                          width: '300px',
                          cursor: 'pointer',
                          borderRadius: '2.6792px',
                          margin: '10px 0px',
                          textAlign: 'left'
                        }}
                        onClick={() => handleTemplate(eachIndex)}
                      >
                        <TemplateThumbnailImage
                          TemplatesThumbnail={TemplatesThumbnail}
                          eachState={eachState}
                        />

                        <Text
                          type={'title'}
                          level={6}
                          weight={'bold'}
                          extraClass={`m-0 mr-3`}
                        >
                          {eachState.title}
                        </Text>
                        <Text
                          type={'title'}
                          level={7}
                          weight={'normal'}
                          extraClass={`m-0 mr-3 ${styles.templateDescription}`}
                        >
                          {eachState.description}
                        </Text>
                      </Col>
                    );
                  })}
                </>
              ) : (
                <Col style={{ width: '100%', textAlign: 'center' }}>
                  <Text
                    type={'title'}
                    level={6}
                    weight={'bold'}
                    extraClass={`m-0 mr-3`}
                  >
                    No templates here yet. Coming soon!
                  </Text>
                </Col>
              )}
            </Row>
          )}
        </Col>
      </Row>
    </>
  );
};

let Step2DashboardTemplateModal = ({
  template,
  setStep,
  setSelectedTemplate,
  allTemplates
}) => {
  const [showMore, setShowMore] = useState(true);
  const [copiedState, setCopiedState] = useState(1);
  const history = useHistory();
  const dispatch = useDispatch();
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
  const { dashboards, activeDashboard } = useSelector(
    (state) => state.dashboard
  );
  const [integrationCheckFailedAt, setIntegrationCheckFailedAt] =
    useState(undefined);
  let integrationChecks = null;

  // This gets called when Click on Any window happens
  const onWindowClick = (index) => {
    setSelectedTemplate(allTemplates[index]);
  };

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
    let integrationResults = integrationChecks.checkRequirements(
      template.required_integrations
    );
    setHaveRequirements(integrationResults.result);
    setIntegrationCheckFailedAt(integrationResults.failedAt);
    setCopiedState(1);
  }, [template]);

  // Below UseEffect gets called once in a lifetime
  useEffect(() => {
    // Below is to check render all the Related Templates
    let temp = [];
    allTemplates &&
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
    return () => {
      setSelectedTemplate(null);
    };
  }, []);

  // If Any error occur and we get unexpected template
  if (template === null)
    return (
      <>
        <Row>
          <Col span={24} style={{ display: 'flex', justifyContent: 'end' }}>
            <Button
              size='large'
              type='text'
              icon={<SVG size={20} name='close' />}
              onClick={() => {
                setStep(1);
                setSelectedTemplate(null);
                dispatch({
                  type: UPDATE_PICKED_FIRST_DASHBOARD_TEMPLATE,
                  payload: null
                });
                dispatch({ type: NEW_DASHBOARD_TEMPLATES_MODAL_CLOSE });
              }}
            />
          </Col>
        </Row>
        <Row>
          <Button onClick={() => setStep(1)}>Go Back</Button> <br />
          Please Select Template
        </Row>
      </>
    );

  const fetchDashboardItems = () => {
    dispatch(fetchDashboards(activeProject.id));
    dispatch(fetchQueries(activeProject.id));
    dispatch(fetchGroups(activeProject.id));
    dispatch(fetchKPIConfig(activeProject.id));
    dispatch(fetchPageUrls(activeProject.id));
    // dispatch(deleteQueryTest())
    fetchEventNames(activeProject.id);
    getUserPropertiesV2(activeProject.id);
    getGroupProperties(activeProject.id);
    dispatch(fetchSmartPropertyRules(activeProject.id));
    fetchWeeklyIngishtsMetaData(activeProject.id);
    dispatch(fetchAttrContentGroups(activeProject.id));
  };

  const HandleConfirmOkay = async () => {
    try {
      setCopiedState(2);
      const res = await createDashboardFromTemplate(
        activeProject.id,
        template.id
      );

      setItemToLocalStorage(DASHBOARD_KEYS.ACTIVE_DASHBOARD_ID, res.data.id);
      setCopiedState(3);
      //   setStep(1);
      //   dispatch({ type: NEW_DASHBOARD_TEMPLATES_MODAL_CLOSE });

      // fetchDashboardItems ();
      if (res) {
        history.push(PathUrls.Dashboard);
      }
    } catch (err) {
      setCopiedState(1);
      console.log(err);
    }
  };

  const onIntegrateNowClick = () => {
    setStep(1);
    dispatch({ type: NEW_DASHBOARD_TEMPLATES_MODAL_CLOSE });
    history.push('/settings/integration');
  };
  return (
    <div>
      <Row className={styles.modalContainerTop}>
        <Col
          span={24}
          style={{ display: 'flex', justifyContent: 'space-between' }}
        >
          <Button
            size='large'
            type='text'
            icon={<ArrowLeftSVG />}
            onClick={() => {
              setSelectedTemplate(null);
              dispatch({
                type: UPDATE_PICKED_FIRST_DASHBOARD_TEMPLATE,
                payload: null
              });
              setStep(1);
            }}
          />
          <Button
            size='large'
            type='text'
            icon={<SVG size={20} name='close' />}
            onClick={() => {
              setStep(1);
              setSelectedTemplate(null);
              dispatch({
                type: UPDATE_PICKED_FIRST_DASHBOARD_TEMPLATE,
                payload: null
              });
              dispatch({ type: NEW_DASHBOARD_TEMPLATES_MODAL_CLOSE });
            }}
          />
        </Col>
      </Row>

      <Row className={styles.modalContainerBody}>
        <Col span={24} className={styles.modalTitleHeader}>
          <Text
            type={'title'}
            level={4}
            weight={'bold'}
            extraClass={`m-0 mr-3`}
          >
            {template.title}{' '}
          </Text>

          <Row gutter={[8, 8]} className={styles.templateSection}>
            <Col span={12}>
              <Text
                type={'title'}
                level={7}
                weight={'normal'}
                extraClass={`m-0 mr-3`}
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
                disabled={haveRequirements ? false : true}
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
                      type={'title'}
                      level={7}
                      weight={'normal'}
                      extraClass={`m-0 mr-3 `}
                    >
                      <>
                        Please complete{' '}
                        {integrationCheckFailedAt != undefined &&
                          integrationCheckFailedAt.map(
                            (eachIntegration, eachIndex) => (
                              <span
                                key={eachIndex}
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
                  type={'title'}
                  level={6}
                  weight={'bold'}
                  extraClass={`m-0 mr-3 `}
                >
                  Included Reports
                </Text>
                {showMore
                  ? template.units.length > 4 &&
                    template?.units
                      ?.slice(0, 4)
                      .map((eachReport, eachIndex) => {
                        return (
                          <Text
                            type={'title'}
                            level={7}
                            weight={'normal'}
                            extraClass={`m-0 mr-3 `}
                          >{`${eachIndex + 1}. ${eachReport.title}`}</Text>
                        );
                      })
                  : template?.units?.map((eachReport, eachIndex) => {
                      return (
                        <Text
                          type={'title'}
                          level={7}
                          weight={'normal'}
                          extraClass={`m-0 mr-3 `}
                        >{`${eachIndex + 1}. ${eachReport.title}`}</Text>
                      );
                    })}

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
                  type={'title'}
                  level={7}
                  weight={'bold'}
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
            onWindowClick={onWindowClick}
          />
        </Col>
      </Row>
    </div>
  );
};

/*
  Main Component Responsible for Rendering of Modal and initial Methods
*/
let DashboardTemplatesModal = ({ apisCalled, getOkText }) => {
  let dispatch = useDispatch();

  let dashboardTemplates = useSelector((state) => state.dashboardTemplates);

  let dashboard_templates_modal_state = useSelector(
    (state) => state.dashboard_templates_Reducer
  );
  let [allTemplates, setAllTemplates] = useState([]);
  const [finalTemplates, setFinalTemplates] = useState([]);
  let [step, setStep] = useState(1);
  let [selectedTemplate, setSelectedTemplate] = useState(null);
  const [searchValue, setSearchValue] = useState('');
  const [searchedTemplates, setSearchedTemplates] = useState([]);
  const [categorySelected, setCategorySelected] = useState(null);

  const [categoryMap, setCategoryMap] = useState(new Map());
  const searchTemplateHandle = (event) => {
    setSearchValue(event.target.value);
  };
  // THis useEffect gets Called when Any Category Selection Change happens
  useEffect(() => {
    if (categorySelected) {
      setFinalTemplates(
        categoryMap.get(categorySelected.toLowerCase().replace(/\s/g, ''))
      );
    } else setFinalTemplates(allTemplates);
  }, [categorySelected]);
  // THis UseEffect callback function gets called when search value is changed
  useEffect(() => {
    let searchResults = allTemplates.filter((item) =>
      item?.title?.toLowerCase().includes(searchValue.toLowerCase())
    );
    setSearchedTemplates(searchResults);
    setCategorySelected(null);
  }, [searchValue]);
  // This useEffect is to Select Templates from First Dashboard Experience
  useEffect(() => {
    if (dashboard_templates_modal_state.pickedFirstTemplate) {
      let temp = allTemplates.find(
        (ele) => ele.id == dashboard_templates_modal_state.pickedFirstTemplate
      );
      dispatch({ type: NEW_DASHBOARD_TEMPLATES_MODAL_OPEN });
      setStep(2);
      setSelectedTemplate(temp);
    }
  }, [dashboard_templates_modal_state.pickedFirstTemplate]);
  useEffect(() => {
    // fetchDashboards(setAllTemplates)
    setAllTemplates(dashboardTemplates.templates.data);
  }, []);
  // This useEffect gets called whenever templates data gets changed in Redux store
  useEffect(() => {
    setAllTemplates(dashboardTemplates.templates.data);
  }, [dashboardTemplates.templates.data]);

  // This is useeffect callback gets called when  Set of searched Templates  changes
  useEffect(() => {
    setFinalTemplates(searchedTemplates);
    setCategorySelected(null);
  }, [searchedTemplates]);

  useEffect(() => {
    setFinalTemplates(allTemplates);
  }, [allTemplates]);
  useEffect(() => {
    // if (selectedTemplate != null) setStep(2);
  }, [selectedTemplate]);

  useEffect(() => {
    allTemplates.forEach((element) => {
      if (element.categories) {
        element.categories.forEach((eachCategory) => {
          if (!categoryMap.has(eachCategory.toLowerCase().replace(/\s/g, ''))) {
            categoryMap.set(eachCategory.toLowerCase().replace(/\s/g, ''), [
              element
            ]);
          } else {
            categoryMap
              .get(eachCategory.toLowerCase().replace(/\s/g, ''))
              .push(element);
          }
        });
      }
    });
  }, [allTemplates]);

  let handleTemplate = (template) => {
    setStep(2);
    setSelectedTemplate(finalTemplates[template]);
  };
  return (
    <>
      {dashboard_templates_modal_state.isNewDashboardTemplateModal ? (
        <Modal
          bodyStyle={{ padding: '0px 0px 24px 0' }}
          title={null}
          visible={dashboard_templates_modal_state.isNewDashboardTemplateModal}
          centered={true}
          zIndex={1005}
          width={1052}
          className={'fa-modal--regular p-4 fa-modal--slideInDown '}
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
              template={selectedTemplate ? selectedTemplate : null}
              setStep={setStep}
              setSelectedTemplate={setSelectedTemplate}
              allTemplates={allTemplates}
            />
          )}
        </Modal>
      ) : (
        ''
      )}
    </>
  );
};
export default React.memo(DashboardTemplatesModal);
