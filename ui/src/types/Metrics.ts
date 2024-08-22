// src/types/Metrics.ts
export interface MemoryUsage {
    total: number;
    available: number;
    used: number;
    usedPercent: number;
}

export interface NetworkUsage {
    bytesRecv: number;
    bytesSent: number;
    name: string;
}

export interface Metrics {
    timestamp: number;
    cpu_usage: number[];
    memory_usage: MemoryUsage;
    network_usage: NetworkUsage[];
}
