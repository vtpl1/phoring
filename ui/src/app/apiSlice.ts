// src/app/apiSlice.ts
import { createApi, fetchBaseQuery } from '@reduxjs/toolkit/query/react';
import { Metrics } from '../types/Metrics';

export const metricsApi = createApi({
    reducerPath: 'metricsApi',
    baseQuery: fetchBaseQuery({ baseUrl: '/api' }),
    endpoints: (builder) => ({
        getRuntimeMetrics: builder.query<Metrics, void>({
            query: () => '/metrics',
        }),
        getMetricsHistory: builder.query<Metrics[], void>({
            query: () => '/metrics/history',
        }),
    }),
});

export const { useGetRuntimeMetricsQuery, useGetMetricsHistoryQuery } = metricsApi;
