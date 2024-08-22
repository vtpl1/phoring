import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { Provider } from "react-redux";
import { Box, ChakraProvider } from "@chakra-ui/react";
import App from './App.tsx'
import { store } from "./app/store";
import './index.css'

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <Provider store={store}>
        <ChakraProvider>
          <Box h={"100vh"} p={0} m={0}>
            <App />
          </Box>
        </ChakraProvider>
    </Provider>
  </StrictMode>,
)
