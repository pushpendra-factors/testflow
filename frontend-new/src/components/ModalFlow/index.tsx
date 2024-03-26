import {
  ArrowLeftOutlined,
  ArrowRightOutlined,
  CloseCircleOutlined,
  CloseOutlined,
  InfoCircleOutlined,
  SearchOutlined
} from '@ant-design/icons';
import { SVG, Text } from 'Components/factorsComponents';
import { DashboardTemplatesControllerType } from 'Reducers/dashboard_templates_modal';
import {
  ADD_DASHBOARD_MODAL_OPEN,
  NEW_DASHBOARD_TEMPLATES_MODAL_CLOSE
} from 'Reducers/types';
import TemplateThumbnailImage from 'Views/Dashboard/AddDashboard/DashboardTemplatesModal/TemplateThumbnailImage';
import {
  Alert,
  Button,
  Card,
  Col,
  Divider,
  Input,
  List,
  Modal,
  Row,
  Spin
} from 'antd';
import React, {
  MouseEventHandler,
  ReactEventHandler,
  useCallback,
  useEffect,
  useMemo,
  useState
} from 'react';
import { DefaultRootState, useDispatch, useSelector } from 'react-redux';
import Paragraph from 'antd/lib/typography/Paragraph';
import { Link } from 'react-router-dom';
import { Integration_Checks } from 'Constants/templates.constants';
import EventGroupBlock from 'Components/QueryComposer/EventGroupBlock';
import QueryBlock from 'Views/Settings/ProjectSettings/Alerts/EventBasedAlert/QueryBlock';
import {
  deleteGroupByForEvent,
  resetGroupBy
} from 'Reducers/coreQuery/middleware';
import { QUERY_TYPE_EVENT } from 'Utils/constants';
import styles from './index.module.scss';

import AlertTemplatesHeader from "./../../assets/images/illustrations/alerttemplatesheader.png"
import DashboardTemplatesHeader from "./../../assets/images/illustrations/dashboardtemplatesheader.png"
          
export type FlowItemType = {
  id: string | number;
  description: string;
  title: string;
  onClick: ReactEventHandler<any> | undefined;
  categories: Array<string>;
  required_integrations?: Array<Array<string>>;
  otherData: object;
  imagePath: string;
  icon?: string;
  backgroundColor?: string;
  color?: string;
  question?: string;
  prepopulate: {
    hubspot?: {
      event: { label: string; group: string };
      filterBy: [
        {
          operator: string;
          props: Array<string>;
          values: Array<any>;
          ref: number;
        }
      ];
    };
    salesforce?: {
      event: { label: string; group: string };
      filterBy: [
        {
          operator: string;
          props: Array<string>;
          values: Array<any>;
          ref: number;
        }
      ];
    };
    noIntegration?: {
      event: { label: string; group: string };
      filterBy: [
        {
          operator: string;
          props: Array<string>;
          values: Array<any>;
          ref: number;
        }
      ];
    };
  };
};

