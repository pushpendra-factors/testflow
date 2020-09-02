import React, {useState, useEffect} from 'react';
import styles from './index.module.scss';

export default function SeqSelector({seq, queryCount, setAnalysisSequence}){

    const [fromSequenceState, setFromState] = useState(() => {
        const fromSeqState = [];
        for(let i = 0; i < queryCount; i++) {
            fromSeqState.push({
                key: i+1,
                enabled: true,
                selected: seq.start === i+1? true : false
            })
        }
        return fromSeqState;
    });

    const [toSequenceState, setToState] = useState(() => {
        const toSeqState = [];
        for(let i = 1; i <= queryCount; i++) {
            toSeqState.push({
                key: i+1,
                enabled: true,
                selected: seq.end === i+1? true : false
            }   
            )
        }
        return toSeqState;
    });

    const setSeqState = (seqType, key) => {
        const fromState = [...fromSeqState];
        
    }

    return (
        <div className={styles.seq_selector}>
            <div className={styles.seq_selector__container}>
                <span className={styles.seq_selector__text}> Choose from event </span>
                <div className={styles.seq_selector__seq}>
                    {
                        fromSequenceState.map(item => {
                            let className = styles.seq_selector__container__seqKey;
                            className += (item.enabled ? styles.seq_selector__container__enabled: styles.seq_selector__container__disabled);
                            className += (item.selected ? styles.seq_selector__container__selected: '');
                            return (<span className={className} onClick={item.enabled? () => setSeqState('from', item.key) : null}>{item.key}</span>)
                        })
                    }
                </div>
            </div>
            <div className={styles.seq_selector__container}>
                <span className={styles.seq_selector__text}> To event</span>
                <div className={styles.seq_selector__seq}>
                { 
                        toSequenceState.map(item => {
                            let className = styles.seq_selector__container__seqKey;
                            className += (item.enabled ? styles.seq_selector__container__enabled: styles.seq_selector__container__disabled);
                            className += (item.selected ? styles.seq_selector__container__selected: '');
                            
                            return (<span className={className} onClick={item.enabled?() => setSeqState('to', item.key) : null}>{item.key}</span>)
                        })
                    }
                </div>
            </div>
        </div>
    );
};