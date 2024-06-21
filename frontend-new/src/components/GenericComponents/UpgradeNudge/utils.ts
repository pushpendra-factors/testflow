interface StatusColor {
  backgroundColor: string;
  progressBarColor: string;
  progressBarBackgroundColor: string;
}

export const getStatusColors = (percentage: number): StatusColor => {
  if (percentage < 75) {
    return {
      backgroundColor: '#F0F5FF',
      progressBarColor: '#597EF7',
      progressBarBackgroundColor: '#D9D9D9'
    };
  }
  if (percentage < 100) {
    return {
      backgroundColor: '#FFF7E6',
      progressBarColor: '#FFA940',
      progressBarBackgroundColor: '#D9D9D9'
    };
  }
  return {
    backgroundColor: '#FFF1F0',
    progressBarColor: '#F5222D',
    progressBarBackgroundColor: '#D9D9D9'
  };
};
