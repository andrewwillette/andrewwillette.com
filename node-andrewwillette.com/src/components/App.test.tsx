import { render, screen } from '@testing-library/react';
import App from './App';
import { MemoryRouter } from "react-router-dom"
import '@testing-library/jest-dom';

test('renders biography paragraph', () => {
  render(<MemoryRouter><App /></MemoryRouter>);
  const linkElement = screen.getByText(/Hi! My name is Andrew Willette. I am a software developer based in Kansas City, Kansas./i);
  expect(linkElement).toBeInTheDocument();
});

test('renders resume link', () => {
  render(<MemoryRouter><App/></MemoryRouter>);
  const linkElement = screen.getByText("CV");
  expect(linkElement).toBeInTheDocument();
});
