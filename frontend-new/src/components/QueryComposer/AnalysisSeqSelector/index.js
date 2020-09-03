import React, { useState, useEffect } from 'react';
import styles from './index.module.scss';

export default function SeqSelector({ seq, queryCount, setAnalysisSequence }) {

    const fromSequenceState = (() => {
        const fromSeqState = [];
        for (let i = 0; i < queryCount - 1; i++) {
            fromSeqState.push({
                key: i + 1,
                enabled: i + 1 < seq.end ? true : false,
                selected: seq.start === i + 1 ? true : false
            })
        }
        return fromSeqState;
    })();

    const toSequenceState = (() => {
        const toSeqState = [];
        for (let i = 1; i < queryCount; i++) {
            toSeqState.push({
                key: i + 1,
                enabled: i + 1 > seq.start ? true : false,
                selected: seq.end === i + 1 ? true : false
            }
            )
        }
        return toSeqState;
    })();


    const setSeqState = (seqType, key) => {
        const newSeq = {
            start: seqType === 'from' ? key : seq.start,
            end: seqType === 'to' ? key : seq.end,
        }
        setAnalysisSequence(newSeq);
    }

    return (
        <div className={styles.seq_selector}>
            <div className={styles.seq_selector__container}>
                <span className={styles.seq_selector__text}> Choose from event </span>
                <div className={styles.seq_selector__container__seq}>
                    {
                        fromSequenceState.map(item => {
                            let classNames = [styles.seq_selector__container__seqKey];
                            item.enabled ? classNames.push(styles.seq_selector__container__enabled) : classNames.push(styles.seq_selector__container__disabled);
                            if (item.selected) {
                                classNames.push(styles.seq_selector__container__selected)
                            }
                            return (<span className={classNames.join(' ')} onClick={item.enabled ? () => setSeqState('from', item.key) : null}>{item.key}</span>)
                        })
                    }
                </div>
            </div>
            <div className={styles.seq_selector__container}>
                <span className={styles.seq_selector__text}> To event</span>
                <div className={styles.seq_selector__container__seq}>
                    {
                        toSequenceState.map(item => {
                            let classNames = [styles.seq_selector__container__seqKey];
                            item.enabled ? classNames.push(styles.seq_selector__container__enabled) : classNames.push(styles.seq_selector__container__disabled);
                            if (item.selected) {
                                classNames.push(styles.seq_selector__container__selected)
                            }
                            return (<span className={classNames.join(' ')} onClick={item.enabled ? () => setSeqState('to', item.key) : null}>{item.key}</span>)
                        })
                    }
                </div>
            </div>
        </div>
    );
};