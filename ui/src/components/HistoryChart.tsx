// src/components/HistoryChart.tsx
import React, { useEffect, useRef } from 'react';
import * as d3 from 'd3';
import { Metrics } from '../types/Metrics';

interface HistoryChartProps {
  data: Metrics[];
}

const HistoryChart: React.FC<HistoryChartProps> = ({ data }) => {
  const chartRef = useRef<SVGSVGElement | null>(null);

  useEffect(() => {
    const svg = d3.select(chartRef.current);
    const width = 600;
    const height = 300;
    const margin = { top: 20, right: 30, bottom: 40, left: 40 };

    // Parse the time
    const parseTime = d3.timeParse('%Y-%m-%dT%H:%M:%S.%LZ');

    // Set up the x and y scales
    const x = d3.scaleTime()
      .domain(d3.extent(data, d => parseTime(d.timestamp.toString()) as Date) as [Date, Date])
      .range([margin.left, width - margin.right]);

    const y = d3.scaleLinear()
      .domain([0, 100])  // Assuming CPU usage max is 100
      .range([height - margin.bottom, margin.top]);

    svg.selectAll('*').remove();

    // Draw the axes
    svg.append('g')
      .attr('transform', `translate(0,${height - margin.bottom})`)
      .call(d3.axisBottom(x));

    svg.append('g')
      .attr('transform', `translate(${margin.left},0)`)
      .call(d3.axisLeft(y));

    // Draw the line
    const line = d3.line<Metrics>()
      .x(d => x(parseTime(d.timestamp.toString()) as Date))
      .y(d => y(d.cpu_usage[0]));  // Plot CPU usage

    svg.append('path')
      .datum(data)
      .attr('fill', 'none')
      .attr('stroke', 'steelblue')
      .attr('stroke-width', 1.5)
      .attr('d', line);
  }, [data]);

  return <svg ref={chartRef} width={600} height={300}></svg>;
};

export default HistoryChart;