function CategoryPill(props: { item: FlowItemType | null }) {
  const { item } = props;
  return (
    <div
      style={{
        padding: '3px 8px',
        display: 'flex',
        borderRadius: '24px',
        alignItems: 'center',
        gap: '10px',
        backgroundColor: item?.backgroundColor,
        color: item?.color,
        width: 'max-content',
        margin: '10px 0 10px 10px'
      }}
    >
      <SVG name={item.icon} color={item.color} />
      {item.categories.join(',')}
    </div>
  );
}
interface FirstScreenPropType {
  data: Array<FlowItemType>;
  onCancel?: ReactEventHandler | undefined;
  startFreshVisible?: boolean;
  FirstScreenIllustration?: JSX.Element;
  handleSelectedItem?: (item: FlowItemType) => void;
  isDashboardTemplatesFlow: boolean;
}
function FirstScreen({
  data,
  onCancel,
  startFreshVisible = false,
  FirstScreenIllustration,
  handleSelectedItem,
  isDashboardTemplatesFlow = false
}: FirstScreenPropType) {
  const dispatch = useDispatch();
  const [categories, setCategories] = useState<string[]>([]);
  const [results, setResults] = useState<Array<FlowItemType>>([]);
  const [selectedCategory, setSelectedCategory] = useState<string>('all');
  const [searchTerm, setSearchTerm] = useState<string>('');
  useEffect(() => {
    setResults(data);
  }, []);
  useEffect(() => {
    if (data && Array.isArray(data)) {
      const tmpSet = new Set();
      data.forEach((eachFlowItem) => {
        eachFlowItem.categories.forEach((eachCategory) => {
          if (eachCategory) tmpSet.add(eachCategory);
        });
      });
      const allCategories = Array.from(tmpSet);
      // converted unknown[] to string[]
      setCategories(allCategories.map((e) => e as string));
    }
  }, [data]);
  useEffect(() => {
    // Here Final Results should be Intersection of AppliedCategories and SearchTerm
    // here only applied search across only title
    if (data && Array.isArray(data)) {
      const tmp = data.filter((e) => {
        // If search term doesn't matches, then never show this results
        if (
          searchTerm.length > 0 &&
          !e.title.toLowerCase().includes(searchTerm.toLowerCase())
        )
          return false;

        // Now checking for Results
        if (e.categories && Array.isArray(e.categories)) {
          if (selectedCategory === 'all') return true;
          const found = e.categories.find((eee) => eee === selectedCategory);
          if (found) return true;
          return false;
        }
        return false;
      });
      setResults(tmp);
    }
  }, [selectedCategory, searchTerm, data]);

  const renderCategories = () => (
    <div className={styles.categories}>
      <Text type='title' level={6} weight='bold'>
        Categories
      </Text>

      <div>
        <div
          className={selectedCategory === 'all' ? styles['selected-item'] : ''}
          onClick={() => setSelectedCategory('all')}
        >
          All Templates
        </div>
        {categories.map((eachCategory) => (
          <div
            className={
              selectedCategory === eachCategory ? styles['selected-item'] : ''
            }
            key={eachCategory}
            onClick={() => setSelectedCategory(eachCategory)}
          >
            {eachCategory}
          </div>
        ))}
      </div>
    </div>
  );
  const renderStartFreshNewDashboard = () => (
    <Row
      gutter={[8, 8]}
      style={{ cursor: 'pointer', marginBottom: '20px' }}
      onClick={() => dispatch({ type: ADD_DASHBOARD_MODAL_OPEN })}
    >
      <Col style={{ width: '50%' }}>
        <TemplateThumbnailImage
          isStartFresh
          eachState={undefined}
          TemplatesThumbnail={undefined}
        />
      </Col>
      <Col
        style={{
          width: '50%',
          display: 'grid',
          placeContent: 'center'
        }}
      >
        <Text type='title' level={6} weight='bold' extraClass='m-0 mr-3'>
          Start Fresh
        </Text>
        <Text
          type='title'
          level={7}
          weight='normal'
          extraClass={`m-0 mr-3 ${styles.startFreshDescription}`}
        >
          Create an empty dashboard, run queries and add widgets to start
          monitoring.{' '}
        </Text>
      </Col>
    </Row>
  );
  const renderSearchInput = () => (
    <div
      style={{
        background: 'white',
        position: 'sticky',
        top: 0,
        zIndex: 1,
        padding: '5px 0'
      }}
    >
      <Input
        prefix={<SearchOutlined />}
        placeholder='Search'
        size='large'
        className='fa-input'
        type='text'
        onChange={(e) => {
          setSearchTerm(e.target.value);
        }}
      />
    </div>
  );
  const renderLists = () => {
    return (
      <List
        className={styles.itemsList}
        grid={{ gutter: 48, column: 2 }}
        style={{ margin: '10px 0', padding: '10px' }}
        dataSource={results}
        renderItem={(item) => (
          <List.Item
            key={item.id}
            onClick={() => handleSelectedItem && handleSelectedItem(item)}
            style={{ cursor: handleSelectedItem ? 'pointer' : 'auto' }}
          >
            <Card
              className={styles.item}
              style={{
                border: !isDashboardTemplatesFlow
                  ? '1px solid #dedede'
                  : 'none',
                padding: !isDashboardTemplatesFlow ? '10px' : 'auto',
                borderRadius: '8px',
                height: !isDashboardTemplatesFlow ? '165.98px' : 'auto'
              }}
            >
              {item?.icon && <CategoryPill item={item} />}
              {item.imagePath && (
                <img
                  alt={item.title}
                  // onLoad={() => setIsLoaded(true)}
                  style={{
                    // display: isLoaded === true ? 'block' : 'none',
                    padding: '5px 0px',
                    margin: '0 auto',
                    borderRadius: '5px',
                    width: '100%'
                  }}
                  src={item.imagePath}
                />
              )}

              <Text type='title' level={6} weight='bold' extraClass='m-0 mr-3'>
                {item.title}
              </Text>
              <Text
                type='title'
                level={7}
                weight='normal'
                extraClass={`m-0 mr-3 ${styles.templateDescription}`}
              >
                {item.description}
              </Text>
            </Card>
          </List.Item>
        )}
      />
    );

    return (
      <div>
        {data.map((eachFlowItem) => (
          <div>{eachFlowItem.title}</div>
        ))}
      </div>
    );
  };
  return (
    <Row className={styles.firstscreencontainer}>
      <Row>
        <div>
          <img src={FirstScreenIllustration ? DashboardTemplatesHeader : AlertTemplatesHeader} />
          <div>
            <Text type='title' level={6} weight='bold'>
              What are you planning today ?
            </Text>

            <Paragraph>
              Discover the perfect dashboard template with ease. Simplify your
              selection process and find the ideal design to elevate your
              project effortlessly.
            </Paragraph>
          </div>
        </div>
        <div>
          {onCancel && (
            <Button onClick={onCancel} type='text' icon={<CloseOutlined />} />
          )}
        </div>
      </Row>
      <Row>
        <Col span={5}>{renderCategories()}</Col>
        <Col span={19} style={{ maxHeight: '80vh', overflow: 'scroll' }}>
          <div style={{ padding: '20px' }}>
            {startFreshVisible && renderStartFreshNewDashboard()}
            {renderSearchInput()}
            {renderLists()}
          </div>
        </Col>
      </Row>
    </Row>
  );
}

