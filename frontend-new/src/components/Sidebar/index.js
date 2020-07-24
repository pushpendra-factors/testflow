import React from 'react';
import styles from './index.module.scss';

function Sidebar() {
  
  return (
    <nav className={styles.sidebarNavigation}>
      <div className={styles.logo}></div>
      <ul>
        <li className={styles.active}>
          <i className="fa fa-share-alt"></i>
          <span className={styles.tooltip}>Connections</span>
        </li>
        <li>
          <i className="fa fa-hdd-o"></i>
          <span className={styles.tooltip}>Devices</span>
        </li>
        <li>
          <i className="fa fa-newspaper-o"></i>
          <span className={styles.tooltip}>Contacts</span>
        </li>
        <li>
          <i className="fa fa-print"></i>
          <span className={styles.tooltip}>Fax</span>
        </li>
        <li>
          <i className="fa fa-sliders"></i>
          <span className={styles.tooltip}>Settings</span>
        </li>
      </ul>
    </nav>
  )
}

export default Sidebar;