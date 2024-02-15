import logger from 'Utils/logger';
import { getHostUrl, get } from 'Utils/request';

const host = getHostUrl();

export const filterProject = (input: string, option: any) => {
  const inputValue = input?.toLowerCase();
  return (
    option?.label?.toString()?.toLowerCase()?.includes(inputValue) ||
    option?.value?.toString()?.includes(inputValue)
  );
};

export const fetchAllProjects = () => {
  try {
    const url = `${host}v1/projects/custom_projects`;
    return get(null, url);
  } catch (error) {
    logger.error(error);
    return null;
  }
};
