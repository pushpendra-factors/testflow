import React from 'react';
import {
  Document,
  Page,
  Text,
  StyleSheet,
  Font,
  View,
  Link,
  Svg,
  Path,
  G,
  Ellipse
} from '@react-pdf/renderer';

import {
  CopyNote,
  CopyOption,
  CopyOption1,
  CopyOption2,
  CopyOption2Desc,
  CopySdkTitle,
  CopyTitle,
  OnboardingSupportLink,
  SDK_FLOW
} from '../../../utils';

Font.register({
  family: 'Inter',
  src: 'https://fonts.gstatic.com/s/inter/v12/UcCO3FwrK3iLTeHuS_fvQtMwCp50KnMw2boKoduKmMEVuLyfAZJhiJ-Ek-_EeAmM.woff2'
});

Font.registerHyphenationCallback((word) => [word]);

const styles = StyleSheet.create({
  page: {
    display: 'flex',
    flexDirection: 'column',
    backgroundColor: '#fff',
    fontFamily: 'Inter',
    color: '#242424'
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
    display: 'flex',
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    backgroundColor: '#E6F7FF',
    padding: '25px 20px',
    color: '#003A8C'
  },
  h1: {
    fontSize: 24,
    fontWeight: 600,
    fontFamily: 'Inter'
  },
  copySdkTitle: {
    fontWeight: 600,
    fontSize: 18,
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
  codeHead: {
    fontSize: 12,
    fontFamily: 'Inter',
    fontWeight: 600,
    color: '#1890ff'
  },
  note: {
    fontWeight: 400,
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
    marginBottom: 30,
    marginTop: 8
  },
  subtitle: { color: '#b7bec8', marginBottom: 8, fontSize: 12, marginTop: 8 },
  subtitle1: {
    marginTop: 20,
    marginBottom: 16,
    fontFamily: 'Inter'
  },
  colpad: {
    display: 'flex',
    flexDirection: 'column',
    padding: '10px 16px',
    gap: 6,
    marginBottom: 16
  }
});

function StepsPdf({ scriptCode }: StepsPdfProps) {
  return (
    <Document>
      <Page size='A4' style={styles.page}>
        <View style={styles.title1}>
          <Text style={styles.h1}>{CopyTitle}</Text>
          <View>
            <Svg width='24' height='30' viewBox='0 0 38 46'>
              <Path
                d='M7.40991 31.0996H18.9965C18.9965 37.4987 13.809 42.6896 7.40991 42.6896V31.0996Z'
                fill='#FF3535'
              />
              <Path
                d='M7.40991 17.2051H30.5899C30.5899 23.6061 25.4009 28.7951 18.9999 28.7951H7.40991V17.2051Z'
                fill='#FF3535'
              />
              <Path
                d='M7.40991 14.9005C7.40991 8.49957 12.5989 3.31055 18.9999 3.31055H30.5899C30.5899 9.71153 25.4009 14.9005 18.9999 14.9005H7.40991Z'
                fill='#FF3535'
              />

              <G>
                <Ellipse
                  opacity='0.1'
                  cx='40.9923'
                  cy='32.7356'
                  rx='30.3577'
                  ry='31.2737'
                  fill='white'
                />
              </G>
            </Svg>
          </View>
        </View>
        <View style={{ padding: '0px 20px' }}>
          <View style={styles.subtitle1}>
            <Text style={styles.copySdkTitle}>{CopySdkTitle}</Text>
          </View>

          <View style={styles.section1}>
            <Text style={styles.codeHead}>{`<script>`}</Text>
            <Text style={styles.code}>{scriptCode}</Text>
            <Text style={styles.codeHead}>{`</script>`}</Text>
          </View>
          <View style={styles.noteText}>
            <Text style={styles.note}>
              <Text style={{ fontWeight: 600 }}>Note:</Text> {CopyNote}
            </Text>
          </View>

          <View style={styles.section2}>
            <Text style={styles.h3}>{CopyOption}</Text>
          </View>
          <View style={styles.section3}>
            <Text style={styles.h4}>{CopyOption1}</Text>
          </View>

          <View style={styles.colpad}>
            <Text style={styles.normalText}>
              1. Sign in to{' '}
              <Link src='https://tagmanager.google.com/' href=''>
                Google Tag Manager
              </Link>
              , select “Workspace”, and “Add a new tag”
            </Text>
            <Text style={styles.normalText}>{SDK_FLOW.GTM.step2}</Text>
            <Text style={styles.normalText}>{SDK_FLOW.GTM.step3}</Text>
            <Text style={styles.normalText}>{SDK_FLOW.GTM.step4}</Text>
            <Text style={styles.normalText}>{SDK_FLOW.GTM.step5}</Text>
            <Text style={styles.normalText}>{SDK_FLOW.GTM.step6}</Text>
            <Text style={styles.normalText}>
              SDK still says "Not verified"? Check these{' '}
              <Link src='https://help.factors.ai/en/articles/7260638-connecting-factors-to-your-website'>
                steps
              </Link>
            </Text>
          </View>

          <View style={styles.col}>
            <Text style={styles.h4}>{CopyOption2}</Text>
          </View>

          <View style={styles.colpad}>
            <Text style={styles.normalText}>1. {CopyOption2Desc}</Text>
          </View>

          <View style={{ marginTop: 40 }}>
            <Text style={styles.normalText}>
              If you have any questions or issues, please reach out to our{' '}
              <Link src={OnboardingSupportLink}>support team</Link> for
              assistance.
            </Text>
          </View>
        </View>
      </Page>
    </Document>
  );
}

interface StepsPdfProps {
  scriptCode: string;
}

export default StepsPdf;
