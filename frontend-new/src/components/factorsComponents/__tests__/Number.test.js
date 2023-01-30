import { render, screen } from '@testing-library/react';
import React from 'react';
import Number from '../Number';

describe('Number Component  tests', () => {
  it('should render with short hand', () => {
    render(<Number number={15300} shortHand />);
    const numberElement = screen.getByText('15.3K');
    expect(numberElement).toBeInTheDocument();
  });

  it('should render number prefix and suffix', () => {
    render(<Number number={15300} shortHand prefix='$' suffix='cash' />);
    const numberElement = screen.getByText('$15.3Kcash');
    expect(numberElement).toBeInTheDocument();
  });

  it('should render number without short hand', () => {
    render(<Number number={15300} />);
    const numberElement = screen.getByText('15,300');
    expect(numberElement).toBeInTheDocument();
  });

  it('should render number without short hand and with prefix and suffix', () => {
    render(<Number number={15300} prefix='$' suffix='cash' />);
    const numberElement = screen.getByText('$15,300cash');
    expect(numberElement).toBeInTheDocument();
  });

  it('should have the passed classname', () => {
    render(<Number number={15300} className='test-class' />);
    const numberElement = screen.getByTestId('number');
    expect(numberElement).toHaveClass('test-class');
  });
});
