// src/app/store.ts
import { configureStore } from '@reduxjs/toolkit';
import { metricsApi } from './apiSlice';

export const store = configureStore({
    reducer: {
        [metricsApi.reducerPath]: metricsApi.reducer,
    },
    middleware: (getDefaultMiddleware) =>
        getDefaultMiddleware().concat(metricsApi.middleware),
});

export type RootState = ReturnType<typeof store.getState>;
export type AppDispatch = typeof store.dispatch;