interface AlertsTemplateStep2ScreenPropType {
  item: FlowItemType | null;
  onCancel?: () => void;
  handleBack?: () => void;
  onFinish?: () => void;
}
function AlertsTemplateStep2Screen(props: AlertsTemplateStep2ScreenPropType) {
  const { item, onCancel, handleBack, onFinish } = props;
  const [currentQuery, setCurrentQuery] = useState<any>([])
  const [currentProperty, setCurrentProperty] = useState<any>([])
  const [integrationState, setIntegrationState] = useState<{
    [key: string]: boolean;
  }>({});
  const [shouldAllow, setShouldAllow] = useState(false)
  const [queries, setQueries] = useState([]);
  const sdkCheck = useSelector(
    (state: any) => state?.global?.projectSettingsV1?.int_completed
  );
  const integration = useSelector(
    (state: any) => state.global.currentProjectSettings
  );
  const { groupBy } = useSelector(
    (state: any) => state?.coreQuery?.groupBy?.event
  );
  const { groups } = useSelector((state: any) => state?.coreQuery);
  const { bingAds, marketo } = useSelector((state: any) => state?.global);
  useEffect(() => {
    const integration_check = new Integration_Checks(
      sdkCheck,
      integration,
      bingAds,
      marketo
    );
    const failedAt = [];
    let finalCheck=false
    const Integration: { [key: string]: boolean } = {}; // ex. hubspot,website_sdk: true
    const IntegrationMapResults : { [key: string]: boolean } = {}; 
    item?.required_integrations?.forEach((eachReq)=>{
      let tmpAns = true;
      eachReq.forEach(e=>Integration[e] = tmpAns && !!integration_check[e])
    })

    let allInts = Object.keys(item?.prepopulate || {})
    allInts.forEach((eachKey:string)=>{
      let allIntsBreak = eachKey.split(',')
      let tmpCheck = true
      allIntsBreak.forEach((e)=> tmpCheck = tmpCheck && Integration[e])
      IntegrationMapResults[eachKey] = tmpCheck
      finalCheck = finalCheck || tmpCheck
    })
    Integration.ok = finalCheck
  
    
    setIntegrationState(Integration); // Integration Object having eachIntegration: boolean value
    if (Integration.ok) {
 
      Integration.noIntegration = true;
      if (!('prepopulate' in item)) return;



      let allIntPairs = Object.keys(item.prepopulate)
      allIntPairs.forEach((eachIntPair)=>{
        if(IntegrationMapResults[eachIntPair]){
          setQueries([
            {
              ...item.prepopulate[eachIntPair].event,
              alias: '',
              filters: item.prepopulate[eachIntPair].filterBy
            }
          ]);
          setCurrentQuery([
            {
              ...item.prepopulate[eachIntPair].event,
              alias: '',
              filters: item.prepopulate[eachIntPair].filterBy
            }
          ])
          setCurrentProperty(item.payload_props[eachIntPair] || [])
        }
      })
 
    }
    
  }, []);
  const queryChange = useCallback(
    (newEvent, index, changeType = 'add', flag = null) => {
      const queryupdated = [...queries];
      if (queryupdated[index]) {
        if (changeType === 'add') {
          if (
            JSON.stringify(queryupdated[index]) !== JSON.stringify(newEvent)
          ) {
            deleteGroupByForEvent(newEvent, index);
            resetGroupBy();
            // setEventPropertyDetails({});
          }
          queryupdated[index] = newEvent;
        } else if (changeType === 'filters_updated') {
          // dont remove group by if filter is changed
          queryupdated[index] = newEvent;
        } else {
          deleteGroupByForEvent(newEvent, index);
          resetGroupBy();
          queryupdated.splice(index, 1);
          // setEventPropertyDetails({});
        }
      } else {
        if (flag) {
          Object.assign(newEvent, { pageViewVal: flag });
        }
        queryupdated.push(newEvent);
      }
      setQueries(queryupdated);
    },
    [queries]
  );

  const groupsList = useMemo(() => {
    const listGroups = [];
    Object.entries(groups?.all_groups || {}).forEach(
      ([group_name, display_name]) => {
        listGroups.push([display_name, group_name]);
      }
    );
    return listGroups;
  }, [groups]);
  const queryList = () => {
    const blockList = [];
    queries.forEach((event, index) => {
      blockList.push(
        <div key={index}>
          <QueryBlock
            availableGroups={groupsList}
            index={index + 1}
            queryType={QUERY_TYPE_EVENT}
            event={event}
            queries={queries}
            eventChange={queryChange}
            groupAnalysis // this can be true or false based on accounts/people
          />
        </div>
      );
    });

    if (queries.length < 1) {
      blockList.push(
        <div key='init'>
          <QueryBlock
            availableGroups={groupsList}
            queryType={QUERY_TYPE_EVENT}
            index={queries.length + 1}
            queries={queries}
            eventChange={queryChange}
            groupBy={[]}
            groupAnalysis
          />
        </div>
      );
    }

    return blockList;
  };
  const handleContinue = ()=>{
    if(onCancel) onCancel()
    onFinish(item, currentQuery, currentProperty)
  }
  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between' }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: '10px' }}>
          <Button
            className='fa-button'
            type='text'
            icon={<ArrowLeftOutlined />}
            onClick={handleBack}
          />
          Go Back to Templates
        </div>
        <div>
          <Button
            onClick={onCancel}
            className='fa-button'
            type='text'
            icon={<CloseOutlined />}
          />
        </div>
      </div>
      <CategoryPill item={item} />
      <div style={{ marginLeft: '10px' }}>
        <Text type='title' level={6} weight='bold' extraClass='m-0 mr-3'>
          {item?.title}
        </Text>
        <Text
          type='title'
          level={7}
          weight='normal'
          extraClass={`m-0 mr-3 ${styles.templateDescription}`}
        >
          {item?.description}
        </Text>

        {integrationState.ok && <>
            <div style={{ padding: '10px 0' }}>
            <Text
              type='title'
              level={7}
              weight='normal'
              extraClass={`m-0 mr-3 mb-2 `}
            >
              {item?.question}
            </Text>
            {item && 'prepopulate' in item && (
              <div className='border--thin-2 px-4 py-2 border-radius--sm'>
                {queryList()}
              </div>
            )}
          </div>
          <div style={{ padding: '10px 0' }}>
            <b>Useful tip</b>: In the following screen, the alert name, message,
            and properties are pre-populated. <br />
            Adding and configuring the destinations is all that is required of
            you.
          </div>
        </>
        }
        {!integrationState.ok && (
          <Alert
            type='warning'
            icon={<InfoCircleOutlined />}
            message={
              <>
                Please complete{' '}
                <b style={{ textTransform: 'capitalize' }}>
                  {Object.keys(integrationState)
                    .filter((eachKey) => eachKey !== 'ok' && integrationState[eachKey] === false)
                    .join(',')}
                </b>{' '}
                integration to use this Template.{' '}
                <Link to='/settings/integration'>
                  Integrate Now <ArrowRightOutlined />
                </Link>
              </>
            }
          />
        )}
      </div>
      <div
        style={{
          display: 'flex',
          justifyContent: 'end',
          gap: '10px',
          marginTop: '5px'
        }}
      >
        <Button type='text' size='large' onClick={handleBack}>
          Go Back to Templates
        </Button>
        <Button type='primary' size='large' onClick={handleContinue} disabled={!integrationState.ok}>
          Continue
        </Button>
      </div>
    </div>
  );
}
interface ModalFlowPropType {
  isDashboardTemplatesFlow: boolean;
  data: Array<any>;
  visible: boolean;
  onCancel?: () => void;

