import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import '@fontsource/inter/400'
import '@fontsource/inter/500'
import '@fontsource/inter/600'
import '@fontsource/inter/700'
import '@fontsource-variable/fraunces'
import App from './App'
import './index.css'

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <App />
  </StrictMode>,
)
