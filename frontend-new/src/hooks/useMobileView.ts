import { Grid } from 'antd';

const { useBreakpoint } = Grid;

const useMobileView = (): boolean => {
  const screens = useBreakpoint();

  //considering xs and sm as mobile screen for now
  return screens?.md ? false : true;
};

export default useMobileView;
