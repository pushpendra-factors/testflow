import React from 'react';
import styles from './index.module.scss';
import { Text } from 'factorsComponents';

export default function SeqSelector({ seq, queryCount, setAnalysisSequence }) {
  const fromSequenceState = (() => {
    const fromSeqState = [];
    for (let i = 0; i < queryCount - 1; i++) {
      fromSeqState.push({
        key: i + 1,
        enabled: i + 1 < seq.end,
        selected: seq.start === i + 1
      });
    }
    return fromSeqState;
  })();

  const toSequenceState = (() => {
    const toSeqState = [];
    for (let i = 1; i < queryCount; i++) {
      toSeqState.push({
        key: i + 1,
        enabled: i + 1 > seq.start,
        selected: seq.end === i + 1
      }
      );
    }
    return toSeqState;
  })();

  const setSeqState = (seqType, key) => {
    const newSeq = {
      start: seqType === 'from' ? key : seq.start,
      end: seqType === 'to' ? key : seq.end
    };
    setAnalysisSequence(newSeq);
  };

  return (
    <div className={`${styles.seq_selector}`}>
      <div className={styles.seq_selector__container}>
        <Text type={'title'} level={6} weight={'bold'} extraClass={'m-0 mb-2'}>Choose from event</Text>
        <div className={styles.seq_selector__container__seq}>
          {
            fromSequenceState.map((item, index) => {
              const classNames = [styles.seq_selector__container__seqKey];
              item.enabled ? classNames.push(styles.seq_selector__container__enabled) : classNames.push(styles.seq_selector__container__disabled);
              if (item.selected) {
                classNames.push(styles.seq_selector__container__selected);
              }
              return (<span key={index} className={classNames.join(' ')} onClick={item.enabled ? () => setSeqState('from', item.key) : () => setSeqState('from', 0) }>{item.key}</span>);
            })
          }
        </div>
      </div>
      <div className={styles.seq_selector__container}>
        <Text type={'title'} level={6} weight={'bold'} extraClass={'m-0 mb-2'}>To event</Text>
        <div className={styles.seq_selector__container__seq}>
          {
            toSequenceState.map((item, index) => {
              const classNames = [styles.seq_selector__container__seqKey];
              item.enabled ? classNames.push(styles.seq_selector__container__enabled) : classNames.push(styles.seq_selector__container__disabled);
              if (item.selected) {
                classNames.push(styles.seq_selector__container__selected);
              }
              return (<span key={index} className={classNames.join(' ')} onClick={item.enabled ? () => setSeqState('to', item.key) : () => setSeqState('to', 0) }>{item.key}</span>);
            })
          }
        </div>
      </div>
    </div>
  );
}
