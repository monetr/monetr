import { createRoot } from 'react-dom/client';
import './styles.css';

import { App } from './App';

const container = document.getElementById('root');
if (container) {
  createRoot(container).render(<App />);
}
