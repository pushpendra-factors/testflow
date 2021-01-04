import React from 'react';
import FaToggleBtn from '../../FaToggleBtn';
import styles from './index.module.scss';

const channels = [
    {label: 'All Channels', value: 'all_ads'}, 
    {label: "Google", value: 'google_ads', icon: 'google_ads'},
    {label: "Facebook", value: 'facebook_ads', icon: 'facebook_ads'},
    {label: "Linkedin", value: 'linkedin_ads', icon: 'linkedin_ads'}
]


const ChannelBlock = ({channel, onChannelSelect}) => {
    return (<div className={styles.block}>
        {channels.map((ch) => {
            return <FaToggleBtn 
                label={ch.label} 
                icon={ch.icon}
                state={ch.value === channel}
                onToggle={() => onChannelSelect(ch.value)}
                > </FaToggleBtn>
        })}
    </div>)
};

export default ChannelBlock;