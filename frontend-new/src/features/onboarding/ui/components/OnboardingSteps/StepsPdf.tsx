import React from 'react';
import {
  Document,
  Page,
  Text,
  StyleSheet,
  Font,
  View
} from '@react-pdf/renderer';

import {
  CopyNote,
  CopyOption,
  CopyOption1,
  CopyOption1Title,
  CopyOption2,
  CopyOption2Desc,
  CopyOption2Title,
  CopySdkTitle,
  CopyTitle,
  SDK_FLOW
} from '../../../utils';

Font.register({
  family: 'Inter',
  src: 'https://fonts.gstatic.com/s/inter/v12/UcCO3FwrK3iLTeHuS_fvQtMwCp50KnMw2boKoduKmMEVuLyfAZJhiJ-Ek-_EeAmM.woff2'
});

const styles = StyleSheet.create({
  page: {
    display: 'flex',
    flexDirection: 'column',
    backgroundColor: '#fff',
    fontFamily: 'Inter',
    padding: 20
  },
  col: {
    display: 'flex',
    flexDirection: 'column'
  },
  section1: {
    display: 'flex',
    flexDirection: 'column',
    padding: '16px 20px',
    marginBottom: 8,
    backgroundColor: '#f5f5f5',
    borderRadius: 8
  },
  section2: {
    flexDirection: 'column',
    display: 'flex',
    marginBottom: 32
  },
  section3: {
    flexDirection: 'column',
    display: 'flex'
  },
  title1: {
    textAlign: 'center',
    marginBottom: 40
  },
  h1: {
    fontSize: 24,
    fontWeight: 600,
    fontFamily: 'Inter'
  },
  h3: {
    fontWeight: 600,
    fontSize: 18,
    fontFamily: 'Inter'
  },
  h4: {
    fontWeight: 600,
    fontSize: 14,
    fontFamily: 'Inter'
  },
  code: {
    fontSize: 12,
    fontFamily: 'Inter',
    fontWeight: 600,
    color: '#3e516c'
  },
  note: {
    fontWeight: 600,
    fontSize: 12,
    fontFamily: 'Inter'
  },
  normalText: {
    fontSize: 12,
    marginBottom: 4,
    fontFamily: 'Inter'
  },
  noteText: {
    display: 'flex',
    marginBottom: 50
  },
  subtitle: { color: '#b7bec8', marginBottom: 8, fontSize: 12, marginTop: 8 },
  subtitle1: {
    marginBottom: 8
  },
  colpad: {
    display: 'flex',
    flexDirection: 'column',
    padding: '0px 16px',
    gap: 6,
    marginBottom: 16
  }
});

const StepsPdf = ({ scriptCode }: StepsPdfProps) => {
  return (
    <Document>
      <Page size='A4' style={styles.page}>
        <View style={styles.title1}>
          <Text style={styles.h1}>{CopyTitle}</Text>
        </View>
        <View style={styles.subtitle1}>
          <Text style={styles.h3}>{CopySdkTitle}</Text>
        </View>

        <View style={styles.section1}>
          <Text style={styles.code}>{`<script>`}</Text>
          <Text style={styles.code}>{scriptCode}</Text>
          <Text style={styles.code}>{`</script>`}</Text>
        </View>
        <View style={styles.noteText}>
          <Text style={styles.note}>Note: {CopyNote}</Text>
        </View>

        <View style={styles.section2}>
          <Text style={styles.h3}>{CopyOption}</Text>
        </View>
        <View style={styles.section3}>
          <Text style={styles.h4}>{CopyOption1}</Text>
        </View>

        <View style={styles.colpad}>
          <Text style={styles.subtitle}>{CopyOption1Title}</Text>
          <Text style={styles.normalText}>{SDK_FLOW.GTM.step1}</Text>
          <Text style={styles.normalText}>{SDK_FLOW.GTM.step2}</Text>
          <Text style={styles.normalText}>{SDK_FLOW.GTM.step3}</Text>
          <Text style={styles.normalText}>{SDK_FLOW.GTM.step4}</Text>
          <Text style={styles.normalText}>{SDK_FLOW.GTM.step5}</Text>
          <Text style={styles.normalText}>{SDK_FLOW.GTM.step6}</Text>
        </View>

        <View style={styles.col}>
          <Text style={styles.h4}>{CopyOption2}</Text>
        </View>

        <View style={styles.colpad}>
          <Text style={styles.subtitle}>{CopyOption2Title}</Text>
          <Text style={styles.normalText}>{CopyOption2Desc}</Text>
        </View>
      </Page>
    </Document>
  );
};

interface StepsPdfProps {
  scriptCode: string;
}

export default StepsPdf;
