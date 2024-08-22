// src/components/MetricsHistory.tsx
import React from 'react';
import { Box, Spinner, Text, Heading } from '@chakra-ui/react';
import { useGetMetricsHistoryQuery } from '../app/apiSlice';
import HistoryChart from "./HistoryChart";

const MetricsHistory: React.FC = () => {
  const { data, error, isLoading } = useGetMetricsHistoryQuery();

  if (isLoading) return <Spinner size="xl" />;
  if (error) return <Text>Error fetching history</Text>;

  return (
    <Box p={5}>
      <Heading mb={4}>Historical Metrics</Heading>
      {data && data.length > 0 ? (
        <HistoryChart data={data} />
      ) : (
        <Text>No history available</Text>
      )}
    </Box>
  );
};

export default MetricsHistory;
