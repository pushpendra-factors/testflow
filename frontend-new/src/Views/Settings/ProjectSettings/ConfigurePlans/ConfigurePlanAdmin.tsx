import { Text } from 'Components/factorsComponents';
import { getFeatureConfigData } from 'Reducers/featureConfig/services';
import {
  FeatureConfig,
  FeatureConfigApiResponse,
  SixSignalInfo
} from 'Reducers/featureConfig/types';
import { AdminLock } from 'Routes/feature';
import { PathUrls } from 'Routes/pathUrls';
import logger from 'Utils/logger';
import { Select, Spin, notification } from 'antd';
import useAgentInfo from 'hooks/useAgentInfo';
import React, { useEffect, useState } from 'react';
import { useHistory } from 'react-router-dom';
import { PLANS, PLANS_V0 } from 'Constants/plans.constants';
import { setShowAnalyticsResult } from 'Reducers/coreQuery/actions';
import { useDispatch } from 'react-redux';
import CustomPlanConfigure from './CustomPlanConfigure';
import { fetchAllProjects, filterProject } from './utils';

function ConfigurePlanAdmin() {
  const [selectedProject, setSelectedProject] = useState<string | null>(null);
  const [projects, setProjects] = useState<{ id: string; name: string }[]>([]);
  const [projectConfig, setProjectConfig] = useState<{
    activeFeatures?: FeatureConfig[] | null;
    addOns?: FeatureConfig[];
    sixSignalInfo?: SixSignalInfo;
    projectId?: number;
    planName?: string;
  }>({});
  const [projectInfoLoading, setProjectInfoLoading] = useState(false);
  const [loading, setLoading] = useState(false);

  const { email } = useAgentInfo();
  const history = useHistory();
  const dispatch = useDispatch();

  const handleProjectChange = (projectId: string) => {
    setSelectedProject(projectId);
  };
  const renderLoader = () => (
    <div className='w-full h-full flex items-center justify-center'>
      <div className='w-full h-64 flex items-center justify-center'>
        <Spin size='large' />
      </div>
    </div>
  );

  const fetchProjectConfig = async (project: string) => {
    try {
      setProjectInfoLoading(true);
      if (!project) return;
      const res = (await getFeatureConfigData(
        project
      )) as FeatureConfigApiResponse;
      if (res?.data) {
        setProjectConfig({
          activeFeatures: res.data?.plan?.feature_list,
          addOns: res.data?.add_ons,
          sixSignalInfo: res.data?.six_signal_info,
          projectId: res.data?.project_id,
          planName: res?.data?.plan?.name
        });
      }
      setProjectInfoLoading(false);
    } catch (error) {
      logger.error('Error in fetching project config', error);
      notification.error({
        message: 'Error in fetching project details',
        description:
          error?.data?.err?.display_message || 'Something went wrong.',
        duration: 2
      });
      setProjectInfoLoading(false);
    }
  };

  const successCallback = () => {
    if (selectedProject) {
      fetchProjectConfig(selectedProject);
    }
  };

  useEffect(() => {
    if (selectedProject) {
      fetchProjectConfig(selectedProject);
    }
  }, [selectedProject]);

  // making this route only accesible to Admin email
  useEffect(() => {
    if (!AdminLock(email)) {
      history.push(PathUrls.ProfileAccounts);
    }
  }, []);

  useEffect(() => {
    const fetchProjects = async () => {
      try {
        setLoading(true);
        const res = await fetchAllProjects();
        if (res?.data && Array.isArray(res.data)) {
          setProjects(res.data);
        }
        setLoading(false);
      } catch (error) {
        logger.error('Error!', error);
        notification.error({
          message: 'Error in fetching projects',
          description:
            error?.data?.err?.display_message || 'Something went wrong.',
          duration: 2
        });
        setLoading(false);
      }
    };
    fetchProjects();
  }, []);

  // hiding top nav bar
  useEffect(() => {
    dispatch(setShowAnalyticsResult(true));

    return () => {
      dispatch(setShowAnalyticsResult(false));
    };
  }, [setShowAnalyticsResult]);

  if (loading) return renderLoader();
  return (
    <div className='p-8 w-full h-full'>
      <div>
        <Text type='title' level={3} weight='bold' extraClass='m-0 m-1'>
          Admin Plan Configuration
        </Text>
      </div>
      <div className='flex items-center gap-3 my-5'>
        <Text type='paragraph' mini>
          Project:
        </Text>
        <Select
          style={{ minWidth: 300 }}
          className='fa-select'
          value={selectedProject}
          filterOption={filterProject}
          showSearch
          showArrow
          onChange={handleProjectChange}
          optionLabelProp='label'
          placeholder='Select a project'
        >
          {projects?.map((project) => (
            <Select.Option
              key={project.id}
              value={project.id}
              label={project?.name}
            >
              {`${project?.name}: ${project?.id}`}
            </Select.Option>
          ))}
        </Select>
      </div>
      {projectInfoLoading && renderLoader()}
      {!projectInfoLoading &&
      selectedProject &&
      projectConfig?.activeFeatures &&
      String(selectedProject) === String(projectConfig?.projectId) ? (
        projectConfig?.planName === PLANS.PLAN_CUSTOM ||
        projectConfig?.planName === PLANS_V0.PLAN_CUSTOM ? (
          <CustomPlanConfigure
            sixSignalInfo={projectConfig?.sixSignalInfo}
            activeFeatures={projectConfig?.activeFeatures}
            addOns={projectConfig?.addOns}
            featureLoading={projectInfoLoading}
            projectId={selectedProject}
            successCallback={successCallback}
          />
        ) : (
          <Text type='paragraph' mini>
            Plan configuration is only allowed for {PLANS.PLAN_CUSTOM} plan
          </Text>
        )
      ) : null}
    </div>
  );
}

export default ConfigurePlanAdmin;
