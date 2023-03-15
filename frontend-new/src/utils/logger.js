const style =
  'color:white; font-family:monospace; font-size:12px; font-weight:bold; background-color:#ee3c3c; padding:2px 3px; border-radius:3px';
const label = '%cDebug';

class logger {
  log(...args) {
    if (window.FACTORS_APP_DEBUG === true) console.log(label, style, ...args);
  }
  error(...args) {
    if (window.FACTORS_APP_DEBUG === true) console.error(label, style, ...args);
  }
  warn(...args) {
    if (window.FACTORS_APP_DEBUG === true) console.warn(label, style, ...args);
  }
}

export default new logger();
