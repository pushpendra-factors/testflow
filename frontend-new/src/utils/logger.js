const style =
  'color:white; font-family:monospace; font-size:12px; font-weight:bold; background-color:#ee3c3c; padding:2px 3px; border-radius:3px';
const label = '%cDebug';

class logger {
  showLogs() {
    return (
      window.FACTORS_APP_DEBUG === true ||
      process.env.NODE_ENV === 'development'
    );
  }

  log(...args) {
    if (this.showLogs()) console.log(label, style, ...args);
  }
  error(...args) {
    if (this.showLogs()) console.error(label, style, ...args);
  }
  warn(...args) {
    if (this.showLogs()) console.warn(label, style, ...args);
  }
}

export default new logger();
