import { customerSupportLink } from './constants';

export const meetLink = (isFreePlan) => {
  if (isFreePlan) return customerSupportLink;
  return 'http://calendly.com/prajwal-007/30min';
};
