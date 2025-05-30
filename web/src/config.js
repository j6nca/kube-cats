const config = {
    api_url: window.env?.API_URL || 'http://localhost:8001',
    environment: process.env.NODE_ENV || 'development',
    max_pods_per_node: process.env.MAX_PODS_PER_NODE || 10,
    show_labels: process.env.SHOW_LABELS || true,
    show_pod_count: process.env.SHOW_POD_COUNT || true,
    show_tooltips: process.env.SHOW_TOOLTIPS || false,
};

export default config;