  Step2Screen?: React.ComponentClass<any>;
  Step1Props?: object;
  startFreshVisible?: boolean;
  FirstScreenIllustration?: JSX.Element;
  handleLastFinish?: () => void;
  defaultSelectedItem?: FlowItemType | null;
}
function ModalFlow({
  isDashboardTemplatesFlow = false,
  data,
  visible,
  onCancel,
  Step2Screen,
  Step1Props,
  startFreshVisible = false,
  handleLastFinish,
  defaultSelectedItem = null,
  ...restProps
}: ModalFlowPropType) {
  const [step, setStep] = useState(1);
  const [defaultData, setDefaultData] = useState<Array<any>>([]);
  const [selectedItem, setSelectedItem] = useState<FlowItemType | null>(null);
  const handleSelectedItem = (item: FlowItemType) => {
    setStep(2);
    setSelectedItem(item);
  };
  const handleCancelModal = () => {
    setStep(1);
    setSelectedItem(null)
    if (onCancel) onCancel();
  };
  const handleBack = () => {
    setStep(1);
    setSelectedItem(null)
  };
  const handleSelectItem = (item: FlowItemType) => {
    setSelectedItem(item);
  };
  useEffect(()=>{
    if(defaultSelectedItem){
      handleSelectedItem(defaultSelectedItem)
    }
  },[defaultSelectedItem])
  return (
    <Modal
      title={null}
      centered
      zIndex={1005}
      width={1052}
      className='fa-modal--regular'
      closable={false}
      visible={visible}
      footer={null}
      onCancel={handleCancelModal}
    >
      {step === 1 ? (
        <FirstScreen
          data={data}
          {...Step1Props}
          onCancel={onCancel}
          startFreshVisible={startFreshVisible}
          handleSelectedItem={handleSelectedItem}
          isDashboardTemplatesFlow={isDashboardTemplatesFlow}
          {...restProps}
        />
      ) : Step2Screen ? (
        <Step2Screen
          template={selectedItem}
          allTemplates={data}
          onCancel={handleCancelModal}
          handleBack={handleBack}
          handleSelectItem={handleSelectItem}
        />
      ) : (
        <AlertsTemplateStep2Screen
          item={selectedItem}
          onCancel={handleCancelModal}
          handleBack={handleBack}
          onFinish={handleLastFinish}
        />
      )}
    </Modal>
  );
}
/*
template,
  setStep,
  setSelectedTemplate,
  allTemplates

*/

export default ModalFlow;
