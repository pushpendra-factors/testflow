export const addShadowToHeader = () => {
  const scrollTop =
    window.pageYOffset !== undefined
      ? window.pageYOffset
      : (
        document.documentElement ||
        document.body.parentNode ||
        document.body
      ).scrollTop;
  if (scrollTop > 0) {
    document.getElementById('app-header').style.filter =
      'drop-shadow(0px 2px 0px rgba(200, 200, 200, 0.25))';
  } else {
    document.getElementById('app-header').style.filter = 'none';
  }
};