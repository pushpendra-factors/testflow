export const EngagementTag: {
  [key: string]: {
    bgColor: string;
    icon: string;
  };
} = {
  Hot: {
    bgColor: '#FFF1F0',
    icon: 'fire'
  },
  Warm: {
    bgColor: '#FFF7E6',
    icon: 'sun'
  },
  Cool: {
    bgColor: '#F0F5FF',
    icon: 'snowflake'
  }
};

export function nearestGreater100(number: number): number {
  const remainder: number = number % 100;
  if (remainder === 0) {
    return number;
  } else {
    return number + (100 - remainder);
  }
}

export function transformDate(yyyymmdd: string): string {
  const year = yyyymmdd.slice(0, 4);
  const month = yyyymmdd.slice(4, 6);
  const day = yyyymmdd.slice(6, 8);

  const monthsMap: { [key: string]: string } = {
    '01': 'Jan',
    '02': 'Feb',
    '03': 'Mar',
    '04': 'Apr',
    '05': 'May',
    '06': 'Jun',
    '07': 'Jul',
    '08': 'Aug',
    '09': 'Sep',
    '10': 'Oct',
    '11': 'Nov',
    '12': 'Dec',
  };

  const monthAbbreviation = monthsMap[month];

  return `${monthAbbreviation} ${parseInt(day, 10)}, ${year}`;
}
